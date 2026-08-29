package jobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPurgeDeletedStoriesUsesOneUTCCutoffStableCursorAndBoundedInputs(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	firstCursor := storydomain.StoryRetentionCursor{
		DeletedAt: time.Date(2026, time.July, 1, 7, 0, 0, 0, time.UTC),
		StoryID:   uuid.New(),
		Valid:     true,
	}
	secondCursor := storydomain.StoryRetentionCursor{
		DeletedAt: firstCursor.DeletedAt.Add(time.Minute),
		StoryID:   uuid.New(),
		Valid:     true,
	}
	store := &deletedStoryRetentionStoreStub{purgeResults: []storydomain.StoryRetentionResult{
		{
			CandidateCount:      deletedStoryRetentionBatchSize,
			DeletedStoryCount:   deletedStoryRetentionBatchSize,
			EnqueuedObjectCount: 3,
			NextCursor:          firstCursor,
		},
		{
			CandidateCount:      2,
			DeletedStoryCount:   2,
			EnqueuedObjectCount: 1,
			NextCursor:          secondCursor,
		},
	}}
	objects := &retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"}

	err := purgeDeletedStoriesAt(context.Background(), store, objects, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Len(t, store.purgeBatches, 2)
	for _, batch := range store.purgeBatches {
		require.Equal(t, now.UTC().Add(-deletedStoryRetention), batch.DeletedBefore)
		require.Equal(t, now.UTC(), batch.EnqueuedAt)
		require.Equal(t, deletedStoryRetentionBatchSize, batch.BatchSize)
		require.Equal(t, deletedStoryAttachmentMaximum, batch.MaximumAttachmentCount)
		require.Equal(t, "aws", batch.StorageProvider)
		require.Equal(t, "attachments", batch.ContainerName)
	}
	require.False(t, store.purgeBatches[0].Cursor.Valid)
	require.Equal(t, firstCursor, store.purgeBatches[1].Cursor)
	require.Empty(t, objects.deletedBlobNames, "daily story purge must only enqueue object work")
}

func TestProcessAttachmentObjectDeletionsCompletesFencedClaimsAndPurgesRetention(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 8, 15, 0, 0, time.UTC)
	outboxID, attachmentID := uuid.New(), uuid.New()
	store := &deletedStoryRetentionStoreStub{
		claimResults: [][]storydomain.AttachmentObjectDeletion{{{
			OutboxID: outboxID, AttachmentID: attachmentID, WorkspaceID: uuid.New(),
			StorageProvider: "aws", ContainerName: "attachments", BlobName: "private/object.png",
			AttemptCount: 1,
		}}},
		completeResults:       []bool{true},
		completedPurgeResults: []int64{4},
	}
	objects := &retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"}

	err := processAttachmentObjectDeletionsAt(context.Background(), store, objects, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Len(t, store.claimBatches, 1)
	claim := store.claimBatches[0]
	require.Equal(t, now, claim.AsOf)
	require.Equal(t, now.Add(-attachmentObjectDeletionLease), claim.LeaseExpiredBefore)
	require.NotEqual(t, uuid.Nil, claim.ClaimToken)
	require.Equal(t, attachmentObjectDeletionBatchSize, claim.BatchSize)
	require.Equal(t, []string{"private/object.png"}, objects.deletedBlobNames)
	require.Equal(t, []storydomain.AttachmentObjectDeletionCompletion{{
		OutboxID: outboxID, ClaimToken: claim.ClaimToken, CompletedAt: now,
	}}, store.completions)
	require.Equal(t, []completedObjectDeletionPurgeCall{{
		completedBefore: now.Add(-completedObjectDeletionRetention),
		batchSize:       completedObjectDeletionPurgeBatchSize,
	}}, store.completedPurgeCalls)
}

func TestProcessAttachmentObjectDeletionsPersistsSafeRetryWithoutLeakingObjectName(t *testing.T) {
	t.Parallel()

	const blobName = "private/customer-sensitive-object.png"
	now := time.Date(2026, time.August, 28, 8, 15, 0, 0, time.UTC)
	outboxID := uuid.New()
	store := &deletedStoryRetentionStoreStub{
		claimResults: [][]storydomain.AttachmentObjectDeletion{{{
			OutboxID: outboxID, AttachmentID: uuid.New(), WorkspaceID: uuid.New(),
			StorageProvider: "azure", ContainerName: "attachments", BlobName: blobName,
			AttemptCount: 3,
		}}},
		failResults:           []bool{true},
		completedPurgeResults: []int64{0},
	}
	objects := &retainedAttachmentObjectStoreStub{
		provider: "azure", container: "attachments",
		deleteErr: errors.New("provider response included " + blobName),
	}
	var logOutput bytes.Buffer
	log := logger.NewWithText(&logOutput, slog.LevelDebug, "story-retention-test")

	err := processAttachmentObjectDeletionsAt(context.Background(), store, objects, log, now)

	require.ErrorIs(t, err, errAttachmentObjectDeletionIncomplete)
	require.NotContains(t, err.Error(), blobName)
	require.NotContains(t, logOutput.String(), blobName)
	require.Equal(t, []storydomain.AttachmentObjectDeletionFailure{{
		OutboxID:      outboxID,
		ClaimToken:    store.claimBatches[0].ClaimToken,
		FailedAt:      now,
		NextAttemptAt: now.Add(4 * time.Minute),
		LastError:     attachmentObjectDeletionSafeFailure,
	}}, store.failures)
}

func TestProcessAttachmentObjectDeletionsRejectsLostCompletionClaim(t *testing.T) {
	t.Parallel()

	store := &deletedStoryRetentionStoreStub{
		claimResults: [][]storydomain.AttachmentObjectDeletion{{{
			OutboxID: uuid.New(), AttachmentID: uuid.New(), WorkspaceID: uuid.New(),
			StorageProvider: "aws", ContainerName: "attachments", BlobName: "object.png",
			AttemptCount: 1,
		}}},
		completeResults:       []bool{false},
		completedPurgeResults: []int64{0},
	}
	objects := &retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"}

	err := processAttachmentObjectDeletionsAt(
		context.Background(), store, objects, newTestJobLogger(), time.Now().UTC(),
	)

	require.ErrorIs(t, err, errAttachmentObjectDeletionClaimLost)
	require.Len(t, objects.deletedBlobNames, 1, "object deletion remains safe because provider delete is idempotent")
}

func TestAttachmentObjectDeletionRefreshesUTCLeaseForEveryClaimBatch(t *testing.T) {
	t.Parallel()

	firstClaimAt := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	secondClaimAt := firstClaimAt.Add(2 * time.Minute)
	retentionAt := secondClaimAt.Add(time.Minute)
	firstBatch := make([]storydomain.AttachmentObjectDeletion, attachmentObjectDeletionBatchSize)
	for index := range firstBatch {
		firstBatch[index] = storydomain.AttachmentObjectDeletion{
			OutboxID: uuid.New(), AttachmentID: uuid.New(), WorkspaceID: uuid.New(),
			StorageProvider: "aws", ContainerName: "attachments",
			BlobName: fmt.Sprintf("object-%d.png", index), AttemptCount: 1,
		}
	}
	store := &deletedStoryRetentionStoreStub{
		claimResults: [][]storydomain.AttachmentObjectDeletion{
			firstBatch,
			{{
				OutboxID: uuid.New(), AttachmentID: uuid.New(), WorkspaceID: uuid.New(),
				StorageProvider: "aws", ContainerName: "attachments", BlobName: "last.png",
				AttemptCount: 1,
			}},
		},
		completedPurgeResults: []int64{0},
	}
	clockValues := []time.Time{firstClaimAt, secondClaimAt, retentionAt}
	clockIndex := 0
	clock := func() time.Time {
		value := clockValues[clockIndex]
		clockIndex++
		return value
	}

	err := processAttachmentObjectDeletionsWithClock(
		context.Background(),
		store,
		&retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"},
		newTestJobLogger(),
		clock,
	)

	require.NoError(t, err)
	require.Len(t, store.claimBatches, 2)
	require.Equal(t, firstClaimAt, store.claimBatches[0].AsOf)
	require.Equal(t, firstClaimAt.Add(-attachmentObjectDeletionLease), store.claimBatches[0].LeaseExpiredBefore)
	require.Equal(t, secondClaimAt, store.claimBatches[1].AsOf)
	require.Equal(t, secondClaimAt.Add(-attachmentObjectDeletionLease), store.claimBatches[1].LeaseExpiredBefore)
	require.Equal(t, firstClaimAt, store.completions[0].CompletedAt)
	require.Equal(t, secondClaimAt, store.completions[len(store.completions)-1].CompletedAt)
	require.Equal(t, retentionAt.Add(-completedObjectDeletionRetention), store.completedPurgeCalls[0].completedBefore)
}

func TestPurgeDeletedStoriesHonorsCancellationAndReportsBoundedBacklog(t *testing.T) {
	t.Parallel()

	objects := &retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelledStore := &deletedStoryRetentionStoreStub{}
	err := purgeDeletedStoriesAt(ctx, cancelledStore, objects, newTestJobLogger(), time.Now().UTC())
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, cancelledStore.purgeBatches)

	results := make([]storydomain.StoryRetentionResult, deletedStoryRetentionMaxBatches)
	deletedAt := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	for index := range results {
		results[index] = storydomain.StoryRetentionResult{
			CandidateCount:    deletedStoryRetentionBatchSize,
			DeletedStoryCount: deletedStoryRetentionBatchSize,
			NextCursor: storydomain.StoryRetentionCursor{
				DeletedAt: deletedAt.Add(time.Duration(index) * time.Second),
				StoryID:   uuid.New(),
				Valid:     true,
			},
		}
	}
	backlogStore := &deletedStoryRetentionStoreStub{purgeResults: results}
	err = purgeDeletedStoriesAt(
		context.Background(), backlogStore, objects, newTestJobLogger(), time.Now().UTC(),
	)
	require.ErrorIs(t, err, errDeletedStoryRetentionBacklog)
	require.Len(t, backlogStore.purgeBatches, deletedStoryRetentionMaxBatches)
}

