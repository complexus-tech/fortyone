package storiesrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// MyStories returns a list of stories.
func (r *repo) MyStories(ctx context.Context) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
	defer span.End()

	myStories := []dbStory{
		{ID: uuid.New(), Title: "Story 1", Description: "This is story 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 2", Description: "This is story 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 3", Description: "This is story 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 4", Description: "This is story 4", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 5", Description: "This is story 5", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 6", Description: "This is story 6", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 7", Description: "This is story 7", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 8", Description: "This is story 8", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 9", Description: "This is story 9", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Story 10", Description: "This is story 10", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	var stories []dbStory
	q := `
		SELECT
			id,
			title,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			stories;
	`

	// stmt, err := r.db.PreparexContext(ctx, q)
	// if err != nil {
	// 	errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
	// 	r.log.Error(ctx, errMsg)
	// 	span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
	// 	return nil, err
	// }
	// defer stmt.Close()

	// r.log.Info(ctx, "Fetching stories.")
	// if err := stmt.SelectContext(ctx, &stories); err != nil {
	// 	errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
	// 	r.log.Error(ctx, errMsg)
	// 	span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
	// 	return nil, err
	// }

	r.log.Info(ctx, "Stories retrieved successfully.")
	span.AddEvent("Stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(myStories), nil
}

// Get returns the story with the specified ID.
func (r *repo) Get(ctx context.Context, id int) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Get")
	defer span.End()

	params := map[string]interface{}{"id": id}
	var story dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, `
		SELECT
			id, title, description, created_at, updated_at, deleted_at
		FROM
			stories
		WHERE
			id = :id;
	`)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", id)
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.Int("story.id", id)))
		return stories.CoreSingleStory{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching story #%d", id), "id", id)
	err = stmt.GetContext(ctx, &story, params)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to retrieve story from database: %s", err), "id", id)
		span.RecordError(errors.New("story not found"), trace.WithAttributes(attribute.Int("story.id", id)))
		return stories.CoreSingleStory{}, err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%d retrieved successfully", id), "id", id)
	span.AddEvent("Story retrieved.", trace.WithAttributes(attribute.Int("story.id", id)))
	return toCoreStory(story), nil
}
