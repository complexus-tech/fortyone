package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const (
	maximumStorageProviderLength = 32
	maximumContainerNameLength   = 255
	maximumBlobNameLength        = 1024
	maximumOutboxErrorLength     = 255
)

var errStoryRetentionRepositoryNotConfigured = errors.New("story retention repository is not configured")

// storyRetentionQueries is deliberately narrower than the generated story
// querier. It documents the complete database surface of retention and keeps
// transaction behavior independently testable.
type storyRetentionQueries interface {
	ListDeletedStoryRetentionCandidates(
		context.Context,
		storyreadsql.ListDeletedStoryRetentionCandidatesParams,
	) ([]storyreadsql.ListDeletedStoryRetentionCandidatesRow, error)
	ListStoryRetentionAttachmentCandidates(
		context.Context,
		storyreadsql.ListStoryRetentionAttachmentCandidatesParams,
	) ([]uuid.UUID, error)
	DeleteStoryRetentionCandidates(
		context.Context,
		storyreadsql.DeleteStoryRetentionCandidatesParams,
	) ([]uuid.UUID, error)
	DeleteUnreferencedStoryRetentionAttachments(
		context.Context,
		storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams,
	) ([]storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow, error)
	InsertAttachmentObjectDeletionOutbox(
		context.Context,
		storyreadsql.InsertAttachmentObjectDeletionOutboxParams,
	) (int64, error)
	ClaimAttachmentObjectDeletions(
		context.Context,
		storyreadsql.ClaimAttachmentObjectDeletionsParams,
	) ([]storyreadsql.ClaimAttachmentObjectDeletionsRow, error)
	CompleteAttachmentObjectDeletion(
		context.Context,
		storyreadsql.CompleteAttachmentObjectDeletionParams,
	) (int64, error)
	FailAttachmentObjectDeletion(
		context.Context,
		storyreadsql.FailAttachmentObjectDeletionParams,
	) (int64, error)
	PurgeCompletedAttachmentObjectDeletions(
		context.Context,
		storyreadsql.PurgeCompletedAttachmentObjectDeletionsParams,
	) (int64, error)
}

