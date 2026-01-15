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

// PurgeDeletedWorkspaces permanently deletes workspaces that have been marked as deleted for 48+ hours
func PurgeDeletedWorkspaces(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedWorkspaces")
	defer span.End()
	log.Info(ctx, "Purging workspaces deleted for more than 48 hours")

	deleteQuery := `
		DELETE FROM workspaces
		WHERE deleted_at IS NOT NULL
		AND deleted_at < NOW() - INTERVAL '48 hours'
	`
	result, err := db.ExecContext(ctx, deleteQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete workspaces: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("workspaces_deleted", trace.WithAttributes(
		attribute.Int("rows_affected", int(rowsAffected)),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d workspaces", rowsAffected))
	return nil
}
