package teamsettingsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestSprintAutomationQueriesAreScopedClockedBoundedAndRetrySafe(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/sprint_automation.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: listsprintautomationteams :many",
		"settings.workspace_id > sqlc.arg(after_workspace_id)",
		"settings.team_id > sqlc.arg(after_team_id)",
		"order by settings.workspace_id, settings.team_id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"-- name: locksprintautomation :exec",
		"pg_advisory_xact_lock",
		"'sprint-automation:'",
		"-- name: countupcomingsprintsforautomation :one",
		"limit cast(sqlc.arg(upcoming_target) as integer)",
		"sprint.team_id = sqlc.arg(team_id)",
		"sprint.workspace_id = sqlc.arg(workspace_id)",
		"-- name: createautomatedsprint :one",
		"cast(sqlc.arg(created_at) as timestamptz)",
		"-- name: advancesprintautomationcounter :execrows",
		"settings.next_auto_sprint_number = cast(sqlc.arg(expected_next_number) as integer)",
		"-- name: listsprintautomationinactivitycandidates :many",
		"-- name: getsprintautomationinactivitysnapshot :one",
		"order by story.created_at desc, story.id",
		"order by activity.created_at desc, activity.activity_id desc",
		"-- name: disablesprintautomationifinactive :execrows",
		"story.created_at >= sqlc.arg(activity_before)",
		"activity.created_at >= sqlc.arg(activity_before)",
		"-- name: insertsprintautomationauditevent :exec",
		"'automation'",
		"cast(sqlc.arg(created_at) as timestamptz)",
	} {
		require.Contains(t, query, contract)
	}
	for _, forbidden := range []string{
		"now()",
		"current_timestamp",
		"current_date",
		"interval '",
		" offset ",
		"select *",
	} {
		require.NotContains(t, query, forbidden)
	}
}

func TestSprintAutomationRepositoryMapsUTCKeysetPages(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	workspaceID := uuid.New()
	teamID := uuid.New()
	queries := &sprintAutomationQueries{
		teamRows: []teamsettingssql.ListSprintAutomationTeamsRow{{WorkspaceID: workspaceID, TeamID: teamID}},
		inactivityRows: []teamsettingssql.ListSprintAutomationInactivityCandidatesRow{{
			WorkspaceID: workspaceID, TeamID: teamID,
		}},
	}
	repository := newWithQueries(queries)
	cursor := teamsettings.SprintAutomationCursor{WorkspaceID: workspaceID, TeamID: teamID, Valid: true}

	teams, err := repository.ListSprintAutomationTeams(
		context.Background(),
		teamsettings.SprintAutomationQuery{Cursor: cursor, BatchSize: 50},
	)
	require.NoError(t, err)
	require.Equal(t, []teamsettings.SprintAutomationTeamRef{{WorkspaceID: workspaceID, TeamID: teamID}}, teams)
	require.Equal(t, teamsettingssql.ListSprintAutomationTeamsParams{
		HasCursor: true, AfterWorkspaceID: workspaceID, AfterTeamID: teamID, BatchSize: 50,
	}, queries.teamParams)

	teamCreatedBefore := time.Date(2026, time.May, 30, 10, 0, 0, 0, location)
	settingsUpdatedBefore := time.Date(2026, time.July, 29, 10, 0, 0, 0, location)
	_, err = repository.ListSprintAutomationInactivityCandidates(
		context.Background(),
		teamsettings.SprintAutomationInactivityQuery{
			TeamCreatedBefore: teamCreatedBefore, SettingsUpdatedBefore: settingsUpdatedBefore,
			Cursor: cursor, BatchSize: 25,
		},
	)
	require.NoError(t, err)
	require.Equal(t, teamsettingssql.ListSprintAutomationInactivityCandidatesParams{
		TeamCreatedBefore: teamCreatedBefore.UTC(), SettingsUpdatedBefore: settingsUpdatedBefore.UTC(),
		HasCursor: true, AfterWorkspaceID: workspaceID, AfterTeamID: teamID, BatchSize: 25,
	}, queries.inactivityParams)
}

