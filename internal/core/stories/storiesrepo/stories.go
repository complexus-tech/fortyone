package storiesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/stories"
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

// Create creates a new story.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	q := `
	INSERT INTO stories (
			id, sequence_id, title, description, description_html, parent_id, objective_id,
			status_id, assignee_id, blocked_by_id, blocking_id, related_id, reporter_id,
			priority, sprint_id, team_id, start_date, end_date, created_at, updated_at
	) VALUES (
			:id, :sequence_id, :title, :description, :description_html, :parent_id, :objective_id,
			:status_id, :assignee_id, :blocked_by_id, :blocking_id, :related_id, :reporter_id,
			:priority, :sprint_id, :team_id, :start_date, :end_date, :created_at, :updated_at
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
	if _, err := stmt.ExecContext(ctx, story); err != nil {
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

func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID) (int, error) {
	nextSequenceID := 1
	return nextSequenceID, nil
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
