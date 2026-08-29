package storiesrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStoryAutoArchiveTransactionLocksBeforeBoundedTransition(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	queries := &storyAutomationQueryStub{archivedIDs: []uuid.UUID{uuid.New(), uuid.New()}}
	repository := storyAutomationTestRepository(queries, nil)

	result, err := repository.ArchiveEligibleStoriesBatch(
		context.Background(),
		storydomain.StoryAutoArchiveBatch{AsOf: asOf, BatchSize: 25},
	)

	require.NoError(t, err)
	require.Equal(t, storydomain.StoryAutoArchiveResult{Archived: 2}, result)
	require.Equal(t, []string{"lock", "archive"}, queries.calls)
	require.Equal(t, storyAutoArchiveLockName, queries.lockParams.LockName)
	require.Equal(t, asOf.UTC(), queries.archiveParams.AsOf)
	require.Equal(t, int32(25), queries.archiveParams.BatchSize)
}

func TestStoryAutoCloseTransactionWritesExactScopedActivityBatch(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	actorID := uuid.New()
	row := storyreadsql.CloseEligibleStoriesBatchRow{
		ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(), StatusID: uuid.New(),
	}
	queries := &storyAutomationQueryStub{closedRows: []storyreadsql.CloseEligibleStoriesBatchRow{row}, closeActivityRows: 1}
	repository := storyAutomationTestRepository(queries, nil)

	result, err := repository.CloseEligibleStoriesBatch(
		context.Background(),
		storydomain.StoryAutoCloseBatch{AsOf: asOf, SystemUserID: actorID, BatchSize: 50},
	)

	require.NoError(t, err)
	require.Equal(t, storydomain.StoryAutoCloseResult{Closed: 1, ActivitiesRecorded: 1}, result)
	require.Equal(t, []string{"lock", "close", "close_activities"}, queries.calls)
	require.Equal(t, storyAutoCloseLockName, queries.lockParams.LockName)
	require.Equal(t, storyreadsql.CloseEligibleStoriesBatchParams{AsOf: asOf, BatchSize: 50}, queries.closeParams)
	require.Equal(t, actorID, queries.closeActivityParams.SystemUserID)
	require.Equal(t, asOf, queries.closeActivityParams.AsOf)
	require.Equal(t, []uuid.UUID{row.ID}, queries.closeActivityParams.StoryIds)
	require.Equal(t, []uuid.UUID{row.WorkspaceID}, queries.closeActivityParams.WorkspaceIds)
	require.Equal(t, []uuid.UUID{row.TeamID}, queries.closeActivityParams.TeamIds)
	require.Equal(t, []uuid.UUID{row.StatusID}, queries.closeActivityParams.StatusIds)
	require.NotNil(t, queries.closeActivityParams.Reason)
	require.Equal(t, storyAutoCloseActivityReason, *queries.closeActivityParams.Reason)
}

func TestStoryAutoCloseSideEffectFailureRollsBackTransition(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("activity insert unavailable")
	queries := &storyAutomationQueryStub{
		closedRows: []storyreadsql.CloseEligibleStoriesBatchRow{{
			ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(), StatusID: uuid.New(),
		}},
		closeActivityErr: wantErr,
	}
	rolledBack := false
	repository := storyAutomationTestRepository(queries, &rolledBack)

	result, err := repository.CloseEligibleStoriesBatch(context.Background(), storydomain.StoryAutoCloseBatch{
		AsOf: time.Now().UTC(), SystemUserID: uuid.New(), BatchSize: 10,
	})

	require.ErrorIs(t, err, wantErr)
	require.True(t, rolledBack)
	require.Zero(t, result, "an aborted transaction must not expose transitioned counts")
	require.Equal(t, []string{"lock", "close", "close_activities"}, queries.calls)
}

func TestStoryAutoCloseSideEffectCountMismatchRollsBackTransition(t *testing.T) {
	t.Parallel()

	queries := &storyAutomationQueryStub{
		closedRows: []storyreadsql.CloseEligibleStoriesBatchRow{{
			ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(), StatusID: uuid.New(),
		}},
		closeActivityRows: 0,
	}
	rolledBack := false
	repository := storyAutomationTestRepository(queries, &rolledBack)

	result, err := repository.CloseEligibleStoriesBatch(context.Background(), storydomain.StoryAutoCloseBatch{
		AsOf: time.Now().UTC(), SystemUserID: uuid.New(), BatchSize: 10,
	})

	require.ErrorContains(t, err, "inserted 0 rows, want 1")
	require.True(t, rolledBack)
	require.Zero(t, result)
}