func TestPurgeDeletedStoriesRejectsACompositeCursorThatDoesNotAdvance(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	firstID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	previousID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	store := &deletedStoryRetentionStoreStub{purgeResults: []storydomain.StoryRetentionResult{
		{
			CandidateCount:    deletedStoryRetentionBatchSize,
			DeletedStoryCount: deletedStoryRetentionBatchSize,
			NextCursor: storydomain.StoryRetentionCursor{
				DeletedAt: deletedAt,
				StoryID:   firstID,
				Valid:     true,
			},
		},
		{
			CandidateCount:    1,
			DeletedStoryCount: 1,
			NextCursor: storydomain.StoryRetentionCursor{
				DeletedAt: deletedAt,
				StoryID:   previousID,
				Valid:     true,
			},
		},
	}}

	err := purgeDeletedStoriesAt(
		context.Background(),
		store,
		&retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"},
		newTestJobLogger(),
		time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
	)

	require.EqualError(t, err, "purge deleted story retention batch: cursor did not advance")
	require.Len(t, store.purgeBatches, 2)
}

func TestAttachmentObjectDeletionRetryDelayIsExponentialAndCapped(t *testing.T) {
	t.Parallel()

	require.Equal(t, time.Minute, attachmentObjectDeletionRetryDelay(0))
	require.Equal(t, time.Minute, attachmentObjectDeletionRetryDelay(1))
	require.Equal(t, 2*time.Minute, attachmentObjectDeletionRetryDelay(2))
	require.Equal(t, 4*time.Minute, attachmentObjectDeletionRetryDelay(3))
	require.Equal(t, attachmentObjectDeletionMaximumBackoff, attachmentObjectDeletionRetryDelay(100))
}

