package jobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	deletedStoryRetention              = 30 * 24 * time.Hour
	deletedStoryRetentionBatchSize     = 100
	deletedStoryRetentionMaxBatches    = 100
	deletedStoryAttachmentMaximum      = 5_000
	attachmentObjectDeletionBatchSize  = 100
	attachmentObjectDeletionMaxBatches = 100
	// The shared handler must retain its four-minute task timeout so no claim can
	// remain active for the full five-minute lease inside one invocation.
	attachmentObjectDeletionLease          = 5 * time.Minute
	attachmentObjectDeletionMaximumBackoff = 24 * time.Hour
	attachmentObjectDeletionSafeFailure    = "object storage deletion failed"
	completedObjectDeletionRetention       = 30 * 24 * time.Hour
	completedObjectDeletionPurgeBatchSize  = 500
	completedObjectDeletionPurgeMaxBatches = 100
)

var (
	errDeletedStoryRetentionBacklog       = errors.New("deleted story retention backlog remains")
	errAttachmentObjectDeletionBacklog    = errors.New("attachment object deletion backlog remains")
	errAttachmentObjectDeletionIncomplete = errors.New("one or more attachment object deletions require retry")
	errAttachmentObjectDeletionClaimLost  = errors.New("attachment object deletion claim is no longer owned")
)

// DeletedStoryRetentionStore is the worker-owned persistence capability for
// atomic story retirement and fenced attachment-object deletion delivery.
type DeletedStoryRetentionStore interface {
	PurgeDeletedStoriesBatch(
		context.Context,
		storydomain.StoryRetentionBatch,
	) (storydomain.StoryRetentionResult, error)
	ClaimAttachmentObjectDeletions(
		context.Context,
		storydomain.AttachmentObjectDeletionClaimBatch,
	) ([]storydomain.AttachmentObjectDeletion, error)
	CompleteAttachmentObjectDeletion(
		context.Context,
		storydomain.AttachmentObjectDeletionCompletion,
	) (bool, error)
	FailAttachmentObjectDeletion(
		context.Context,
		storydomain.AttachmentObjectDeletionFailure,
	) (bool, error)
	PurgeCompletedAttachmentObjectDeletions(context.Context, time.Time, int) (int64, error)
}

// RetainedAttachmentObjectStore exposes only the configured, credential-free
// route and idempotent delete operation required by retention.
type RetainedAttachmentObjectStore interface {
	RetainedObjectStorage() (provider string, container string, err error)
	DeleteRetainedObject(ctx context.Context, provider string, container string, blobName string) error
}

// PurgeDeletedStories permanently deletes stories retained in trash for more
// than 30 days and atomically enqueues every resulting object deletion.
func PurgeDeletedStories(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
) error {
	return purgeDeletedStoriesAt(ctx, store, objects, log, time.Now().UTC())
}