func TestSprintStoryMigrationTransactionOrdersTransitionActivityAndAudit(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	actorID := uuid.New()
	row := storyreadsql.MigrateEligibleSprintStoriesBatchRow{
		ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(),
		PreviousSprintID: uuid.New(), NewSprintID: uuid.New(),
	}
	queries := &storyAutomationQueryStub{
		migratedRows:          []storyreadsql.MigrateEligibleSprintStoriesBatchRow{row},
		migrationActivityRows: 1,
		migrationAuditRows:    1,
	}
	repository := storyAutomationTestRepository(queries, nil)

	result, err := repository.MigrateEligibleSprintStoriesBatch(
		context.Background(),
		storydomain.SprintStoryMigrationBatch{AsOf: asOf, SystemUserID: actorID, BatchSize: 40},
	)

	require.NoError(t, err)
	require.Equal(t, storydomain.SprintStoryMigrationResult{
		Migrated: 1, ActivitiesRecorded: 1, AuditEventsRecorded: 1,
	}, result)
	require.Equal(t, []string{"lock", "migrate", "migration_activities", "migration_audits"}, queries.calls)
	require.Equal(t, storySprintMigrationLockName, queries.lockParams.LockName)
	require.Equal(t, storyreadsql.MigrateEligibleSprintStoriesBatchParams{AsOf: asOf, BatchSize: 40}, queries.migrationParams)
	require.Equal(t, actorID, queries.migrationActivityParams.SystemUserID)
	require.Equal(t, []uuid.UUID{row.ID}, queries.migrationActivityParams.StoryIds)
	require.Equal(t, []uuid.UUID{row.WorkspaceID}, queries.migrationActivityParams.WorkspaceIds)
	require.Equal(t, []uuid.UUID{row.TeamID}, queries.migrationActivityParams.TeamIds)
	require.Equal(t, []uuid.UUID{row.PreviousSprintID}, queries.migrationActivityParams.PreviousSprintIds)
	require.Equal(t, []uuid.UUID{row.NewSprintID}, queries.migrationActivityParams.NewSprintIds)
	require.NotNil(t, queries.migrationActivityParams.Reason)
	require.Equal(t, sprintMigrationActivityReason, *queries.migrationActivityParams.Reason)
	require.NotNil(t, queries.migrationAuditParams.SystemUserID)
	require.Equal(t, actorID, *queries.migrationAuditParams.SystemUserID)
	require.Equal(t, queries.migrationActivityParams.StoryIds, queries.migrationAuditParams.StoryIds)
	require.Equal(t, queries.migrationActivityParams.WorkspaceIds, queries.migrationAuditParams.WorkspaceIds)
	require.Equal(t, queries.migrationActivityParams.TeamIds, queries.migrationAuditParams.TeamIds)
	require.Equal(t, queries.migrationActivityParams.PreviousSprintIds, queries.migrationAuditParams.PreviousSprintIds)
	require.Equal(t, queries.migrationActivityParams.NewSprintIds, queries.migrationAuditParams.NewSprintIds)
}

func TestSprintStoryMigrationAuditFailureRollsBackTransitionAndActivity(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("audit insert unavailable")
	queries := &storyAutomationQueryStub{
		migratedRows: []storyreadsql.MigrateEligibleSprintStoriesBatchRow{{
			ID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New(),
			PreviousSprintID: uuid.New(), NewSprintID: uuid.New(),
		}},
		migrationActivityRows: 1,
		migrationAuditErr:     wantErr,
	}
	rolledBack := false
	repository := storyAutomationTestRepository(queries, &rolledBack)

	result, err := repository.MigrateEligibleSprintStoriesBatch(
		context.Background(),
		storydomain.SprintStoryMigrationBatch{
			AsOf: time.Now().UTC(), SystemUserID: uuid.New(), BatchSize: 10,
		},
	)

	require.ErrorIs(t, err, wantErr)
	require.True(t, rolledBack)
	require.Zero(t, result)
	require.Equal(t, []string{"lock", "migrate", "migration_activities", "migration_audits"}, queries.calls)
}

func TestStoryAutomationRepositoryRejectsInvalidBoundaries(t *testing.T) {
	t.Parallel()

	configured := storyAutomationTestRepository(&storyAutomationQueryStub{}, nil)
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "missing repository",
			run: func() error {
				_, err := (&repo{}).ArchiveEligibleStoriesBatch(context.Background(), storydomain.StoryAutoArchiveBatch{
					AsOf: time.Now().UTC(), BatchSize: 1,
				})
				return err
			},
		},
		{
			name: "missing clock",
			run: func() error {
				_, err := configured.ArchiveEligibleStoriesBatch(context.Background(), storydomain.StoryAutoArchiveBatch{BatchSize: 1})
				return err
			},
		},
		{
			name: "oversized batch",
			run: func() error {
				_, err := configured.ArchiveEligibleStoriesBatch(context.Background(), storydomain.StoryAutoArchiveBatch{
					AsOf: time.Now().UTC(), BatchSize: maximumStoryAutomationBatchSize + 1,
				})
				return err
			},
		},
		{
			name: "missing activity actor",
			run: func() error {
				_, err := configured.CloseEligibleStoriesBatch(context.Background(), storydomain.StoryAutoCloseBatch{
					AsOf: time.Now().UTC(), BatchSize: 1,
				})
				return err
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, test.run())
		})
	}
}

