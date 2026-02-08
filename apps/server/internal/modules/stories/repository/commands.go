package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error) {
	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := `
		INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence) 
		VALUES (:workspace_id, :team_id, 0) 
		ON CONFLICT (workspace_id, team_id) 
		DO UPDATE SET current_sequence = team_story_sequences.current_sequence + 1 
		RETURNING current_sequence
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceId,
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return 0, nil, nil, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var currentSequence int
	err = stmt.GetContext(ctx, &currentSequence, params)
	if err != nil {
		tx.Rollback()
		return 0, nil, nil, fmt.Errorf("failed to get/update sequence: %w", err)
	}

	commit := func() error {
		return tx.Commit()
	}

	rollback := func() error {
		return tx.Rollback()
	}

	return currentSequence, commit, rollback, nil
}

// Create creates a new story with automatic sequence recovery on conflicts.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	// Validate status belongs to the same team
	if story.Status != nil {
		if err := r.validateStatusTeam(ctx, *story.Status, story.Team); err != nil {
			return stories.CoreSingleStory{}, err
		}
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastSequence, commit, rollback, err := r.GetNextSequenceID(ctx, story.Team, story.Workspace)
		if err != nil {
			return stories.CoreSingleStory{}, fmt.Errorf("failed to get next sequence ID: %w", err)
		}
		story.SequenceID = lastSequence + 1

		cs, err := r.insertStory(ctx, story)
		if err != nil {
			rollback()

			// Check if this is a duplicate sequence ID error
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
				strings.Contains(err.Error(), "unique_team_sequence") {
				r.log.Info(ctx, "sequence out of sync, retrying with corrected sequence",
					"attempt", attempt,
					"team_id", story.Team,
					"tried_sequence", story.SequenceID)

				// Sync the sequence to the correct value
				if syncErr := r.syncSequence(ctx, story.Team, story.Workspace); syncErr != nil {
					r.log.Error(ctx, "failed to sync sequence", "error", syncErr)
					return stories.CoreSingleStory{}, fmt.Errorf("failed to sync sequence: %w", syncErr)
				}

				lastErr = err
				continue // Retry
			}

			// Different error, return immediately
			return stories.CoreSingleStory{}, fmt.Errorf("failed to insert story: %w", err)
		}

		if err := commit(); err != nil {
			return stories.CoreSingleStory{}, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return toCoreStory(cs), nil
	}

	// Exhausted retries
	return stories.CoreSingleStory{}, fmt.Errorf("failed to create story after %d retries: %w", maxRetries, lastErr)
}

// syncSequence syncs the team_story_sequences table with the actual max sequence_id in the stories table.
func (r *repo) syncSequence(ctx context.Context, teamID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.syncSequence")
	defer span.End()

	// Get the actual max sequence_id from the stories table
	maxSeqQuery := `
		SELECT COALESCE(MAX(sequence_id), 0) 
		FROM stories 
		WHERE team_id = :team_id AND workspace_id = :workspace_id
	`
	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	var maxSequence int
	stmt, err := r.db.PrepareNamedContext(ctx, maxSeqQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare max sequence query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &maxSequence, params); err != nil {
		return fmt.Errorf("failed to get max sequence: %w", err)
	}

	// Update the team_story_sequences table
	updateQuery := `
		UPDATE team_story_sequences 
		SET current_sequence = :max_sequence 
		WHERE team_id = :team_id AND workspace_id = :workspace_id
	`
	updateParams := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
		"max_sequence": maxSequence,
	}

	updateStmt, err := r.db.PrepareNamedContext(ctx, updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare update sequence query: %w", err)
	}
	defer updateStmt.Close()

	if _, err := updateStmt.ExecContext(ctx, updateParams); err != nil {
		return fmt.Errorf("failed to update sequence: %w", err)
	}

	r.log.Info(ctx, "sequence synced successfully",
		"team_id", teamID,
		"workspace_id", workspaceID,
		"new_sequence", maxSequence)

	return nil
}

func (r *repo) insertStory(ctx context.Context, story *stories.CoreSingleStory) (dbStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.insertStory")
	defer span.End()

	q := `
			INSERT INTO stories (
					sequence_id, title, description, description_html,
					parent_id, objective_id, status_id, assignee_id, 
					blocked_by_id, blocking_id, related_id, reporter_id,
					priority, sprint_id, key_result_id, team_id, workspace_id, start_date, 
					end_date, created_at, updated_at
			) VALUES (
					:sequence_id, :title, :description, :description_html,
					:parent_id, :objective_id, :status_id, :assignee_id, :blocked_by_id,
					:blocking_id, :related_id, :reporter_id, :priority, :sprint_id,
					:key_result_id, :team_id, :workspace_id, :start_date, :end_date, :created_at, :updated_at
			) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.sprint_id, stories.key_result_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
		`

	var cs dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbStory{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "creating story.")
	if err := stmt.GetContext(ctx, &cs, toDBStory(*story)); err != nil {
		errMsg := fmt.Sprintf("failed to create story: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create story"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbStory{}, err
	}

	r.log.Info(ctx, "Story created successfully.")
	span.AddEvent("Story created.", trace.WithAttributes(
		attribute.String("story.title", story.Title),
	))

	return cs, err
}

func (r *repo) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateLabels")
	defer span.End()

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First, delete all existing labels for the story
	deleteQuery := `
		DELETE FROM story_labels 
		WHERE story_id = :story_id 
		AND story_id IN (
			SELECT id FROM stories 
			WHERE id = :story_id 
			AND workspace_id = :workspace_id
		)
	`
	params := map[string]any{
		"story_id":     id,
		"workspace_id": workspaceId,
	}

	if _, err = tx.NamedExecContext(ctx, deleteQuery, params); err != nil {
		return fmt.Errorf("failed to delete existing labels: %w", err)
	}

	// If we have new labels to insert
	if len(labels) > 0 {
		// Prepare values for bulk insert
		values := make([]string, len(labels))
		args := make([]any, 0, len(labels)*2)
		for i, labelID := range labels {
			values[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			args = append(args, id, labelID)
		}

		// Insert new labels
		insertQuery := fmt.Sprintf(`
			INSERT INTO story_labels (story_id, label_id)
			VALUES %s
		`, strings.Join(values, ","))

		if _, err = tx.ExecContext(ctx, insertQuery, args...); err != nil {
			return fmt.Errorf("failed to insert new labels: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns a list of stories for a workspace with additional filters.

// MyStories returns a list of stories.

func (r *repo) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Delete")
	defer span.End()
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	stmt, err := r.db.PrepareNamedContext(ctx, `
		UPDATE stories 
		SET deleted_at = NOW(),
				updated_at = NOW() 
		WHERE id = :id
		AND workspace_id = :workspace_id;
	`)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting story #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s deleted successfully", id), "id", id)
	span.AddEvent("Story deleted.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// List returns a list of stories for a workspace with additional filters.

func (r *repo) BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkDelete")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
        UPDATE stories
        SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = ANY(:ids) AND workspace_id = :workspace_id;
    `

	r.log.Info(ctx, fmt.Sprintf("Deleting stories: %v", ids), "ids", ids)
	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v deleted successfully", ids), "ids", ids)
	span.AddEvent("Stories deleted.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// HardBulkDelete performs permanent removal of the stories with the specified IDs.
func (r *repo) HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.HardBulkDelete")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}
	query := `
		DELETE FROM stories
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id;
	`

	r.log.Info(ctx, fmt.Sprintf("Hard deleting stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to hard delete stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to hard delete", "ids", ids)
		return fmt.Errorf("no stories found to delete")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories hard deleted successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories hard deleted.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// Restore restores a story with the specified ID.
func (r *repo) Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Restore")
	defer span.End()

	query := `
			UPDATE stories 
			SET deleted_at = NULL, 
					updated_at = NOW() 
			WHERE id = :id
			AND workspace_id = :workspace_id;
	`
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare restore statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("restoring story #%s", id), "id", id)
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to restore story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("story #%s restored successfully", id), "id", id)
	span.AddEvent("story restored.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (r *repo) BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkRestore")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
				UPDATE stories
				SET deleted_at = NULL, updated_at = NOW()
				WHERE id = ANY(:ids)
				AND workspace_id = :workspace_id;
			`

	r.log.Info(ctx, fmt.Sprintf("restoring stories: %v", ids), "ids", ids)
	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to restore stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("stories: %v restored successfully", ids), "ids", ids)
	span.AddEvent("stories restored.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// BulkUnarchive unarchives the stories with the specified IDs.
func (r *repo) BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUnarchive")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
		UPDATE stories
		SET archived_at = NULL, updated_at = NOW()
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id
			AND archived_at IS NOT NULL;
	`

	r.log.Info(ctx, fmt.Sprintf("Bulk unarchiving stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to bulk unarchive stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to unarchive", "ids", ids)
		return fmt.Errorf("no stories found to unarchive")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories unarchived successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories unarchived.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// BulkArchive archives the stories with the specified IDs.
func (r *repo) BulkArchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkArchive")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
		UPDATE stories
		SET archived_at = NOW()
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id
			AND archived_at IS NULL;
	`

	r.log.Info(ctx, fmt.Sprintf("Bulk archiving stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to bulk archive stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to archive", "ids", ids)
		return fmt.Errorf("no stories found to archive")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories archived successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories archived.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// Update updates the story with the specified ID.
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	r.log.Info(ctx, "business.repository.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Update")
	defer span.End()

	// If status is being updated, validate it belongs to the same team
	if statusId, ok := updates["status_id"].(uuid.UUID); ok {
		// We need to get the story's team ID first
		var teamId uuid.UUID
		q := `SELECT team_id FROM stories WHERE id = :story_id AND workspace_id = :workspace_id`
		params := map[string]any{
			"story_id":     id,
			"workspace_id": workspaceId,
		}
		stmt, err := r.db.PrepareNamedContext(ctx, q)
		if err != nil {
			errMsg := fmt.Sprintf("failed to prepare team query statement: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
		defer stmt.Close()

		if err := stmt.GetContext(ctx, &teamId, params); err != nil {
			errMsg := fmt.Sprintf("failed to get story team: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}

		if err := r.validateStatusTeam(ctx, statusId, teamId); err != nil {
			return err
		}
	}

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named update statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating story #%s", id), "id", id)
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to update story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s updated successfully", id), "id", id)
	span.AddEvent("Story updated.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkUpdate updates the stories with the specified IDs.
func (r *repo) BulkUpdate(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUpdate")
	defer span.End()

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id IN (:ids) AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named bulk update statement: %s", err), "ids", ids)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating stories: %v", ids), "ids", ids)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to update stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v updated successfully", ids), "ids", ids)
	span.AddEvent("Stories updated.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

func (r *repo) RecordActivities(ctx context.Context, activities []stories.CoreActivity) ([]stories.CoreActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.RecordActivities")
	defer span.End()

	dbActivities, err := r.recordActivities(ctx, activities)
	if err != nil {
		return nil, fmt.Errorf("failed to insert activities: %w", err)
	}

	return toCoreActivities(dbActivities), nil
}

func (r *repo) recordActivities(ctx context.Context, activities []stories.CoreActivity) ([]dbActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.recordActivity")
	defer span.End()

	// Prepare the base query for bulk insert
	q := `
		INSERT INTO story_activities (
			story_id, 
			activity_type, 
			field_changed, 
			current_value,
			old_value,
			new_value,
			user_id,
			workspace_id
		)
		VALUES (
			:story_id, 
			:activity_type, 
			:field_changed, 
			:current_value,
			:old_value,
			:new_value,
			:user_id,
			:workspace_id
		)
		RETURNING story_activities.*;
	`

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the statement
	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	// Insert each activity and collect results
	var result []dbActivity
	for _, activity := range activities {
		var da dbActivity
		if err := stmt.GetContext(ctx, &da, toDBActivity(activity)); err != nil {
			errMsg := fmt.Sprintf("Failed to insert activity: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to insert activity"), trace.WithAttributes(attribute.String("error", errMsg)))
			return nil, err
		}
		result = append(result, da)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, fmt.Sprintf("Successfully created %d activities", len(activities)))
	span.AddEvent("Activities created.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
	))

	return result, nil
}

// GetActivitiesWithUser returns activities for a given story ID with user details and pagination.

func (r *repo) CreateComment(ctx context.Context, cnc stories.CoreNewComment) (comments.CoreComment, error) {
	r.log.Info(ctx, "business.repository.stories.CreateComment")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CreateComment")
	defer span.End()

	q := `
	INSERT INTO story_comments (
		content, story_id, commenter_id, parent_id
	) VALUES (
		:content, :story_id, :commenter_id, :parent_id
	) RETURNING story_comments.*;
`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return comments.CoreComment{}, err
	}
	defer stmt.Close()

	var comment commentsrepository.DbComment
	if err := stmt.GetContext(ctx, &comment, toDBNewComment(cnc)); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to insert comment: %s", err))
		return comments.CoreComment{}, err
	}

	return toCoreComment(comment), nil
}

func (r *repo) validateStatusTeam(ctx context.Context, statusId, teamId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.validateStatusTeam")
	defer span.End()

	q := `
		SELECT EXISTS (
			SELECT 1 FROM statuses 
			WHERE status_id = :status_id 
			AND team_id = :team_id
		)
	`
	params := map[string]any{
		"status_id": statusId,
		"team_id":   teamId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare validation statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		errMsg := fmt.Sprintf("failed to validate status team: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if !exists {
		return errors.New("status does not belong to the story's team")
	}

	return nil
}

// DuplicateStory creates a copy of an existing story with a new sequence ID
func (r *repo) DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.DuplicateStory")
	defer span.End()

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get the original story
	originalStory, err := r.getStoryById(ctx, originalStoryID, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get original story: %w", err)
	}

	// Get new sequence ID
	lastSequence, commit, rollback, err := r.GetNextSequenceID(ctx, originalStory.Team, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get next sequence ID: %w", err)
	}
	defer rollback()

	// Prepare the insert query for the new story
	q := `
		INSERT INTO stories (
			sequence_id,
			title,
			description,
			description_html,
			team_id,
			objective_id,
			status_id,
			assignee_id,
			priority,
			sprint_id,
			workspace_id,
			reporter_id,
			created_at,
			updated_at
		) VALUES (
			:sequence_id,
			:title,
			:description,
			:description_html,
			:team_id,
			:objective_id,
			:status_id,
			:assignee_id,
			:priority,
			:sprint_id,
			:workspace_id,
			:reporter_id,
			NOW(),
			NOW()
		) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.sprint_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
	`

	// Prepare parameters for the new story
	params := map[string]any{
		"sequence_id":      lastSequence + 1,
		"title":            "Copy of " + originalStory.Title,
		"description":      originalStory.Description,
		"description_html": originalStory.DescriptionHTML,
		"team_id":          originalStory.Team,
		"objective_id":     originalStory.Objective,
		"status_id":        originalStory.Status,
		"assignee_id":      originalStory.Assignee,
		"priority":         originalStory.Priority,
		"sprint_id":        originalStory.Sprint,
		"workspace_id":     workspaceId,
		"reporter_id":      userID,
	}

	// Execute the insert
	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var newStory dbStory
	if err := stmt.GetContext(ctx, &newStory, params); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to create duplicate story: %w", err)
	}

	// Commit the sequence ID transaction
	if err := commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit sequence ID transaction: %w", err)
	}

	// Commit the main transaction
	if err := tx.Commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, fmt.Sprintf("Successfully duplicated story #%s", originalStoryID))
	span.AddEvent("Story duplicated.", trace.WithAttributes(
		attribute.String("original_story.id", originalStoryID.String()),
		attribute.String("new_story.id", newStory.ID.String()),
	))

	return toCoreStory(newStory), nil
}

// CountStoriesInWorkspace returns the count of stories in a workspace.

func (r *repo) CreateStoryFromIssue(ctx context.Context, workspaceID, teamID uuid.UUID, title, description string, reporterID uuid.UUID) (uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CreateStoryFromIssue")
	defer span.End()

	// 1) Get default status for team (category 'unstarted')
	var statusID uuid.UUID
	statusQuery := `SELECT status_id FROM statuses WHERE team_id = :team_id AND category = 'unstarted' LIMIT 1`
	stmtStatus, err := r.db.PrepareNamedContext(ctx, statusQuery)
	if err != nil {
		return uuid.Nil, fmt.Errorf("prepare status query: %w", err)
	}
	if err := stmtStatus.GetContext(ctx, &statusID, map[string]any{"team_id": teamID}); err != nil {
		stmtStatus.Close()
		return uuid.Nil, fmt.Errorf("fetch default status: %w", err)
	}
	stmtStatus.Close()

	// 2) Get next sequence id
	sequenceID, commit, rollback, err := r.GetNextSequenceID(ctx, teamID, workspaceID)
	if err != nil {
		return uuid.Nil, err
	}

	// 3) Insert story
	insertQuery := `
		INSERT INTO stories (
			sequence_id, title, description, description_html, status_id, priority, team_id, workspace_id, reporter_id, created_at, updated_at
		) VALUES (
			:sequence_id, :title, :description, :description_html, :status_id, :priority, :team_id, :workspace_id, :reporter_id, NOW(), NOW()
		) RETURNING id`

	params := map[string]any{
		"sequence_id":      sequenceID + 1,
		"title":            title,
		"description":      description,
		"description_html": description,
		"status_id":        statusID,
		"team_id":          teamID,
		"workspace_id":     workspaceID,
		"reporter_id":      reporterID,
		"priority":         "No Priority",
	}

	stmt, err := r.db.PrepareNamedContext(ctx, insertQuery)
	if err != nil {
		rollback()
		return uuid.Nil, fmt.Errorf("prepare insert story: %w", err)
	}
	var storyID uuid.UUID
	if err := stmt.GetContext(ctx, &storyID, params); err != nil {
		stmt.Close()
		rollback()
		return uuid.Nil, fmt.Errorf("insert story: %w", err)
	}
	stmt.Close()

	if err := commit(); err != nil {
		return uuid.Nil, err
	}
	return storyID, nil
}

// UpdateStoryStatus updates only the status of a story - used for automated transitions
func (r *repo) UpdateStoryStatus(ctx context.Context, storyID uuid.UUID, workspaceID uuid.UUID, statusID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateStoryStatus")
	defer span.End()

	updates := map[string]any{
		"status_id": statusID,
	}

	return r.Update(ctx, storyID, workspaceID, updates)
}

// GetStatusCategory returns the category for a given status ID

func (r *repo) AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error) {
	query := `
		INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id)
		VALUES (:from_story_id, :to_story_id, :association_type, :workspace_id)
		RETURNING id
	`

	params := map[string]any{
		"from_story_id":    fromID,
		"to_story_id":      toID,
		"association_type": associationType,
		"workspace_id":     workspaceID,
	}

	var id uuid.UUID
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to prepare insert association: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &id, params); err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to insert association: %w", err)
	}

	// Fetch the 'to' story details to return complete object
	// We use the full getStoryById to ensure we get a consistent view of the story including its own associations/substories if needed,
	// limiting potential inconsistency.
	toStory, err := r.getStoryById(ctx, toID, workspaceID)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to fetch target story details: %w", err)
	}

	coreToStory := toCoreStory(toStory)

	return stories.CoreStoryAssociation{
		ID:          id,
		FromStoryID: fromID,
		ToStoryID:   toID,
		Type:        associationType,
		Story: stories.CoreStoryList{
			ID:          coreToStory.ID,
			SequenceID:  coreToStory.SequenceID,
			Title:       coreToStory.Title,
			Status:      coreToStory.Status,
			Priority:    coreToStory.Priority,
			Assignee:    coreToStory.Assignee,
			Reporter:    coreToStory.Reporter,
			Workspace:   coreToStory.Workspace,
			Team:        coreToStory.Team,
			CreatedAt:   coreToStory.CreatedAt,
			UpdatedAt:   coreToStory.UpdatedAt,
			CompletedAt: coreToStory.CompletedAt,
			DeletedAt:   coreToStory.DeletedAt,
			ArchivedAt:  coreToStory.ArchivedAt,
			Labels:      coreToStory.Labels,
			SubStories:  toCoreStoryList(coreToStory.SubStories),
		},
	}, nil
}

// RemoveAssociation removes an association between two stories.
func (r *repo) RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) error {
	query := `
		DELETE FROM story_associations
		WHERE id = :id AND workspace_id = :workspace_id
	`

	params := map[string]any{
		"id":           associationID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare delete association: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		return fmt.Errorf("failed to delete association: %w", err)
	}

	return nil
}

// Helper to convert []CoreStoryList (from CoreSingleStory.SubStories) back to []CoreStoryList
// This seems redundant but CoreSingleStory uses []CoreStoryList for SubStories.
func toCoreStoryList(stories []stories.CoreStoryList) []stories.CoreStoryList {
	return stories
}
