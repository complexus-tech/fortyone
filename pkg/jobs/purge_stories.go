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
