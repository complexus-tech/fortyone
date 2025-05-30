package storiesrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/repo/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/linksrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// GetNextSequenceID returns the next sequence ID for a team.
func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error) {
	var currentSequence int

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := `
		UPDATE team_story_sequences
		SET current_sequence = current_sequence + 1
		WHERE workspace_id = :workspace_id AND team_id = :team_id
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

	err = stmt.GetContext(ctx, &currentSequence, params)
	if errors.Is(err, sql.ErrNoRows) {
		// If no record exists, insert a new one starting from 1
		q := `
			INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence)
			VALUES (:workspace_id, :team_id, 0)
			RETURNING current_sequence
		`
		stmt, err = tx.PrepareNamedContext(ctx, q)
		if err != nil {
			tx.Rollback()
			return 0, nil, nil, fmt.Errorf("failed to prepare named statement for insert: %w", err)
		}
		defer stmt.Close()

		err = stmt.GetContext(ctx, &currentSequence, params)
	}

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

// Create creates a new story.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	// Validate status belongs to the same team
	if story.Status != nil {
		if err := r.validateStatusTeam(ctx, *story.Status, story.Team); err != nil {
			return stories.CoreSingleStory{}, err
		}
	}

	lastSequence, commit, rollback, err := r.GetNextSequenceID(ctx, story.Team, story.Workspace)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get next sequence ID: %w", err)
	}
	story.SequenceID = lastSequence + 1

	cs, err := r.insertStory(ctx, story)
	if err != nil {
		rollback()
		return stories.CoreSingleStory{}, fmt.Errorf("failed to insert story: %w", err)
	}

	if err := commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return toCoreStory(cs), nil
}

func (r *repo) insertStory(ctx context.Context, story *stories.CoreSingleStory) (dbStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.insertStory")
	defer span.End()

	q := `
			INSERT INTO stories (
					sequence_id, title, description, description_html,
					parent_id, objective_id, status_id, assignee_id, 
					blocked_by_id, blocking_id, related_id, reporter_id,
					priority, sprint_id, team_id, workspace_id, start_date, 
					end_date, created_at, updated_at
			) VALUES (
					:sequence_id, :title, :description, :description_html,
					:parent_id, :objective_id, :status_id, :assignee_id, :blocked_by_id,
					:blocking_id, :related_id, :reporter_id, :priority, :sprint_id,
					:team_id, :workspace_id, :start_date, :end_date, :created_at, :updated_at
			) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.sprint_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
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

