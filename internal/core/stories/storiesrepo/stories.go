package storiesrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/comments/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/links/linksrepo"
	"github.com/complexus-tech/projects-api/internal/core/stories"
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
	params := map[string]interface{}{
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
			VALUES (:workspace_id, :team_id, 1)
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
			) RETURNING stories.*;
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

	params := map[string]interface{}{
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
	params := map[string]interface{}{
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
		args := make([]interface{}, 0, len(labels)*2)
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
func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
	defer span.End()

	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.description,
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
	`
	var setClauses []string

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	filters["workspace_id"] = workspaceId

	query += " WHERE " + strings.Join(setClauses, " AND ") + " AND deleted_at IS NULL AND workspace_id = :workspace_id;"

	var stories []dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "fetching stories.")
	if err := stmt.SelectContext(ctx, &stories, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "stories retrieved successfully.")
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", query),
	))

	return toCoreStories(stories), nil
}

// MyStories returns a list of stories.
func (r *repo) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
	defer span.End()

	currentUser, _ := mid.GetUserID(ctx)

	params := map[string]interface{}{
		"workspace_id": workspaceId,
		"current_user": currentUser,
	}

	var stories []dbStory
	// TODO: Add check for current user in watchers list

	q := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.description,
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
		WHERE s.workspace_id = :workspace_id AND s.deleted_at IS NULL 
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

	params := map[string]interface{}{
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
	params := map[string]interface{}{"id": id, "workspace_id": workspaceId}

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

	params := map[string]interface{}{"ids": ids, "workspace_id": workspaceId}

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
	params := map[string]interface{}{"id": id, "workspace_id": workspaceId}

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

	params := map[string]interface{}{"ids": ids, "workspace_id": workspaceId}

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
		params := map[string]interface{}{
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
	params := map[string]interface{}{
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
	params := map[string]interface{}{"story_id": storyID}

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

	params := map[string]interface{}{"story_id": storyID}

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

func (r *repo) validateStatusTeam(ctx context.Context, statusId, teamId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.validateStatusTeam")
	defer span.End()

	q := `
		SELECT EXISTS (
			SELECT 1 FROM story_statuses 
			WHERE status_id = :status_id 
			AND team_id = :team_id
		)
	`
	params := map[string]interface{}{
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
