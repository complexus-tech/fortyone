package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PurgeExpiredTokens permanently deletes verification tokens older than 7 days
func PurgeExpiredTokens(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeExpiredTokens")
	defer span.End()

	log.Info(ctx, "Purging verification tokens older than 7 days")

	deleteQuery := `
		DELETE FROM verification_tokens
		WHERE created_at < NOW() - INTERVAL '7 days'
	`

	result, err := db.ExecContext(ctx, deleteQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("tokens_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", rowsAffected),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d expired verification tokens", rowsAffected))
	return nil
}

// PurgeDeletedStories permanently deletes stories that have been marked as deleted for 30+ days
func PurgeDeletedStories(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedStories")
	defer span.End()
	log.Info(ctx, "Purging stories deleted for more than 30 days")

	deleteQuery := `
		DELETE FROM stories
		WHERE deleted_at IS NOT NULL
		AND deleted_at < NOW() - INTERVAL '30 days'
	`
	result, err := db.ExecContext(ctx, deleteQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete stories: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("stories_deleted", trace.WithAttributes(
		attribute.Int("rows_affected", int(rowsAffected)),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d stories", rowsAffected))
	return nil
}

// PurgeOldStripeWebhookEvents permanently deletes stripe webhook events older than 30 days
func PurgeOldStripeWebhookEvents(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeOldStripeWebhookEvents")
	defer span.End()

	log.Info(ctx, "Purging webhook events older than 30 days")

	deleteQuery := `
		DELETE FROM stripe_webhook_events
		WHERE processed_at < NOW() - INTERVAL '30 days'
	`

	result, err := db.ExecContext(ctx, deleteQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete old webhook events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("webhook_events_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", rowsAffected),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d old webhook events", rowsAffected))
	return nil
}