func (r *repo) GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error) {
	r.log.Info(ctx, "business.repository.stories.GetStoryLinks")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetStoryLinks")
	defer span.End()

	span.SetAttributes(attribute.String("storyId", storyID.String()))

	var dbLinks []linksrepo.DbLink
	query := `
		SELECT 
			link_id,
			title,
			url,
			story_id,
			created_at,
			updated_at
		FROM story_links
		WHERE story_id = :story_id
	`

	params := map[string]any{
		"story_id": storyID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbLinks, params); err != nil {
		r.log.Error(ctx, "error selecting links", err)
		return nil, err
	}

	return linksrepo.ToCoreLinks(dbLinks), nil
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
func (r *repo) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.MyStories")
	defer span.End()

	currentUser, _ := mid.GetUserID(ctx)

	params := map[string]any{
		"workspace_id": workspaceId,
		"current_user": currentUser,
	}

	var stories []dbStory

	q := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'labels', '[]'		
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id 
						AND sub.deleted_at IS NULL
				), '[]'
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user
		WHERE s.workspace_id = :workspace_id 
		AND s.deleted_at IS NULL
		AND s.parent_id IS NULL
		AND (s.assignee_id = :current_user OR s.reporter_id = :current_user)
		ORDER BY s.created_at DESC;
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "fetching stories.")
	if err := stmt.SelectContext(ctx, &stories, params); err != nil {
		errMsg := fmt.Sprintf("failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "stories retrieved successfully.")
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// Get returns the story with the specified ID.
func (r *repo) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Get")
	defer span.End()

	story, err := r.getStoryById(ctx, id, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, err
	}

	return toCoreStory(story), nil
}

func (r *repo) getStoryById(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (dbStory, error) {
	query := `
        SELECT
					s.id,
					s.title,
					s.priority,
					s.sequence_id,
					s.status_id,
					s.description,
					s.description_html,
					s.team_id,
					s.objective_id,
					s.sprint_id,
					s.workspace_id,
					s.assignee_id,
					s.reporter_id,
					s.start_date,
					s.end_date,
					s.created_at,
					s.updated_at,
					s.deleted_at,
					COALESCE(
							(
									SELECT
											json_agg(sub.*)
									FROM
											stories sub
									WHERE
											sub.parent_id = s.id AND sub.deleted_at IS NULL
							), '[]'
					) AS sub_stories,
					COALESCE(
						(
								SELECT
										json_agg(l.label_id)
								FROM
										labels l
										INNER JOIN story_labels sl ON sl.label_id = l.label_id
								WHERE
										sl.story_id = s.id
						), '[]') AS labels
				FROM
					stories s
				WHERE
					s.id =:id
					AND s.workspace_id =:workspace_id;
    `

	params := map[string]any{
		"id":           id,
		"workspace_id": workspaceId,
	}

	var story dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "id", id)
		return dbStory{}, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &story, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbStory{}, errors.New("story not found")
		}
		r.log.Error(ctx, fmt.Sprintf("failed to execute query: %s", err), "id", id)
		return dbStory{}, err
	}

	return story, nil
}

// Delete deletes the story with the specified ID.
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

// BulkDelete removes the stories with the specified IDs.
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

func (r *repo) GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetSubStories")
	defer span.End()

	subStories, err := r.getSubStories(ctx, parentId, workspaceId)
	if err != nil {
		return nil, err
	}

	return toCoreStories(subStories), nil
}

func (r *repo) getSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]dbStory, error) {
	query := `
        SELECT
          id,
					sequence_id,
					title,
					priority,
					description,
					status_id,
					start_date,
					end_date,
					sprint_id,
					team_id,
					objective_id,
					workspace_id,
					assignee_id,
					reporter_id,
					created_at,
					updated_at
        FROM
            stories
        WHERE
            parent_id = :parent_id
					AND workspace_id = :workspace_id
					AND deleted_at IS NULL
        ORDER BY sequence_id ASC;
    `
	params := map[string]any{
		"parent_id":    parentId,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", parentId)
		return nil, err
	}
	defer stmt.Close()

	var subStories []dbStory
	if err := stmt.SelectContext(ctx, &subStories, params); err != nil {
		return nil, err
	}

	return subStories, nil
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
			user_id,
			workspace_id
		)
		VALUES (
			:story_id, 
			:activity_type, 
			:field_changed, 
			:current_value,
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

// GetActivities returns all activities for a given story ID.
func (r *repo) GetActivities(ctx context.Context, storyID uuid.UUID) ([]stories.CoreActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetActivities")
	defer span.End()
	params := map[string]any{"story_id": storyID}

	q := `
		SELECT 
			activity_id,
			story_id,
			user_id,
			activity_type,
			field_changed,
			current_value,
			created_at,
			workspace_id
		FROM story_activities
		WHERE story_id = :story_id
		ORDER BY created_at
	`

	var activities []dbActivity

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &activities, params); err != nil {
		errMsg := fmt.Sprintf("Failed to get activities: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get activities"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return toCoreActivities(activities), nil
}

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

	var comment commentsrepo.DbComment
	if err := stmt.GetContext(ctx, &comment, toDBNewComment(cnc)); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to insert comment: %s", err))
		return comments.CoreComment{}, err
	}

	return toCoreComment(comment), nil
}