func purgeDeletedStoriesAt(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("deleted story retention context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedStories")
	defer span.End()
	if store == nil {
		return errors.New("deleted story retention store is required")
	}
	if objects == nil {
		return errors.New("attachment object store is required")
	}
	if log == nil {
		return errors.New("deleted story retention logger is required")
	}
	if now.IsZero() {
		return errors.New("deleted story retention clock is required")
	}

	now = now.UTC()
	provider, container, err := objects.RetainedObjectStorage()
	if err != nil {
		return fmt.Errorf("resolve attachment object storage route: %w", err)
	}
	if provider == "" || container == "" {
		return errors.New("resolve attachment object storage route: route is incomplete")
	}

	log.Info(ctx, "Purging stories deleted for more than 30 days")
	startedAt := time.Now()
	deletedBefore := now.Add(-deletedStoryRetention)
	var cursor storydomain.StoryRetentionCursor
	var totalCandidates int64
	var totalStoriesDeleted int64
	var totalObjectsEnqueued int64
	storyBatches := 0
	for ; storyBatches < deletedStoryRetentionMaxBatches; storyBatches++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("purge deleted stories after %d candidates: %w", totalCandidates, err)
		}

		result, batchErr := store.PurgeDeletedStoriesBatch(ctx, storydomain.StoryRetentionBatch{
			DeletedBefore:          deletedBefore,
			EnqueuedAt:             now,
			Cursor:                 cursor,
			BatchSize:              deletedStoryRetentionBatchSize,
			MaximumAttachmentCount: deletedStoryAttachmentMaximum,
			StorageProvider:        provider,
			ContainerName:          container,
		})
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("purge deleted story retention batch: %w", batchErr)
		}
		if err := validateStoryRetentionResult(result); err != nil {
			return err
		}
		if result.CandidateCount == 0 {
			break
		}
		if cursor.Valid && !storyRetentionCursorAdvances(cursor, result.NextCursor) {
			return errors.New("purge deleted story retention batch: cursor did not advance")
		}

		totalCandidates += int64(result.CandidateCount)
		totalStoriesDeleted += result.DeletedStoryCount
		totalObjectsEnqueued += result.EnqueuedObjectCount
		cursor = result.NextCursor
		span.AddEvent("deleted_story_retention_batch", trace.WithAttributes(
			attribute.Int("batch", storyBatches+1),
			attribute.Int("stories.candidates", result.CandidateCount),
			attribute.Int64("stories.deleted", result.DeletedStoryCount),
			attribute.Int64("objects.enqueued", result.EnqueuedObjectCount),
		))
		if result.CandidateCount < deletedStoryRetentionBatchSize {
			storyBatches++
			break
		}
	}
	if storyBatches >= deletedStoryRetentionMaxBatches {
		return fmt.Errorf(
			"purge deleted stories after %d candidates: %w",
			totalCandidates,
			errDeletedStoryRetentionBacklog,
		)
	}

	duration := time.Since(startedAt)
	span.AddEvent("deleted_story_retention_completed", trace.WithAttributes(
		attribute.Int64("stories.candidates", totalCandidates),
		attribute.Int64("stories.deleted", totalStoriesDeleted),
		attribute.Int64("objects.enqueued", totalObjectsEnqueued),
		attribute.Int("story_batches", storyBatches),
	))
	log.Info(
		ctx,
		"Deleted story retention job completed",
		"stories_candidates", totalCandidates,
		"stories_deleted", totalStoriesDeleted,
		"objects_enqueued", totalObjectsEnqueued,
		"story_batches", storyBatches,
		"duration", duration,
	)
	return nil
}

// ProcessAttachmentObjectDeletions is scheduled independently from the daily
// story purge so due retries and expired leases are serviced at minute-level
// cadence without making object-store availability part of the DB transaction.
func ProcessAttachmentObjectDeletions(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
) error {
	return processAttachmentObjectDeletionsWithClock(ctx, store, objects, log, func() time.Time {
		return time.Now().UTC()
	})
}

func processAttachmentObjectDeletionsAt(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
	now time.Time,
) error {
	return processAttachmentObjectDeletionsWithClock(ctx, store, objects, log, func() time.Time {
		return now
	})
}