func TestSprintAutomationRepositoryIsIdempotentAfterCommit(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	asOf := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)
	workspaceID := uuid.New()
	teamID := uuid.New()
	queries := &sprintAutomationQueries{
		lockedSettings: automatedSprintSettingsRow(workspaceID, teamID),
		upcomingCounts: []int32{0, 2},
		boundaryErrs:   []error{pgx.ErrNoRows},
		createdIDs:     []uuid.UUID{uuid.New(), uuid.New()},
		advanceRows:    1,
	}
	repository := sprintAutomationRepositoryWithTransaction(queries)
	ref := teamsettings.SprintAutomationTeamRef{WorkspaceID: workspaceID, TeamID: teamID}

	first, err := repository.RunSprintAutomationForTeam(context.Background(), ref, asOf)
	require.NoError(t, err)
	require.Equal(t, teamsettings.SprintAutomationRunResult{Created: 2}, first)
	second, err := repository.RunSprintAutomationForTeam(context.Background(), ref, asOf)
	require.NoError(t, err)
	require.Equal(t, teamsettings.SprintAutomationRunResult{}, second)

	require.Equal(t, 2, queries.lockAutomationCalls)
	require.Equal(t, 2, queries.lockSettingsCalls)
	require.Len(t, queries.createdParams, 2)
	require.Len(t, queries.auditParams, 2)
	require.Len(t, queries.advanceParams, 1)
	require.Equal(t, asOf.UTC(), queries.advanceParams[0].UpdatedAt)
	require.Equal(t, int32(1), queries.advanceParams[0].ExpectedNextNumber)
	require.Equal(t, int32(2), queries.advanceParams[0].CreatedCount)

	wantNames := []string{"Sprint 1", "Sprint 2"}
	for index, params := range queries.createdParams {
		require.Equal(t, asOf.UTC(), params.CreatedAt)
		require.Equal(t, asOf.UTC(), params.UpdatedAt)
		require.Equal(t, wantNames[index], params.Name)
	}
	require.Equal(t, "2026-08-31", queries.createdParams[0].StartDate.Format(time.DateOnly))
	require.Equal(t, "2026-09-13", queries.createdParams[0].EndDate.Format(time.DateOnly))
	require.Equal(t, "2026-09-14", queries.createdParams[1].StartDate.Format(time.DateOnly))
	require.Equal(t, "2026-09-27", queries.createdParams[1].EndDate.Format(time.DateOnly))
	for _, audit := range queries.auditParams {
		require.Equal(t, asOf.UTC(), audit.CreatedAt)
		require.Equal(t, "sprint.auto_created", audit.EventType)
	}
}

func TestSprintAutomationRepositoryReturnsNoRolledBackCounts(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("audit unavailable")
	workspaceID := uuid.New()
	teamID := uuid.New()
	queries := &sprintAutomationQueries{
		lockedSettings: automatedSprintSettingsRow(workspaceID, teamID),
		upcomingCounts: []int32{0},
		boundaryErrs:   []error{pgx.ErrNoRows},
		createdIDs:     []uuid.UUID{uuid.New()},
		auditErr:       wantErr,
	}
	repository := sprintAutomationRepositoryWithTransaction(queries)

	result, err := repository.RunSprintAutomationForTeam(
		context.Background(),
		teamsettings.SprintAutomationTeamRef{WorkspaceID: workspaceID, TeamID: teamID},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, teamsettings.SprintAutomationRunResult{}, result)
	require.Empty(t, queries.advanceParams)
}

func TestSprintAutomationRepositoryRejectsOversizedManagedSchedule(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	managed := make([]teamsettingssql.ListManagedFutureSprintsForUpdateRow, maxManagedSprintScheduleRows+1)
	queries := &sprintAutomationQueries{
		lockedSettings: automatedSprintSettingsRow(workspaceID, teamID),
		managedRows:    managed,
	}
	repository := sprintAutomationRepositoryWithTransaction(queries)

	_, err := repository.RunSprintAutomationForTeam(
		context.Background(),
		teamsettings.SprintAutomationTeamRef{WorkspaceID: workspaceID, TeamID: teamID},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.ErrorIs(t, err, teamsettings.ErrSprintScheduleTooLarge)
	require.Equal(t, int32(maxManagedSprintScheduleRows+1), queries.managedParams.RowLimit)
	require.Empty(t, queries.countParams)
}

func TestSprintAutomationRepositoryRechecksInactivityAndAuditsAtomically(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	disabledAt := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)
	teamCreatedBefore := disabledAt.AddDate(0, 0, -90)
	settingsUpdatedBefore := disabledAt.AddDate(0, 0, -30)
	workspaceID := uuid.New()
	teamID := uuid.New()
	lastActivity := disabledAt.AddDate(0, 0, -120)
	queries := &sprintAutomationQueries{
		lockedSettings: automatedSprintSettingsRow(workspaceID, teamID),
		inactivitySnapshot: teamsettingssql.GetSprintAutomationInactivitySnapshotRow{
			Name: "Platform", TeamCreatedAt: teamCreatedBefore.Add(-time.Hour),
			SettingsUpdatedAt:   settingsUpdatedBefore.Add(-time.Hour),
			HasLatestHumanStory: true, LatestHumanStoryAt: lastActivity,
		},
		disableRows: 1,
	}
	repository := sprintAutomationRepositoryWithTransaction(queries)
	eligibility := teamsettings.SprintAutomationInactivityEligibility{
		WorkspaceID: workspaceID, TeamID: teamID,
		TeamCreatedBefore: teamCreatedBefore, SettingsUpdatedBefore: settingsUpdatedBefore,
		ActivityBefore: teamCreatedBefore, DisabledAt: disabledAt,
		Reason:         "  No human sprint planning activity in the last 90 days  ",
		InactivityDays: 90, GraceDays: 30,
	}

	disabled, err := repository.DisableSprintAutomationIfInactive(context.Background(), eligibility)

	require.NoError(t, err)
	require.True(t, disabled)
	require.Len(t, queries.disableParams, 1)
	require.Equal(t, disabledAt.UTC(), queries.disableParams[0].DisabledAt)
	require.Equal(t, "No human sprint planning activity in the last 90 days", *queries.disableParams[0].DisabledReason)
	require.Len(t, queries.auditParams, 1)
	require.Equal(t, "sprint_automation.disabled", queries.auditParams[0].EventType)
	var metadata disabledSprintAutomationAuditMetadata
	require.NoError(t, json.Unmarshal(queries.auditParams[0].Metadata, &metadata))
	require.Equal(t, "Platform", metadata.TeamName)
	require.NotNil(t, metadata.LastActivityAt)
	require.Equal(t, lastActivity.UTC(), metadata.LastActivityAt.UTC())
}

