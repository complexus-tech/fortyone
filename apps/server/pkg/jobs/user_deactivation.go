package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// InactiveUserDeactivator is the worker-owned persistence capability for
// bounded inactivity-policy enforcement.
type InactiveUserDeactivator interface {
	DeactivateInactiveUsers(context.Context, time.Time, time.Time, time.Time, int) (int64, error)
}

// ProcessUserDeactivation deactivates users that received inactivity warnings 30+ days ago and are still inactive
func ProcessUserDeactivation(ctx context.Context, store InactiveUserDeactivator, log *logger.Logger) error {
	return processUserDeactivationAt(ctx, store, log, time.Now().UTC())
}

func processUserDeactivationAt(
	ctx context.Context,
	store InactiveUserDeactivator,
	log *logger.Logger,
	now time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessUserDeactivation")
	defer span.End()
	if store == nil {
		return errors.New("user maintenance store is required")
	}
	if log == nil {
		return errors.New("user maintenance logger is required")
	}
	if now.IsZero() {
		return errors.New("user maintenance clock is required")
	}
	now = now.UTC()
	inactiveBefore := now.AddDate(0, -8, 0)
	warningSentBefore := now.Add(-30 * 24 * time.Hour)

	log.Info(ctx, "Deactivating users inactive for 8+ months with 30-day grace period")
	deactivated, err := drainMaintenanceBatches(ctx, "deactivate inactive users", func(ctx context.Context, batchSize int) (int64, error) {
		return store.DeactivateInactiveUsers(ctx, inactiveBefore, warningSentBefore, now, batchSize)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("users_deactivated", trace.WithAttributes(
		attribute.Int64("rows_affected", deactivated),
	))
	log.Info(ctx, "Deactivated inactive users", "rows_affected", deactivated)
	return nil
}
