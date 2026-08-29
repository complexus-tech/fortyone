package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	maintenancePurgeBatchSize  = 500
	maintenancePurgeMaxBatches = 100
)

var (
	errInvalidMaintenancePurgeResult = errors.New("maintenance purge returned an invalid row count")
	errMaintenanceBacklogRemaining   = errors.New("maintenance purge backlog remains")
)

// ChatSessionPurger is the worker-owned persistence capability for retiring
// expired soft-deleted chat sessions in bounded batches.
type ChatSessionPurger interface {
	PurgeDeletedChatSessions(context.Context, time.Time, int) (int64, error)
}

// PurgeDeletedChatSessions permanently deletes chat sessions marked as deleted
// for 30+ days while retaining any session that still anchors an unresolved
// mutation approval quarantine.
func PurgeDeletedChatSessions(ctx context.Context, store ChatSessionPurger, log *logger.Logger) error {
	return purgeDeletedChatSessionsAt(ctx, store, log, time.Now().UTC())
}

func purgeDeletedChatSessionsAt(
	ctx context.Context,
	store ChatSessionPurger,
	log *logger.Logger,
	now time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedChatSessions")
	defer span.End()
	if store == nil {
		return errors.New("chat session maintenance store is required")
	}
	if log == nil {
		return errors.New("chat session maintenance logger is required")
	}
	if now.IsZero() {
		return errors.New("chat session maintenance clock is required")
	}
	now = now.UTC()
	deletedBefore := now.Add(-30 * 24 * time.Hour)

	log.Info(ctx, "Purging chat sessions deleted for more than 30 days")
	deleted, err := drainMaintenanceBatches(ctx, "purge deleted chat sessions", func(ctx context.Context, batchSize int) (int64, error) {
		return store.PurgeDeletedChatSessions(ctx, deletedBefore, batchSize)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("chat_sessions_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", deleted),
	))
	log.Info(ctx, "Permanently deleted chat sessions", "rows_affected", deleted)
	return nil
}

func drainMaintenanceBatches(
	ctx context.Context,
	operation string,
	purge func(context.Context, int) (int64, error),
) (int64, error) {
	if ctx == nil {
		return 0, errors.New("maintenance purge context is required")
	}
	if purge == nil {
		return 0, errors.New("maintenance purge operation is required")
	}

	var total int64
	for batch := 0; batch < maintenancePurgeMaxBatches; batch++ {
		if err := ctx.Err(); err != nil {
			return total, fmt.Errorf("%s interrupted after %d rows: %w", operation, total, err)
		}

		deleted, err := purge(ctx, maintenancePurgeBatchSize)
		if err != nil {
			return total, fmt.Errorf("%s: %w", operation, err)
		}
		if deleted < 0 || deleted > int64(maintenancePurgeBatchSize) {
			return total, fmt.Errorf(
				"%s: %w: got %d, want 0..%d",
				operation,
				errInvalidMaintenancePurgeResult,
				deleted,
				maintenancePurgeBatchSize,
			)
		}
		total += deleted
		if deleted < int64(maintenancePurgeBatchSize) {
			return total, nil
		}
	}

	return total, fmt.Errorf(
		"%s after %d rows: %w",
		operation,
		total,
		errMaintenanceBacklogRemaining,
	)
}