func storyAutomationTestRepository(queries storyAutomationQueries, rolledBack *bool) *repo {
	repository := &repo{}
	repository.runStoryAutomationTransaction = func(
		_ context.Context,
		operation func(storyAutomationQueries) error,
	) error {
		err := operation(queries)
		if rolledBack != nil {
			*rolledBack = err != nil
		}
		return err
	}
	return repository
}

type storyAutomationQueryStub struct {
	lockParams storyreadsql.LockStoryAutomationParams
	lockErr    error

	archiveParams storyreadsql.ArchiveEligibleStoriesBatchParams
	archivedIDs   []uuid.UUID
	archiveErr    error

	closeParams         storyreadsql.CloseEligibleStoriesBatchParams
	closedRows          []storyreadsql.CloseEligibleStoriesBatchRow
	closeErr            error
	closeActivityParams storyreadsql.InsertStoryAutoCloseActivitiesParams
	closeActivityRows   int64
	closeActivityErr    error

	migrationParams         storyreadsql.MigrateEligibleSprintStoriesBatchParams
	migratedRows            []storyreadsql.MigrateEligibleSprintStoriesBatchRow
	migrationErr            error
	migrationActivityParams storyreadsql.InsertSprintMigrationActivitiesParams
	migrationActivityRows   int64
	migrationActivityErr    error
	migrationAuditParams    storyreadsql.InsertSprintMigrationAuditEventsParams
	migrationAuditRows      int64
	migrationAuditErr       error

	calls []string
}

func (queries *storyAutomationQueryStub) LockStoryAutomation(
	_ context.Context,
	params storyreadsql.LockStoryAutomationParams,
) error {
	queries.calls = append(queries.calls, "lock")
	queries.lockParams = params
	return queries.lockErr
}

func (queries *storyAutomationQueryStub) ArchiveEligibleStoriesBatch(
	_ context.Context,
	params storyreadsql.ArchiveEligibleStoriesBatchParams,
) ([]uuid.UUID, error) {
	queries.calls = append(queries.calls, "archive")
	queries.archiveParams = params
	return queries.archivedIDs, queries.archiveErr
}

func (queries *storyAutomationQueryStub) CloseEligibleStoriesBatch(
	_ context.Context,
	params storyreadsql.CloseEligibleStoriesBatchParams,
) ([]storyreadsql.CloseEligibleStoriesBatchRow, error) {
	queries.calls = append(queries.calls, "close")
	queries.closeParams = params
	return queries.closedRows, queries.closeErr
}

func (queries *storyAutomationQueryStub) InsertStoryAutoCloseActivities(
	_ context.Context,
	params storyreadsql.InsertStoryAutoCloseActivitiesParams,
) (int64, error) {
	queries.calls = append(queries.calls, "close_activities")
	queries.closeActivityParams = params
	return queries.closeActivityRows, queries.closeActivityErr
}

func (queries *storyAutomationQueryStub) MigrateEligibleSprintStoriesBatch(
	_ context.Context,
	params storyreadsql.MigrateEligibleSprintStoriesBatchParams,
) ([]storyreadsql.MigrateEligibleSprintStoriesBatchRow, error) {
	queries.calls = append(queries.calls, "migrate")
	queries.migrationParams = params
	return queries.migratedRows, queries.migrationErr
}

func (queries *storyAutomationQueryStub) InsertSprintMigrationActivities(
	_ context.Context,
	params storyreadsql.InsertSprintMigrationActivitiesParams,
) (int64, error) {
	queries.calls = append(queries.calls, "migration_activities")
	queries.migrationActivityParams = params
	return queries.migrationActivityRows, queries.migrationActivityErr
}

func (queries *storyAutomationQueryStub) InsertSprintMigrationAuditEvents(
	_ context.Context,
	params storyreadsql.InsertSprintMigrationAuditEventsParams,
) (int64, error) {
	queries.calls = append(queries.calls, "migration_audits")
	queries.migrationAuditParams = params
	return queries.migrationAuditRows, queries.migrationAuditErr
}
