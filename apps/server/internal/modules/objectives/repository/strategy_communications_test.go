package objectivesrepository

import (
	"context"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type strategyCommunicationQueriesStub struct {
	objectivessql.Querier
	administratorParams   objectivessql.ListStrategyCommunicationAdministratorsParams
	administrators        []objectivessql.ListStrategyCommunicationAdministratorsRow
	weeklyRecipientParams objectivessql.ListStrategyWeeklyCommunicationRecipientsParams
	weeklyRecipients      []objectivessql.ListStrategyWeeklyCommunicationRecipientsRow
	foundationParams      objectivessql.GetStrategyCommunicationFoundationParams
	foundation            objectivessql.GetStrategyCommunicationFoundationRow
	monthlyParams         objectivessql.GetStrategyCommunicationMonthlySummaryParams
	monthly               objectivessql.GetStrategyCommunicationMonthlySummaryRow
	signalParams          objectivessql.ListStrategyWeeklyCommunicationRecordsParams
	signals               []objectivessql.ListStrategyWeeklyCommunicationRecordsRow
}

func (stub *strategyCommunicationQueriesStub) ListStrategyCommunicationAdministrators(
	_ context.Context,
	params objectivessql.ListStrategyCommunicationAdministratorsParams,
) ([]objectivessql.ListStrategyCommunicationAdministratorsRow, error) {
	stub.administratorParams = params
	return stub.administrators, nil
}

func (stub *strategyCommunicationQueriesStub) ListStrategyWeeklyCommunicationRecipients(
	_ context.Context,
	params objectivessql.ListStrategyWeeklyCommunicationRecipientsParams,
) ([]objectivessql.ListStrategyWeeklyCommunicationRecipientsRow, error) {
	stub.weeklyRecipientParams = params
	return stub.weeklyRecipients, nil
}

func (stub *strategyCommunicationQueriesStub) GetStrategyCommunicationFoundation(
	_ context.Context,
	params objectivessql.GetStrategyCommunicationFoundationParams,
) (objectivessql.GetStrategyCommunicationFoundationRow, error) {
	stub.foundationParams = params
	return stub.foundation, nil
}

func (stub *strategyCommunicationQueriesStub) GetStrategyCommunicationMonthlySummary(
	_ context.Context,
	params objectivessql.GetStrategyCommunicationMonthlySummaryParams,
) (objectivessql.GetStrategyCommunicationMonthlySummaryRow, error) {
	stub.monthlyParams = params
	return stub.monthly, nil
}

func (stub *strategyCommunicationQueriesStub) ListStrategyWeeklyCommunicationRecords(
	_ context.Context,
	params objectivessql.ListStrategyWeeklyCommunicationRecordsParams,
) ([]objectivessql.ListStrategyWeeklyCommunicationRecordsRow, error) {
	stub.signalParams = params
	return stub.signals, nil
}

func TestListStrategyCommunicationAdministratorsUsesLookaheadAndMapsCursor(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	lookaheadUserID := uuid.New()
	stub := &strategyCommunicationQueriesStub{administrators: []objectivessql.ListStrategyCommunicationAdministratorsRow{
		{UserID: firstUserID, WorkspaceID: workspaceID, Timezone: "UTC"},
		{UserID: secondUserID, WorkspaceID: workspaceID, Timezone: "Africa/Harare"},
		{UserID: lookaheadUserID, WorkspaceID: workspaceID, Timezone: "UTC"},
	}}
	repository := newWithQueries(stub)

	page, err := repository.ListStrategyCommunicationAdministrators(
		context.Background(),
		&objectivesdomain.StrategyCommunicationCursor{WorkspaceID: workspaceID, UserID: firstUserID},
		2,
	)

	require.NoError(t, err)
	require.True(t, page.HasMore)
	require.Len(t, page.Recipients, 2)
	require.Equal(t, int32(3), stub.administratorParams.ResultLimit)
	require.True(t, stub.administratorParams.HasCursor)
	require.Equal(t, workspaceID, stub.administratorParams.AfterWorkspaceID)
	require.Equal(t, firstUserID, stub.administratorParams.AfterUserID)
	require.Equal(t, secondUserID, page.Recipients[1].UserID)
}

func TestListStrategyWeeklyCommunicationRecipientsDoesNotRequireSignals(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	stub := &strategyCommunicationQueriesStub{weeklyRecipients: []objectivessql.ListStrategyWeeklyCommunicationRecipientsRow{{
		UserID: userID, WorkspaceID: workspaceID, Timezone: "Invalid/But-Preserved",
	}}}
	repository := newWithQueries(stub)

	page, err := repository.ListStrategyWeeklyCommunicationRecipients(context.Background(), nil, 100)

	require.NoError(t, err)
	require.False(t, page.HasMore)
	require.Equal(t, int32(101), stub.weeklyRecipientParams.ResultLimit)
	require.Equal(t, "Invalid/But-Preserved", page.Recipients[0].Timezone)
}

func TestGetStrategyCommunicationFoundationMapsTypedCountsAndDates(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	periodStart := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.FixedZone("test", 2*60*60))
	periodEnd := periodStart.AddDate(0, 3, 0)
	stub := &strategyCommunicationQueriesStub{foundation: objectivessql.GetStrategyCommunicationFoundationRow{
		HasUltimateGoal: true,
		PillarCount:     4,
		ObjectiveCount:  12,
	}}
	repository := newWithQueries(stub)

	foundation, err := repository.GetStrategyCommunicationFoundation(context.Background(), workspaceID, periodStart, periodEnd)

	require.NoError(t, err)
	require.Equal(t, periodStart.UTC(), stub.foundationParams.PeriodStart)
	require.Equal(t, periodEnd.UTC(), stub.foundationParams.PeriodEnd)
	require.Equal(t, objectivesdomain.StrategyCommunicationFoundation{HasUltimateGoal: true, PillarCount: 4, ObjectiveCount: 12}, foundation)
}

