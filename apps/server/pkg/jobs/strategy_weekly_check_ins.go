package jobs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
)

func processStrategyWeeklyCheckIns(ctx context.Context, store StrategyCommunicationsStore, notifier StrategyNotificationCreator, systemUserID uuid.UUID, now time.Time) error {
	staleBefore := now.UTC().Add(-time.Duration(strategyStaleAfterDays) * 24 * time.Hour)
	return processStrategyRecipientPages(ctx, store.ListStrategyWeeklyCommunicationRecipients, func(recipient objectivesdomain.StrategyCommunicationRecipient) error {
		localNow, due := strategyWeeklyLocalTime(now, recipient.Timezone)
		if !due {
			return nil
		}

		checkIns, err := getStrategyWeeklyCheckIns(ctx, store, recipient, staleBefore)
		if err != nil {
			return err
		}
		for _, checkIn := range checkIns {
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
				return createErr
			}
		}
		return nil
	})
}

func strategyWeeklyLocalTime(now time.Time, timezone string) (time.Time, bool) {
	localNow := now.UTC().In(strategyLocation(timezone))
	return localNow, localNow.Hour() == strategyCommunicationHour && localNow.Weekday() == time.Wednesday
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

func getStrategyWeeklyCheckIns(
	ctx context.Context,
	store StrategyCommunicationsStore,
	recipient objectivesdomain.StrategyCommunicationRecipient,
	staleBefore time.Time,
) ([]strategyCheckIn, error) {
	var records []strategyCheckInRecord
	var cursor *objectivesdomain.StrategyWeeklySignalCursor
	for batch := 0; batch < strategyWeeklySignalMaxBatches; batch++ {
		page, err := store.ListStrategyWeeklyCommunicationSignals(
			ctx,
			staleBefore.UTC(),
			recipient.UserID,
			recipient.WorkspaceID,
			cursor,
			strategyWeeklySignalBatchSize,
		)
		if err != nil {
			return nil, fmt.Errorf("load weekly strategy check-in for recipient %s in workspace %s: %w", recipient.UserID, recipient.WorkspaceID, err)
		}
		if len(page.Records) > strategyWeeklySignalBatchSize {
			return nil, fmt.Errorf("weekly strategy signal page exceeds limit: got %d, limit %d", len(page.Records), strategyWeeklySignalBatchSize)
		}
		for _, record := range page.Records {
			records = append(records, strategyCheckInRecordFromDomain(record))
		}
		if !page.HasMore {
			return buildStrategyCheckIns(records), nil
		}
		if page.NextCursor == nil || len(page.Records) == 0 {
			return nil, errors.New("weekly strategy signal page reported more work without a cursor row")
		}
		if cursor != nil && !strategyWeeklySignalCursorAdvances(*cursor, *page.NextCursor) {
			return nil, errors.New("weekly strategy signal cursor did not advance")
		}
		cursor = page.NextCursor
	}
	return nil, errStrategyCommunicationBacklog
}

func strategyWeeklySignalCursorAdvances(previous, next objectivesdomain.StrategyWeeklySignalCursor) bool {
	if previous.ObjectiveID != next.ObjectiveID {
		return previous.ObjectiveID.String() < next.ObjectiveID.String()
	}
	if previous.KeyResultNullRank != next.KeyResultNullRank {
		return previous.KeyResultNullRank < next.KeyResultNullRank
	}
	return previous.KeyResultID.String() < next.KeyResultID.String()
}

func strategyCheckInRecordFromDomain(record objectivesdomain.StrategyWeeklySignalRecord) strategyCheckInRecord {
	return strategyCheckInRecord{
		strategyRecipient:        strategyRecipientFromDomain(record.Recipient),
		ObjectiveID:              record.ObjectiveID,
		TeamID:                   record.TeamID,
		ObjectiveName:            record.ObjectiveName,
		ObjectiveHealth:          record.ObjectiveHealth,
		ObjectiveStatusID:        record.ObjectiveStatusID,
		ObjectiveStatusName:      record.ObjectiveStatusName,
		ObjectiveStatusCategory:  record.ObjectiveStatusCategory,
		ObjectiveStartDate:       record.ObjectiveStartDate,
		ObjectiveEndDate:         record.ObjectiveEndDate,
		ObjectiveUpdatedAt:       record.ObjectiveUpdatedAt,
		IsStaleObjective:         record.IsStaleObjective,
		IsAtRiskObjective:        record.IsAtRiskObjective,
		KeyResultID:              record.KeyResultID,
		KeyResultName:            record.KeyResultName,
		KeyResultMeasurementType: record.KeyResultMeasurementType,
		KeyResultStartValue:      record.KeyResultStartValue,
		KeyResultCurrentValue:    record.KeyResultCurrentValue,
		KeyResultTargetValue:     record.KeyResultTargetValue,
		KeyResultStartDate:       record.KeyResultStartDate,
		KeyResultEndDate:         record.KeyResultEndDate,
		KeyResultUpdatedAt:       record.KeyResultUpdatedAt,
	}
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
