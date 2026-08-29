package taskhandlers

import (
	"context"
	"errors"
	"fmt"

	platformidempotency "github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

const idempotencyBacklogWarningInterval = 100

var errInvalidIdempotencyPurgeResult = errors.New("idempotency purge returned an invalid row count")

// IdempotencyReceiptPurger is the narrow worker port used to retire expired
// API receipts. Implementations must delete at most batchSize rows atomically.
type IdempotencyReceiptPurger interface {
	PurgeExpired(context.Context, int) (int64, error)
}

// IdempotencyCleanupHandler drains bounded database batches until it observes
// a short batch. The Asynq timeout remains the outer runtime bound, so a large
// backlog is retried without one database statement or lock growing unbounded.
type IdempotencyCleanupHandler struct {
	log      *logger.Logger
	receipts IdempotencyReceiptPurger
}

func NewIdempotencyCleanupHandler(
	log *logger.Logger,
	receipts IdempotencyReceiptPurger,
) *IdempotencyCleanupHandler {
	return &IdempotencyCleanupHandler{log: log, receipts: receipts}
}

func (h *IdempotencyCleanupHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	if h == nil || h.log == nil || h.receipts == nil {
		return errors.New("idempotency cleanup dependencies are required")
	}

	var total int64
	for batch := 1; ; batch++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("idempotency receipt cleanup interrupted after %d rows: %w", total, err)
		}

		deleted, err := h.receipts.PurgeExpired(ctx, platformidempotency.MaxPurgeBatchSize)
		if err != nil {
			return fmt.Errorf("purge expired idempotency receipts: %w", err)
		}
		if deleted < 0 || deleted > platformidempotency.MaxPurgeBatchSize {
			return fmt.Errorf("%w: got %d", errInvalidIdempotencyPurgeResult, deleted)
		}
		total += deleted

		if batch%idempotencyBacklogWarningInterval == 0 {
			h.log.Warn(ctx, "API idempotency receipt cleanup is draining a large backlog", "batches", batch, "deleted", total)
		}
		if deleted < platformidempotency.MaxPurgeBatchSize {
			h.log.Info(ctx, "Expired API idempotency receipts purged", "batches", batch, "deleted", total)
			return nil
		}
	}
}