// PurgeDeletedStoriesBatch permanently removes one stable page of expired
// stories and durably records every object that became unreferenced. All
// database mutations occur in one transaction.
func (r *repo) PurgeDeletedStoriesBatch(
	ctx context.Context,
	batch storydomain.StoryRetentionBatch,
) (storydomain.StoryRetentionResult, error) {
	if ctx == nil {
		return storydomain.StoryRetentionResult{}, errors.New("story retention context is required")
	}
	if r == nil || r.runRetentionTransaction == nil {
		return storydomain.StoryRetentionResult{}, errStoryRetentionRepositoryNotConfigured
	}
	batchSize, attachmentLookahead, err := validateStoryRetentionBatch(batch)
	if err != nil {
		return storydomain.StoryRetentionResult{}, err
	}

	deletedBefore := batch.DeletedBefore.UTC()
	enqueuedAt := batch.EnqueuedAt.UTC()
	afterDeletedAt := deletedBefore
	afterStoryID := uuid.Nil
	if batch.Cursor.Valid {
		afterDeletedAt = batch.Cursor.DeletedAt.UTC()
		afterStoryID = batch.Cursor.StoryID
	}

	var result storydomain.StoryRetentionResult
	err = r.runRetentionTransaction(ctx, func(queries storyRetentionQueries) error {
		candidates, queryErr := queries.ListDeletedStoryRetentionCandidates(
			ctx,
			storyreadsql.ListDeletedStoryRetentionCandidatesParams{
				DeletedBefore:  &deletedBefore,
				HasCursor:      batch.Cursor.Valid,
				AfterDeletedAt: &afterDeletedAt,
				AfterStoryID:   afterStoryID,
				BatchSize:      batchSize,
			},
		)
		if queryErr != nil {
			return fmt.Errorf("list deleted story retention candidates: %w", queryErr)
		}
		if len(candidates) == 0 {
			result = storydomain.StoryRetentionResult{}
			return nil
		}
		if len(candidates) > int(batchSize) {
			return fmt.Errorf("list deleted story retention candidates: returned %d rows, want at most %d", len(candidates), batchSize)
		}

		storyIDs := make([]uuid.UUID, 0, len(candidates))
		for _, candidate := range candidates {
			if candidate.ID == uuid.Nil || candidate.DeletedAt == nil || candidate.DeletedAt.IsZero() {
				return errors.New("list deleted story retention candidates: database returned an invalid cursor")
			}
			storyIDs = append(storyIDs, candidate.ID)
		}

		attachmentIDs, queryErr := queries.ListStoryRetentionAttachmentCandidates(
			ctx,
			storyreadsql.ListStoryRetentionAttachmentCandidatesParams{
				MaximumAttachmentCount: attachmentLookahead,
				StoryIds:               storyIDs,
			},
		)
		if queryErr != nil {
			return fmt.Errorf("list story retention attachment candidates: %w", queryErr)
		}
		if len(attachmentIDs) > batch.MaximumAttachmentCount {
			return fmt.Errorf(
				"list story retention attachment candidates: exceeded maximum of %d",
				batch.MaximumAttachmentCount,
			)
		}

		deletedStoryIDs, queryErr := queries.DeleteStoryRetentionCandidates(
			ctx,
			storyreadsql.DeleteStoryRetentionCandidatesParams{
				StoryIds:      storyIDs,
				DeletedBefore: &deletedBefore,
			},
		)
		if queryErr != nil {
			return fmt.Errorf("delete story retention candidates: %w", queryErr)
		}
		if len(deletedStoryIDs) != len(storyIDs) {
			return fmt.Errorf(
				"delete story retention candidates: deleted %d of %d selected stories",
				len(deletedStoryIDs),
				len(storyIDs),
			)
		}

		var retiredAttachments []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow
		if len(attachmentIDs) > 0 {
			retiredAttachments, queryErr = queries.DeleteUnreferencedStoryRetentionAttachments(
				ctx,
				storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams{AttachmentIds: attachmentIDs},
			)
			if queryErr != nil {
				return fmt.Errorf("delete unreferenced story retention attachments: %w", queryErr)
			}
		}

		for _, attachment := range retiredAttachments {
			if attachment.AttachmentID == uuid.Nil || attachment.WorkspaceID == uuid.Nil ||
				strings.TrimSpace(attachment.BlobName) == "" || len(attachment.BlobName) > maximumBlobNameLength {
				return errors.New("delete unreferenced story retention attachments: database returned invalid routing metadata")
			}
			inserted, insertErr := queries.InsertAttachmentObjectDeletionOutbox(
				ctx,
				storyreadsql.InsertAttachmentObjectDeletionOutboxParams{
					AttachmentID:    attachment.AttachmentID,
					WorkspaceID:     attachment.WorkspaceID,
					StorageProvider: strings.TrimSpace(batch.StorageProvider),
					ContainerName:   strings.TrimSpace(batch.ContainerName),
					BlobName:        attachment.BlobName,
					EnqueuedAt:      enqueuedAt,
				},
			)
			if insertErr != nil {
				return fmt.Errorf("enqueue attachment object deletion: %w", insertErr)
			}
			if inserted != 1 {
				return fmt.Errorf("enqueue attachment object deletion: inserted %d rows, want 1", inserted)
			}
		}

		last := candidates[len(candidates)-1]
		result = storydomain.StoryRetentionResult{
			CandidateCount:      len(candidates),
			DeletedStoryCount:   int64(len(deletedStoryIDs)),
			EnqueuedObjectCount: int64(len(retiredAttachments)),
			NextCursor: storydomain.StoryRetentionCursor{
				DeletedAt: last.DeletedAt.UTC(),
				StoryID:   last.ID,
				Valid:     true,
			},
		}
		return nil
	})
	if err != nil {
		return storydomain.StoryRetentionResult{}, fmt.Errorf("purge deleted story retention batch: %w", err)
	}
	return result, nil
}

