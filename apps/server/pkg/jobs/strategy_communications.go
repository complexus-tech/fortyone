package jobs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	strategyCommunicationHour = 9
	strategySnapshotVersion   = 1
	strategyStaleAfterDays    = 7
	strategyWeeklyDetailLimit = 10
)

// StrategyNotificationCreator is the notification boundary used by strategy jobs.
// The concrete service persists the notification, publishes it in-app, and queues
// the existing coalesced email digest.
type StrategyNotificationCreator interface {
	Create(context.Context, notifications.CoreNewNotification) (notifications.CoreNotification, error)
}

type strategyRecipient struct {
	UserID      uuid.UUID `db:"user_id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Timezone    string    `db:"timezone"`
}

type strategyFoundation struct {
	HasUltimateGoal bool `db:"has_ultimate_goal"`
	PillarCount     int  `db:"pillar_count"`
	ObjectiveCount  int  `db:"objective_count"`
}

type strategyCheckIn struct {
	strategyRecipient
	StaleObjectives  int
	AtRiskObjectives int
	StaleKeyResults  int
	Objectives       []notifications.StrategyObjectiveSnapshot
	KeyResults       []notifications.StrategyKeyResultSnapshot
}

type strategyCheckInRecord struct {
	strategyRecipient
	ObjectiveID              uuid.UUID  `db:"objective_id"`
	TeamID                   uuid.UUID  `db:"team_id"`
	ObjectiveName            string     `db:"objective_name"`
	ObjectiveHealth          *string    `db:"objective_health"`
	ObjectiveStatusID        *uuid.UUID `db:"objective_status_id"`
	ObjectiveStatusName      *string    `db:"objective_status_name"`
	ObjectiveStatusCategory  *string    `db:"objective_status_category"`
	ObjectiveStartDate       *time.Time `db:"objective_start_date"`
	ObjectiveEndDate         *time.Time `db:"objective_end_date"`
	ObjectiveUpdatedAt       time.Time  `db:"objective_updated_at"`
	IsStaleObjective         bool       `db:"is_stale_objective"`
	IsAtRiskObjective        bool       `db:"is_at_risk_objective"`
	KeyResultID              *uuid.UUID `db:"key_result_id"`
	KeyResultName            *string    `db:"key_result_name"`
	KeyResultMeasurementType *string    `db:"key_result_measurement_type"`
	KeyResultStartValue      *float64   `db:"key_result_start_value"`
	KeyResultCurrentValue    *float64   `db:"key_result_current_value"`
	KeyResultTargetValue     *float64   `db:"key_result_target_value"`
	KeyResultStartDate       *time.Time `db:"key_result_start_date"`
	KeyResultEndDate         *time.Time `db:"key_result_end_date"`
	KeyResultUpdatedAt       *time.Time `db:"key_result_updated_at"`
}

type strategyMonthlySummary struct {
	PillarCount          int      `db:"pillar_count"`
	PillarsNeedingReview int      `db:"pillars_needing_review"`
	ObjectiveCount       int      `db:"objective_count"`
	AtRiskObjectives     int      `db:"at_risk_objectives"`
	UnalignedObjectives  int      `db:"unaligned_objectives"`
	KeyResultCount       int      `db:"key_result_count"`
	KeyResultProgress    *float64 `db:"key_result_progress"`
	CompletedStories     int      `db:"completed_stories"`
}

// ProcessStrategyCommunications creates due local-time planning reminders,
// weekly owner check-ins, and monthly leadership summaries.
func ProcessStrategyCommunications(
	ctx context.Context,
	db *sqlx.DB,
	log *logger.Logger,
	notifier StrategyNotificationCreator,
	systemUserID uuid.UUID,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStrategyCommunications")
	defer span.End()

	now := time.Now().UTC()
	var processingErrors []error

	if err := processStrategyPlanningReminders(ctx, db, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("planning reminders: %w", err))
	}
	if err := processStrategyWeeklyCheckIns(ctx, db, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("weekly check-ins: %w", err))
	}
	if err := processStrategyMonthlySummaries(ctx, db, notifier, systemUserID, now); err != nil {
		processingErrors = append(processingErrors, fmt.Errorf("monthly summaries: %w", err))
	}

	if len(processingErrors) > 0 {
		err := errors.Join(processingErrors...)
		span.RecordError(err)
		log.Error(ctx, "strategy communications completed with errors", "error", err)
		return err
	}

	return nil
}

func processStrategyPlanningReminders(ctx context.Context, db *sqlx.DB, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	recipients, err := getStrategyAdminRecipients(ctx, db)
	if err != nil {
		return err
	}

	var processingErrors []error
	for _, recipient := range recipients {
		localNow := now.In(strategyLocation(recipient.Timezone))
		if localNow.Hour() != strategyCommunicationHour {
			continue
		}

		quarterStart := nextQuarterStart(localNow)
		daysUntil := calendarDaysBetween(localNow, quarterStart)
		if daysUntil != 21 && daysUntil != 7 {
			continue
		}

		foundation, loadErr := getStrategyFoundation(ctx, db, recipient.WorkspaceID, quarterStart)
		if loadErr != nil {
			processingErrors = append(processingErrors, loadErr)
			continue
		}
		if foundation.HasUltimateGoal && foundation.PillarCount > 0 && foundation.ObjectiveCount > 0 {
			continue
		}

		period := fmt.Sprintf("Q%d", quarterForMonth(quarterStart.Month()))
		if quarterStart.Month() == time.January {
			period = fmt.Sprintf("%d strategy", quarterStart.Year())
		}
		missing := missingStrategyElements(foundation)
		notification := notifications.CoreNewNotification{
			DedupeKey:   fmt.Sprintf("strategy:planning:%s:%s:%s:%d", recipient.WorkspaceID, recipient.UserID, quarterStart.Format("2006-01-02"), daysUntil),
			RecipientID: recipient.UserID,
			WorkspaceID: recipient.WorkspaceID,
			Type:        "strategy_update",
			EntityType:  "strategy",
			EntityID:    recipient.WorkspaceID,
			ActorID:     systemUserID,
			Title:       fmt.Sprintf("Plan your %s", period),
			Message: notifications.NotificationMessage{
				Template: "Your next planning period starts in {days}. Review {missing} while there is still time to align the team.",
				Variables: map[string]notifications.Variable{
					"days":    {Value: fmt.Sprintf("%d days", daysUntil), Type: "date"},
					"missing": {Value: missing, Type: "value"},
				},
				Strategy: &notifications.StrategyNotificationSnapshot{
					Version:     strategySnapshotVersion,
					Kind:        notifications.StrategyNotificationKindPlanningReminder,
					GeneratedAt: now.UTC(),
					Planning: &notifications.StrategyPlanningSnapshot{
						Period:          period,
						StartsAt:        quarterStart,
						DaysUntil:       daysUntil,
						HasUltimateGoal: foundation.HasUltimateGoal,
						PillarCount:     foundation.PillarCount,
						ObjectiveCount:  foundation.ObjectiveCount,
						MissingElements: missingStrategyElementKeys(foundation),
					},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
}

func processStrategyWeeklyCheckIns(ctx context.Context, db *sqlx.DB, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	checkIns, err := getStrategyWeeklyCheckIns(ctx, db, now)
	if err != nil {
		return err
	}

	var processingErrors []error
	for _, checkIn := range checkIns {
		localNow := now.In(strategyLocation(checkIn.Timezone))
		if localNow.Hour() != strategyCommunicationHour || localNow.Weekday() != time.Wednesday {
			continue
		}
		totalSignals := checkIn.StaleObjectives + checkIn.AtRiskObjectives + checkIn.StaleKeyResults
		if totalSignals == 0 {
			continue
		}

		isoYear, week := localNow.ISOWeek()
		summary := strategyCheckInSummary(checkIn)
		objectives, keyResults, omittedDetails := boundedStrategyCheckInDetails(checkIn, strategyWeeklyDetailLimit)
		notification := notifications.CoreNewNotification{
			DedupeKey:   strategyWeeklyCheckInDedupeKey(checkIn.WorkspaceID, checkIn.UserID, isoYear, week),
			RecipientID: checkIn.UserID,
			WorkspaceID: checkIn.WorkspaceID,
			Type:        "strategy_update",
			EntityType:  "strategy",
			EntityID:    checkIn.WorkspaceID,
			ActorID:     systemUserID,
			Title:       "Your weekly strategy check-in",
			Message: notifications.NotificationMessage{
				Template: "A quick review will keep execution connected to strategy: {summary}.",
				Variables: map[string]notifications.Variable{
					"summary": {Value: summary, Type: "value"},
				},
				Strategy: &notifications.StrategyNotificationSnapshot{
					Version:     strategySnapshotVersion,
					Kind:        notifications.StrategyNotificationKindWeeklyCheckIn,
					GeneratedAt: now.UTC(),
					WeeklyCheckIn: &notifications.StrategyWeeklyCheckInSnapshot{
						StaleAfterDays: strategyStaleAfterDays,
						Counts: notifications.StrategyWeeklyCheckInCounts{
							AtRiskObjectives: checkIn.AtRiskObjectives,
							StaleObjectives:  checkIn.StaleObjectives,
							StaleKeyResults:  checkIn.StaleKeyResults,
							UniqueObjectives: uniqueStrategyObjectiveCount(checkIn),
						},
						TeamCounts:     strategyWeeklyTeamCounts(checkIn, objectives, keyResults),
						Objectives:     objectives,
						KeyResults:     keyResults,
						OmittedDetails: omittedDetails,
					},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
}

func strategyWeeklyCheckInDedupeKey(workspaceID, userID uuid.UUID, isoYear, isoWeek int) string {
	return fmt.Sprintf("strategy:check-in:%s:%s:%d-%02d", workspaceID, userID, isoYear, isoWeek)
}

type strategyWeeklyTeamCountAccumulator struct {
	counts           notifications.StrategyWeeklyCheckInCounts
	objectiveIDs     map[uuid.UUID]struct{}
	objectiveDetails int
	keyResultDetails int
}

func strategyWeeklyTeamCounts(
	checkIn strategyCheckIn,
	selectedObjectives []notifications.StrategyObjectiveSnapshot,
	selectedKeyResults []notifications.StrategyKeyResultSnapshot,
) []notifications.StrategyWeeklyCheckInTeamCountsSnapshot {
	byTeam := make(map[uuid.UUID]*strategyWeeklyTeamCountAccumulator)
	teamAccumulator := func(teamID uuid.UUID) *strategyWeeklyTeamCountAccumulator {
		accumulator, ok := byTeam[teamID]
		if !ok {
			accumulator = &strategyWeeklyTeamCountAccumulator{
				objectiveIDs: make(map[uuid.UUID]struct{}),
			}
			byTeam[teamID] = accumulator
		}
		return accumulator
	}

	for _, objective := range checkIn.Objectives {
		accumulator := teamAccumulator(objective.TeamID)
		accumulator.objectiveDetails++
		accumulator.objectiveIDs[objective.ID] = struct{}{}
		if containsString(objective.Reasons, notifications.StrategySignalReasonAtRisk) {
			accumulator.counts.AtRiskObjectives++
		}
		if containsString(objective.Reasons, notifications.StrategySignalReasonStale) {
			accumulator.counts.StaleObjectives++
		}
	}
	for _, keyResult := range checkIn.KeyResults {
		accumulator := teamAccumulator(keyResult.TeamID)
		accumulator.keyResultDetails++
		accumulator.counts.StaleKeyResults++
		accumulator.objectiveIDs[keyResult.ObjectiveID] = struct{}{}
	}
	selectedObjectiveDetails := make(map[uuid.UUID]int)
	for _, objective := range selectedObjectives {
		selectedObjectiveDetails[objective.TeamID]++
	}
	selectedKeyResultDetails := make(map[uuid.UUID]int)
	for _, keyResult := range selectedKeyResults {
		selectedKeyResultDetails[keyResult.TeamID]++
	}

	teamIDs := make([]uuid.UUID, 0, len(byTeam))
	for teamID := range byTeam {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(i, j int) bool {
		return teamIDs[i].String() < teamIDs[j].String()
	})

	teamCounts := make([]notifications.StrategyWeeklyCheckInTeamCountsSnapshot, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		accumulator := byTeam[teamID]
		accumulator.counts.UniqueObjectives = len(accumulator.objectiveIDs)
		omittedObjectives := max(0, accumulator.objectiveDetails-selectedObjectiveDetails[teamID])
		omittedKeyResults := max(0, accumulator.keyResultDetails-selectedKeyResultDetails[teamID])
		var omittedDetails *notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot
		if omittedObjectives > 0 || omittedKeyResults > 0 {
			omittedDetails = &notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot{
				Objectives: omittedObjectives,
				KeyResults: omittedKeyResults,
			}
		}
		teamCounts = append(teamCounts, notifications.StrategyWeeklyCheckInTeamCountsSnapshot{
			TeamID:         teamID,
			Counts:         accumulator.counts,
			OmittedDetails: omittedDetails,
		})
	}
	return teamCounts
}

func boundedStrategyCheckInDetails(
	checkIn strategyCheckIn,
	limit int,
) (
	[]notifications.StrategyObjectiveSnapshot,
	[]notifications.StrategyKeyResultSnapshot,
	*notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot,
) {
	if limit < 0 {
		limit = 0
	}

	objectiveLimit := min(len(checkIn.Objectives), (limit+1)/2)
	keyResultLimit := min(len(checkIn.KeyResults), limit-objectiveLimit)
	objectiveLimit = min(len(checkIn.Objectives), limit-keyResultLimit)
	if remaining := limit - objectiveLimit - keyResultLimit; remaining > 0 {
		keyResultLimit = min(len(checkIn.KeyResults), keyResultLimit+remaining)
	}

	objectives := checkIn.Objectives[:objectiveLimit]
	keyResults := checkIn.KeyResults[:keyResultLimit]
	omittedObjectives := len(checkIn.Objectives) - len(objectives)
	omittedKeyResults := len(checkIn.KeyResults) - len(keyResults)
	if omittedObjectives == 0 && omittedKeyResults == 0 {
		return objectives, keyResults, nil
	}

	return objectives, keyResults, &notifications.StrategyWeeklyCheckInOmittedDetailsSnapshot{
		Objectives: omittedObjectives,
		KeyResults: omittedKeyResults,
	}
}

func uniqueStrategyObjectiveCount(checkIn strategyCheckIn) int {
	objectiveIDs := make(map[uuid.UUID]struct{}, len(checkIn.Objectives)+len(checkIn.KeyResults))
	for _, objective := range checkIn.Objectives {
		objectiveIDs[objective.ID] = struct{}{}
	}
	for _, keyResult := range checkIn.KeyResults {
		objectiveIDs[keyResult.ObjectiveID] = struct{}{}
	}
	return len(objectiveIDs)
}

func getStrategyWeeklyCheckIns(ctx context.Context, db *sqlx.DB, now time.Time) ([]strategyCheckIn, error) {
	cutoff := now.UTC().Add(-time.Duration(strategyStaleAfterDays) * 24 * time.Hour)
	var records []strategyCheckInRecord
	// The membership predicate mirrors the team route: a recipient must belong
	// to the objective's team or be a workspace administrator. That keeps both
	// the snapshot contents and its eventual deep links within the user's access.
	if err := db.SelectContext(ctx, &records, strategyWeeklyCheckInsQuery(), cutoff, now.UTC(), strategyCommunicationHour); err != nil {
		return nil, fmt.Errorf("load weekly strategy check-ins: %w", err)
	}

	return buildStrategyCheckIns(records), nil
}

func strategyWeeklyCheckInsQuery() string {
	return `
		-- Materialize the small, local-time-due recipient set before touching
		-- objective and key-result signals. Unknown timezone names intentionally
		-- fall back to UTC, matching strategyLocation.
		WITH due_recipients AS MATERIALIZED (
			SELECT
				u.user_id,
				wm.workspace_id,
				wm.role,
				COALESCE(timezone_names.name, 'UTC') AS timezone
			FROM workspace_members wm
			JOIN users u ON u.user_id = wm.user_id AND u.is_active = true AND u.is_system = false
			JOIN workspaces w ON w.workspace_id = wm.workspace_id AND w.deleted_at IS NULL
			LEFT JOIN pg_timezone_names timezone_names
				ON timezone_names.name = NULLIF(TRIM(u.timezone), '')
			WHERE EXTRACT(ISODOW FROM CAST($2 AS TIMESTAMPTZ) AT TIME ZONE COALESCE(timezone_names.name, 'UTC')) = 3
				AND EXTRACT(HOUR FROM CAST($2 AS TIMESTAMPTZ) AT TIME ZONE COALESCE(timezone_names.name, 'UTC')) = $3
		),
		eligible_objectives AS (
			SELECT
				recipient.user_id,
				o.workspace_id,
				recipient.timezone,
				o.objective_id,
				o.team_id,
				o.name AS objective_name,
				CAST(o.health AS TEXT) AS objective_health,
				os.status_id AS objective_status_id,
				os.name AS objective_status_name,
				os.category AS objective_status_category,
				o.start_date AS objective_start_date,
				o.end_date AS objective_end_date,
				o.updated_at AS objective_updated_at,
				o.updated_at < $1 AS is_stale_objective,
				COALESCE(CAST(o.health AS TEXT) IN ('At Risk', 'Off Track'), false) AS is_at_risk_objective
			FROM due_recipients recipient
			JOIN objectives o ON o.workspace_id = recipient.workspace_id AND o.lead_user_id = recipient.user_id
			LEFT JOIN objective_statuses os ON os.status_id = o.status_id
			WHERE o.team_id IS NOT NULL
				AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled', 'paused')
				AND (
					recipient.role = 'admin'
					OR EXISTS (
						SELECT 1
						FROM team_members tm
						WHERE tm.team_id = o.team_id AND tm.user_id = recipient.user_id
					)
				)
		),
		stale_key_results AS (
			SELECT
				kr.id,
				kr.objective_id,
				kr.name,
				CAST(kr.measurement_type AS TEXT) AS measurement_type,
				kr.start_value,
				kr.current_value,
				kr.target_value,
				kr.start_date,
				kr.end_date,
				kr.updated_at
			FROM key_results kr
			JOIN eligible_objectives eo ON eo.objective_id = kr.objective_id
			WHERE kr.updated_at < $1
				AND NOT COALESCE(
					CASE
						WHEN CAST(kr.measurement_type AS TEXT) IN ('percentage', 'number')
							AND kr.target_value >= kr.start_value
							THEN kr.current_value >= kr.target_value
						WHEN CAST(kr.measurement_type AS TEXT) IN ('percentage', 'number')
							AND kr.target_value < kr.start_value
							THEN kr.current_value <= kr.target_value
						WHEN CAST(kr.measurement_type AS TEXT) = 'boolean'
							THEN kr.current_value = kr.target_value
						ELSE false
					END,
					false
				)
		)
		SELECT
			eo.user_id,
			eo.workspace_id,
			eo.timezone,
			eo.objective_id,
			eo.team_id,
			eo.objective_name,
			eo.objective_health,
			eo.objective_status_id,
			eo.objective_status_name,
			eo.objective_status_category,
			eo.objective_start_date,
			eo.objective_end_date,
			eo.objective_updated_at,
			eo.is_stale_objective,
			eo.is_at_risk_objective,
			kr.id AS key_result_id,
			kr.name AS key_result_name,
			kr.measurement_type AS key_result_measurement_type,
			kr.start_value AS key_result_start_value,
			kr.current_value AS key_result_current_value,
			kr.target_value AS key_result_target_value,
			kr.start_date AS key_result_start_date,
			kr.end_date AS key_result_end_date,
			kr.updated_at AS key_result_updated_at
		FROM eligible_objectives eo
		LEFT JOIN stale_key_results kr ON kr.objective_id = eo.objective_id
		WHERE eo.is_stale_objective OR eo.is_at_risk_objective OR kr.id IS NOT NULL
		ORDER BY
			eo.workspace_id,
			eo.user_id,
			CASE eo.objective_health WHEN 'Off Track' THEN 0 WHEN 'At Risk' THEN 1 ELSE 2 END,
			eo.objective_updated_at,
			eo.objective_id,
			kr.updated_at,
			kr.id
	`
}

type strategyCheckInKey struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
}

type strategyCheckInAccumulator struct {
	checkIn    strategyCheckIn
	objectives map[uuid.UUID]notifications.StrategyObjectiveSnapshot
	keyResults map[uuid.UUID]notifications.StrategyKeyResultSnapshot
}

func buildStrategyCheckIns(records []strategyCheckInRecord) []strategyCheckIn {
	accumulators := make(map[strategyCheckInKey]*strategyCheckInAccumulator)
	for _, record := range records {
		key := strategyCheckInKey{UserID: record.UserID, WorkspaceID: record.WorkspaceID}
		accumulator, ok := accumulators[key]
		if !ok {
			accumulator = &strategyCheckInAccumulator{
				checkIn: strategyCheckIn{
					strategyRecipient: record.strategyRecipient,
					Objectives:        make([]notifications.StrategyObjectiveSnapshot, 0),
					KeyResults:        make([]notifications.StrategyKeyResultSnapshot, 0),
				},
				objectives: make(map[uuid.UUID]notifications.StrategyObjectiveSnapshot),
				keyResults: make(map[uuid.UUID]notifications.StrategyKeyResultSnapshot),
			}
			accumulators[key] = accumulator
		}

		objectiveReasons := make([]string, 0, 2)
		if record.IsAtRiskObjective {
			objectiveReasons = append(objectiveReasons, notifications.StrategySignalReasonAtRisk)
		}
		if record.IsStaleObjective {
			objectiveReasons = append(objectiveReasons, notifications.StrategySignalReasonStale)
		}
		status := strategyObjectiveStatus(record)
		if len(objectiveReasons) > 0 {
			accumulator.objectives[record.ObjectiveID] = notifications.StrategyObjectiveSnapshot{
				ID:        record.ObjectiveID,
				TeamID:    record.TeamID,
				Name:      record.ObjectiveName,
				Health:    copyString(record.ObjectiveHealth),
				Status:    status,
				StartDate: copyTime(record.ObjectiveStartDate),
				EndDate:   copyTime(record.ObjectiveEndDate),
				UpdatedAt: record.ObjectiveUpdatedAt,
				Reasons:   objectiveReasons,
			}
		}

		if record.KeyResultID != nil {
			accumulator.keyResults[*record.KeyResultID] = notifications.StrategyKeyResultSnapshot{
				ID:              *record.KeyResultID,
				ObjectiveID:     record.ObjectiveID,
				TeamID:          record.TeamID,
				Name:            stringValue(record.KeyResultName),
				ObjectiveName:   record.ObjectiveName,
				ObjectiveHealth: copyString(record.ObjectiveHealth),
				ObjectiveStatus: status,
				MeasurementType: stringValue(record.KeyResultMeasurementType),
				StartValue:      copyFloat(record.KeyResultStartValue),
				CurrentValue:    copyFloat(record.KeyResultCurrentValue),
				TargetValue:     copyFloat(record.KeyResultTargetValue),
				StartDate:       copyTime(record.KeyResultStartDate),
				EndDate:         copyTime(record.KeyResultEndDate),
				UpdatedAt:       timeValue(record.KeyResultUpdatedAt),
				Reasons: []string{
					notifications.StrategySignalReasonStale,
					notifications.StrategySignalReasonIncomplete,
				},
			}
		}
	}

	keys := make([]strategyCheckInKey, 0, len(accumulators))
	for key := range accumulators {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].WorkspaceID != keys[j].WorkspaceID {
			return keys[i].WorkspaceID.String() < keys[j].WorkspaceID.String()
		}
		return keys[i].UserID.String() < keys[j].UserID.String()
	})

	checkIns := make([]strategyCheckIn, 0, len(keys))
	for _, key := range keys {
		accumulator := accumulators[key]
		for _, objective := range accumulator.objectives {
			accumulator.checkIn.Objectives = append(accumulator.checkIn.Objectives, objective)
			if containsString(objective.Reasons, notifications.StrategySignalReasonAtRisk) {
				accumulator.checkIn.AtRiskObjectives++
			}
			if containsString(objective.Reasons, notifications.StrategySignalReasonStale) {
				accumulator.checkIn.StaleObjectives++
			}
		}
		for _, keyResult := range accumulator.keyResults {
			accumulator.checkIn.KeyResults = append(accumulator.checkIn.KeyResults, keyResult)
		}
		accumulator.checkIn.StaleKeyResults = len(accumulator.checkIn.KeyResults)

		sort.Slice(accumulator.checkIn.Objectives, func(i, j int) bool {
			left := accumulator.checkIn.Objectives[i]
			right := accumulator.checkIn.Objectives[j]
			if strategyHealthRank(left.Health) != strategyHealthRank(right.Health) {
				return strategyHealthRank(left.Health) < strategyHealthRank(right.Health)
			}
			if !left.UpdatedAt.Equal(right.UpdatedAt) {
				return left.UpdatedAt.Before(right.UpdatedAt)
			}
			return left.ID.String() < right.ID.String()
		})
		sort.Slice(accumulator.checkIn.KeyResults, func(i, j int) bool {
			left := accumulator.checkIn.KeyResults[i]
			right := accumulator.checkIn.KeyResults[j]
			if !left.UpdatedAt.Equal(right.UpdatedAt) {
				return left.UpdatedAt.Before(right.UpdatedAt)
			}
			return left.ID.String() < right.ID.String()
		})

		checkIns = append(checkIns, accumulator.checkIn)
	}
	return checkIns
}

func strategyObjectiveStatus(record strategyCheckInRecord) *notifications.StrategyObjectiveStatusSnapshot {
	if record.ObjectiveStatusID == nil || record.ObjectiveStatusName == nil {
		return nil
	}
	return &notifications.StrategyObjectiveStatusSnapshot{
		ID:       *record.ObjectiveStatusID,
		Name:     *record.ObjectiveStatusName,
		Category: stringValue(record.ObjectiveStatusCategory),
	}
}

func strategyHealthRank(health *string) int {
	if health == nil {
		return 3
	}
	switch *health {
	case "Off Track":
		return 0
	case "At Risk":
		return 1
	default:
		return 2
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func copyString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func processStrategyMonthlySummaries(ctx context.Context, db *sqlx.DB, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	recipients, err := getStrategyAdminRecipients(ctx, db)
	if err != nil {
		return err
	}

	var processingErrors []error
	for _, recipient := range recipients {
		localNow := now.In(strategyLocation(recipient.Timezone))
		if localNow.Hour() != strategyCommunicationHour || localNow.Day() != 1 {
			continue
		}

		monthStart := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, localNow.Location())
		previousMonthStart := monthStart.AddDate(0, -1, 0)
		summary, loadErr := getStrategyMonthlySummary(ctx, db, recipient.WorkspaceID, previousMonthStart.UTC(), monthStart.UTC())
		if loadErr != nil {
			processingErrors = append(processingErrors, loadErr)
			continue
		}
		if summary.PillarCount == 0 && summary.ObjectiveCount == 0 {
			continue
		}

		title := fmt.Sprintf("%s strategy summary", previousMonthStart.Format("January"))
		summaryText := strategyMonthlySummaryText(summary)
		if summary.PillarsNeedingReview > 0 {
			summaryText = fmt.Sprintf("%d pillars need review; %s", summary.PillarsNeedingReview, summaryText)
		}

		notification := notifications.CoreNewNotification{
			DedupeKey:   fmt.Sprintf("strategy:monthly:%s:%s:%s", recipient.WorkspaceID, recipient.UserID, previousMonthStart.Format("2006-01")),
			RecipientID: recipient.UserID,
			WorkspaceID: recipient.WorkspaceID,
			Type:        "strategy_update",
			EntityType:  "strategy",
			EntityID:    recipient.WorkspaceID,
			ActorID:     systemUserID,
			Title:       title,
			Message: notifications.NotificationMessage{
				Template: "Here is the current strategy snapshot and last month's linked delivery: {summary}.",
				Variables: map[string]notifications.Variable{
					"summary": {Value: summaryText, Type: "value"},
				},
				Strategy: &notifications.StrategyNotificationSnapshot{
					Version:     strategySnapshotVersion,
					Kind:        notifications.StrategyNotificationKindMonthlySummary,
					GeneratedAt: now.UTC(),
					MonthlySummary: &notifications.StrategyMonthlySummarySnapshot{
						PeriodStart:          previousMonthStart,
						PeriodEnd:            monthStart,
						PillarCount:          summary.PillarCount,
						PillarsNeedingReview: summary.PillarsNeedingReview,
						ObjectiveCount:       summary.ObjectiveCount,
						AtRiskObjectives:     summary.AtRiskObjectives,
						UnalignedObjectives:  summary.UnalignedObjectives,
						KeyResultCount:       summary.KeyResultCount,
						KeyResultProgress:    summary.KeyResultProgress,
						CompletedStories:     summary.CompletedStories,
					},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
}

func strategyMonthlySummaryText(summary strategyMonthlySummary) string {
	parts := make([]string, 0, 4)
	switch {
	case summary.KeyResultProgress != nil && summary.KeyResultCount > 0:
		parts = append(parts, fmt.Sprintf(
			"%.0f%% average progress across %d %s",
			*summary.KeyResultProgress,
			summary.KeyResultCount,
			pluralize(summary.KeyResultCount, "key result", "key results"),
		))
	case summary.KeyResultProgress != nil:
		// Older persisted snapshots did not carry a count. Preserve their real
		// progress instead of treating a missing additive field as no data.
		parts = append(parts, fmt.Sprintf("%.0f%% average key-result progress", *summary.KeyResultProgress))
	case summary.KeyResultCount > 0:
		parts = append(parts, fmt.Sprintf(
			"progress is unavailable for %d %s",
			summary.KeyResultCount,
			pluralize(summary.KeyResultCount, "key result", "key results"),
		))
	default:
		parts = append(parts, "no key results in the current snapshot")
	}
	parts = append(parts,
		fmt.Sprintf("%d objectives needing attention", summary.AtRiskObjectives),
		fmt.Sprintf("%d unaligned objectives", summary.UnalignedObjectives),
		fmt.Sprintf("%d linked stories completed last month", summary.CompletedStories),
	)
	return strings.Join(parts, ", ")
}

func getStrategyAdminRecipients(ctx context.Context, db *sqlx.DB) ([]strategyRecipient, error) {
	var recipients []strategyRecipient
	if err := db.SelectContext(ctx, &recipients, `
		SELECT
			u.user_id,
			wm.workspace_id,
			COALESCE(NULLIF(TRIM(u.timezone), ''), 'UTC') AS timezone
		FROM workspace_members wm
		JOIN users u ON u.user_id = wm.user_id AND u.is_active = true AND u.is_system = false
		JOIN workspaces w ON w.workspace_id = wm.workspace_id AND w.deleted_at IS NULL
		WHERE wm.role = 'admin'
		ORDER BY wm.workspace_id, u.user_id
	`); err != nil {
		return nil, fmt.Errorf("load strategy administrators: %w", err)
	}
	return recipients, nil
}

func getStrategyFoundation(ctx context.Context, db *sqlx.DB, workspaceID uuid.UUID, quarterStart time.Time) (strategyFoundation, error) {
	quarterEnd := quarterStart.AddDate(0, 3, 0)
	var foundation strategyFoundation
	if err := db.GetContext(ctx, &foundation, strategyFoundationQuery(), workspaceID, quarterStart.Format("2006-01-02"), quarterEnd.Format("2006-01-02")); err != nil {
		return strategyFoundation{}, fmt.Errorf("load strategy foundation for workspace %s: %w", workspaceID, err)
	}
	return foundation, nil
}

func strategyFoundationQuery() string {
	return `
		SELECT
			EXISTS (
				SELECT 1 FROM workspace_strategies ws
				WHERE ws.workspace_id = $1 AND NULLIF(TRIM(ws.ultimate_goal), '') IS NOT NULL
			) AS has_ultimate_goal,
			(SELECT COUNT(*) FROM strategic_pillars sp WHERE sp.workspace_id = $1) AS pillar_count,
			(
				SELECT COUNT(*)
				FROM objectives o
				LEFT JOIN objective_statuses os ON os.status_id = o.status_id
				WHERE o.workspace_id = $1
					AND (o.start_date IS NULL OR o.start_date < $3)
					AND (o.end_date IS NULL OR o.end_date >= $2)
					AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled', 'paused')
			) AS objective_count
	`
}

func getStrategyMonthlySummary(ctx context.Context, db *sqlx.DB, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (strategyMonthlySummary, error) {
	var summary strategyMonthlySummary
	if err := db.GetContext(ctx, &summary, strategyMonthlySummaryQuery(), workspaceID, periodStart, periodEnd); err != nil {
		return strategyMonthlySummary{}, fmt.Errorf("load monthly strategy summary for workspace %s: %w", workspaceID, err)
	}
	return summary, nil
}

func strategyMonthlySummaryQuery() string {
	return `
		WITH objective_data AS (
			SELECT
				o.objective_id,
				o.health,
				soa.pillar_id
			FROM objectives o
			LEFT JOIN objective_statuses os ON os.status_id = o.status_id
			LEFT JOIN strategy_objective_alignments soa ON soa.objective_id = o.objective_id
			WHERE o.workspace_id = $1
				AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled', 'paused')
		),
		key_result_data AS (
			SELECT
				CASE
					WHEN CAST(kr.measurement_type AS TEXT) = 'percentage' THEN
						GREATEST(0.0, LEAST(100.0, kr.current_value))
					WHEN CAST(kr.measurement_type AS TEXT) = 'number' AND kr.target_value = kr.start_value THEN
						CASE WHEN kr.current_value = kr.target_value THEN 100.0 ELSE 0.0 END
					WHEN CAST(kr.measurement_type AS TEXT) = 'number' THEN
						GREATEST(0.0, LEAST(100.0,
						((kr.current_value - kr.start_value) / NULLIF(kr.target_value - kr.start_value, 0)) * 100.0
					))
					WHEN CAST(kr.measurement_type AS TEXT) = 'boolean' THEN
						CASE WHEN kr.current_value = kr.target_value THEN 100.0 ELSE 0.0 END
					ELSE NULL
				END AS progress
			FROM key_results kr
			JOIN objective_data objective ON objective.objective_id = kr.objective_id
		)
		SELECT
			(SELECT COUNT(*) FROM strategic_pillars WHERE workspace_id = $1) AS pillar_count,
			(
				SELECT COUNT(DISTINCT pillar_id) FROM objective_data
				WHERE pillar_id IS NOT NULL AND health IN ('At Risk', 'Off Track')
			) AS pillars_needing_review,
			(SELECT COUNT(*) FROM objective_data) AS objective_count,
			(SELECT COUNT(*) FROM objective_data WHERE health IN ('At Risk', 'Off Track')) AS at_risk_objectives,
			(SELECT COUNT(*) FROM objective_data WHERE pillar_id IS NULL) AS unaligned_objectives,
			(SELECT COUNT(*) FROM key_result_data) AS key_result_count,
			(SELECT AVG(progress) FROM key_result_data) AS key_result_progress,
			(
				SELECT COUNT(*) FROM stories s
				WHERE s.workspace_id = $1
					AND (s.objective_id IS NOT NULL OR s.key_result_id IS NOT NULL)
					AND s.completed_at >= $2
					AND s.completed_at < $3
					AND s.deleted_at IS NULL
			) AS completed_stories
	`
}

func strategyLocation(timezone string) *time.Location {
	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.UTC
	}
	return location
}

func nextQuarterStart(now time.Time) time.Time {
	currentQuarter := (int(now.Month()) - 1) / 3
	nextMonth := time.Month((currentQuarter+1)*3 + 1)
	year := now.Year()
	if nextMonth > time.December {
		nextMonth = time.January
		year++
	}
	return time.Date(year, nextMonth, 1, 0, 0, 0, 0, now.Location())
}

func quarterForMonth(month time.Month) int {
	return (int(month)-1)/3 + 1
}

func calendarDaysBetween(from, to time.Time) int {
	// UTC is intentional here: only the local calendar components matter, and
	// using the local offset would make DST boundaries appear 23 or 25 hours long.
	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(toDate.Sub(fromDate).Hours() / 24)
}

func missingStrategyElements(foundation strategyFoundation) string {
	missing := make([]string, 0, 3)
	if !foundation.HasUltimateGoal {
		missing = append(missing, "your ultimate goal")
	}
	if foundation.PillarCount == 0 {
		missing = append(missing, "strategic pillars")
	}
	if foundation.ObjectiveCount == 0 {
		missing = append(missing, "the next period's objectives")
	}
	return strings.Join(missing, ", ")
}

func missingStrategyElementKeys(foundation strategyFoundation) []string {
	missing := make([]string, 0, 3)
	if !foundation.HasUltimateGoal {
		missing = append(missing, notifications.StrategyMissingElementUltimateGoal)
	}
	if foundation.PillarCount == 0 {
		missing = append(missing, notifications.StrategyMissingElementPillars)
	}
	if foundation.ObjectiveCount == 0 {
		missing = append(missing, notifications.StrategyMissingElementObjectives)
	}
	return missing
}

func strategyCheckInSummary(checkIn strategyCheckIn) string {
	parts := make([]string, 0, 3)
	if checkIn.AtRiskObjectives > 0 {
		parts = append(parts, fmt.Sprintf("%d at-risk objectives", checkIn.AtRiskObjectives))
	}
	if checkIn.StaleObjectives > 0 {
		parts = append(parts, fmt.Sprintf("%d objectives without a recent update", checkIn.StaleObjectives))
	}
	if checkIn.StaleKeyResults > 0 {
		parts = append(parts, fmt.Sprintf("%d stalled key results", checkIn.StaleKeyResults))
	}
	return strings.Join(parts, ", ")
}