func TestSprintAutomationRepositoryKeepsRecentlyActiveTeamEnabled(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	workspaceID := uuid.New()
	teamID := uuid.New()
	activityBefore := asOf.AddDate(0, 0, -90)
	queries := &sprintAutomationQueries{
		lockedSettings: automatedSprintSettingsRow(workspaceID, teamID),
		inactivitySnapshot: teamsettingssql.GetSprintAutomationInactivitySnapshotRow{
			Name: "Platform", TeamCreatedAt: activityBefore.Add(-time.Hour),
			SettingsUpdatedAt:          asOf.AddDate(0, 0, -31),
			HasLatestHumanSprintChange: true, LatestHumanSprintChangeAt: activityBefore,
		},
	}
	repository := sprintAutomationRepositoryWithTransaction(queries)

	disabled, err := repository.DisableSprintAutomationIfInactive(
		context.Background(),
		teamsettings.SprintAutomationInactivityEligibility{
			WorkspaceID: workspaceID, TeamID: teamID,
			TeamCreatedBefore: activityBefore, SettingsUpdatedBefore: asOf.AddDate(0, 0, -30),
			ActivityBefore: activityBefore, DisabledAt: asOf,
			Reason: "inactive", InactivityDays: 90, GraceDays: 30,
		},
	)

	require.NoError(t, err)
	require.False(t, disabled)
	require.Empty(t, queries.disableParams)
	require.Empty(t, queries.auditParams)
}

type sprintAutomationQueries struct {
	teamsettingssql.Querier
	teamParams       teamsettingssql.ListSprintAutomationTeamsParams
	teamRows         []teamsettingssql.ListSprintAutomationTeamsRow
	inactivityParams teamsettingssql.ListSprintAutomationInactivityCandidatesParams
	inactivityRows   []teamsettingssql.ListSprintAutomationInactivityCandidatesRow

	lockAutomationCalls int
	lockSettingsCalls   int
	lockedSettings      teamsettingssql.LockSprintSettingsRow
	lockSettingsErr     error
	managedParams       teamsettingssql.ListManagedFutureSprintsForUpdateParams
	managedRows         []teamsettingssql.ListManagedFutureSprintsForUpdateRow
	countParams         []teamsettingssql.CountUpcomingSprintsForAutomationParams
	upcomingCounts      []int32
	boundaryErrs        []error
	createdIDs          []uuid.UUID
	createdParams       []teamsettingssql.CreateAutomatedSprintParams
	auditParams         []teamsettingssql.InsertSprintAutomationAuditEventParams
	auditErr            error
	advanceParams       []teamsettingssql.AdvanceSprintAutomationCounterParams
	advanceRows         int64

	inactivitySnapshot teamsettingssql.GetSprintAutomationInactivitySnapshotRow
	inactivityErr      error
	disableParams      []teamsettingssql.DisableSprintAutomationIfInactiveParams
	disableRows        int64
}

func (queries *sprintAutomationQueries) ListSprintAutomationTeams(
	_ context.Context,
	params teamsettingssql.ListSprintAutomationTeamsParams,
) ([]teamsettingssql.ListSprintAutomationTeamsRow, error) {
	queries.teamParams = params
	return queries.teamRows, nil
}

