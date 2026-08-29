package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const deletedWorkspaceRetention = 48 * time.Hour

// SoftDeletedWorkspacePurger is the worker-owned persistence capability for
// bounded, integration-safe workspace trash purges.
type SoftDeletedWorkspacePurger interface {
	PurgeSoftDeletedWorkspacesBatch(
		context.Context,
		workspacedomain.DeletedWorkspacePurgeBatch,
	) (workspacedomain.DeletedWorkspacePurgeResult, error)
}

// PurgeDeletedWorkspaces permanently deletes workspaces that have been marked as deleted for 48+ hours
func PurgeDeletedWorkspaces(ctx context.Context, store SoftDeletedWorkspacePurger, log *logger.Logger) error {
	return purgeDeletedWorkspacesAt(ctx, store, log, time.Now().UTC())
}

func purgeDeletedWorkspacesAt(
	ctx context.Context,
	store SoftDeletedWorkspacePurger,
	log *logger.Logger,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("soft-deleted workspace purge context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedWorkspaces")
	defer span.End()
	if store == nil {
		return errors.New("soft-deleted workspace purge store is required")
	}
	if log == nil {
		return errors.New("soft-deleted workspace purge logger is required")
	}
	if now.IsZero() {
		return errors.New("soft-deleted workspace purge clock is required")
	}

	now = now.UTC()
	deletedBefore := now.Add(-deletedWorkspaceRetention)
	var cursor workspacedomain.DeletedWorkspacePurgeCursor
	var totalCandidates int64
	var totalDeleted int64
	var totalBlocked int64
	batchCount := 0

	log.Info(ctx, "Purging workspaces deleted for more than 48 hours")
	for batchIndex := 0; batchIndex < maintenancePurgeMaxBatches; batchIndex++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("purge soft-deleted workspaces after %d candidates: %w", totalCandidates, err)
		}

		result, err := store.PurgeSoftDeletedWorkspacesBatch(
			ctx,
			workspacedomain.DeletedWorkspacePurgeBatch{
				DeletedBefore:               deletedBefore,
				ProcessedAt:                 now,
				Cursor:                      cursor,
				BatchSize:                   maintenancePurgeBatchSize,
				IntegrationLifecycleLockKey: slackrepository.SlackInstallationLifecycleAdvisoryKey,
			},
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("purge soft-deleted workspace batch: %w", err)
		}
		if err := validateDeletedWorkspacePurgeResult(result, maintenancePurgeBatchSize); err != nil {
			return err
		}
		if result.CandidateCount == 0 {
			recordDeletedWorkspacePurgeCompletion(
				ctx, span, log, totalCandidates, totalDeleted, totalBlocked, batchCount,
			)
			return nil
		}

		batchCount++
		totalCandidates += int64(result.CandidateCount)
		totalDeleted += result.Deleted
		totalBlocked += result.Blocked
		span.AddEvent("soft_deleted_workspace_purge_batch", trace.WithAttributes(
			attribute.Int("batch", batchCount),
			attribute.Int("candidates", result.CandidateCount),
			attribute.Int64("deleted", result.Deleted),
			attribute.Int64("blocked", result.Blocked),
		))
		cursor = result.Cursor
		if result.CandidateCount < maintenancePurgeBatchSize {
			recordDeletedWorkspacePurgeCompletion(
				ctx, span, log, totalCandidates, totalDeleted, totalBlocked, batchCount,
			)
			return nil
		}
	}

	return fmt.Errorf(
		"purge soft-deleted workspaces after %d candidates: %w",
		totalCandidates,
		errMaintenanceBacklogRemaining,
	)
}

func validateDeletedWorkspacePurgeResult(
	result workspacedomain.DeletedWorkspacePurgeResult,
	batchSize int,
) error {
	if result.CandidateCount < 0 || result.CandidateCount > batchSize {
		return fmt.Errorf(
			"purge soft-deleted workspace batch: invalid candidate count %d, want 0..%d",
			result.CandidateCount,
			batchSize,
		)
	}
	if result.Deleted < 0 || result.Blocked < 0 ||
		result.Deleted+result.Blocked != int64(result.CandidateCount) {
		return fmt.Errorf(
			"purge soft-deleted workspace batch: invalid result candidates=%d deleted=%d blocked=%d",
			result.CandidateCount,
			result.Deleted,
			result.Blocked,
		)
	}
	if result.CandidateCount > 0 &&
		(!result.Cursor.Valid || result.Cursor.DeletedAt.IsZero() || result.Cursor.WorkspaceID == uuid.Nil) {
		return errors.New("purge soft-deleted workspace batch: non-empty result requires a cursor")
	}
	return nil
}

func recordDeletedWorkspacePurgeCompletion(
	ctx context.Context,
	span trace.Span,
	log *logger.Logger,
	totalCandidates, totalDeleted, totalBlocked int64,
	batchCount int,
) {
	if totalBlocked > 0 {
		log.Warn(ctx, "Deferred workspace purge until Slack credentials are encrypted", "workspaces", totalBlocked)
	}
	span.AddEvent("workspaces_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", totalDeleted),
		attribute.Int64("candidates", totalCandidates),
		attribute.Int64("blocked", totalBlocked),
		attribute.Int("batches", batchCount),
	))
	log.Info(
		ctx,
		"Permanently deleted soft-deleted workspaces",
		"rows_affected", totalDeleted,
		"candidates", totalCandidates,
		"blocked", totalBlocked,
		"batches", batchCount,
	)
}