func processAttachmentObjectDeletionsWithClock(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
	clock func() time.Time,
) error {
	if ctx == nil {
		return errors.New("attachment object deletion context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessAttachmentObjectDeletions")
	defer span.End()
	if store == nil {
		return errors.New("attachment object deletion store is required")
	}
	if objects == nil {
		return errors.New("attachment object store is required")
	}
	if log == nil {
		return errors.New("attachment object deletion logger is required")
	}
	if clock == nil {
		return errors.New("attachment object deletion clock is required")
	}

	stats, deliveryErr := drainAttachmentObjectDeletions(ctx, store, objects, log, clock)
	retentionNow := clock().UTC()
	if retentionNow.IsZero() {
		return errors.Join(deliveryErr, errors.New("attachment object deletion clock is required"))
	}
	completedRowsPurged, retentionErr := purgeCompletedAttachmentObjectDeletionRows(
		ctx,
		store,
		retentionNow.Add(-completedObjectDeletionRetention),
	)
	span.AddEvent("attachment_object_deletions_processed", trace.WithAttributes(
		attribute.Int64("objects.claimed", stats.claimed),
		attribute.Int64("objects.completed", stats.completed),
		attribute.Int64("objects.failed", stats.failed),
		attribute.Int64("outbox.completed_rows_purged", completedRowsPurged),
	))
	if deliveryErr != nil {
		span.RecordError(deliveryErr)
	}
	if retentionErr != nil {
		span.RecordError(retentionErr)
	}
	if err := errors.Join(deliveryErr, retentionErr); err != nil {
		return err
	}
	log.Info(
		ctx,
		"Attachment object deletion job completed",
		"objects_claimed", stats.claimed,
		"objects_completed", stats.completed,
		"completed_outbox_rows_purged", completedRowsPurged,
	)
	return nil
}

type attachmentObjectDeletionStats struct {
	claimed   int64
	completed int64
	failed    int64
}

func drainAttachmentObjectDeletions(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	objects RetainedAttachmentObjectStore,
	log *logger.Logger,
	clock func() time.Time,
) (attachmentObjectDeletionStats, error) {
	var stats attachmentObjectDeletionStats
	for batchIndex := 0; batchIndex < attachmentObjectDeletionMaxBatches; batchIndex++ {
		if err := ctx.Err(); err != nil {
			return stats, err
		}
		claimNow := clock().UTC()
		if claimNow.IsZero() {
			return stats, errors.New("attachment object deletion clock is required")
		}
		claimToken, err := uuid.NewRandom()
		if err != nil {
			return stats, errors.New("create attachment object deletion claim token")
		}
		deletions, err := store.ClaimAttachmentObjectDeletions(
			ctx,
			storydomain.AttachmentObjectDeletionClaimBatch{
				AsOf:               claimNow,
				LeaseExpiredBefore: claimNow.Add(-attachmentObjectDeletionLease),
				ClaimToken:         claimToken,
				BatchSize:          attachmentObjectDeletionBatchSize,
			},
		)
		if err != nil {
			return stats, fmt.Errorf("claim attachment object deletions: %w", err)
		}
		if len(deletions) > attachmentObjectDeletionBatchSize {
			return stats, fmt.Errorf(
				"claim attachment object deletions: returned %d rows, want at most %d",
				len(deletions),
				attachmentObjectDeletionBatchSize,
			)
		}
		if len(deletions) == 0 {
			if stats.failed > 0 {
				return stats, errAttachmentObjectDeletionIncomplete
			}
			return stats, nil
		}

		stats.claimed += int64(len(deletions))
		for _, deletion := range deletions {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
			if deletion.OutboxID == uuid.Nil || deletion.AttachmentID == uuid.Nil ||
				deletion.ClaimToken != claimToken || deletion.AttemptCount <= 0 {
				return stats, errors.New("claim attachment object deletions: store returned an invalid claim")
			}

			if err := objects.DeleteRetainedObject(
				ctx,
				deletion.StorageProvider,
				deletion.ContainerName,
				deletion.BlobName,
			); err != nil {
				failed, failErr := store.FailAttachmentObjectDeletion(
					ctx,
					storydomain.AttachmentObjectDeletionFailure{
						OutboxID:      deletion.OutboxID,
						ClaimToken:    deletion.ClaimToken,
						FailedAt:      claimNow,
						NextAttemptAt: claimNow.Add(attachmentObjectDeletionRetryDelay(deletion.AttemptCount)),
						LastError:     attachmentObjectDeletionSafeFailure,
					},
				)
				if failErr != nil {
					return stats, fmt.Errorf("record attachment object deletion failure: %w", failErr)
				}
				if !failed {
					return stats, errAttachmentObjectDeletionClaimLost
				}
				stats.failed++
				log.Warn(
					ctx,
					"Attachment object deletion scheduled for retry",
					"outbox_id", deletion.OutboxID,
					"attachment_id", deletion.AttachmentID,
					"attempt", deletion.AttemptCount,
				)
				continue
			}

			completed, completeErr := store.CompleteAttachmentObjectDeletion(
				ctx,
				storydomain.AttachmentObjectDeletionCompletion{
					OutboxID:    deletion.OutboxID,
					ClaimToken:  deletion.ClaimToken,
					CompletedAt: claimNow,
				},
			)
			if completeErr != nil {
				return stats, fmt.Errorf("complete attachment object deletion: %w", completeErr)
			}
			if !completed {
				return stats, errAttachmentObjectDeletionClaimLost
			}
			stats.completed++
		}

		if len(deletions) < attachmentObjectDeletionBatchSize {
			if stats.failed > 0 {
				return stats, errAttachmentObjectDeletionIncomplete
			}
			return stats, nil
		}
	}
	return stats, errAttachmentObjectDeletionBacklog
}

func validateStoryRetentionResult(result storydomain.StoryRetentionResult) error {
	if result.CandidateCount < 0 || result.CandidateCount > deletedStoryRetentionBatchSize ||
		result.DeletedStoryCount < 0 || result.DeletedStoryCount != int64(result.CandidateCount) ||
		result.EnqueuedObjectCount < 0 || result.EnqueuedObjectCount > deletedStoryAttachmentMaximum {
		return fmt.Errorf(
			"purge deleted story retention batch: invalid result candidates=%d deleted=%d objects=%d",
			result.CandidateCount,
			result.DeletedStoryCount,
			result.EnqueuedObjectCount,
		)
	}
	if result.CandidateCount > 0 &&
		(!result.NextCursor.Valid || result.NextCursor.DeletedAt.IsZero() || result.NextCursor.StoryID == uuid.Nil) {
		return errors.New("purge deleted story retention batch: non-empty result requires a cursor")
	}
	return nil
}

func storyRetentionCursorAdvances(
	previous storydomain.StoryRetentionCursor,
	next storydomain.StoryRetentionCursor,
) bool {
	if next.DeletedAt.After(previous.DeletedAt) {
		return true
	}
	if !next.DeletedAt.Equal(previous.DeletedAt) {
		return false
	}
	// PostgreSQL compares uuid values by their 16-byte representation. Matching
	// that order keeps the worker guard aligned with the SQL keyset predicate.
	return bytes.Compare(next.StoryID[:], previous.StoryID[:]) > 0
}

func attachmentObjectDeletionRetryDelay(attempt int) time.Duration {
	delay := time.Minute
	for retry := 1; retry < attempt && delay < attachmentObjectDeletionMaximumBackoff; retry++ {
		if delay >= attachmentObjectDeletionMaximumBackoff/2 {
			return attachmentObjectDeletionMaximumBackoff
		}
		delay *= 2
	}
	if delay > attachmentObjectDeletionMaximumBackoff {
		return attachmentObjectDeletionMaximumBackoff
	}
	return delay
}

func purgeCompletedAttachmentObjectDeletionRows(
	ctx context.Context,
	store DeletedStoryRetentionStore,
	completedBefore time.Time,
) (int64, error) {
	var total int64
	for batchIndex := 0; batchIndex < completedObjectDeletionPurgeMaxBatches; batchIndex++ {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		purged, err := store.PurgeCompletedAttachmentObjectDeletions(
			ctx,
			completedBefore,
			completedObjectDeletionPurgeBatchSize,
		)
		if err != nil {
			return total, fmt.Errorf("purge completed attachment object deletion rows: %w", err)
		}
		if purged < 0 || purged > completedObjectDeletionPurgeBatchSize {
			return total, fmt.Errorf(
				"purge completed attachment object deletion rows: deleted %d rows, want 0..%d",
				purged,
				completedObjectDeletionPurgeBatchSize,
			)
		}
		total += purged
		if purged < completedObjectDeletionPurgeBatchSize {
			return total, nil
		}
	}
	return total, errAttachmentObjectDeletionBacklog
}
