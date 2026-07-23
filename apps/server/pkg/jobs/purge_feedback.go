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

// PurgeDeletedFeedback permanently deletes feedback after its 30-day recovery window.
func PurgeDeletedFeedback(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedFeedback")
	defer span.End()
	log.Info(ctx, "Purging feedback deleted for more than 30 days")

	result, err := db.ExecContext(ctx, `
		DELETE FROM feedback_items
		WHERE deleted_at IS NOT NULL
			AND deleted_at < NOW() - INTERVAL '30 days'
	`)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete feedback: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("feedback_deleted", trace.WithAttributes(
		attribute.Int("rows_affected", int(rowsAffected)),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d feedback items", rowsAffected))
	return nil
}
