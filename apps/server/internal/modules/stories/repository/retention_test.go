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

func TestStoryRetentionTransactionEnqueuesOnlyAttachmentsThatBecameOrphaned(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	deletedAt := now.Add(-31 * 24 * time.Hour)
	storyID, sharedAttachmentID, orphanAttachmentID := uuid.New(), uuid.New(), uuid.New()
	workspaceID := uuid.New()
	queries := &storyRetentionQueryStub{
		candidateRows: []storyreadsql.ListDeletedStoryRetentionCandidatesRow{{
			ID: storyID, DeletedAt: &deletedAt,
		}},
		attachmentIDs:   []uuid.UUID{sharedAttachmentID, orphanAttachmentID},
		deletedStoryIDs: []uuid.UUID{storyID},
		retiredAttachmentRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{{
			AttachmentID: orphanAttachmentID,
			WorkspaceID:  workspaceID,
			BlobName:     "orphan.png",
		}},
		insertRows: 1,
	}
	repository := &repo{retention: queries}
	repository.runRetentionTransaction = func(
		_ context.Context,
		operation func(storyRetentionQueries) error,
	) error {
		return operation(queries)
	}

	result, err := repository.PurgeDeletedStoriesBatch(context.Background(), storydomain.StoryRetentionBatch{
		DeletedBefore:          now.Add(-30 * 24 * time.Hour),
		EnqueuedAt:             now,
		BatchSize:              10,
		MaximumAttachmentCount: 5,
		StorageProvider:        "aws",
		ContainerName:          "attachments",
	})

	require.NoError(t, err)
	require.Equal(t, storydomain.StoryRetentionResult{
		CandidateCount:      1,
		DeletedStoryCount:   1,
		EnqueuedObjectCount: 1,
		NextCursor: storydomain.StoryRetentionCursor{
			DeletedAt: deletedAt,
			StoryID:   storyID,
			Valid:     true,
		},
	}, result)
	require.Equal(t, int32(6), queries.attachmentParams.MaximumAttachmentCount, "adapter must request one bounded look-ahead row")
	require.Equal(t, []uuid.UUID{storyID}, queries.attachmentParams.StoryIds)
	require.Equal(t, []uuid.UUID{sharedAttachmentID, orphanAttachmentID}, queries.deleteAttachmentParams.AttachmentIds)
	require.Equal(t, []storyreadsql.InsertAttachmentObjectDeletionOutboxParams{{
		AttachmentID:    orphanAttachmentID,
		WorkspaceID:     workspaceID,
		StorageProvider: "aws",
		ContainerName:   "attachments",
		BlobName:        "orphan.png",
		EnqueuedAt:      now,
	}}, queries.insertParams)
}

func TestStoryRetentionTransactionErrorReturnsNoCursorOrCommittedResult(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("outbox unavailable")
	now := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	deletedAt := now.Add(-31 * 24 * time.Hour)
	storyID, attachmentID := uuid.New(), uuid.New()
	queries := &storyRetentionQueryStub{
		candidateRows:   []storyreadsql.ListDeletedStoryRetentionCandidatesRow{{ID: storyID, DeletedAt: &deletedAt}},
		attachmentIDs:   []uuid.UUID{attachmentID},
		deletedStoryIDs: []uuid.UUID{storyID},
		retiredAttachmentRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{{
			AttachmentID: attachmentID, WorkspaceID: uuid.New(), BlobName: "rollback.png",
		}},
		insertErr: wantErr,
	}
	rolledBack := false
	repository := &repo{retention: queries}
	repository.runRetentionTransaction = func(
		_ context.Context,
		operation func(storyRetentionQueries) error,
	) error {
		err := operation(queries)
		rolledBack = err != nil
		return err
	}

	result, err := repository.PurgeDeletedStoriesBatch(context.Background(), storydomain.StoryRetentionBatch{
		DeletedBefore: now.Add(-30 * 24 * time.Hour), EnqueuedAt: now,
		BatchSize: 10, MaximumAttachmentCount: 10,
		StorageProvider: "aws", ContainerName: "attachments",
	})

	require.ErrorIs(t, err, wantErr)
	require.True(t, rolledBack)
	require.Zero(t, result, "an aborted transaction must never expose a cursor for the caller to advance")
	require.Equal(t, []string{"list_stories", "list_attachments", "delete_stories", "delete_attachments", "insert_outbox"}, queries.calls)
}

