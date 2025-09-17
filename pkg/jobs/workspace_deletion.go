package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessWorkspaceDeletion permanently deletes workspaces that received inactivity warnings 30+ days ago and are still inactive
func ProcessWorkspaceDeletion(ctx context.Context, db *sqlx.DB, log *logger.Logger, systemUserID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWorkspaceDeletion")
	defer span.End()

	log.Info(ctx, "Permanently deleting workspaces inactive for 6+ months with 30-day grace period")

	deleteQuery := `
		DELETE FROM workspaces
		WHERE last_accessed_at < NOW() - INTERVAL '6 months 30 days'
		AND deleted_at IS NULL
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
