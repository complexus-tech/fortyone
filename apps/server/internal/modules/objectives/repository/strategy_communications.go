package objectivesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const maximumStrategyCommunicationBatchSize = 100

var (
	errInvalidStrategyCommunicationRecipient = errors.New("strategy communication query returned an invalid recipient")
	errInvalidStrategyWeeklySignal           = errors.New("strategy communication query returned an invalid weekly signal")
)

// ListStrategyCommunicationAdministrators returns a stable, bounded page of
// active administrators in non-deleted workspaces.
func (repository *Repository) ListStrategyCommunicationAdministrators(
	ctx context.Context,
	cursor *objectivesdomain.StrategyCommunicationCursor,
	limit int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	params, err := strategyCommunicationRecipientParams(cursor, limit)
	if err != nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, err
	}
	if repository == nil || repository.queries == nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, errObjectiveRepositoryNotConfigured
	}

	rows, err := repository.queries.ListStrategyCommunicationAdministrators(ctx, params)
	if err != nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, fmt.Errorf("list strategy communication administrators: %w", mapDatabaseError(err))
	}
	return strategyCommunicationAdministratorPage(rows, limit)
}

// ListStrategyWeeklyCommunicationRecipients returns a stable, bounded page of
// active workspace members. Local-time due evaluation and signal loading happen
// in the job after the page is read, keeping this query bounded and portable.
func (repository *Repository) ListStrategyWeeklyCommunicationRecipients(
	ctx context.Context,
	cursor *objectivesdomain.StrategyCommunicationCursor,
	limit int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	params, err := strategyCommunicationWeeklyRecipientParams(cursor, limit)
	if err != nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, err
	}
	if repository == nil || repository.queries == nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, errObjectiveRepositoryNotConfigured
	}

	rows, err := repository.queries.ListStrategyWeeklyCommunicationRecipients(ctx, params)
	if err != nil {
		return objectivesdomain.StrategyCommunicationRecipientPage{}, fmt.Errorf("list weekly strategy communication recipients: %w", mapDatabaseError(err))
	}
	return strategyCommunicationWeeklyRecipientPage(rows, limit)
}

func strategyCommunicationRecipientParams(
	cursor *objectivesdomain.StrategyCommunicationCursor,
	limit int,
) (objectivessql.ListStrategyCommunicationAdministratorsParams, error) {
	rowLimit, err := strategyCommunicationLookaheadLimit(limit)
	if err != nil {
		return objectivessql.ListStrategyCommunicationAdministratorsParams{}, err
	}
	params := objectivessql.ListStrategyCommunicationAdministratorsParams{ResultLimit: rowLimit}
	if cursor == nil {
		return params, nil
	}
	if cursor.WorkspaceID == uuid.Nil || cursor.UserID == uuid.Nil {
		return objectivessql.ListStrategyCommunicationAdministratorsParams{}, fmt.Errorf("%w: strategy communication cursor requires workspace and user", objectivesdomain.ErrInvalid)
	}
	params.HasCursor = true
	params.AfterWorkspaceID = cursor.WorkspaceID
	params.AfterUserID = cursor.UserID
	return params, nil
}

func strategyCommunicationWeeklyRecipientParams(
	cursor *objectivesdomain.StrategyCommunicationCursor,
	limit int,
) (objectivessql.ListStrategyWeeklyCommunicationRecipientsParams, error) {
	rowLimit, err := strategyCommunicationLookaheadLimit(limit)
	if err != nil {
		return objectivessql.ListStrategyWeeklyCommunicationRecipientsParams{}, err
	}
	params := objectivessql.ListStrategyWeeklyCommunicationRecipientsParams{ResultLimit: rowLimit}
	if cursor == nil {
		return params, nil
	}
	if cursor.WorkspaceID == uuid.Nil || cursor.UserID == uuid.Nil {
		return objectivessql.ListStrategyWeeklyCommunicationRecipientsParams{}, fmt.Errorf("%w: weekly strategy communication cursor requires workspace and user", objectivesdomain.ErrInvalid)
	}
	params.HasCursor = true
	params.AfterWorkspaceID = cursor.WorkspaceID
	params.AfterUserID = cursor.UserID
	return params, nil
}

func strategyCommunicationLookaheadLimit(limit int) (int32, error) {
	if limit < 1 || limit > maximumStrategyCommunicationBatchSize {
		return 0, fmt.Errorf("%w: strategy communication limit must be between 1 and %d", objectivesdomain.ErrInvalid, maximumStrategyCommunicationBatchSize)
	}
	rowLimit, err := safecast.Int32(limit + 1)
	if err != nil {
		return 0, fmt.Errorf("%w: strategy communication limit: %v", objectivesdomain.ErrInvalid, err)
	}
	return rowLimit, nil
}

