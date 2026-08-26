package storiesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const activityCompactionWindowSQL = "30 seconds"

func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	currentSequence, err := r.getNextSequenceID(ctx, tx, teamID, workspaceId)
	if err != nil {
		_ = tx.Rollback()
		return 0, nil, nil, err
	}

	return currentSequence, tx.Commit, tx.Rollback, nil
}

func (r *repo) getNextSequenceID(ctx context.Context, tx *sqlx.Tx, teamID, workspaceID uuid.UUID) (int, error) {
	query := `
		INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence) 
		VALUES (:workspace_id, :team_id, 0) 
		ON CONFLICT (workspace_id, team_id) 
		DO UPDATE SET current_sequence = team_story_sequences.current_sequence + 1 
		RETURNING current_sequence
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var currentSequence int
	if err := stmt.GetContext(ctx, &currentSequence, params); err != nil {
		return 0, fmt.Errorf("failed to get/update sequence: %w", err)
	}

	return currentSequence, nil
}

type storyCreateTransaction interface {
	NextSequence(ctx context.Context, teamID, workspaceID uuid.UUID) (int, error)
	InsertStory(ctx context.Context, story *stories.CoreSingleStory) (dbStory, error)
	InsertLabels(ctx context.Context, storyID, workspaceID, teamID uuid.UUID, labelIDs []uuid.UUID) error
	Commit() error
	Rollback() error
}

type sqlStoryCreateTransaction struct {
	repo *repo
	tx   *sqlx.Tx
}

func (tx *sqlStoryCreateTransaction) NextSequence(ctx context.Context, teamID, workspaceID uuid.UUID) (int, error) {
	return tx.repo.getNextSequenceID(ctx, tx.tx, teamID, workspaceID)
}

func (tx *sqlStoryCreateTransaction) InsertStory(ctx context.Context, story *stories.CoreSingleStory) (dbStory, error) {
	return tx.repo.insertStory(ctx, tx.tx, story)
}

func (tx *sqlStoryCreateTransaction) InsertLabels(
	ctx context.Context,
	storyID, workspaceID, teamID uuid.UUID,
	labelIDs []uuid.UUID,
) error {
	return tx.repo.insertStoryLabels(ctx, tx.tx, storyID, workspaceID, teamID, labelIDs)
}

func (tx *sqlStoryCreateTransaction) Commit() error {
	return tx.tx.Commit()
}

func (tx *sqlStoryCreateTransaction) Rollback() error {
	return tx.tx.Rollback()
}

// Create creates a new story with automatic sequence recovery on conflicts.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) (stories.CoreSingleStory, error) {
	created, _, err := r.CreateIdempotent(ctx, story)
	return created, err
}

// CreateIdempotent creates a story once when ExternalCreationKey is set and
// reports whether this call performed the insert. A concurrent retry blocks on
// the unique key and then returns the already-committed story.
func (r *repo) CreateIdempotent(ctx context.Context, story *stories.CoreSingleStory) (stories.CoreSingleStory, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	// Validate status belongs to the same team
	if story.Status != nil {
		if err := r.validateStatusTeam(ctx, *story.Status, story.Team); err != nil {
			return stories.CoreSingleStory{}, false, err
		}
	}

	const maxRetries = 3
	var lastErr error
	labelIDs := deduplicateLabelIDs(story.Labels)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cs, err := r.createStoryAttempt(ctx, story, labelIDs)
		if err != nil {
			if isExternalCreationKeyConflict(err) && story.CreationKey != nil {
				existing, lookupErr := r.findByExternalCreationKey(ctx, story.Workspace, *story.CreationKey)
				if lookupErr != nil {
					return stories.CoreSingleStory{}, false, fmt.Errorf("load idempotent story: %w", lookupErr)
				}
				return existing, false, nil
			}
			if isStorySequenceConflict(err) {
				r.log.Info(ctx, "sequence out of sync, retrying with corrected sequence",
					"attempt", attempt,
					"team_id", story.Team,
					"tried_sequence", story.SequenceID)

				// Sync the sequence to the correct value
				if syncErr := r.syncSequence(ctx, story.Team, story.Workspace); syncErr != nil {
					r.log.Error(ctx, "failed to sync sequence", "error", syncErr)
					return stories.CoreSingleStory{}, false, fmt.Errorf("failed to sync sequence: %w", syncErr)
				}

				lastErr = err
				continue // Retry
			}

			return stories.CoreSingleStory{}, false, fmt.Errorf("failed to create story: %w", err)
		}

		createdStory := toCoreStory(cs)
		createdStory.Labels = labelIDs
		createdStory.CreatedNow = true
		return createdStory, true, nil
	}

	// Exhausted retries
	return stories.CoreSingleStory{}, false, fmt.Errorf("failed to create story after %d retries: %w", maxRetries, lastErr)
}

func (r *repo) createStoryAttempt(
	ctx context.Context,
	story *stories.CoreSingleStory,
	labelIDs []uuid.UUID,
) (dbStory, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return dbStory{}, fmt.Errorf("begin story creation transaction: %w", err)
	}

	return executeStoryCreateTransaction(ctx, &sqlStoryCreateTransaction{repo: r, tx: tx}, story, labelIDs)
}

func executeStoryCreateTransaction(
	ctx context.Context,
	tx storyCreateTransaction,
	story *stories.CoreSingleStory,
	labelIDs []uuid.UUID,
) (created dbStory, err error) {
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	lastSequence, err := tx.NextSequence(ctx, story.Team, story.Workspace)
	if err != nil {
		return dbStory{}, fmt.Errorf("advance story sequence: %w", err)
	}
	story.SequenceID = lastSequence + 1

	created, err = tx.InsertStory(ctx, story)
	if err != nil {
		return dbStory{}, fmt.Errorf("insert story: %w", err)
	}
	if err := tx.InsertLabels(ctx, created.ID, story.Workspace, story.Team, labelIDs); err != nil {
		return dbStory{}, fmt.Errorf("insert story labels: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return dbStory{}, fmt.Errorf("commit story creation transaction: %w", err)
	}
	committed = true

	return created, nil
}

func isStorySequenceConflict(err error) bool {
	if err == nil {
		return false
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code == "23505" && postgresError.ConstraintName == "unique_team_sequence"
	}

	return strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
		strings.Contains(err.Error(), "unique_team_sequence")
}

func isExternalCreationKeyConflict(err error) bool {
	if err == nil {
		return false
	}
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == "stories_external_creation_key_key"
}

func (r *repo) findByExternalCreationKey(ctx context.Context, workspaceID uuid.UUID, key string) (stories.CoreSingleStory, error) {
	var row dbStory
	err := r.db.GetContext(ctx, &row, `
		SELECT s.id, s.sequence_id, s.title, s.description, s.description_html,
		       s.parent_id, s.objective_id, s.status_id, s.assignee_id,
		       s.blocked_by_id, s.blocking_id, s.related_id, s.reporter_id,
		       s.priority, s.estimate_unit, s.estimated_duration_minutes,
		       s.minimum_focus_block_minutes, s.auto_scheduling_enabled,
		       s.auto_scheduling_locked, s.auto_scheduling_status,
		       s.auto_scheduling_reason, s.auto_scheduling_updated_at,
		       s.sprint_id, s.key_result_id,
		       s.team_id, s.workspace_id, s.start_date, s.end_date,
		       s.created_at, s.updated_at, s.external_creation_key,
		       COALESCE(
		           (SELECT json_agg(sl.label_id) FROM story_labels sl WHERE sl.story_id = s.id),
		           CAST('[]' AS json)
		       ) AS labels
		FROM stories s
		WHERE s.workspace_id = $1
		  AND s.external_creation_key = $2
		LIMIT 1
	`, workspaceID, strings.TrimSpace(key))
	if err != nil {
		return stories.CoreSingleStory{}, err
	}
	return toCoreStory(row), nil
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

func (r *repo) insertStory(ctx context.Context, tx *sqlx.Tx, story *stories.CoreSingleStory) (dbStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.insertStory")
	defer span.End()

	q := `
			INSERT INTO stories (
					sequence_id, title, description, description_html,
					parent_id, objective_id, status_id, assignee_id, 
					blocked_by_id, blocking_id, related_id, reporter_id,
					priority, estimate_unit, estimated_duration_minutes, minimum_focus_block_minutes,
					auto_scheduling_enabled, auto_scheduling_locked, auto_scheduling_status,
					auto_scheduling_reason, auto_scheduling_updated_at,
					sprint_id, key_result_id, team_id, workspace_id, start_date,
					end_date, external_creation_key, created_at, updated_at
			) VALUES (
					:sequence_id, :title, :description, :description_html,
					:parent_id, :objective_id, :status_id, :assignee_id, :blocked_by_id,
					:blocking_id, :related_id, :reporter_id, :priority, :estimate_unit,
					:estimated_duration_minutes, :minimum_focus_block_minutes,
					:auto_scheduling_enabled, :auto_scheduling_locked, :auto_scheduling_status,
					:auto_scheduling_reason, :auto_scheduling_updated_at, :sprint_id,
					:key_result_id, :team_id, :workspace_id, :start_date, :end_date, :external_creation_key, :created_at, :updated_at
			) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.estimate_unit, stories.estimated_duration_minutes, stories.minimum_focus_block_minutes, stories.auto_scheduling_enabled, stories.auto_scheduling_locked, stories.auto_scheduling_status, stories.auto_scheduling_reason, stories.auto_scheduling_updated_at, stories.sprint_id, stories.key_result_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.external_creation_key, stories.created_at, stories.updated_at;
		`

	var cs dbStory
	stmt, err := tx.PrepareNamedContext(ctx, q)
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

func deduplicateLabelIDs(labelIDs []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(labelIDs))
	unique := make([]uuid.UUID, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		if _, exists := seen[labelID]; exists {
			continue
		}
		seen[labelID] = struct{}{}
		unique = append(unique, labelID)
	}
	return unique
}

func (r *repo) insertStoryLabels(
	ctx context.Context,
	tx *sqlx.Tx,
	storyID, workspaceID, teamID uuid.UUID,
	labelIDs []uuid.UUID,
) error {
	labelIDs = deduplicateLabelIDs(labelIDs)
	if len(labelIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.Named(`
		INSERT INTO story_labels (story_id, label_id)
		SELECT :story_id, labels.label_id
		FROM labels
		WHERE labels.label_id IN (:label_ids)
			AND labels.workspace_id = :workspace_id
			AND (labels.team_id = :team_id OR labels.team_id IS NULL)
		RETURNING label_id
	`, map[string]any{
		"story_id":     storyID,
		"workspace_id": workspaceID,
		"team_id":      teamID,
		"label_ids":    labelIDs,
	})
	if err != nil {
		return fmt.Errorf("bind story labels: %w", err)
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return fmt.Errorf("expand story labels: %w", err)
	}

	rows, err := tx.QueryxContext(ctx, tx.Rebind(query), args...)
	if err != nil {
		return fmt.Errorf("insert story labels: %w", err)
	}
	defer rows.Close()

	inserted := 0
	for rows.Next() {
		var labelID uuid.UUID
		if err := rows.Scan(&labelID); err != nil {
			return fmt.Errorf("scan inserted story label: %w", err)
		}
		inserted++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read inserted story labels: %w", err)
	}
	if inserted != len(labelIDs) {
		return fmt.Errorf("%w: %d of %d labels are authorized", stories.ErrInvalidStoryLabels, inserted, len(labelIDs))
	}

	return nil
}

func (r *repo) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateLabels")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var teamID uuid.UUID
	if err := tx.GetContext(ctx, &teamID, `
		SELECT team_id
		FROM stories
		WHERE id = $1 AND workspace_id = $2
		FOR UPDATE
	`, id, workspaceId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stories.ErrNotFound
		}
		return fmt.Errorf("load story team: %w", err)
	}

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

	if err := r.insertStoryLabels(ctx, tx, id, workspaceId, teamID, labels); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

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

// HardBulkDelete permanently removes the stories and returns inline-media
// attachments that no feature references after the deletion.
func (r *repo) HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) ([]uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.HardBulkDelete")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin hard story deletion: %w", err)
	}
	defer tx.Rollback()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}
	orphanedAttachmentIDs := []uuid.UUID{}
	orphanQuery := `
		SELECT DISTINCT inline_media.attachment_id
		FROM story_inline_attachments inline_media
		JOIN stories target_story ON target_story.id = inline_media.story_id
		JOIN attachments attachment ON attachment.attachment_id = inline_media.attachment_id
		WHERE target_story.id = ANY(:ids)
			AND target_story.workspace_id = :workspace_id
			AND attachment.workspace_id = target_story.workspace_id
			AND NOT EXISTS (
				SELECT 1
				FROM story_inline_attachments other_inline_media
				JOIN stories other_inline_story ON other_inline_story.id = other_inline_media.story_id
				WHERE other_inline_media.attachment_id = inline_media.attachment_id
					AND NOT (
						other_inline_story.workspace_id = :workspace_id
						AND other_inline_story.id = ANY(:ids)
					)
			)
			AND NOT EXISTS (
				SELECT 1
				FROM story_attachments other_story_media
				JOIN stories other_story ON other_story.id = other_story_media.story_id
				WHERE other_story_media.attachment_id = inline_media.attachment_id
					AND NOT (
						other_story.workspace_id = :workspace_id
						AND other_story.id = ANY(:ids)
					)
			)
			AND NOT EXISTS (
				SELECT 1
				FROM document_attachments document_media
				WHERE document_media.attachment_id = inline_media.attachment_id
			)`
	orphanStmt, err := tx.PrepareNamedContext(ctx, orphanQuery)
	if err != nil {
		return nil, fmt.Errorf("prepare orphaned story media query: %w", err)
	}
	defer orphanStmt.Close()
	if err := orphanStmt.SelectContext(ctx, &orphanedAttachmentIDs, params); err != nil {
		return nil, fmt.Errorf("select orphaned story media: %w", err)
	}

	deleteQuery := `
			DELETE FROM stories
			WHERE id = ANY(:ids)
				AND workspace_id = :workspace_id;
		`

	r.log.Info(ctx, fmt.Sprintf("Hard deleting stories: %v", ids), "ids", ids)

	result, err := tx.NamedExecContext(ctx, deleteQuery, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to hard delete stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return nil, err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to hard delete", "ids", ids)
		return nil, fmt.Errorf("no stories found to delete")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit hard story deletion: %w", err)
	}

	r.log.Info(ctx, fmt.Sprintf("Stories hard deleted successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories hard deleted.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return orphanedAttachmentIDs, nil
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

	query := "WITH updated_story AS (UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id AND workspace_id = :workspace_id RETURNING id, assignee_id) "
	query += "DELETE FROM story_collaborators sc USING updated_story us WHERE sc.story_id = us.id AND sc.user_id = us.assignee_id;"

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

// UpdateIfUnchanged atomically applies updates only when updated_at still
// matches the version inspected by the caller. It returns false without
// changing the story when another writer won the race.
func (r *repo) UpdateIfUnchanged(
	ctx context.Context,
	id, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) (bool, error) {
	r.log.Info(ctx, "business.repository.stories.UpdateIfUnchanged")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateIfUnchanged")
	defer span.End()

	if len(updates) == 0 {
		return false, errors.New("conditional story update requires at least one field")
	}
	if statusID, ok := updates["status_id"].(uuid.UUID); ok {
		var teamID uuid.UUID
		if err := r.db.GetContext(ctx, &teamID, `SELECT team_id FROM stories WHERE id = $1 AND workspace_id = $2`, id, workspaceID); err != nil {
			span.RecordError(err)
			return false, fmt.Errorf("load story team for conditional update: %w", err)
		}
		if err := r.validateStatusTeam(ctx, statusID, teamID); err != nil {
			return false, err
		}
	}

	fields := make([]string, 0, len(updates))
	for field := range updates {
		if !isConditionalStoryUpdateField(field) {
			return false, fmt.Errorf("unsupported conditional story update field %q", field)
		}
		fields = append(fields, field)
	}
	sort.Strings(fields)
	setClauses := make([]string, 0, len(fields)+1)
	params := map[string]any{
		"id":                  id,
		"workspace_id":        workspaceID,
		"expected_updated_at": expectedUpdatedAt.UTC(),
	}
	for _, field := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = updates[field]
	}
	setClauses = append(setClauses, "updated_at = NOW()")

	query := `WITH updated_story AS (
		UPDATE stories
		SET ` + strings.Join(setClauses, ", ") + `
		WHERE id = :id
		  AND workspace_id = :workspace_id
		  AND updated_at = :expected_updated_at
		RETURNING id, assignee_id
	), deleted_collaborators AS (
		DELETE FROM story_collaborators sc
		USING updated_story us
		WHERE sc.story_id = us.id
		  AND sc.user_id = us.assignee_id
		RETURNING sc.story_id
	)
	SELECT EXISTS (SELECT 1 FROM updated_story)`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("prepare conditional story update: %w", err)
	}
	defer stmt.Close()

	var updated bool
	if err := stmt.GetContext(ctx, &updated, params); err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("execute conditional story update: %w", err)
	}
	return updated, nil
}

// MayaScheduleBlocksExist reports whether there is a committed, Maya-managed
// schedule to lock. User-created calendar blocks never satisfy this contract.
func (r *repo) MayaScheduleBlocksExist(ctx context.Context, storyID, workspaceID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM calendar_schedule_blocks
			WHERE story_id = $1
				AND workspace_id = $2
				AND source = 'maya'
				AND completed_at IS NULL
		)
	`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, storyID, workspaceID); err != nil {
		return false, fmt.Errorf("check Maya schedule blocks: %w", err)
	}
	return exists, nil
}