func TestAttachmentObjectDeletionAdapterMapsTypedClaimAndFencing(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	claimToken := uuid.New()
	queries := &storyRetentionQueryStub{
		claimRows: []storyreadsql.ClaimAttachmentObjectDeletionsRow{{
			OutboxID: uuid.New(), AttachmentID: uuid.New(), WorkspaceID: uuid.New(),
			StorageProvider: "azure", ContainerName: "attachments", BlobName: "opaque.png",
			ClaimToken: &claimToken, AttemptCount: 2,
		}},
		completeRows: 1,
		failRows:     0,
	}
	repository := &repo{retention: queries}

	claims, err := repository.ClaimAttachmentObjectDeletions(
		context.Background(),
		storydomain.AttachmentObjectDeletionClaimBatch{
			AsOf: now, LeaseExpiredBefore: now.Add(-5 * time.Minute),
			ClaimToken: claimToken, BatchSize: 25,
		},
	)
	require.NoError(t, err)
	require.Len(t, claims, 1)
	require.Equal(t, claimToken, claims[0].ClaimToken)
	require.Equal(t, int32(25), queries.claimParams.BatchSize)
	require.Equal(t, now, *queries.claimParams.AsOf)

	completed, err := repository.CompleteAttachmentObjectDeletion(
		context.Background(),
		storydomain.AttachmentObjectDeletionCompletion{
			OutboxID: claims[0].OutboxID, ClaimToken: claimToken, CompletedAt: now,
		},
	)
	require.NoError(t, err)
	require.True(t, completed)

	failed, err := repository.FailAttachmentObjectDeletion(
		context.Background(),
		storydomain.AttachmentObjectDeletionFailure{
			OutboxID: claims[0].OutboxID, ClaimToken: claimToken,
			FailedAt: now, NextAttemptAt: now.Add(time.Minute), LastError: "safe failure",
		},
	)
	require.NoError(t, err)
	require.False(t, failed, "zero affected rows communicates a lost claim fence")
}

type storyRetentionQueryStub struct {
	candidateRows          []storyreadsql.ListDeletedStoryRetentionCandidatesRow
	candidateErr           error
	attachmentIDs          []uuid.UUID
	attachmentErr          error
	attachmentParams       storyreadsql.ListStoryRetentionAttachmentCandidatesParams
	deletedStoryIDs        []uuid.UUID
	deleteStoryErr         error
	deleteAttachmentParams storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams
	retiredAttachmentRows  []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow
	deleteAttachmentErr    error
	insertParams           []storyreadsql.InsertAttachmentObjectDeletionOutboxParams
	insertRows             int64
	insertErr              error
	claimParams            storyreadsql.ClaimAttachmentObjectDeletionsParams
	claimRows              []storyreadsql.ClaimAttachmentObjectDeletionsRow
	claimErr               error
	completeRows           int64
	completeErr            error
	failRows               int64
	failErr                error
	purgeRows              int64
	purgeErr               error
	calls                  []string
}

func (queries *storyRetentionQueryStub) ListDeletedStoryRetentionCandidates(
	_ context.Context,
	_ storyreadsql.ListDeletedStoryRetentionCandidatesParams,
) ([]storyreadsql.ListDeletedStoryRetentionCandidatesRow, error) {
	queries.calls = append(queries.calls, "list_stories")
	return queries.candidateRows, queries.candidateErr
}

func (queries *storyRetentionQueryStub) ListStoryRetentionAttachmentCandidates(
	_ context.Context,
	params storyreadsql.ListStoryRetentionAttachmentCandidatesParams,
) ([]uuid.UUID, error) {
	queries.calls = append(queries.calls, "list_attachments")
	queries.attachmentParams = params
	return queries.attachmentIDs, queries.attachmentErr
}

func (queries *storyRetentionQueryStub) DeleteStoryRetentionCandidates(
	_ context.Context,
	_ storyreadsql.DeleteStoryRetentionCandidatesParams,
) ([]uuid.UUID, error) {
	queries.calls = append(queries.calls, "delete_stories")
	return queries.deletedStoryIDs, queries.deleteStoryErr
}

func (queries *storyRetentionQueryStub) DeleteUnreferencedStoryRetentionAttachments(
	_ context.Context,
	params storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams,
) ([]storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow, error) {
	queries.calls = append(queries.calls, "delete_attachments")
	queries.deleteAttachmentParams = params
	return queries.retiredAttachmentRows, queries.deleteAttachmentErr
}

func (queries *storyRetentionQueryStub) InsertAttachmentObjectDeletionOutbox(
	_ context.Context,
	params storyreadsql.InsertAttachmentObjectDeletionOutboxParams,
) (int64, error) {
	queries.calls = append(queries.calls, "insert_outbox")
	queries.insertParams = append(queries.insertParams, params)
	return queries.insertRows, queries.insertErr
}

func (queries *storyRetentionQueryStub) ClaimAttachmentObjectDeletions(
	_ context.Context,
	params storyreadsql.ClaimAttachmentObjectDeletionsParams,
) ([]storyreadsql.ClaimAttachmentObjectDeletionsRow, error) {
	queries.claimParams = params
	return queries.claimRows, queries.claimErr
}

func (queries *storyRetentionQueryStub) CompleteAttachmentObjectDeletion(
	context.Context,
	storyreadsql.CompleteAttachmentObjectDeletionParams,
) (int64, error) {
	return queries.completeRows, queries.completeErr
}

func (queries *storyRetentionQueryStub) FailAttachmentObjectDeletion(
	context.Context,
	storyreadsql.FailAttachmentObjectDeletionParams,
) (int64, error) {
	return queries.failRows, queries.failErr
}

func (queries *storyRetentionQueryStub) PurgeCompletedAttachmentObjectDeletions(
	context.Context,
	storyreadsql.PurgeCompletedAttachmentObjectDeletionsParams,
) (int64, error) {
	return queries.purgeRows, queries.purgeErr
}