func (r *repo) ClaimAttachmentObjectDeletions(
	ctx context.Context,
	batch storydomain.AttachmentObjectDeletionClaimBatch,
) ([]storydomain.AttachmentObjectDeletion, error) {
	if ctx == nil {
		return nil, errors.New("attachment object deletion claim context is required")
	}
	if err := r.retentionConfigured(); err != nil {
		return nil, err
	}
	batchSize, err := validateAttachmentObjectDeletionClaimBatch(batch)
	if err != nil {
		return nil, err
	}
	asOf, leaseExpiredBefore, claimToken := batch.AsOf.UTC(), batch.LeaseExpiredBefore.UTC(), batch.ClaimToken
	rows, err := r.retention.ClaimAttachmentObjectDeletions(
		ctx,
		storyreadsql.ClaimAttachmentObjectDeletionsParams{
			AsOf:               &asOf,
			ClaimToken:         &claimToken,
			LeaseExpiredBefore: &leaseExpiredBefore,
			BatchSize:          batchSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("claim attachment object deletions: %w", err)
	}
	if len(rows) > int(batchSize) {
		return nil, fmt.Errorf("claim attachment object deletions: returned %d rows, want at most %d", len(rows), batchSize)
	}

	deletions := make([]storydomain.AttachmentObjectDeletion, 0, len(rows))
	for _, row := range rows {
		if row.OutboxID == uuid.Nil || row.AttachmentID == uuid.Nil || row.WorkspaceID == uuid.Nil ||
			row.ClaimToken == nil || *row.ClaimToken != claimToken || row.AttemptCount <= 0 ||
			strings.TrimSpace(row.StorageProvider) == "" || strings.TrimSpace(row.ContainerName) == "" ||
			strings.TrimSpace(row.BlobName) == "" {
			return nil, errors.New("claim attachment object deletions: database returned invalid routing metadata")
		}
		attemptCount, convertErr := safecast.Int64(int64(row.AttemptCount))
		if convertErr != nil {
			return nil, fmt.Errorf("claim attachment object deletions: convert attempt count: %w", convertErr)
		}
		deletions = append(deletions, storydomain.AttachmentObjectDeletion{
			OutboxID:        row.OutboxID,
			AttachmentID:    row.AttachmentID,
			WorkspaceID:     row.WorkspaceID,
			StorageProvider: row.StorageProvider,
			ContainerName:   row.ContainerName,
			BlobName:        row.BlobName,
			ClaimToken:      *row.ClaimToken,
			AttemptCount:    attemptCount,
		})
	}
	return deletions, nil
}

func (r *repo) CompleteAttachmentObjectDeletion(
	ctx context.Context,
	completion storydomain.AttachmentObjectDeletionCompletion,
) (bool, error) {
	if ctx == nil {
		return false, errors.New("attachment object deletion completion context is required")
	}
	if err := r.retentionConfigured(); err != nil {
		return false, err
	}
	if completion.OutboxID == uuid.Nil || completion.ClaimToken == uuid.Nil || completion.CompletedAt.IsZero() {
		return false, errors.New("attachment object deletion completion is invalid")
	}
	completedAt, claimToken := completion.CompletedAt.UTC(), completion.ClaimToken
	rows, err := r.retention.CompleteAttachmentObjectDeletion(
		ctx,
		storyreadsql.CompleteAttachmentObjectDeletionParams{
			CompletedAt: &completedAt,
			OutboxID:    completion.OutboxID,
			ClaimToken:  &claimToken,
		},
	)
	if err != nil {
		return false, fmt.Errorf("complete attachment object deletion: %w", err)
	}
	return exactlyOneAffected("complete attachment object deletion", rows)
}

func (r *repo) FailAttachmentObjectDeletion(
	ctx context.Context,
	failure storydomain.AttachmentObjectDeletionFailure,
) (bool, error) {
	if ctx == nil {
		return false, errors.New("attachment object deletion failure context is required")
	}
	if err := r.retentionConfigured(); err != nil {
		return false, err
	}
	lastError := strings.TrimSpace(failure.LastError)
	if failure.OutboxID == uuid.Nil || failure.ClaimToken == uuid.Nil || failure.FailedAt.IsZero() ||
		failure.NextAttemptAt.IsZero() || !failure.NextAttemptAt.After(failure.FailedAt) ||
		lastError == "" || len(lastError) > maximumOutboxErrorLength {
		return false, errors.New("attachment object deletion failure is invalid")
	}
	failedAt, nextAttemptAt, claimToken := failure.FailedAt.UTC(), failure.NextAttemptAt.UTC(), failure.ClaimToken
	rows, err := r.retention.FailAttachmentObjectDeletion(
		ctx,
		storyreadsql.FailAttachmentObjectDeletionParams{
			NextAttemptAt: nextAttemptAt,
			LastError:     &lastError,
			FailedAt:      failedAt,
			OutboxID:      failure.OutboxID,
			ClaimToken:    &claimToken,
		},
	)
	if err != nil {
		return false, fmt.Errorf("fail attachment object deletion: %w", err)
	}
	return exactlyOneAffected("fail attachment object deletion", rows)
}

func (r *repo) PurgeCompletedAttachmentObjectDeletions(
	ctx context.Context,
	completedBefore time.Time,
	batchSize int,
) (int64, error) {
	if ctx == nil {
		return 0, errors.New("completed attachment object deletion purge context is required")
	}
	if err := r.retentionConfigured(); err != nil {
		return 0, err
	}
	if completedBefore.IsZero() || batchSize <= 0 {
		return 0, errors.New("completed attachment object deletion purge input is invalid")
	}
	convertedBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, fmt.Errorf("convert completed attachment object deletion purge batch size: %w", err)
	}
	completedBefore = completedBefore.UTC()
	rows, err := r.retention.PurgeCompletedAttachmentObjectDeletions(
		ctx,
		storyreadsql.PurgeCompletedAttachmentObjectDeletionsParams{
			CompletedBefore: &completedBefore,
			BatchSize:       convertedBatchSize,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("purge completed attachment object deletions: %w", err)
	}
	if rows < 0 || rows > int64(batchSize) {
		return 0, fmt.Errorf("purge completed attachment object deletions: deleted %d rows, want 0..%d", rows, batchSize)
	}
	return rows, nil
}

func (r *repo) retentionConfigured() error {
	if r == nil || r.retention == nil {
		return errStoryRetentionRepositoryNotConfigured
	}
	return nil
}

func validateStoryRetentionBatch(batch storydomain.StoryRetentionBatch) (int32, int32, error) {
	if batch.DeletedBefore.IsZero() || batch.EnqueuedAt.IsZero() ||
		!batch.DeletedBefore.Before(batch.EnqueuedAt) || batch.BatchSize <= 0 ||
		batch.MaximumAttachmentCount <= 0 || batch.MaximumAttachmentCount >= math.MaxInt32 {
		return 0, 0, errors.New("story retention batch is invalid")
	}
	provider, container := strings.TrimSpace(batch.StorageProvider), strings.TrimSpace(batch.ContainerName)
	if provider == "" || len(provider) > maximumStorageProviderLength ||
		container == "" || len(container) > maximumContainerNameLength {
		return 0, 0, errors.New("story retention storage route is invalid")
	}
	if batch.Cursor.Valid && (batch.Cursor.DeletedAt.IsZero() || batch.Cursor.StoryID == uuid.Nil) {
		return 0, 0, errors.New("story retention cursor is invalid")
	}
	batchSize, err := safecast.Int32(batch.BatchSize)
	if err != nil {
		return 0, 0, fmt.Errorf("convert story retention batch size: %w", err)
	}
	attachmentLookahead, err := safecast.Int32(batch.MaximumAttachmentCount + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("convert story retention attachment lookahead: %w", err)
	}
	return batchSize, attachmentLookahead, nil
}

func validateAttachmentObjectDeletionClaimBatch(
	batch storydomain.AttachmentObjectDeletionClaimBatch,
) (int32, error) {
	if batch.AsOf.IsZero() || batch.LeaseExpiredBefore.IsZero() ||
		batch.LeaseExpiredBefore.After(batch.AsOf) || batch.ClaimToken == uuid.Nil || batch.BatchSize <= 0 {
		return 0, errors.New("attachment object deletion claim batch is invalid")
	}
	batchSize, err := safecast.Int32(batch.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("convert attachment object deletion claim batch size: %w", err)
	}
	return batchSize, nil
}

func exactlyOneAffected(operation string, rows int64) (bool, error) {
	if rows < 0 || rows > 1 {
		return false, fmt.Errorf("%s: affected %d rows, want 0 or 1", operation, rows)
	}
	return rows == 1, nil
}
