package storiesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
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

// MyStories returns a list of stories.
func (r *repo) MyStories(ctx context.Context) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
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

// Get returns the story with the specified ID.
func (r *repo) Get(ctx context.Context, id string) (stories.CoreSingleStory, error) {
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
			description,
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
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.String("story.id", id)))
		return stories.CoreSingleStory{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching story #%s", id), "id", id)
	err = stmt.GetContext(ctx, &story, params)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to retrieve story from database: %s", err), "id", id)
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.String("story.id", id)))
		return stories.CoreSingleStory{}, err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s retrieved successfully", id), "id", id)
	span.AddEvent("Story retrieved.", trace.WithAttributes(attribute.String("story.id", id)))
	return toCoreStory(story), nil
}
