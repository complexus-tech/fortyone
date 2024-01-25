package issues

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/web"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	List(ctx context.Context) ([]Issue, error)
	Get(ctx context.Context, id int) (Issue, error)
}

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func NewRepository(log *logger.Logger, db *sqlx.DB) Repository {
	return &repo{
		db:  db,
		log: log,
	}
}

// List returns all known issues.
func (r *repo) List(ctx context.Context) ([]Issue, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.issues.List")
	defer span.End()

	var issues []Issue
	q := `
		SELECT
			id,
			title,
			description,
			created_at,
			updated_at,
			deleted_at
		FROM
			issues;
	`

	stmt, err := r.db.PreparexContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching issues.")
	if err := stmt.SelectContext(ctx, &issues); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve issues from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("issues not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Issues retrieved successfully.")
	span.AddEvent("Issues retrieved.", trace.WithAttributes(
		attribute.Int("issue.count", len(issues)),
		attribute.String("query", q),
	))

	return issues, nil
}

// Get returns the issue with the specified ID.
func (r *repo) Get(ctx context.Context, id int) (Issue, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.issues.Get")
	defer span.End()

	params := map[string]interface{}{"id": id}
	var issue Issue

	stmt, err := r.db.PrepareNamedContext(ctx, `
		SELECT
			id, title, description, created_at, updated_at, deleted_at
		FROM
			issues
		WHERE
			id = :id;
	`)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", id)
		span.RecordError(errors.New("issue not found"), trace.WithAttributes(attribute.Int("issue.id", id)))
		return Issue{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching issue #%d", id), "id", id)
	err = stmt.GetContext(ctx, &issue, params)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to retrieve issue from database: %s", err), "id", id)
		span.RecordError(errors.New("issue not found"), trace.WithAttributes(attribute.Int("issue.id", id)))
		return Issue{}, err
	}

	r.log.Info(ctx, fmt.Sprintf("Issue #%d retrieved successfully", id), "id", id)
	span.AddEvent("Issue retrieved.", trace.WithAttributes(attribute.Int("issue.id", id)))
	return issue, nil
}
