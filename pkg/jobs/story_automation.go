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

// ActivityRecord represents a story activity for bulk insertion
type ActivityRecord struct {
	StoryID      uuid.UUID `db:"story_id"`
	UserID       uuid.UUID `db:"user_id"`
	ActivityType string    `db:"activity_type"`
	FieldChanged string    `db:"field_changed"`
	CurrentValue uuid.UUID `db:"current_value"`
	WorkspaceID  uuid.UUID `db:"workspace_id"`
}

// recordActivitiesBatch bulk inserts activity records for closed stories using VALUES clause
func recordActivitiesBatch(ctx context.Context, tx *sqlx.Tx, stories []ClosedStory, systemUserID uuid.UUID, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.recordActivitiesBatch")
	defer span.End()

	if len(stories) == 0 {
		return nil
	}

	// Build structured activity records (much more efficient than maps)
	activities := make([]ActivityRecord, len(stories))
	for i, story := range stories {
		activities[i] = ActivityRecord{
			StoryID:      story.ID,
			UserID:       systemUserID,
			ActivityType: "update",
			FieldChanged: "status_id",
			CurrentValue: story.StatusID,
			WorkspaceID:  story.WorkspaceID,
		}
	}

	// Use VALUES clause for true bulk insert (single database round-trip)
	activityQuery := `
		INSERT INTO story_activities (
			story_id, 
			user_id, 
			activity_type, 
			field_changed, 
			current_value, 
			workspace_id
		) VALUES (:story_id, :user_id, :activity_type, :field_changed, :current_value, :workspace_id)`

	// Execute single bulk insert with all records
	result, err := tx.NamedExecContext(ctx, activityQuery, activities)
	if err != nil {
		span.RecordError(err)
		log.Error(ctx, "Failed to bulk insert activity records", "error", err, "count", len(activities))
		return fmt.Errorf("failed to bulk insert %d activities: %w", len(activities), err)
	}

	// Verify all records were inserted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn(ctx, "Could not verify activity insert count", "error", err)
	} else if int(rowsAffected) != len(activities) {
		log.Warn(ctx, "Activity insert count mismatch", "expected", len(activities), "actual", rowsAffected)
	}

	span.AddEvent("activities recorded", trace.WithAttributes(
		attribute.Int("activities.count", len(activities)),
		attribute.Int64("rows.affected", rowsAffected),
	))

	log.Info(ctx, fmt.Sprintf("Successfully bulk inserted %d auto-close activities", len(activities)))
	return nil
}

// ProcessSprintStoryMigration moves incomplete stories from ended sprints to next available sprints
// for teams with MoveIncompleteStoriesEnabled = true
func ProcessSprintStoryMigration(ctx context.Context, db *sqlx.DB, log *logger.Logger, systemUserID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessSprintStoryMigration")
	defer span.End()

	log.Info(ctx, "Processing sprint story migration for ended sprints")
	startTime := time.Now()

	const batchSize = 500 // Process 500 sprints at a time
	totalProcessed := 0
	totalActivitiesRecorded := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing migration batch %d", batchCount))

		// Process one batch
		processed, activitiesRecorded, err := processSprintMigrationBatch(ctx, db, log, systemUserID, batchSize)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to process migration batch %d: %w", batchCount, err)
		}

		totalProcessed += processed
		totalActivitiesRecorded += activitiesRecorded

		log.Info(ctx, fmt.Sprintf("Migration batch %d completed: %d stories migrated, %d activities recorded",
			batchCount, processed, activitiesRecorded))

		// If we processed fewer than batch size, we're done
		if processed < batchSize {
			break
		}

		// Small delay to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)

	span.AddEvent("sprint story migration completed", trace.WithAttributes(
		attribute.Int64("stories.migrated", int64(totalProcessed)),
		attribute.Int64("activities.recorded", int64(totalActivitiesRecorded)),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Sprint migration job completed: %d stories migrated, %d activities recorded in %d batches over %v",
		totalProcessed, totalActivitiesRecorded, batchCount, duration))

	return nil
}

// MigratedStory represents a story that was migrated between sprints
type MigratedStory struct {
	ID               uuid.UUID `db:"id"`
	WorkspaceID      uuid.UUID `db:"workspace_id"`
	TeamID           uuid.UUID `db:"team_id"`
	PreviousSprintID uuid.UUID `db:"previous_sprint_id"`
	NewSprintID      uuid.UUID `db:"new_sprint_id"`
}

