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

// InactiveWorkspaceDeleter is the worker-owned persistence capability for
// bounded, integration-safe inactivity deletion transactions.
type InactiveWorkspaceDeleter interface {
	DeleteInactiveWorkspacesBatch(
		context.Context,
		workspacedomain.InactivityDeletionBatch,
	) (workspacedomain.InactivityDeletionResult, error)
}

// ProcessWorkspaceDeletion permanently deletes workspaces that remain
// inactive after six calendar months and whose warning grace period elapsed at
// least 30 days ago.
func ProcessWorkspaceDeletion(
	ctx context.Context,
	store InactiveWorkspaceDeleter,
	log *logger.Logger,
) error {
	return processWorkspaceDeletionAt(ctx, store, log, time.Now().UTC())
}

func processWorkspaceDeletionAt(
	ctx context.Context,
	store InactiveWorkspaceDeleter,
	log *logger.Logger,
	now time.Time,
) error {
	if ctx == nil {
		return errors.New("inactive workspace deletion context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWorkspaceDeletion")
	defer span.End()
	if store == nil {
		return errors.New("inactive workspace deletion store is required")
	}
	if log == nil {
		return errors.New("inactive workspace deletion logger is required")
	}
	if now.IsZero() {
		return errors.New("inactive workspace deletion clock is required")
	}

	now = now.UTC()
	inactiveBefore := now.AddDate(0, -6, 0)
	warningSentBefore := now.Add(-30 * 24 * time.Hour)
	var cursor workspacedomain.InactivityCursor
	var totalCandidates int64
	var totalDeleted int64
	var totalBlocked int64
	batchCount := 0

	log.Info(ctx, "Permanently deleting workspaces inactive for 6+ months with 30-day warning grace period")
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("delete inactive workspaces after %d candidates: %w", totalCandidates, err)
		}

		result, err := store.DeleteInactiveWorkspacesBatch(
			ctx,
			workspacedomain.InactivityDeletionBatch{
				InactiveBefore:              inactiveBefore,
				WarningSentBefore:           warningSentBefore,
				ProcessedAt:                 now,
				Cursor:                      cursor,
				BatchSize:                   maintenancePurgeBatchSize,
				IntegrationLifecycleLockKey: slackrepository.SlackInstallationLifecycleAdvisoryKey,
			},
		)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("delete inactive workspace batch: %w", err)
		}
		if err := validateInactivityDeletionResult(result, maintenancePurgeBatchSize); err != nil {
			return err
		}
		if result.CandidateCount == 0 {
			break
		}

		batchCount++
		totalCandidates += int64(result.CandidateCount)
		totalDeleted += result.Deleted
		totalBlocked += result.Blocked
		span.AddEvent("inactive_workspace_deletion_batch", trace.WithAttributes(
			attribute.Int("batch", batchCount),
			attribute.Int("candidates", result.CandidateCount),
			attribute.Int64("deleted", result.Deleted),
			attribute.Int64("blocked", result.Blocked),
		))
		cursor = result.Cursor
		if result.CandidateCount < maintenancePurgeBatchSize {
			break
		}
	}

	if totalBlocked > 0 {
		log.Warn(
			ctx,
			"Deferred inactive workspace deletion until Slack credentials are encrypted",
			"workspaces", totalBlocked,
		)
	}
	span.AddEvent("workspaces_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", totalDeleted),
		attribute.Int64("candidates", totalCandidates),
		attribute.Int64("blocked", totalBlocked),
		attribute.Int("batches", batchCount),
	))
	log.Info(
		ctx,
		"Permanently deleted inactive workspaces",
		"rows_affected", totalDeleted,
		"candidates", totalCandidates,
		"blocked", totalBlocked,
		"batches", batchCount,
	)
	return nil
}

func validateInactivityDeletionResult(
	result workspacedomain.InactivityDeletionResult,
	batchSize int,
) error {
	if result.CandidateCount < 0 || result.CandidateCount > batchSize {
		return fmt.Errorf(
			"delete inactive workspace batch: invalid candidate count %d, want 0..%d",
			result.CandidateCount,
			batchSize,
		)
	}
	if result.Deleted < 0 || result.Blocked < 0 ||
		result.Deleted+result.Blocked != int64(result.CandidateCount) {
		return fmt.Errorf(
			"delete inactive workspace batch: invalid result candidates=%d deleted=%d blocked=%d",
			result.CandidateCount,
			result.Deleted,
			result.Blocked,
		)
	}
	if result.CandidateCount > 0 &&
		(!result.Cursor.Valid || result.Cursor.LastAccessedAt.IsZero() || result.Cursor.WorkspaceID == uuid.Nil) {
		return errors.New("delete inactive workspace batch: non-empty result requires a cursor")
	}
	return nil
}