func (queries *sprintAutomationQueries) ListSprintAutomationInactivityCandidates(
	_ context.Context,
	params teamsettingssql.ListSprintAutomationInactivityCandidatesParams,
) ([]teamsettingssql.ListSprintAutomationInactivityCandidatesRow, error) {
	queries.inactivityParams = params
	return queries.inactivityRows, nil
}

func (queries *sprintAutomationQueries) LockSprintAutomation(
	context.Context,
	teamsettingssql.LockSprintAutomationParams,
) error {
	queries.lockAutomationCalls++
	return nil
}

func (queries *sprintAutomationQueries) LockSprintSettings(
	_ context.Context,
	_ teamsettingssql.LockSprintSettingsParams,
) (teamsettingssql.LockSprintSettingsRow, error) {
	queries.lockSettingsCalls++
	return queries.lockedSettings, queries.lockSettingsErr
}

func (queries *sprintAutomationQueries) ListManagedFutureSprintsForUpdate(
	_ context.Context,
	params teamsettingssql.ListManagedFutureSprintsForUpdateParams,
) ([]teamsettingssql.ListManagedFutureSprintsForUpdateRow, error) {
	queries.managedParams = params
	return queries.managedRows, nil
}

func (queries *sprintAutomationQueries) CountUpcomingSprintsForAutomation(
	_ context.Context,
	params teamsettingssql.CountUpcomingSprintsForAutomationParams,
) (int32, error) {
	queries.countParams = append(queries.countParams, params)
	if len(queries.upcomingCounts) == 0 {
		return 0, nil
	}
	count := queries.upcomingCounts[0]
	queries.upcomingCounts = queries.upcomingCounts[1:]
	return count, nil
}

func (queries *sprintAutomationQueries) GetSprintAutomationScheduleBoundary(
	context.Context,
	teamsettingssql.GetSprintAutomationScheduleBoundaryParams,
) (time.Time, error) {
	if len(queries.boundaryErrs) == 0 {
		return time.Time{}, pgx.ErrNoRows
	}
	err := queries.boundaryErrs[0]
	queries.boundaryErrs = queries.boundaryErrs[1:]
	return time.Time{}, err
}

func (queries *sprintAutomationQueries) CreateAutomatedSprint(
	_ context.Context,
	params teamsettingssql.CreateAutomatedSprintParams,
) (uuid.UUID, error) {
	queries.createdParams = append(queries.createdParams, params)
	if len(queries.createdIDs) == 0 {
		return uuid.New(), nil
	}
	id := queries.createdIDs[0]
	queries.createdIDs = queries.createdIDs[1:]
	return id, nil
}

func (queries *sprintAutomationQueries) InsertSprintAutomationAuditEvent(
	_ context.Context,
	params teamsettingssql.InsertSprintAutomationAuditEventParams,
) error {
	queries.auditParams = append(queries.auditParams, params)
	return queries.auditErr
}

func (queries *sprintAutomationQueries) AdvanceSprintAutomationCounter(
	_ context.Context,
	params teamsettingssql.AdvanceSprintAutomationCounterParams,
) (int64, error) {
	queries.advanceParams = append(queries.advanceParams, params)
	return queries.advanceRows, nil
}

func (queries *sprintAutomationQueries) GetSprintAutomationInactivitySnapshot(
	context.Context,
	teamsettingssql.GetSprintAutomationInactivitySnapshotParams,
) (teamsettingssql.GetSprintAutomationInactivitySnapshotRow, error) {
	return queries.inactivitySnapshot, queries.inactivityErr
}

func (queries *sprintAutomationQueries) DisableSprintAutomationIfInactive(
	_ context.Context,
	params teamsettingssql.DisableSprintAutomationIfInactiveParams,
) (int64, error) {
	queries.disableParams = append(queries.disableParams, params)
	return queries.disableRows, nil
}

func sprintAutomationRepositoryWithTransaction(queries *sprintAutomationQueries) *repo {
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(teamsettingssql.Querier) error) error {
		return operation(queries)
	}
	return repository
}

func automatedSprintSettingsRow(workspaceID, teamID uuid.UUID) teamsettingssql.LockSprintSettingsRow {
	return teamsettingssql.LockSprintSettingsRow{
		TeamID: teamID, WorkspaceID: workspaceID, AutoCreateSprints: true,
		UpcomingSprintsCount: 2, SprintDurationWeeks: 2, SprintStartDay: "Monday",
		NextAutoSprintNumber: 1,
	}
}

var _ teamsettingssql.Querier = (*sprintAutomationQueries)(nil)
