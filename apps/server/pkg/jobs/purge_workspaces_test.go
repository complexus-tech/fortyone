package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type softDeletedWorkspacePurgerStub struct {
	results []workspacedomain.DeletedWorkspacePurgeResult
	batches []workspacedomain.DeletedWorkspacePurgeBatch
	err     error
}

func (store *softDeletedWorkspacePurgerStub) PurgeSoftDeletedWorkspacesBatch(
	_ context.Context,
	batch workspacedomain.DeletedWorkspacePurgeBatch,
) (workspacedomain.DeletedWorkspacePurgeResult, error) {
	store.batches = append(store.batches, batch)
	if store.err != nil {
		return workspacedomain.DeletedWorkspacePurgeResult{}, store.err
	}
	if len(store.results) == 0 {
		return workspacedomain.DeletedWorkspacePurgeResult{}, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}

func TestPurgeDeletedWorkspacesUsesOneUTCCutoffAndStableCursor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	firstCursor := workspacedomain.DeletedWorkspacePurgeCursor{
		DeletedAt:   time.Date(2026, time.August, 24, 8, 0, 0, 0, time.UTC),
		WorkspaceID: uuid.New(), Valid: true,
	}
	secondCursor := workspacedomain.DeletedWorkspacePurgeCursor{
		DeletedAt: firstCursor.DeletedAt.Add(time.Hour), WorkspaceID: uuid.New(), Valid: true,
	}
	store := &softDeletedWorkspacePurgerStub{results: []workspacedomain.DeletedWorkspacePurgeResult{
		{CandidateCount: maintenancePurgeBatchSize, Deleted: maintenancePurgeBatchSize - 1, Blocked: 1, Cursor: firstCursor},
		{CandidateCount: 2, Deleted: 2, Cursor: secondCursor},
	}}

	err := purgeDeletedWorkspacesAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Len(t, store.batches, 2)
	for _, batch := range store.batches {
		require.Equal(t, now.UTC().Add(-deletedWorkspaceRetention), batch.DeletedBefore)
		require.Equal(t, now.UTC(), batch.ProcessedAt)
		require.Equal(t, maintenancePurgeBatchSize, batch.BatchSize)
		require.Equal(t, slackrepository.SlackInstallationLifecycleAdvisoryKey, batch.IntegrationLifecycleLockKey)
	}
	require.False(t, store.batches[0].Cursor.Valid)
	require.Equal(t, firstCursor, store.batches[1].Cursor)
}

func TestPurgeDeletedWorkspacesRejectsInvalidResult(t *testing.T) {
	t.Parallel()

	store := &softDeletedWorkspacePurgerStub{results: []workspacedomain.DeletedWorkspacePurgeResult{{
		CandidateCount: 2,
		Deleted:        1,
		Cursor: workspacedomain.DeletedWorkspacePurgeCursor{
			DeletedAt: time.Now().UTC(), WorkspaceID: uuid.New(), Valid: true,
		},
	}}}

	err := purgeDeletedWorkspacesAt(context.Background(), store, newTestJobLogger(), time.Now().UTC())

	require.ErrorContains(t, err, "invalid result")
}

func TestPurgeDeletedWorkspacesHonorsCancellationBeforeDatabaseWork(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &softDeletedWorkspacePurgerStub{}

	err := purgeDeletedWorkspacesAt(ctx, store, newTestJobLogger(), time.Now().UTC())

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, store.batches)
}

func TestPurgeDeletedWorkspacesReportsBoundedBacklog(t *testing.T) {
	t.Parallel()

	results := make([]workspacedomain.DeletedWorkspacePurgeResult, maintenancePurgeMaxBatches)
	deletedAt := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	for index := range results {
		results[index] = workspacedomain.DeletedWorkspacePurgeResult{
			CandidateCount: maintenancePurgeBatchSize,
			Deleted:        maintenancePurgeBatchSize,
			Cursor: workspacedomain.DeletedWorkspacePurgeCursor{
				DeletedAt:   deletedAt.Add(time.Duration(index) * time.Second),
				WorkspaceID: uuid.New(), Valid: true,
			},
		}
	}
	store := &softDeletedWorkspacePurgerStub{results: results}

	err := purgeDeletedWorkspacesAt(context.Background(), store, newTestJobLogger(), time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC))

	require.ErrorIs(t, err, errMaintenanceBacklogRemaining)
	require.Len(t, store.batches, maintenancePurgeMaxBatches)
}

func TestPurgeDeletedWorkspacesValidatesDependenciesAndPropagatesStoreError(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	log := newTestJobLogger()
	store := &softDeletedWorkspacePurgerStub{}
	var nilContext context.Context
	require.EqualError(t, purgeDeletedWorkspacesAt(nilContext, store, log, now), "soft-deleted workspace purge context is required")
	require.EqualError(t, purgeDeletedWorkspacesAt(context.Background(), nil, log, now), "soft-deleted workspace purge store is required")
	require.EqualError(t, purgeDeletedWorkspacesAt(context.Background(), store, nil, now), "soft-deleted workspace purge logger is required")
	require.EqualError(t, purgeDeletedWorkspacesAt(context.Background(), store, log, time.Time{}), "soft-deleted workspace purge clock is required")

	wantErr := errors.New("database unavailable")
	err := purgeDeletedWorkspacesAt(context.Background(), &softDeletedWorkspacePurgerStub{err: wantErr}, log, now)
	require.ErrorIs(t, err, wantErr)
}
