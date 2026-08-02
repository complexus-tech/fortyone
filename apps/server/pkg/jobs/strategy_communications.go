package jobs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	_ "time/tzdata"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const strategyCommunicationHour = 9

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
	StaleObjectives  int `db:"stale_objectives"`
	AtRiskObjectives int `db:"at_risk_objectives"`
	StaleKeyResults  int `db:"stale_key_results"`
}

type strategyMonthlySummary struct {
	PillarCount          int     `db:"pillar_count"`
	PillarsNeedingReview int     `db:"pillars_needing_review"`
	ObjectiveCount       int     `db:"objective_count"`
	AtRiskObjectives     int     `db:"at_risk_objectives"`
	UnalignedObjectives  int     `db:"unaligned_objectives"`
	KeyResultProgress    float64 `db:"key_result_progress"`
	CompletedStories     int     `db:"completed_stories"`
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
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
}

func processStrategyWeeklyCheckIns(ctx context.Context, db *sqlx.DB, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	var checkIns []strategyCheckIn
	if err := db.SelectContext(ctx, &checkIns, `
		SELECT
			u.user_id,
			o.workspace_id,
			COALESCE(NULLIF(TRIM(u.timezone), ''), 'UTC') AS timezone,
			COUNT(DISTINCT o.objective_id) FILTER (
				WHERE o.updated_at < NOW() - INTERVAL '7 days'
				AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled')
			) AS stale_objectives,
			COUNT(DISTINCT o.objective_id) FILTER (
				WHERE o.health IN ('At Risk', 'Off Track')
				AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled')
			) AS at_risk_objectives,
			COUNT(DISTINCT kr.id) FILTER (
				WHERE kr.updated_at < NOW() - INTERVAL '7 days'
				AND COALESCE(os.category, '') NOT IN ('completed', 'cancelled')
				AND kr.current_value IS DISTINCT FROM kr.target_value
			) AS stale_key_results
		FROM objectives o
		JOIN users u ON u.user_id = o.lead_user_id AND u.is_active = true AND u.is_system = false
		JOIN workspace_members wm ON wm.workspace_id = o.workspace_id AND wm.user_id = u.user_id
		JOIN workspaces w ON w.workspace_id = o.workspace_id AND w.deleted_at IS NULL
		LEFT JOIN objective_statuses os ON os.status_id = o.status_id
		LEFT JOIN key_results kr ON kr.objective_id = o.objective_id
		GROUP BY u.user_id, o.workspace_id, u.timezone
	`); err != nil {
		return fmt.Errorf("load weekly strategy check-ins: %w", err)
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

		_, week := localNow.ISOWeek()
		summary := strategyCheckInSummary(checkIn)
		notification := notifications.CoreNewNotification{
			DedupeKey:   fmt.Sprintf("strategy:check-in:%s:%s:%d-%02d", checkIn.WorkspaceID, checkIn.UserID, localNow.Year(), week),
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
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
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

		progress := int(math.Round(summary.KeyResultProgress))
		title := fmt.Sprintf("%s strategy summary", previousMonthStart.Format("January"))
		summaryText := fmt.Sprintf(
			"%d%% key result progress, %d objectives needing attention, %d unaligned objectives, and %d linked stories completed",
			progress,
			summary.AtRiskObjectives,
			summary.UnalignedObjectives,
			summary.CompletedStories,
		)
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
				Template: "Here is the strategy-to-execution picture for last month: {summary}.",
				Variables: map[string]notifications.Variable{
					"summary": {Value: summaryText, Type: "value"},
				},
			},
		}
		if _, createErr := notifier.Create(ctx, notification); createErr != nil {
			processingErrors = append(processingErrors, createErr)
		}
	}
	return errors.Join(processingErrors...)
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
	if err := db.GetContext(ctx, &foundation, `
		SELECT
			EXISTS (
				SELECT 1 FROM workspace_strategies ws
				WHERE ws.workspace_id = $1 AND NULLIF(TRIM(ws.ultimate_goal), '') IS NOT NULL
			) AS has_ultimate_goal,
			(SELECT COUNT(*) FROM strategic_pillars sp WHERE sp.workspace_id = $1) AS pillar_count,
			(
				SELECT COUNT(*) FROM objectives o
				WHERE o.workspace_id = $1
					AND (o.start_date IS NULL OR o.start_date < $3)
					AND (o.end_date IS NULL OR o.end_date >= $2)
			) AS objective_count
	`, workspaceID, quarterStart.Format("2006-01-02"), quarterEnd.Format("2006-01-02")); err != nil {
		return strategyFoundation{}, fmt.Errorf("load strategy foundation for workspace %s: %w", workspaceID, err)
	}
	return foundation, nil
}

func getStrategyMonthlySummary(ctx context.Context, db *sqlx.DB, workspaceID uuid.UUID, periodStart, periodEnd time.Time) (strategyMonthlySummary, error) {
	var summary strategyMonthlySummary
	if err := db.GetContext(ctx, &summary, `
		WITH objective_data AS (
			SELECT
				o.objective_id,
				o.health,
				soa.pillar_id
			FROM objectives o
			LEFT JOIN strategy_objective_alignments soa ON soa.objective_id = o.objective_id
			WHERE o.workspace_id = $1
		),
		key_result_data AS (
			SELECT
				CASE
					WHEN kr.target_value = kr.start_value THEN
						CASE WHEN kr.current_value = kr.target_value THEN 100.0 ELSE 0.0 END
					ELSE GREATEST(0.0, LEAST(100.0,
						((kr.current_value - kr.start_value) / NULLIF(kr.target_value - kr.start_value, 0)) * 100.0
					))
				END AS progress
			FROM key_results kr
			JOIN objectives o ON o.objective_id = kr.objective_id
			WHERE o.workspace_id = $1
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
			COALESCE((SELECT AVG(progress) FROM key_result_data), 0) AS key_result_progress,
			(
				SELECT COUNT(*) FROM stories s
				WHERE s.workspace_id = $1
					AND (s.objective_id IS NOT NULL OR s.key_result_id IS NOT NULL)
					AND s.completed_at >= $2
					AND s.completed_at < $3
					AND s.deleted_at IS NULL
			) AS completed_stories
	`, workspaceID, periodStart, periodEnd); err != nil {
		return strategyMonthlySummary{}, fmt.Errorf("load monthly strategy summary for workspace %s: %w", workspaceID, err)
	}
	return summary, nil
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
