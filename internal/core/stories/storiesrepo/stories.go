package storiesrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

// Create creates a new story.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	q := `
	INSERT INTO stories (
			id, sequence_id, title, description, description_html, parent_id, objective_id,
			status_id, assignee_id, blocked_by_id, blocking_id, related_id, reporter_id,
			priority, sprint_id, team_id, workspace_id, start_date, end_date, created_at, updated_at
	) VALUES (
			:id, :sequence_id, :title, :description, :description_html, :parent_id, :objective_id,
			:status_id, :assignee_id, :blocked_by_id, :blocking_id, :related_id, :reporter_id,
			:priority, :sprint_id, :team_id, :workspace_id, :start_date, :end_date, :created_at, :updated_at
	);
`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Creating story.")
	if _, err := stmt.ExecContext(ctx, toDBStory(*story)); err != nil {
		errMsg := fmt.Sprintf("Failed to create story: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create story"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "Story created successfully.")
	span.AddEvent("Story created.", trace.WithAttributes(
		attribute.String("story.title", story.Title),
	))

	return nil
}

// GetNextSequenceID returns the next sequence ID for a team.
func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID) (int, error) {
	var currentSequence int

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback in case of error

	// Try to update the existing record and get the new sequence
	query := `
		UPDATE team_story_sequences
		SET current_sequence = current_sequence + 1
		WHERE team_id = :team_id
		RETURNING current_sequence
	`
	params := map[string]interface{}{
		"team_id": teamID,
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &currentSequence, params)
	if err == sql.ErrNoRows {
		// If no record exists, insert a new one starting from 1
		q := `
			INSERT INTO team_story_sequences (team_id, current_sequence)
			VALUES (:team_id, 1)
			RETURNING current_sequence
		`
		stmt, err = tx.PrepareNamedContext(ctx, q)
		if err != nil {
			return 0, fmt.Errorf("failed to prepare named statement for insert: %w", err)
		}
		defer stmt.Close()

		err = stmt.GetContext(ctx, &currentSequence, params)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get/update sequence: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return currentSequence, nil
}

// TeamStories returns a list of stories for a team.
func (r *repo) TeamStories(ctx context.Context, teamId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.TeamStories")
	defer span.End()

	var stories []dbStory
	q := `
		SELECT
			id,
			sequence_id,
			title,
			priority,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories;
	`

	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching stories.")
	if err := stmt.SelectContext(ctx, &stories); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Stories retrieved successfully.")
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// ObjectiveStories returns a list of stories for an objective.
func (r *repo) ObjectiveStories(ctx context.Context, objectiveId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ObjectiveStories")
	defer span.End()
	var stories []dbStory
	q := `
		SELECT
			id,
			sequence_id,
			title,
			priority,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories
		WHERE
			objective_id = :objective_id;
	`
	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching stories for objective #%s", objectiveId), "objectiveID", objectiveId)
	if err := stmt.SelectContext(ctx, &stories, objectiveId); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories for objective #%s retrieved successfully", objectiveId), "objectiveID", objectiveId)
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// SprintStories returns a list of stories for a sprint.
func (r *repo) SprintStories(ctx context.Context, sprintId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.SprintStories")
	defer span.End()
	var stories []dbStory
	q := `
		SELECT
			id,
			sequence_id,
			title,
			priority,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories
		WHERE
			sprint_id = :sprint_id;
	`
	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching stories for sprint #%s", sprintId), "sprintId", sprintId)
	if err := stmt.SelectContext(ctx, &stories, sprintId); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories for sprint #%s retrieved successfully", sprintId), "sprintId", sprintId)
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// EpicStories returns a list of stories for an epic.
func (r *repo) EpicStories(ctx context.Context, epicId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.EpicStories")
	defer span.End()
	var stories []dbStory
	q := `
		SELECT
			id,
			sequence_id,
			title,
			priority,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories
		WHERE
			epic_id = :epic_id;
	`
	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching stories for epic #%s", epicId), "epicId", epicId)
	if err := stmt.SelectContext(ctx, &stories, epicId); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories for epic #%s retrieved successfully", epicId), "epicId", epicId)
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// MyStories returns a list of stories.
func (r *repo) MyStories(ctx context.Context) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
	defer span.End()

	var stories []dbStory
	// filter where assignee_id = current user or reporter_id = current user or current user in in the watchers list
	q := `
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
			created_at,
			updated_at
		FROM
			stories
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC;
	`

	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching stories.")
	if err := stmt.SelectContext(ctx, &stories); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Stories retrieved successfully.")
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// Get returns the story with the specified ID.
func (r *repo) Get(ctx context.Context, id uuid.UUID) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Get")
	defer span.End()

	params := map[string]interface{}{"id": id}
	var story dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, `
		SELECT
			id,
			title,
			priority,
			sequence_id,
			status_id,
			description,
			description_html,
			team_id,
			objective_id,
			sprint_id,
			workspace_id,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories
		WHERE
			id = :id;
	`)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", id)
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.String("story.id", id.String())))
		return stories.CoreSingleStory{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching story #%s", id), "id", id)
	err = stmt.GetContext(ctx, &story, params)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to retrieve story from database: %s", err), "id", id)
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.String("story.id", id.String())))
		return stories.CoreSingleStory{}, err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s retrieved successfully", id), "id", id)
	span.AddEvent("Story retrieved.", trace.WithAttributes(attribute.String("story.id", id.String())))
	return toCoreStory(story), nil
}

// Delete deletes the story with the specified ID.
func (r *repo) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Delete")
	defer span.End()
	params := map[string]interface{}{"id": id}

	stmt, err := r.db.PrepareNamedContext(ctx, `
		UPDATE stories 
		SET deleted_at = NOW(),
				updated_at = NOW() 
		WHERE id = :id;
	`)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting story #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to delete story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s deleted successfully", id), "id", id)
	span.AddEvent("Story deleted.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkDelete removes the stories with the specified IDs.
func (r *repo) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkDelete")
	defer span.End()

	query := `
        UPDATE stories 
        SET deleted_at = NOW(), updated_at = NOW() 
        WHERE id = ANY($1)
    `

	r.log.Info(ctx, fmt.Sprintf("Deleting stories: %v", ids), "ids", ids)

	_, err := r.db.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to delete stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v deleted successfully", ids), "ids", ids)
	span.AddEvent("Stories deleted.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// Restore rrestores a story with the specified ID.
func (r *repo) Restore(ctx context.Context, id uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Restore")
	defer span.End()

	query := `
			UPDATE stories 
			SET deleted_at = NULL, 
					updated_at = NOW() 
			WHERE id = :id
	`
	params := map[string]interface{}{"id": id}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare restore statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Restoring story #%s", id), "id", id)
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to restore story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s restored successfully", id), "id", id)
	span.AddEvent("Story restored.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (r *repo) BulkRestore(ctx context.Context, ids []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkRestore")
	defer span.End()

	query := `
				UPDATE stories
				SET deleted_at = NULL, updated_at = NOW()
				WHERE id = ANY($1)
			`

	r.log.Info(ctx, fmt.Sprintf("Restoring stories: %v", ids), "ids", ids)
	_, err := r.db.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to restore stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v restored successfully", ids), "ids", ids)
	span.AddEvent("Stories restored.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// Update updates the story with the specified ID.
func (r *repo) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Update")
	defer span.End()

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"id": id}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id"

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
func (r *repo) BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUpdate")
	defer span.End()

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"ids": ids}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id IN (:ids)"

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