func strategyCommunicationAdministratorPage(
	rows []objectivessql.ListStrategyCommunicationAdministratorsRow,
	limit int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	recipients := make([]objectivesdomain.StrategyCommunicationRecipient, 0, len(rows))
	for _, row := range rows {
		recipient, err := strategyCommunicationRecipient(row.UserID, row.WorkspaceID, row.Timezone)
		if err != nil {
			return objectivesdomain.StrategyCommunicationRecipientPage{}, err
		}
		recipients = append(recipients, recipient)
	}
	return objectivesdomain.StrategyCommunicationRecipientPage{Recipients: recipients, HasMore: hasMore}, nil
}

func strategyCommunicationWeeklyRecipientPage(
	rows []objectivessql.ListStrategyWeeklyCommunicationRecipientsRow,
	limit int,
) (objectivesdomain.StrategyCommunicationRecipientPage, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	recipients := make([]objectivesdomain.StrategyCommunicationRecipient, 0, len(rows))
	for _, row := range rows {
		recipient, err := strategyCommunicationRecipient(row.UserID, row.WorkspaceID, row.Timezone)
		if err != nil {
			return objectivesdomain.StrategyCommunicationRecipientPage{}, err
		}
		recipients = append(recipients, recipient)
	}
	return objectivesdomain.StrategyCommunicationRecipientPage{Recipients: recipients, HasMore: hasMore}, nil
}

func strategyCommunicationRecipient(userID, workspaceID uuid.UUID, timezone string) (objectivesdomain.StrategyCommunicationRecipient, error) {
	if userID == uuid.Nil || workspaceID == uuid.Nil {
		return objectivesdomain.StrategyCommunicationRecipient{}, errInvalidStrategyCommunicationRecipient
	}
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	return objectivesdomain.StrategyCommunicationRecipient{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Timezone:    timezone,
	}, nil
}