func (r *repo) GetComments(ctx context.Context, storyID uuid.UUID) ([]comments.CoreComment, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetComments")
	defer span.End()

	q := `
		SELECT 
			sc.comment_id,
			sc.story_id,
			sc.commenter_id,
			sc.content,
			sc.parent_id,
			sc.created_at,
			sc.updated_at,
			COALESCE(
				(
					SELECT
						json_agg(sub.*)
					FROM
						story_comments sub
					WHERE
						sub.parent_id = sc.comment_id			
				), '[]'
			) AS sub_comments
		FROM story_comments sc WHERE sc.story_id = :story_id AND sc.parent_id IS NULL ORDER BY sc.created_at ASC
	`

	params := map[string]any{"story_id": storyID}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return nil, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var comments []commentsrepo.DbComment
	if err := stmt.SelectContext(ctx, &comments, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to get comments: %s", err))
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return toCoreComments(comments), nil
}

func (r *repo) GetComment(ctx context.Context, commentID uuid.UUID) (comments.CoreComment, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetComment")
	defer span.End()

	q := `
		SELECT comment_id, story_id, commenter_id, content, parent_id, created_at, updated_at
		FROM story_comments 
		WHERE comment_id = :comment_id
	`

	params := map[string]any{"comment_id": commentID}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return comments.CoreComment{}, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var comment commentsrepo.DbComment
	if err := stmt.GetContext(ctx, &comment, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to get comment: %s", err))
		return comments.CoreComment{}, fmt.Errorf("failed to get comment: %w", err)
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
func (r *repo) CountStoriesInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CountStoriesInWorkspace")
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM stories
		WHERE workspace_id = :workspace_id
		AND deleted_at IS NULL AND archived_at IS NULL
	`

	params := map[string]any{
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count stories statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count stories in workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to count stories"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	span.AddEvent("stories counted.", trace.WithAttributes(
		attribute.Int("stories.count", count),
	))

	return count, nil
}

// ListGroupedStories returns stories grouped by the specified field with limited stories per group
func (r *repo) ListGroupedStories(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ListGroupedStories")
	defer span.End()

	// Use SQL-based grouping for better performance with large datasets
	if r.shouldUseSQLGrouping(query) {
		return r.listGroupedStoriesSQL(ctx, query)
	}

	// Fall back to Go-based grouping for complex cases
	return r.listGroupedStoriesGo(ctx, query)
}

// shouldUseSQLGrouping determines whether to use SQL-based or Go-based grouping
func (r *repo) shouldUseSQLGrouping(query stories.CoreStoryQuery) bool {
	// Always use optimized SQL grouping for better performance
	switch query.GroupBy {
	case "status", "assignee", "priority", "team", "sprint":
		return true
	default:
		return false
	}
}

// listGroupedStoriesSQL performs grouping directly in SQL for optimal performance
func (r *repo) listGroupedStoriesSQL(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.listGroupedStoriesSQL")
	defer span.End()

	// Build optimized query that fetches storiesPerGroup+1 to check if more exist
	groupColumn := r.getGroupColumn(query.GroupBy)
	limit := query.StoriesPerGroup + 1 // Fetch one extra to check if more exist

	// Simplified query without expensive COUNT() operations
	sqlQuery := fmt.Sprintf(`
		WITH grouped_stories AS (
			SELECT 
				s.id,
				s.sequence_id,
				s.title,
				s.priority,
				s.status_id,
				s.assignee_id,
				s.team_id,
				s.workspace_id,
				s.created_at,
				s.updated_at,
				COALESCE(CAST(%s AS text), 'null') as group_key,
				ROW_NUMBER() OVER (PARTITION BY COALESCE(CAST(%s AS text), 'null') ORDER BY s.created_at DESC) as row_num
			FROM stories s
			%s
		),
		limited_stories AS (
			SELECT *
			FROM grouped_stories
			WHERE row_num <= %d
		)
		SELECT 
			group_key,
			COALESCE(
				json_agg(
					json_build_object(
						'id', id,
						'sequence_id', sequence_id,
						'title', title,
						'priority', priority,
						'status_id', status_id,
						'assignee_id', assignee_id,
						'team_id', team_id,
						'workspace_id', workspace_id,
						'created_at', created_at,
						'updated_at', updated_at
					) ORDER BY created_at DESC
				) FILTER (WHERE id IS NOT NULL), 
				CAST('[]' AS json)
			) as stories_json
		FROM limited_stories
		GROUP BY group_key
		ORDER BY group_key`,
		groupColumn, groupColumn,
		r.buildSimpleWhereClause(query.Filters),
		limit)

	params := r.buildQueryParams(query.Filters)

	type groupResult struct {
		GroupKey    string          `db:"group_key"`
		StoriesJSON json.RawMessage `db:"stories_json"`
	}

	var results []groupResult
	stmt, err := r.db.PrepareNamedContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare group query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &results, params); err != nil {
		return nil, fmt.Errorf("failed to execute group query: %w", err)
	}

	// Convert results to CoreStoryGroup
	var groups []stories.CoreStoryGroup
	for _, result := range results {
		var storyMaps []map[string]interface{}
		if err := json.Unmarshal(result.StoriesJSON, &storyMaps); err != nil {
			r.log.Error(ctx, "failed to unmarshal stories JSON", "error", err)
			continue
		}

		// Convert story maps to CoreStoryList
		var coreStories []stories.CoreStoryList
		for _, storyMap := range storyMaps {
			coreStory := r.mapToStoryList(storyMap)
			coreStories = append(coreStories, coreStory)
		}

		// Check if we have more stories (we fetched storiesPerGroup+1)
		hasMore := len(coreStories) > query.StoriesPerGroup
		if hasMore {
			// Remove the extra story we fetched
			coreStories = coreStories[:query.StoriesPerGroup]
		}

		loadedCount := len(coreStories)

		group := stories.CoreStoryGroup{
			Key:         result.GroupKey,
			LoadedCount: loadedCount,
			HasMore:     hasMore,
			Stories:     coreStories,
			NextPage:    2, // Next page for load more
		}
		groups = append(groups, group)
	}

	span.AddEvent("SQL grouped stories retrieved.", trace.WithAttributes(
		attribute.Int("groups.count", len(groups)),
		attribute.String("group.by", query.GroupBy),
	))

	return groups, nil
}

// buildSimpleWhereClause builds a simplified WHERE clause without subqueries
func (r *repo) buildSimpleWhereClause(filters stories.CoreStoryFilters) string {
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.deleted_at IS NULL",
		"s.parent_id IS NULL",
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.Epic != nil {
		whereClauses = append(whereClauses, "s.epic_id = :epic_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	if filters.AssignedToMe != nil && *filters.AssignedToMe {
		whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
	}

	if filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
	}

	return "WHERE " + strings.Join(whereClauses, " AND ")
}

// mapToStoryList converts a map to CoreStoryList
func (r *repo) mapToStoryList(storyMap map[string]interface{}) stories.CoreStoryList {
	story := stories.CoreStoryList{
		Labels:     []uuid.UUID{},             // Empty for now since we're focusing on performance
		SubStories: []stories.CoreStoryList{}, // Empty for now since we're focusing on performance
	}

	// Helper function to safely convert values
	if id, ok := storyMap["id"].(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			story.ID = parsed
		}
	}

	if seqID, ok := storyMap["sequence_id"].(float64); ok {
		story.SequenceID = int(seqID)
	}

	if title, ok := storyMap["title"].(string); ok {
		story.Title = title
	}

	if priority, ok := storyMap["priority"].(string); ok {
		story.Priority = priority
	}

	if statusID, ok := storyMap["status_id"].(string); ok {
		if parsed, err := uuid.Parse(statusID); err == nil {
			story.Status = &parsed
		}
	}

	if assigneeID, ok := storyMap["assignee_id"].(string); ok && assigneeID != "" {
		if parsed, err := uuid.Parse(assigneeID); err == nil {
			story.Assignee = &parsed
		}
	}

	if reporterID, ok := storyMap["reporter_id"].(string); ok && reporterID != "" {
		if parsed, err := uuid.Parse(reporterID); err == nil {
			story.Reporter = &parsed
		}
	}

	if objectiveID, ok := storyMap["objective_id"].(string); ok && objectiveID != "" {
		if parsed, err := uuid.Parse(objectiveID); err == nil {
			story.Objective = &parsed
		}
	}

	if sprintID, ok := storyMap["sprint_id"].(string); ok && sprintID != "" {
		if parsed, err := uuid.Parse(sprintID); err == nil {
			story.Sprint = &parsed
		}
	}

	if teamID, ok := storyMap["team_id"].(string); ok {
		if parsed, err := uuid.Parse(teamID); err == nil {
			story.Team = parsed
		}
	}

	if workspaceID, ok := storyMap["workspace_id"].(string); ok {
		if parsed, err := uuid.Parse(workspaceID); err == nil {
			story.Workspace = parsed
		}
	}

	// Handle time fields
	if createdAt, ok := storyMap["created_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, createdAt); err == nil {
			story.CreatedAt = parsed
		}
	}

	if updatedAt, ok := storyMap["updated_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			story.UpdatedAt = parsed
		}
	}

	// Handle nullable time fields
	if startDate, ok := storyMap["start_date"].(string); ok && startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			story.StartDate = &parsed
		}
	}

	if endDate, ok := storyMap["end_date"].(string); ok && endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			story.EndDate = &parsed
		}
	}

	return story
}

// getGroupColumn returns the column name for grouping
func (r *repo) getGroupColumn(groupBy string) string {
	switch groupBy {
	case "status":
		return "s.status_id"
	case "assignee":
		return "s.assignee_id"
	case "priority":
		return "s.priority"
	case "team":
		return "s.team_id"
	case "sprint":
		return "s.sprint_id"
	default:
		return "s.id"
	}
}

// buildQueryParams builds the parameter map for the SQL query
func (r *repo) buildQueryParams(filters stories.CoreStoryFilters) map[string]any {
	params := map[string]any{
		"workspace_id":    filters.WorkspaceID,
		"current_user_id": filters.CurrentUserID,
	}

	if len(filters.StatusIDs) > 0 {
		// Convert to PostgreSQL array format
		statusArray := "{"
		for i, id := range filters.StatusIDs {
			if i > 0 {
				statusArray += ","
			}
			statusArray += id.String()
		}
		statusArray += "}"
		params["status_ids"] = statusArray
	}

	if len(filters.AssigneeIDs) > 0 {
		assigneeArray := "{"
		for i, id := range filters.AssigneeIDs {
			if i > 0 {
				assigneeArray += ","
			}
			assigneeArray += id.String()
		}
		assigneeArray += "}"
		params["assignee_ids"] = assigneeArray
	}

	if len(filters.ReporterIDs) > 0 {
		reporterArray := "{"
		for i, id := range filters.ReporterIDs {
			if i > 0 {
				reporterArray += ","
			}
			reporterArray += id.String()
		}
		reporterArray += "}"
		params["reporter_ids"] = reporterArray
	}

	if len(filters.Priorities) > 0 {
		priorityArray := "{"
		for i, priority := range filters.Priorities {
			if i > 0 {
				priorityArray += ","
			}
			priorityArray += "\"" + priority + "\""
		}
		priorityArray += "}"
		params["priorities"] = priorityArray
	}

	if len(filters.TeamIDs) > 0 {
		teamArray := "{"
		for i, id := range filters.TeamIDs {
			if i > 0 {
				teamArray += ","
			}
			teamArray += id.String()
		}
		teamArray += "}"
		params["team_ids"] = teamArray
	}

	if len(filters.SprintIDs) > 0 {
		sprintArray := "{"
		for i, id := range filters.SprintIDs {
			if i > 0 {
				sprintArray += ","
			}
			sprintArray += id.String()
		}
		sprintArray += "}"
		params["sprint_ids"] = sprintArray
	}

	if len(filters.LabelIDs) > 0 {
		labelArray := "{"
		for i, id := range filters.LabelIDs {
			if i > 0 {
				labelArray += ","
			}
			labelArray += id.String()
		}
		labelArray += "}"
		params["label_ids"] = labelArray
	}

	if filters.Parent != nil {
		params["parent_id"] = *filters.Parent
	}

	if filters.Objective != nil {
		params["objective_id"] = *filters.Objective
	}

	if filters.Epic != nil {
		params["epic_id"] = *filters.Epic
	}

	return params
}

// buildStoriesQuery builds the base SQL query for stories with filters
func (r *repo) buildStoriesQuery(filters stories.CoreStoryFilters) string {
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id 
						AND sub.deleted_at IS NULL
				), CAST('[]' AS json)
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), CAST('[]' AS json)
			) AS labels
		FROM
			stories s
	`

	// Add team member join if needed for assigned_to_me or created_by_me filters
	needsTeamJoin := filters.AssignedToMe != nil || filters.CreatedByMe != nil
	if needsTeamJoin {
		query += `
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user_id
		`
	}

	// Build WHERE clauses
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.deleted_at IS NULL",
		"s.parent_id IS NULL",
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if len(filters.LabelIDs) > 0 {
		query += `
			INNER JOIN story_labels sl_filter ON sl_filter.story_id = s.id
		`
		whereClauses = append(whereClauses, "sl_filter.label_id = ANY(:label_ids)")
	}

	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.Epic != nil {
		whereClauses = append(whereClauses, "s.epic_id = :epic_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	if filters.AssignedToMe != nil && *filters.AssignedToMe {
		whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
	}

	if filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
	}

	query += " WHERE " + strings.Join(whereClauses, " AND ")

	return query
}

// listGroupedStoriesGo performs grouping in Go (original implementation)
func (r *repo) listGroupedStoriesGo(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.listGroupedStoriesGo")
	defer span.End()

	// For now, implement a simple approach - get all stories and group in Go
	// TODO: Optimize with SQL-based grouping for better performance

	baseQuery := r.buildStoriesQuery(query.Filters)
	baseQuery += " ORDER BY s.created_at DESC"

	params := r.buildQueryParams(query.Filters)

	var allStories []dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &allStories, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	// Group stories in Go
	groups := r.groupStories(allStories, query.GroupBy, query.StoriesPerGroup)

	span.AddEvent("Go grouped stories retrieved.", trace.WithAttributes(
		attribute.Int("groups.count", len(groups)),
		attribute.String("group.by", query.GroupBy),
	))

	return groups, nil
}

// groupStories groups stories by the specified field and limits stories per group
func (r *repo) groupStories(allStories []dbStory, groupBy string, storiesPerGroup int) []stories.CoreStoryGroup {
	groupMap := make(map[string]*stories.CoreStoryGroup)

	for _, story := range allStories {
		key := r.getGroupKey(story, groupBy)

		// Initialize group if not exists
		if groupMap[key] == nil {
			groupMap[key] = &stories.CoreStoryGroup{
				Key:         key,
				LoadedCount: 0,
				HasMore:     false,
				Stories:     []stories.CoreStoryList{},
				NextPage:    2, // Next page for load more
			}
		}

		group := groupMap[key]

		// Only add to Stories slice if under the limit
		if group.LoadedCount < storiesPerGroup {
			coreStories := toCoreStories([]dbStory{story})
			if len(coreStories) > 0 {
				group.Stories = append(group.Stories, coreStories[0])
				group.LoadedCount++
			}
		} else if group.LoadedCount == storiesPerGroup {
			// We found one more story beyond the limit, so there are more
			group.HasMore = true
		}
	}

	// Convert map to slice
	var groups []stories.CoreStoryGroup
	for _, group := range groupMap {
		groups = append(groups, *group)
	}

	return groups
}

// getGroupKey extracts the group key from a story
func (r *repo) getGroupKey(story dbStory, groupBy string) string {
	switch groupBy {
	case "status":
		if story.Status != nil {
			return story.Status.String()
		}
		return "null"
	case "assignee":
		if story.Assignee != nil {
			return story.Assignee.String()
		}
		return "null"
	case "priority":
		return story.Priority
	case "team":
		return story.Team.String()
	case "sprint":
		if story.Sprint != nil {
			return story.Sprint.String()
		}
		return "null"
	default:
		return "all"
	}
}

// ListGroupStories returns more stories for a specific group (for load more functionality)
func (r *repo) ListGroupStories(ctx context.Context, groupKey string, query stories.CoreStoryQuery) ([]stories.CoreStoryList, int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ListGroupStories")
	defer span.End()

	// Build simplified query for list view (no N+1 subqueries)
	baseQuery := r.buildSimpleStoriesQuery(query.Filters)

	// Add group-specific filter based on groupBy type
	if groupKey == "null" || groupKey == "" {
		// Handle NULL values specially
		baseQuery += fmt.Sprintf(" AND %s IS NULL", r.getGroupColumn(query.GroupBy))
	} else {
		baseQuery += fmt.Sprintf(" AND %s = :group_key", r.getGroupColumn(query.GroupBy))
	}

	// Get paginated stories - fetch one extra to check if there are more
	pageSize := query.PageSize
	offset := (query.Page - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT %d OFFSET %d", pageSize+1, offset)

	params := r.buildQueryParams(query.Filters)
	if groupKey != "null" && groupKey != "" {
		params["group_key"] = groupKey
	}

	var dbStories []dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare stories query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbStories, params); err != nil {
		return nil, 0, fmt.Errorf("failed to get stories: %w", err)
	}

	// Check if there are more stories
	hasMore := len(dbStories) > pageSize
	if hasMore {
		// Remove the extra story we fetched
		dbStories = dbStories[:pageSize]
	}

	coreStories := toCoreStories(dbStories)

	// For load more, we return -1 as total count to indicate we don't need exact count
	totalCount := -1

	span.AddEvent("group stories retrieved.", trace.WithAttributes(
		attribute.Int("stories.count", len(coreStories)),
		attribute.String("group.key", groupKey),
		attribute.Bool("has.more", hasMore),
	))

	return coreStories, totalCount, nil
}

// buildSimpleStoriesQuery builds a simplified query for list views without N+1 subqueries
func (r *repo) buildSimpleStoriesQuery(filters stories.CoreStoryFilters) string {
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			CAST('[]' AS json) AS sub_stories,
			CAST('[]' AS json) AS labels
		FROM
			stories s
	`

	// Add team member join if needed for assigned_to_me or created_by_me filters
	needsTeamJoin := filters.AssignedToMe != nil || filters.CreatedByMe != nil
	if needsTeamJoin {
		query += `
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user_id
		`
	}

	// Add label join if needed for filtering (but don't fetch label data)
	if len(filters.LabelIDs) > 0 {
		query += `
			INNER JOIN story_labels sl_filter ON sl_filter.story_id = s.id
		`
	}

	// Build WHERE clauses (same as buildStoriesQuery)
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.deleted_at IS NULL",
		"s.parent_id IS NULL",
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if len(filters.LabelIDs) > 0 {
		whereClauses = append(whereClauses, "sl_filter.label_id = ANY(:label_ids)")
	}

	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.Epic != nil {
		whereClauses = append(whereClauses, "s.epic_id = :epic_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	if filters.AssignedToMe != nil && *filters.AssignedToMe {
		whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
	}

	if filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
	}

	query += " WHERE " + strings.Join(whereClauses, " AND ")

	return query
}

