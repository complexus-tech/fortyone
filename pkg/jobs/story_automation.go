package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessStoryAutoArchive archives completed stories that have been in completed status
// for longer than each team's configured auto_archive_months setting
func ProcessStoryAutoArchive(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStoryAutoArchive")
	defer span.End()

	log.Info(ctx, "Processing auto-archive for completed stories")

	startTime := time.Now()

	// Use a single set-based SQL operation to archive all eligible stories across all teams
	query := `
		UPDATE stories 
		SET 
			archived_at = CURRENT_TIMESTAMP
		FROM statuses st
		JOIN team_story_automation_settings tsas ON st.team_id = tsas.team_id
		WHERE stories.status_id = st.status_id
			AND st.category = 'completed'
			AND stories.updated_at < (CURRENT_DATE - INTERVAL '1 month' * tsas.auto_archive_months)
			AND stories.deleted_at IS NULL
			AND stories.archived_at IS NULL
			AND tsas.auto_archive_enabled = true`

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		log.Error(ctx, "Failed to auto-archive completed stories", "error", err)
		return fmt.Errorf("failed to auto-archive completed stories: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn(ctx, "Could not get rows affected count", "error", err)
		rowsAffected = -1
	}

	duration := time.Since(startTime)

	span.AddEvent("story auto-archive completed", trace.WithAttributes(
		attribute.Int64("stories.archived", rowsAffected),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Auto-archive completed stories job finished: %d stories archived in %v",
		rowsAffected, duration))

	return nil
}