type deletedStoryRetentionStoreStub struct {
	purgeResults          []storydomain.StoryRetentionResult
	purgeBatches          []storydomain.StoryRetentionBatch
	purgeErr              error
	claimResults          [][]storydomain.AttachmentObjectDeletion
	claimBatches          []storydomain.AttachmentObjectDeletionClaimBatch
	claimErr              error
	completeResults       []bool
	completions           []storydomain.AttachmentObjectDeletionCompletion
	completeErr           error
	failResults           []bool
	failures              []storydomain.AttachmentObjectDeletionFailure
	failErr               error
	completedPurgeResults []int64
	completedPurgeCalls   []completedObjectDeletionPurgeCall
	completedPurgeErr     error
}

func (store *deletedStoryRetentionStoreStub) PurgeDeletedStoriesBatch(
	_ context.Context,
	batch storydomain.StoryRetentionBatch,
) (storydomain.StoryRetentionResult, error) {
	store.purgeBatches = append(store.purgeBatches, batch)
	if store.purgeErr != nil {
		return storydomain.StoryRetentionResult{}, store.purgeErr
	}
	if len(store.purgeResults) == 0 {
		return storydomain.StoryRetentionResult{}, nil
	}
	result := store.purgeResults[0]
	store.purgeResults = store.purgeResults[1:]
	return result, nil
}