// GetStrategyCommunicationFoundation loads the planning-period foundation for
// one workspace using explicit application-owned calendar boundaries.
func (repository *Repository) GetStrategyCommunicationFoundation(
	ctx context.Context,
	workspaceID uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) (objectivesdomain.StrategyCommunicationFoundation, error) {
	if repository == nil || repository.queries == nil {
		return objectivesdomain.StrategyCommunicationFoundation{}, errObjectiveRepositoryNotConfigured
	}
	periodStart, periodEnd, err := validateStrategyCommunicationPeriod(workspaceID, periodStart, periodEnd)
	if err != nil {
		return objectivesdomain.StrategyCommunicationFoundation{}, err
	}
	row, err := repository.queries.GetStrategyCommunicationFoundation(ctx, objectivessql.GetStrategyCommunicationFoundationParams{
		WorkspaceID: workspaceID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return objectivesdomain.StrategyCommunicationFoundation{}, fmt.Errorf("get strategy communication foundation: %w", mapDatabaseError(err))
	}
	pillarCount, err := safecast.Int64(row.PillarCount)
	if err != nil {
		return objectivesdomain.StrategyCommunicationFoundation{}, fmt.Errorf("map strategy pillar count: %w", err)
	}
	objectiveCount, err := safecast.Int64(row.ObjectiveCount)
	if err != nil {
		return objectivesdomain.StrategyCommunicationFoundation{}, fmt.Errorf("map strategy objective count: %w", err)
	}
	return objectivesdomain.StrategyCommunicationFoundation{
		HasUltimateGoal: row.HasUltimateGoal,
		PillarCount:     pillarCount,
		ObjectiveCount:  objectiveCount,
	}, nil
}

// GetStrategyCommunicationMonthlySummary loads one local-month delivery window
// and the workspace's current strategy signals.
func (repository *Repository) GetStrategyCommunicationMonthlySummary(
	ctx context.Context,
	workspaceID uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) (objectivesdomain.StrategyCommunicationMonthlySummary, error) {
	if repository == nil || repository.queries == nil {
		return objectivesdomain.StrategyCommunicationMonthlySummary{}, errObjectiveRepositoryNotConfigured
	}
	periodStart, periodEnd, err := validateStrategyCommunicationPeriod(workspaceID, periodStart, periodEnd)
	if err != nil {
		return objectivesdomain.StrategyCommunicationMonthlySummary{}, err
	}
	row, err := repository.queries.GetStrategyCommunicationMonthlySummary(ctx, objectivessql.GetStrategyCommunicationMonthlySummaryParams{
		WorkspaceID: workspaceID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return objectivesdomain.StrategyCommunicationMonthlySummary{}, fmt.Errorf("get strategy communication monthly summary: %w", mapDatabaseError(err))
	}

	counts, err := strategyCommunicationSummaryCounts(row)
	if err != nil {
		return objectivesdomain.StrategyCommunicationMonthlySummary{}, err
	}
	if row.KeyResultProgressCount < 0 || row.KeyResultProgressCount > row.KeyResultCount {
		return objectivesdomain.StrategyCommunicationMonthlySummary{}, fmt.Errorf("%w: inconsistent key-result progress count", errInvalidStrategyWeeklySignal)
	}
	if row.KeyResultProgressCount > 0 {
		progress := row.KeyResultProgress
		counts.KeyResultProgress = &progress
	}
	return counts, nil
}

func strategyCommunicationSummaryCounts(
	row objectivessql.GetStrategyCommunicationMonthlySummaryRow,
) (objectivesdomain.StrategyCommunicationMonthlySummary, error) {
	values := []struct {
		name  string
		value int64
	}{
		{name: "pillar count", value: row.PillarCount},
		{name: "pillars needing review", value: row.PillarsNeedingReview},
		{name: "objective count", value: row.ObjectiveCount},
		{name: "at-risk objective count", value: row.AtRiskObjectives},
		{name: "unaligned objective count", value: row.UnalignedObjectives},
		{name: "key-result count", value: row.KeyResultCount},
		{name: "completed story count", value: row.CompletedStories},
	}
	converted := make([]int, len(values))
	for index, value := range values {
		count, err := safecast.Int64(value.value)
		if err != nil {
			return objectivesdomain.StrategyCommunicationMonthlySummary{}, fmt.Errorf("map strategy communication %s: %w", value.name, err)
		}
		converted[index] = count
	}
	return objectivesdomain.StrategyCommunicationMonthlySummary{
		PillarCount:          converted[0],
		PillarsNeedingReview: converted[1],
		ObjectiveCount:       converted[2],
		AtRiskObjectives:     converted[3],
		UnalignedObjectives:  converted[4],
		KeyResultCount:       converted[5],
		CompletedStories:     converted[6],
	}, nil
}

func validateStrategyCommunicationPeriod(workspaceID uuid.UUID, periodStart, periodEnd time.Time) (time.Time, time.Time, error) {
	if workspaceID == uuid.Nil || periodStart.IsZero() || periodEnd.IsZero() || !periodStart.Before(periodEnd) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: strategy communication workspace and increasing period are required", objectivesdomain.ErrInvalid)
	}
	return periodStart.UTC(), periodEnd.UTC(), nil
}

// ListStrategyWeeklyCommunicationSignals returns one bounded, stable page of
// current weekly signals for one still-authorized objective lead.
func (repository *Repository) ListStrategyWeeklyCommunicationSignals(
	ctx context.Context,
	staleBefore time.Time,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	cursor *objectivesdomain.StrategyWeeklySignalCursor,
	limit int,
) (objectivesdomain.StrategyWeeklySignalPage, error) {
	if repository == nil || repository.queries == nil {
		return objectivesdomain.StrategyWeeklySignalPage{}, errObjectiveRepositoryNotConfigured
	}
	params, err := strategyWeeklySignalParams(staleBefore, userID, workspaceID, cursor, limit)
	if err != nil {
		return objectivesdomain.StrategyWeeklySignalPage{}, err
	}
	rows, err := repository.queries.ListStrategyWeeklyCommunicationRecords(ctx, params)
	if err != nil {
		return objectivesdomain.StrategyWeeklySignalPage{}, fmt.Errorf("list weekly strategy communication signals: %w", mapDatabaseError(err))
	}
	return strategyWeeklySignalPage(rows, userID, workspaceID, limit)
}

func strategyWeeklySignalParams(
	staleBefore time.Time,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	cursor *objectivesdomain.StrategyWeeklySignalCursor,
	limit int,
) (objectivessql.ListStrategyWeeklyCommunicationRecordsParams, error) {
	rowLimit, err := strategyCommunicationLookaheadLimit(limit)
	if err != nil {
		return objectivessql.ListStrategyWeeklyCommunicationRecordsParams{}, err
	}
	if staleBefore.IsZero() || userID == uuid.Nil || workspaceID == uuid.Nil {
		return objectivessql.ListStrategyWeeklyCommunicationRecordsParams{}, fmt.Errorf("%w: weekly strategy cutoff, user, and workspace are required", objectivesdomain.ErrInvalid)
	}
	params := objectivessql.ListStrategyWeeklyCommunicationRecordsParams{
		ResultLimit: rowLimit,
		UserID:      userID,
		WorkspaceID: workspaceID,
		StaleBefore: staleBefore.UTC(),
	}
	if cursor == nil {
		return params, nil
	}
	if cursor.ObjectiveID == uuid.Nil || cursor.KeyResultNullRank < 0 || cursor.KeyResultNullRank > 1 ||
		(cursor.KeyResultNullRank == 0 && cursor.KeyResultID == uuid.Nil) ||
		(cursor.KeyResultNullRank == 1 && cursor.KeyResultID != uuid.Nil) {
		return objectivessql.ListStrategyWeeklyCommunicationRecordsParams{}, fmt.Errorf("%w: weekly strategy signal cursor is invalid", objectivesdomain.ErrInvalid)
	}
	params.HasCursor = true
	params.AfterObjectiveID = cursor.ObjectiveID
	params.AfterKeyResultNullRank, err = safecast.Int32(cursor.KeyResultNullRank)
	if err != nil {
		return objectivessql.ListStrategyWeeklyCommunicationRecordsParams{}, fmt.Errorf("%w: weekly strategy signal cursor rank: %v", objectivesdomain.ErrInvalid, err)
	}
	params.AfterKeyResultID = cursor.KeyResultID
	return params, nil
}

func strategyWeeklySignalPage(
	rows []objectivessql.ListStrategyWeeklyCommunicationRecordsRow,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	limit int,
) (objectivesdomain.StrategyWeeklySignalPage, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	records := make([]objectivesdomain.StrategyWeeklySignalRecord, 0, len(rows))
	for _, row := range rows {
		record, err := strategyWeeklySignalRecord(row, userID, workspaceID)
		if err != nil {
			return objectivesdomain.StrategyWeeklySignalPage{}, err
		}
		records = append(records, record)
	}
	page := objectivesdomain.StrategyWeeklySignalPage{Records: records, HasMore: hasMore}
	if hasMore {
		last := rows[len(rows)-1]
		page.NextCursor = &objectivesdomain.StrategyWeeklySignalCursor{
			ObjectiveID:       last.ObjectiveID,
			KeyResultNullRank: int(last.KeyResultNullRank),
			KeyResultID:       last.KeyResultSortID,
		}
	}
	return page, nil
}

func strategyWeeklySignalRecord(
	row objectivessql.ListStrategyWeeklyCommunicationRecordsRow,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) (objectivesdomain.StrategyWeeklySignalRecord, error) {
	if row.UserID != userID || row.WorkspaceID != workspaceID || row.ObjectiveID == uuid.Nil ||
		row.TeamID == nil || *row.TeamID == uuid.Nil || strings.TrimSpace(row.ObjectiveName) == "" {
		return objectivesdomain.StrategyWeeklySignalRecord{}, errInvalidStrategyWeeklySignal
	}
	if row.KeyResultID == nil {
		if row.KeyResultNullRank != 1 || row.KeyResultSortID != uuid.Nil {
			return objectivesdomain.StrategyWeeklySignalRecord{}, errInvalidStrategyWeeklySignal
		}
	} else if row.KeyResultNullRank != 0 || row.KeyResultSortID != *row.KeyResultID ||
		row.KeyResultName == nil || row.KeyResultMeasurementType == nil ||
		row.KeyResultStartDate == nil || row.KeyResultEndDate == nil || row.KeyResultUpdatedAt == nil {
		return objectivesdomain.StrategyWeeklySignalRecord{}, errInvalidStrategyWeeklySignal
	}
	recipient, err := strategyCommunicationRecipient(row.UserID, row.WorkspaceID, row.Timezone)
	if err != nil {
		return objectivesdomain.StrategyWeeklySignalRecord{}, err
	}
	var objectiveHealth *string
	if row.ObjectiveHealthIsSet {
		health := row.ObjectiveHealth
		objectiveHealth = &health
	}
	return objectivesdomain.StrategyWeeklySignalRecord{
		Recipient:                recipient,
		ObjectiveID:              row.ObjectiveID,
		TeamID:                   *row.TeamID,
		ObjectiveName:            row.ObjectiveName,
		ObjectiveHealth:          objectiveHealth,
		ObjectiveStatusID:        row.ObjectiveStatusID,
		ObjectiveStatusName:      row.ObjectiveStatusName,
		ObjectiveStatusCategory:  row.ObjectiveStatusCategory,
		ObjectiveStartDate:       utcTimePointer(row.ObjectiveStartDate),
		ObjectiveEndDate:         utcTimePointer(row.ObjectiveEndDate),
		ObjectiveUpdatedAt:       row.ObjectiveUpdatedAt.UTC(),
		IsStaleObjective:         row.IsStaleObjective,
		IsAtRiskObjective:        row.IsAtRiskObjective,
		KeyResultID:              row.KeyResultID,
		KeyResultName:            row.KeyResultName,
		KeyResultMeasurementType: row.KeyResultMeasurementType,
		KeyResultStartValue:      row.KeyResultStartValue,
		KeyResultCurrentValue:    row.KeyResultCurrentValue,
		KeyResultTargetValue:     row.KeyResultTargetValue,
		KeyResultStartDate:       utcTimePointer(row.KeyResultStartDate),
		KeyResultEndDate:         utcTimePointer(row.KeyResultEndDate),
		KeyResultUpdatedAt:       utcTimePointer(row.KeyResultUpdatedAt),
	}, nil
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
