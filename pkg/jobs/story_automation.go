package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessStoryAutoArchive archives completed and cancelled stories that have been in those statuses
// for longer than each team's configured auto_archive_months setting
func ProcessStoryAutoArchive(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStoryAutoArchive")
	defer span.End()

	log.Info(ctx, "Processing auto-archive for completed and cancelled stories")

	startTime := time.Now()

	// Use a single set-based SQL operation to archive all eligible stories across all teams
	query := `
		UPDATE stories 
		SET 
			archived_at = CURRENT_TIMESTAMP
		FROM statuses st
		JOIN team_story_automation_settings tsas ON st.team_id = tsas.team_id
		WHERE stories.status_id = st.status_id
			AND st.category IN ('completed', 'cancelled')
			AND stories.updated_at < (CURRENT_DATE - INTERVAL '1 month' * tsas.auto_archive_months)
			AND stories.deleted_at IS NULL
			AND stories.archived_at IS NULL
			AND tsas.auto_archive_enabled = true`

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		log.Error(ctx, "Failed to auto-archive completed and cancelled stories", "error", err)
		return fmt.Errorf("failed to auto-archive completed and cancelled stories: %w", err)
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

	log.Info(ctx, fmt.Sprintf("Auto-archive completed and cancelled stories job finished: %d stories archived in %v",
		rowsAffected, duration))

	return nil
}

// ProcessStoryAutoClose closes inactive unstarted and started stories by moving them to 'cancelled' status
// for longer than each team's configured auto_close_inactive_months setting
func ProcessStoryAutoClose(ctx context.Context, db *sqlx.DB, log *logger.Logger, systemUserID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStoryAutoClose")
	defer span.End()

	log.Info(ctx, "Processing auto-close for inactive unstarted and started stories")
	startTime := time.Now()

	const batchSize = 1000 // Process 1000 stories at a time
	totalProcessed := 0
	totalActivitiesRecorded := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing batch %d", batchCount))

		// Process one batch
		processed, activitiesRecorded, err := processAutoCloseBatch(ctx, db, log, systemUserID, batchSize)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to process batch %d: %w", batchCount, err)
		}

		totalProcessed += processed
		totalActivitiesRecorded += activitiesRecorded

		log.Info(ctx, fmt.Sprintf("Batch %d completed: %d stories closed, %d activities recorded",
			batchCount, processed, activitiesRecorded))

		// If we processed fewer than batch size, we're done
		if processed < batchSize {
			break
		}

		// Small delay to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)

	span.AddEvent("story auto-close completed", trace.WithAttributes(
		attribute.Int64("stories.closed", int64(totalProcessed)),
		attribute.Int64("activities.recorded", int64(totalActivitiesRecorded)),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Auto-close job completed: %d stories closed, %d activities recorded in %d batches over %v",
		totalProcessed, totalActivitiesRecorded, batchCount, duration))

	return nil
}

// ClosedStory represents a story that was closed by the auto-close process
type ClosedStory struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	TeamID      uuid.UUID `db:"team_id"`
	StatusID    uuid.UUID `db:"status_id"`
}

// processAutoCloseBatch processes a single batch of stories for auto-close
func processAutoCloseBatch(ctx context.Context, db *sqlx.DB, log *logger.Logger, systemUserID uuid.UUID, batchSize int) (int, int, error) {
	ctx, span := web.AddSpan(ctx, "jobs.processAutoCloseBatch")
	defer span.End()

	// Use UPDATE ... RETURNING to get closed stories and update in one query
	selectAndUpdateQuery := `
		WITH stories_to_close AS (
			SELECT 
				stories.id,
				stories.workspace_id,
				stories.team_id,
				stories.status_id
			FROM stories
			JOIN statuses st ON stories.status_id = st.status_id
			JOIN team_story_automation_settings tsas ON st.team_id = tsas.team_id
			WHERE st.category IN ('unstarted', 'started')
				AND stories.updated_at < (CURRENT_DATE - INTERVAL '1 month' * tsas.auto_close_inactive_months)
				AND stories.deleted_at IS NULL
				AND stories.archived_at IS NULL
				AND tsas.auto_close_inactive_enabled = true
			LIMIT :batch_size
		)
		UPDATE stories 
		SET 
			status_id = (
				SELECT status_id 
				FROM statuses 
				WHERE category = 'cancelled' 
				AND team_id = stories.team_id 
				LIMIT 1
			),
			updated_at = CURRENT_TIMESTAMP
		FROM stories_to_close
		WHERE stories.id = stories_to_close.id
		RETURNING stories.id, stories.workspace_id, stories.team_id, stories.status_id`

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the named statement for selecting and updating stories
	stmt, err := tx.PrepareNamedContext(ctx, selectAndUpdateQuery)
	if err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to prepare select and update statement: %w", err)
	}
	defer stmt.Close()

	// Execute the query to close stories and get the closed story details
	var closedStories []ClosedStory
	params := map[string]any{
		"batch_size": batchSize,
	}

	if err := stmt.SelectContext(ctx, &closedStories, params); err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to close stories batch: %w", err)
	}

	activitiesRecorded := 0
	if len(closedStories) > 0 {
		if err := recordActivitiesBatch(ctx, tx, closedStories, systemUserID, log); err != nil {
			log.Error(ctx, "Failed to record activities for batch", "error", err, "stories_count", len(closedStories))
			// Continue without failing the job - story closing is more important than activity recording
		} else {
			activitiesRecorded = len(closedStories)
		}
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to commit batch transaction: %w", err)
	}

	span.AddEvent("batch processed", trace.WithAttributes(
		attribute.Int("stories.closed", len(closedStories)),
		attribute.Int("activities.recorded", activitiesRecorded),
	))

	return len(closedStories), activitiesRecorded, nil
}

// recordActivitiesBatch bulk inserts activity records for closed stories using prepared named statements
func recordActivitiesBatch(ctx context.Context, tx *sqlx.Tx, stories []ClosedStory, systemUserID uuid.UUID, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.recordActivitiesBatch")
	defer span.End()

	if len(stories) == 0 {
		return nil
	}

	// Prepare named statement for bulk activity insertion
	activityQuery := `
		INSERT INTO story_activities (
			story_id, 
			user_id, 
			activity_type, 
			field_changed, 
			current_value, 
			workspace_id
		) VALUES (
			:story_id,
			:user_id,
			:activity_type,
			:field_changed,
			:current_value,
			:workspace_id
		)`

	stmt, err := tx.PrepareNamedContext(ctx, activityQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to prepare activity insert statement: %w", err)
	}
	defer stmt.Close()

	// Build activity records for batch insertion
	activities := make([]map[string]any, len(stories))
	for i, story := range stories {
		activities[i] = map[string]any{
			"story_id":      story.ID,
			"user_id":       systemUserID,
			"activity_type": "update",
			"field_changed": "status_id",
			"current_value": story.StatusID,
			"workspace_id":  story.WorkspaceID,
		}
	}

	// Execute batch insert using the prepared named statement
	for _, activity := range activities {
		if _, err := stmt.ExecContext(ctx, activity); err != nil {
			span.RecordError(err)
			log.Error(ctx, "Failed to insert activity record", "error", err, "story_id", activity["story_id"])
			return fmt.Errorf("failed to insert activity for story %v: %w", activity["story_id"], err)
		}
	}

	span.AddEvent("activities recorded", trace.WithAttributes(
		attribute.Int("activities.count", len(activities)),
	))

	log.Info(ctx, fmt.Sprintf("Successfully recorded %d auto-close activities", len(activities)))
	return nil
}