func (store *deletedStoryRetentionStoreStub) ClaimAttachmentObjectDeletions(
	_ context.Context,
	batch storydomain.AttachmentObjectDeletionClaimBatch,
) ([]storydomain.AttachmentObjectDeletion, error) {
	store.claimBatches = append(store.claimBatches, batch)
	if store.claimErr != nil {
		return nil, store.claimErr
	}
	if len(store.claimResults) == 0 {
		return nil, nil
	}
	result := append([]storydomain.AttachmentObjectDeletion(nil), store.claimResults[0]...)
	store.claimResults = store.claimResults[1:]
	for index := range result {
		result[index].ClaimToken = batch.ClaimToken
	}
	return result, nil
}

func (store *deletedStoryRetentionStoreStub) CompleteAttachmentObjectDeletion(
	_ context.Context,
	completion storydomain.AttachmentObjectDeletionCompletion,
) (bool, error) {
	store.completions = append(store.completions, completion)
	if store.completeErr != nil {
		return false, store.completeErr
	}
	if len(store.completeResults) == 0 {
		return true, nil
	}
	result := store.completeResults[0]
	store.completeResults = store.completeResults[1:]
	return result, nil
}

func (store *deletedStoryRetentionStoreStub) FailAttachmentObjectDeletion(
	_ context.Context,
	failure storydomain.AttachmentObjectDeletionFailure,
) (bool, error) {
	store.failures = append(store.failures, failure)
	if store.failErr != nil {
		return false, store.failErr
	}
	if len(store.failResults) == 0 {
		return true, nil
	}
	result := store.failResults[0]
	store.failResults = store.failResults[1:]
	return result, nil
}

func (store *deletedStoryRetentionStoreStub) PurgeCompletedAttachmentObjectDeletions(
	_ context.Context,
	completedBefore time.Time,
	batchSize int,
) (int64, error) {
	store.completedPurgeCalls = append(store.completedPurgeCalls, completedObjectDeletionPurgeCall{
		completedBefore: completedBefore,
		batchSize:       batchSize,
	})
	if store.completedPurgeErr != nil {
		return 0, store.completedPurgeErr
	}
	if len(store.completedPurgeResults) == 0 {
		return 0, nil
	}
	result := store.completedPurgeResults[0]
	store.completedPurgeResults = store.completedPurgeResults[1:]
	return result, nil
}

type completedObjectDeletionPurgeCall struct {
	completedBefore time.Time
	batchSize       int
}

type retainedAttachmentObjectStoreStub struct {
	provider         string
	container        string
	routeErr         error
	deleteErr        error
	deletedBlobNames []string
}

func (store *retainedAttachmentObjectStoreStub) RetainedObjectStorage() (string, string, error) {
	return store.provider, store.container, store.routeErr
}

func (store *retainedAttachmentObjectStoreStub) DeleteRetainedObject(
	_ context.Context,
	_, _, blobName string,
) error {
	store.deletedBlobNames = append(store.deletedBlobNames, blobName)
	return store.deleteErr
}
