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

// StripeWebhookEventPurger is the worker-owned persistence capability for
// bounded terminal webhook-receipt retention.
type StripeWebhookEventPurger interface {
	PurgeTerminalStripeWebhookEvents(context.Context, time.Time, int) (int64, error)
}

// PurgeOldStripeWebhookEvents permanently deletes stripe webhook events older than 30 days
func PurgeOldStripeWebhookEvents(ctx context.Context, store StripeWebhookEventPurger, log *logger.Logger) error {
	return purgeOldStripeWebhookEventsAt(ctx, store, log, time.Now().UTC())
}

func purgeOldStripeWebhookEventsAt(
	ctx context.Context,
	store StripeWebhookEventPurger,
	log *logger.Logger,
	now time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeOldStripeWebhookEvents")
	defer span.End()
	if store == nil {
		return errors.New("stripe webhook maintenance store is required")
	}
	if log == nil {
		return errors.New("stripe webhook maintenance logger is required")
	}
	if now.IsZero() {
		return errors.New("stripe webhook maintenance clock is required")
	}
	now = now.UTC()
	terminalBefore := now.Add(-30 * 24 * time.Hour)

	log.Info(ctx, "Purging webhook events older than 30 days")
	deleted, err := drainMaintenanceBatches(ctx, "purge terminal Stripe webhook events", func(ctx context.Context, batchSize int) (int64, error) {
		return store.PurgeTerminalStripeWebhookEvents(ctx, terminalBefore, batchSize)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("webhook_events_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", deleted),
	))
	log.Info(ctx, "Permanently deleted terminal Stripe webhook events", "rows_affected", deleted)
	return nil
}