// processSprintMigrationBatch processes a single batch of ended sprints for story migration
func processSprintMigrationBatch(ctx context.Context, db *sqlx.DB, log *logger.Logger, systemUserID uuid.UUID, batchSize int) (int, int, error) {
	ctx, span := web.AddSpan(ctx, "jobs.processSprintMigrationBatch")
	defer span.End()

	// Find ended sprints with incomplete stories and migrate them in one query
	migrationQuery := `
		WITH ended_sprints AS (
			SELECT DISTINCT 
				s.sprint_id,
				s.team_id,
				s.workspace_id,
				s.end_date
			FROM sprints s
			JOIN team_sprint_settings tss ON s.team_id = tss.team_id
			WHERE s.end_date < CURRENT_DATE
				AND tss.move_incomplete_stories_enabled = true
				AND EXISTS (
					SELECT 1 
					FROM stories st 
					JOIN statuses stat ON st.status_id = stat.status_id
					WHERE st.sprint_id = s.sprint_id 
						AND stat.category IN ('unstarted', 'started')
					LIMIT 1
				)
			ORDER BY s.end_date ASC
			LIMIT :batch_size
		),
		next_sprints AS (
			SELECT DISTINCT ON (es.team_id)
				es.sprint_id as ended_sprint_id,
				es.team_id,
				es.workspace_id,
				ns.sprint_id as next_sprint_id
			FROM ended_sprints es
			LEFT JOIN sprints ns ON es.team_id = ns.team_id 
				AND ns.start_date > CURRENT_DATE
			WHERE ns.sprint_id IS NOT NULL
			ORDER BY es.team_id, ns.start_date ASC
		)
		UPDATE stories 
		SET 
			sprint_id = next_sprints.next_sprint_id,
			updated_at = CURRENT_TIMESTAMP
		FROM next_sprints
		JOIN statuses stat ON stories.status_id = stat.status_id
		WHERE stories.sprint_id = next_sprints.ended_sprint_id
			AND stat.team_id = next_sprints.team_id
			AND stat.category IN ('unstarted', 'started')
			AND stories.deleted_at IS NULL
			AND stories.archived_at IS NULL
		RETURNING 
			stories.id, 
			stories.workspace_id, 
			stories.team_id,
			next_sprints.ended_sprint_id as previous_sprint_id,
			stories.sprint_id as new_sprint_id`

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the named statement for migration
	stmt, err := tx.PrepareNamedContext(ctx, migrationQuery)
	if err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to prepare migration statement: %w", err)
	}
	defer stmt.Close()

	// Execute the query to migrate stories and get the migrated story details
	var migratedStories []MigratedStory
	params := map[string]any{
		"batch_size": batchSize,
	}

	if err := stmt.SelectContext(ctx, &migratedStories, params); err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to migrate stories batch: %w", err)
	}

	activitiesRecorded := 0
	if len(migratedStories) > 0 {
		if err := recordMigrationActivitiesBatch(ctx, tx, migratedStories, systemUserID, log); err != nil {
			log.Error(ctx, "Failed to record migration activities for batch", "error", err, "stories_count", len(migratedStories))
			// Continue without failing the job - story migration is more important than activity recording
		} else {
			activitiesRecorded = len(migratedStories)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	span.AddEvent("batch processed", trace.WithAttributes(
		attribute.Int("stories.migrated", len(migratedStories)),
		attribute.Int("activities.recorded", activitiesRecorded),
	))

	return len(migratedStories), activitiesRecorded, nil
}

// MigrationActivityRecord represents a story activity for sprint migration bulk insertion
type MigrationActivityRecord struct {
	StoryID      uuid.UUID `db:"story_id"`
	UserID       uuid.UUID `db:"user_id"`
	ActivityType string    `db:"activity_type"`
	FieldChanged string    `db:"field_changed"`
	CurrentValue uuid.UUID `db:"current_value"`
	WorkspaceID  uuid.UUID `db:"workspace_id"`
}

// recordMigrationActivitiesBatch bulk inserts activity records for migrated stories using VALUES clause
func recordMigrationActivitiesBatch(ctx context.Context, tx *sqlx.Tx, stories []MigratedStory, systemUserID uuid.UUID, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.recordMigrationActivitiesBatch")
	defer span.End()

	if len(stories) == 0 {
		return nil
	}

	// Build structured activity records (much more efficient than maps)
	activities := make([]MigrationActivityRecord, len(stories))
	for i, story := range stories {
		activities[i] = MigrationActivityRecord{
			StoryID:      story.ID,
			UserID:       systemUserID,
			ActivityType: "update",
			FieldChanged: "sprint_id",
			CurrentValue: story.NewSprintID,
			WorkspaceID:  story.WorkspaceID,
		}
	}

	// Use VALUES clause for true bulk insert (single database round-trip)
	activityQuery := `
		INSERT INTO story_activities (
			story_id, 
			user_id, 
			activity_type, 
			field_changed, 
			current_value, 
			workspace_id
		) VALUES (:story_id, :user_id, :activity_type, :field_changed, :current_value, :workspace_id)`

	// Execute single bulk insert with all records
	result, err := tx.NamedExecContext(ctx, activityQuery, activities)
	if err != nil {
		span.RecordError(err)
		log.Error(ctx, "Failed to bulk insert migration activity records", "error", err, "count", len(activities))
		return fmt.Errorf("failed to bulk insert %d migration activities: %w", len(activities), err)
	}

	// Verify all records were inserted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn(ctx, "Could not verify migration activity insert count", "error", err)
	} else if int(rowsAffected) != len(activities) {
		log.Warn(ctx, "Migration activity insert count mismatch", "expected", len(activities), "actual", rowsAffected)
	}

	span.AddEvent("migration activities recorded", trace.WithAttributes(
		attribute.Int("activities.count", len(activities)),
		attribute.Int64("rows.affected", rowsAffected),
	))

	log.Info(ctx, fmt.Sprintf("Successfully bulk inserted %d sprint migration activities", len(activities)))
	return nil
}
