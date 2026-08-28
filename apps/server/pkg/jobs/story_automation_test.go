package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStoryAutoArchiveUsesOneUTCSnapshotAcrossBoundedBatches(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	store := &storyAutomationStoreStub{archiveResults: []storydomain.StoryAutoArchiveResult{
		{Archived: storyAutoArchiveBatchSize},
		{Archived: 3},
	}}

	err := processStoryAutoArchiveAt(context.Background(), store, newTestJobLogger(), asOf)

	require.NoError(t, err)
	require.Len(t, store.archiveBatches, 2)
	for _, batch := range store.archiveBatches {
		require.Equal(t, asOf.UTC(), batch.AsOf)
		require.Equal(t, storyAutoArchiveBatchSize, batch.BatchSize)
	}
}

func TestStoryAutoCloseUsesOneActorAndRejectsIncompleteActivities(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	actorID := uuid.New()
	store := &storyAutomationStoreStub{closeResults: []storydomain.StoryAutoCloseResult{
		{Closed: storyAutoCloseBatchSize, ActivitiesRecorded: storyAutoCloseBatchSize},
		{Closed: 2, ActivitiesRecorded: 1},
	}}

	err := processStoryAutoCloseAt(context.Background(), store, newTestJobLogger(), actorID, asOf)

	require.ErrorContains(t, err, "incomplete side effects")
	require.Len(t, store.closeBatches, 2)
	for _, batch := range store.closeBatches {
		require.Equal(t, asOf, batch.AsOf)
		require.Equal(t, actorID, batch.SystemUserID)
		require.Equal(t, storyAutoCloseBatchSize, batch.BatchSize)
	}
}

func TestSprintStoryMigrationPreservesStoreErrorAndExplicitClock(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("migration transaction unavailable")
	asOf := time.Date(2026, time.August, 28, 11, 0, 0, 0, time.FixedZone("CAT", 2*60*60))
	actorID := uuid.New()
	store := &storyAutomationStoreStub{migrationErr: wantErr}

	err := processSprintStoryMigrationAt(context.Background(), store, newTestJobLogger(), actorID, asOf)

	require.ErrorIs(t, err, wantErr)
	require.Len(t, store.migrationBatches, 1)
	require.Equal(t, storydomain.SprintStoryMigrationBatch{
		AsOf: asOf.UTC(), SystemUserID: actorID, BatchSize: sprintStoryMigrationBatchSize,
	}, store.migrationBatches[0])
}

func TestStoryAutomationHonorsCancellationBeforePersistence(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &storyAutomationStoreStub{}

	err := processStoryAutoArchiveAt(ctx, store, newTestJobLogger(), time.Now().UTC())

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, store.archiveBatches)
}

func TestStoryAutomationReportsBoundedBacklog(t *testing.T) {
	t.Parallel()

	results := make([]storydomain.StoryAutoArchiveResult, storyAutomationMaxBatches)
	for index := range results {
		results[index] = storydomain.StoryAutoArchiveResult{Archived: storyAutoArchiveBatchSize}
	}
	store := &storyAutomationStoreStub{archiveResults: results}

	err := processStoryAutoArchiveAt(
		context.Background(), store, newTestJobLogger(), time.Now().UTC(),
	)

	require.ErrorIs(t, err, errStoryAutomationBacklogRemaining)
	require.Len(t, store.archiveBatches, storyAutomationMaxBatches)
}

func TestStoryAutomationRejectsMissingDependenciesAndActor(t *testing.T) {
	t.Parallel()

	asOf := time.Now().UTC()
	store := &storyAutomationStoreStub{}
	log := newTestJobLogger()
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "archive store",
			run: func() error {
				return processStoryAutoArchiveAt(context.Background(), nil, log, asOf)
			},
		},
		{
			name: "archive logger",
			run: func() error {
				return processStoryAutoArchiveAt(context.Background(), store, nil, asOf)
			},
		},
		{
			name: "archive clock",
			run: func() error {
				return processStoryAutoArchiveAt(context.Background(), store, log, time.Time{})
			},
		},
		{
			name: "auto-close actor",
			run: func() error {
				return processStoryAutoCloseAt(context.Background(), store, log, uuid.Nil, asOf)
			},
		},
		{
			name: "migration actor",
			run: func() error {
				return processSprintStoryMigrationAt(context.Background(), store, log, uuid.Nil, asOf)
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

type storyAutomationStoreStub struct {
	archiveResults []storydomain.StoryAutoArchiveResult
	archiveErr     error
	archiveBatches []storydomain.StoryAutoArchiveBatch

	closeResults []storydomain.StoryAutoCloseResult
	closeErr     error
	closeBatches []storydomain.StoryAutoCloseBatch

	migrationResults []storydomain.SprintStoryMigrationResult
	migrationErr     error
	migrationBatches []storydomain.SprintStoryMigrationBatch
}

func (store *storyAutomationStoreStub) ArchiveEligibleStoriesBatch(
	_ context.Context,
	batch storydomain.StoryAutoArchiveBatch,
) (storydomain.StoryAutoArchiveResult, error) {
	store.archiveBatches = append(store.archiveBatches, batch)
	if store.archiveErr != nil {
		return storydomain.StoryAutoArchiveResult{}, store.archiveErr
	}
	index := len(store.archiveBatches) - 1
	if index >= len(store.archiveResults) {
		return storydomain.StoryAutoArchiveResult{}, nil
	}
	return store.archiveResults[index], nil
}

func (store *storyAutomationStoreStub) CloseEligibleStoriesBatch(
	_ context.Context,
	batch storydomain.StoryAutoCloseBatch,
) (storydomain.StoryAutoCloseResult, error) {
	store.closeBatches = append(store.closeBatches, batch)
	if store.closeErr != nil {
		return storydomain.StoryAutoCloseResult{}, store.closeErr
	}
	index := len(store.closeBatches) - 1
	if index >= len(store.closeResults) {
		return storydomain.StoryAutoCloseResult{}, nil
	}
	return store.closeResults[index], nil
}

func (store *storyAutomationStoreStub) MigrateEligibleSprintStoriesBatch(
	_ context.Context,
	batch storydomain.SprintStoryMigrationBatch,
) (storydomain.SprintStoryMigrationResult, error) {
	store.migrationBatches = append(store.migrationBatches, batch)
	if store.migrationErr != nil {
		return storydomain.SprintStoryMigrationResult{}, store.migrationErr
	}
	index := len(store.migrationBatches) - 1
	if index >= len(store.migrationResults) {
		return storydomain.SprintStoryMigrationResult{}, nil
	}
	return store.migrationResults[index], nil
}