func TestGetStrategyCommunicationMonthlySummaryPreservesUnavailableProgress(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	periodStart := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)
	stub := &strategyCommunicationQueriesStub{monthly: objectivessql.GetStrategyCommunicationMonthlySummaryRow{
		PillarCount:          2,
		PillarsNeedingReview: 1,
		ObjectiveCount:       4,
		AtRiskObjectives:     2,
		UnalignedObjectives:  1,
		KeyResultCount:       3,
		KeyResultProgress:    0,
		CompletedStories:     8,
	}}
	repository := newWithQueries(stub)

	summary, err := repository.GetStrategyCommunicationMonthlySummary(context.Background(), workspaceID, periodStart, periodEnd)

	require.NoError(t, err)
	require.Nil(t, summary.KeyResultProgress)
	require.Equal(t, 3, summary.KeyResultCount)
	require.Equal(t, 8, summary.CompletedStories)
}

func TestListStrategyWeeklyCommunicationSignalsMapsLookaheadCursorAndNullableHealth(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	teamID := uuid.New()
	firstObjectiveID := uuid.New()
	secondObjectiveID := uuid.New()
	lookaheadObjectiveID := uuid.New()
	keyResultID := uuid.New()
	updatedAt := time.Date(2026, time.August, 18, 8, 0, 0, 0, time.UTC)
	keyResultName := "Protect the release gate"
	measurementType := "number"
	keyResultStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	keyResultEnd := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	stub := &strategyCommunicationQueriesStub{signals: []objectivessql.ListStrategyWeeklyCommunicationRecordsRow{
		{
			UserID: userID, WorkspaceID: workspaceID, Timezone: "UTC", ObjectiveID: firstObjectiveID,
			TeamID: &teamID, ObjectiveName: "First", ObjectiveUpdatedAt: updatedAt, IsStaleObjective: true,
			KeyResultNullRank: 1,
		},
		{
			UserID: userID, WorkspaceID: workspaceID, Timezone: "UTC", ObjectiveID: secondObjectiveID,
			TeamID: &teamID, ObjectiveName: "Second", ObjectiveHealthIsSet: true, ObjectiveHealth: "At Risk",
			ObjectiveUpdatedAt: updatedAt, IsAtRiskObjective: true, KeyResultID: &keyResultID,
			KeyResultName: &keyResultName, KeyResultMeasurementType: &measurementType,
			KeyResultStartDate: &keyResultStart, KeyResultEndDate: &keyResultEnd, KeyResultUpdatedAt: &updatedAt,
			KeyResultNullRank: 0, KeyResultSortID: keyResultID,
		},
		{
			UserID: userID, WorkspaceID: workspaceID, Timezone: "UTC", ObjectiveID: lookaheadObjectiveID,
			TeamID: &teamID, ObjectiveName: "Lookahead", ObjectiveUpdatedAt: updatedAt, IsStaleObjective: true,
			KeyResultNullRank: 1,
		},
	}}
	repository := newWithQueries(stub)
	staleBefore := time.Date(2026, time.August, 19, 9, 0, 0, 0, time.UTC)

	page, err := repository.ListStrategyWeeklyCommunicationSignals(context.Background(), staleBefore, userID, workspaceID, nil, 2)

	require.NoError(t, err)
	require.True(t, page.HasMore)
	require.Len(t, page.Records, 2)
	require.Nil(t, page.Records[0].ObjectiveHealth)
	require.Equal(t, "At Risk", *page.Records[1].ObjectiveHealth)
	require.Equal(t, &objectivesdomain.StrategyWeeklySignalCursor{
		ObjectiveID: secondObjectiveID, KeyResultNullRank: 0, KeyResultID: keyResultID,
	}, page.NextCursor)
	require.Equal(t, int32(3), stub.signalParams.ResultLimit)
	require.Equal(t, staleBefore, stub.signalParams.StaleBefore)
}