// UpdateAutoSchedulingStateIfUnchanged persists scheduler-owned state without
// advancing stories.updated_at. That timestamp is the planner's input-version
// watermark; changing it after a successful block reconciliation would make
// the provider outbox reject the schedule that was just committed.
func (r *repo) UpdateAutoSchedulingStateIfUnchanged(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	stateUpdatedAt time.Time,
	locked *bool,
) (bool, error) {
	const query = `
		UPDATE stories
		SET auto_scheduling_status = $4,
			auto_scheduling_reason = $5,
			auto_scheduling_updated_at = $6,
			auto_scheduling_locked = COALESCE($7, auto_scheduling_locked)
		WHERE id = $1
			AND workspace_id = $2
			AND updated_at = $3
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		storyID,
		workspaceID,
		expectedUpdatedAt.UTC(),
		status,
		reason,
		stateUpdatedAt.UTC(),
		locked,
	)
	if err != nil {
		return false, fmt.Errorf("update auto-scheduling state: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read auto-scheduling state update result: %w", err)
	}
	return rows == 1, nil
}

func isConditionalStoryUpdateField(field string) bool {
	switch field {
	case "title",
		"estimate_unit",
		"estimated_duration_minutes",
		"minimum_focus_block_minutes",
		"auto_scheduling_enabled",
		"auto_scheduling_locked",
		"auto_scheduling_status",
		"auto_scheduling_reason",
		"auto_scheduling_updated_at",
		"description",
		"description_html",
		"parent_id",
		"objective_id",
		"status_id",
		"assignee_id",
		"priority",
		"sprint_id",
		"key_result_id",
		"start_date",
		"end_date",
		"completed_at":
		return true
	default:
		return false
	}
}

// BulkUpdate updates the stories with the specified IDs.
func (r *repo) BulkUpdate(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUpdate")
	defer span.End()

	query := "WITH updated_stories AS (UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id IN (:ids) AND workspace_id = :workspace_id RETURNING id, assignee_id) "
	query += "DELETE FROM story_collaborators sc USING updated_stories us WHERE sc.story_id = us.id AND sc.user_id = us.assignee_id;"

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

	insertQuery := `
		INSERT INTO story_activities (
			activity_id,
			story_id, 
			activity_type, 
			field_changed, 
			current_value,
			old_value,
			new_value,
			reason,
			user_id,
			workspace_id
		)
		VALUES (
			:activity_id,
			:story_id, 
			:activity_type, 
			:field_changed, 
			:current_value,
			:old_value,
			:new_value,
			:reason,
			:user_id,
			:workspace_id
		)
		ON CONFLICT (activity_id) DO UPDATE
		SET activity_id = EXCLUDED.activity_id
		RETURNING story_activities.*;
	`
	updateQuery := `
		UPDATE story_activities
		SET
			current_value = :current_value,
			new_value = :new_value,
			reason = COALESCE(:reason, reason),
			created_at = NOW()
		WHERE activity_id = (
			SELECT activity_id
			FROM story_activities
			WHERE story_id = :story_id
				AND user_id = :user_id
				AND workspace_id IS NOT DISTINCT FROM :workspace_id
				AND activity_type = :activity_type
				AND field_changed = :field_changed
				AND created_at >= NOW() - INTERVAL '` + activityCompactionWindowSQL + `'
			ORDER BY created_at DESC
			LIMIT 1
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

	insertStmt, err := tx.PrepareNamedContext(ctx, insertQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer insertStmt.Close()

	updateStmt, err := tx.PrepareNamedContext(ctx, updateQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare activity compaction statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare activity compaction statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer updateStmt.Close()

	// Insert each activity and collect results
	var result []dbActivity
	for _, activity := range activities {
		var da dbActivity
		compact := shouldCompactActivity(activity)
		if activity.ID == uuid.Nil {
			activity.ID = uuid.New()
		}
		dbActivity := toDBActivity(activity)
		if compact {
			err := updateStmt.GetContext(ctx, &da, dbActivity)
			if err == nil {
				result = append(result, da)
				continue
			}
			if !errors.Is(err, sql.ErrNoRows) {
				errMsg := fmt.Sprintf("Failed to compact activity: %s", err)
				r.log.Error(ctx, errMsg)
				span.RecordError(errors.New("failed to compact activity"), trace.WithAttributes(attribute.String("error", errMsg)))
				return nil, err
			}
		}

		if err := insertStmt.GetContext(ctx, &da, dbActivity); err != nil {
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

func shouldCompactActivity(activity stories.CoreActivity) bool {
	return activity.ID == uuid.Nil && activity.Type == "update" && activity.Field != ""
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
	defer tx.Rollback()

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
			estimate_unit,
			estimated_duration_minutes,
			minimum_focus_block_minutes,
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
			:estimate_unit,
			:estimated_duration_minutes,
			:minimum_focus_block_minutes,
			:sprint_id,
			:workspace_id,
			:reporter_id,
			NOW(),
			NOW()
		) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.estimate_unit, stories.estimated_duration_minutes, stories.minimum_focus_block_minutes, stories.auto_scheduling_enabled, stories.auto_scheduling_locked, stories.auto_scheduling_status, stories.auto_scheduling_reason, stories.auto_scheduling_updated_at, stories.sprint_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
	`

	// Prepare parameters for the new story
	params := map[string]any{
		"sequence_id":                 lastSequence + 1,
		"title":                       "Copy of " + originalStory.Title,
		"description":                 originalStory.Description,
		"description_html":            originalStory.DescriptionHTML,
		"team_id":                     originalStory.Team,
		"objective_id":                originalStory.Objective,
		"status_id":                   originalStory.Status,
		"assignee_id":                 originalStory.Assignee,
		"priority":                    originalStory.Priority,
		"estimate_unit":               originalStory.EstimateValue,
		"estimated_duration_minutes":  originalStory.EstimatedDurationMinutes,
		"minimum_focus_block_minutes": originalStory.MinimumFocusBlockMinutes,
		"sprint_id":                   originalStory.Sprint,
		"workspace_id":                workspaceId,
		"reporter_id":                 userID,
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

	originalMediaPath := storyMediaPath(originalStoryID)
	duplicatedMediaPath := storyMediaPath(newStory.ID)
	if _, err := tx.ExecContext(ctx, `
		UPDATE stories
		SET description_html = REPLACE(description_html, $1, $2)
		WHERE id = $3 AND workspace_id = $4`, originalMediaPath, duplicatedMediaPath, newStory.ID, workspaceId); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("rewrite duplicated story media URLs: %w", err)
	}
	newStory.DescriptionHTML = rewriteStoryMediaHTML(newStory.DescriptionHTML, originalStoryID, newStory.ID)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO story_inline_attachments (story_id, attachment_id, created_by)
		SELECT $1, media.attachment_id, $2
		FROM story_inline_attachments media
		JOIN stories source_story ON source_story.id = media.story_id
		JOIN attachments attachment ON attachment.attachment_id = media.attachment_id
		WHERE media.story_id = $3
			AND source_story.workspace_id = $4
			AND attachment.workspace_id = source_story.workspace_id
		ON CONFLICT (story_id, attachment_id) DO NOTHING`, newStory.ID, userID, originalStoryID, workspaceId); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("copy duplicated story media links: %w", err)
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

func storyMediaPath(storyID uuid.UUID) string {
	return "/stories/" + storyID.String() + "/media/"
}

func rewriteStoryMediaHTML(contentHTML *string, originalStoryID, duplicatedStoryID uuid.UUID) *string {
	if contentHTML == nil {
		return nil
	}
	rewritten := strings.ReplaceAll(*contentHTML, storyMediaPath(originalStoryID), storyMediaPath(duplicatedStoryID))
	return &rewritten
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
			sequence_id, title, description, description_html, status_id, priority, estimate_unit, team_id, workspace_id, reporter_id, created_at, updated_at
		) VALUES (
			:sequence_id, :title, :description, :description_html, :status_id, :priority, :estimate_unit, :team_id, :workspace_id, :reporter_id, NOW(), NOW()
		) RETURNING id`

	params := map[string]any{
		"sequence_id":      sequenceID + 1,
		"title":            title,
		"description":      description,
		"description_html": description,
		"status_id":        statusID,
		"estimate_unit":    nil,
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
		WITH inserted AS (
			INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id)
			VALUES (:from_story_id, :to_story_id, :association_type, :workspace_id)
			RETURNING id, from_story_id, to_story_id
		)
		SELECT
			inserted.id,
			from_story.title AS from_story_title,
			to_story.title AS to_story_title
		FROM inserted
		INNER JOIN stories from_story ON from_story.id = inserted.from_story_id
		INNER JOIN stories to_story ON to_story.id = inserted.to_story_id
	`

	params := map[string]any{
		"from_story_id":    fromID,
		"to_story_id":      toID,
		"association_type": associationType,
		"workspace_id":     workspaceID,
	}

	var association struct {
		ID             uuid.UUID `db:"id"`
		FromStoryTitle string    `db:"from_story_title"`
		ToStoryTitle   string    `db:"to_story_title"`
	}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to prepare insert association: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &association, params); err != nil {
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
		ID:             association.ID,
		FromStoryID:    fromID,
		ToStoryID:      toID,
		Type:           associationType,
		FromStoryTitle: association.FromStoryTitle,
		ToStoryTitle:   association.ToStoryTitle,
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

// UpdateAssociation updates an association between two stories.
func (r *repo) UpdateAssociation(ctx context.Context, associationID, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error) {
	query := `
			WITH previous AS (
				SELECT association_type
				FROM story_associations
				WHERE id = :id AND workspace_id = :workspace_id
			),
			updated AS (
				UPDATE story_associations
				SET from_story_id = :from_story_id,
					to_story_id = :to_story_id,
					association_type = :association_type
				WHERE id = :id AND workspace_id = :workspace_id
				RETURNING id, from_story_id, to_story_id
			)
			SELECT
				updated.id,
				previous.association_type AS previous_type,
				from_story.title AS from_story_title,
				to_story.title AS to_story_title
			FROM updated
			CROSS JOIN previous
			INNER JOIN stories from_story ON from_story.id = updated.from_story_id
			INNER JOIN stories to_story ON to_story.id = updated.to_story_id
		`

	params := map[string]any{
		"id":               associationID,
		"from_story_id":    fromID,
		"to_story_id":      toID,
		"association_type": associationType,
		"workspace_id":     workspaceID,
	}

	var result struct {
		ID             uuid.UUID `db:"id"`
		PreviousType   string    `db:"previous_type"`
		FromStoryTitle string    `db:"from_story_title"`
		ToStoryTitle   string    `db:"to_story_title"`
	}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to prepare update association: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to update association: %w", err)
	}

	toStory, err := r.getStoryById(ctx, toID, workspaceID)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to fetch target story details: %w", err)
	}

	coreToStory := toCoreStory(toStory)

	return stories.CoreStoryAssociation{
		ID:             result.ID,
		FromStoryID:    fromID,
		ToStoryID:      toID,
		Type:           associationType,
		PreviousType:   &result.PreviousType,
		FromStoryTitle: result.FromStoryTitle,
		ToStoryTitle:   result.ToStoryTitle,
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
func (r *repo) RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error) {
	query := `
		WITH deleted AS (
			DELETE FROM story_associations
			WHERE id = :id AND workspace_id = :workspace_id
			RETURNING id, from_story_id, to_story_id, association_type
		)
		SELECT
			deleted.id,
			deleted.from_story_id,
			deleted.to_story_id,
			deleted.association_type,
			from_story.title AS from_story_title,
			to_story.title AS to_story_title
		FROM deleted
		INNER JOIN stories from_story ON from_story.id = deleted.from_story_id
		INNER JOIN stories to_story ON to_story.id = deleted.to_story_id
	`

	params := map[string]any{
		"id":           associationID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to prepare delete association: %w", err)
	}
	defer stmt.Close()

	var association struct {
		ID             uuid.UUID `db:"id"`
		FromStoryID    uuid.UUID `db:"from_story_id"`
		ToStoryID      uuid.UUID `db:"to_story_id"`
		Type           string    `db:"association_type"`
		FromStoryTitle string    `db:"from_story_title"`
		ToStoryTitle   string    `db:"to_story_title"`
	}
	if err := stmt.GetContext(ctx, &association, params); err != nil {
		return stories.CoreStoryAssociation{}, fmt.Errorf("failed to delete association: %w", err)
	}

	return stories.CoreStoryAssociation{
		ID:             association.ID,
		FromStoryID:    association.FromStoryID,
		ToStoryID:      association.ToStoryID,
		Type:           association.Type,
		FromStoryTitle: association.FromStoryTitle,
		ToStoryTitle:   association.ToStoryTitle,
	}, nil
}

// Helper to convert []CoreStoryList (from CoreSingleStory.SubStories) back to []CoreStoryList
// This seems redundant but CoreSingleStory uses []CoreStoryList for SubStories.
func toCoreStoryList(stories []stories.CoreStoryList) []stories.CoreStoryList {
	return stories
}
