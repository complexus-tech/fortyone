package issuesrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/issues"
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

// MyIssues returns a list of issues.
func (r *repo) MyIssues(ctx context.Context) ([]issues.CoreIssueList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.issues.List")
	defer span.End()

	myIssues := []dbIssue{
		{ID: uuid.New(), Title: "Issue 1", Description: "This is issue 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 2", Description: "This is issue 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 3", Description: "This is issue 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 4", Description: "This is issue 4", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 5", Description: "This is issue 5", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 6", Description: "This is issue 6", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 7", Description: "This is issue 7", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 8", Description: "This is issue 8", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 9", Description: "This is issue 9", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Title: "Issue 10", Description: "This is issue 10", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	var issues []dbIssue
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

	// stmt, err := r.db.PreparexContext(ctx, q)
	// if err != nil {
	// 	errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
	// 	r.log.Error(ctx, errMsg)
	// 	span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
	// 	return nil, err
	// }
	// defer stmt.Close()

	// r.log.Info(ctx, "Fetching issues.")
	// if err := stmt.SelectContext(ctx, &issues); err != nil {
	// 	errMsg := fmt.Sprintf("Failed to retrieve issues from the database: %s", err)
	// 	r.log.Error(ctx, errMsg)
	// 	span.RecordError(errors.New("issues not found"), trace.WithAttributes(attribute.String("error", errMsg)))
	// 	return nil, err
	// }

	r.log.Info(ctx, "Issues retrieved successfully.")
	span.AddEvent("Issues retrieved.", trace.WithAttributes(
		attribute.Int("issue.count", len(issues)),
		attribute.String("query", q),
	))

	return toCoreIssues(myIssues), nil
}

// Get returns the issue with the specified ID.
func (r *repo) Get(ctx context.Context, id int) (issues.CoreSingleIssue, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.issues.Get")
	defer span.End()

	params := map[string]interface{}{"id": id}
	var issue dbIssue

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
		return issues.CoreSingleIssue{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Fetching issue #%d", id), "id", id)
	err = stmt.GetContext(ctx, &issue, params)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to retrieve issue from database: %s", err), "id", id)
		span.RecordError(errors.New("issue not found"), trace.WithAttributes(attribute.Int("issue.id", id)))
		return issues.CoreSingleIssue{}, err
	}

	r.log.Info(ctx, fmt.Sprintf("Issue #%d retrieved successfully", id), "id", id)
	span.AddEvent("Issue retrieved.", trace.WithAttributes(attribute.Int("issue.id", id)))
	return toCoreIssue(issue), nil
}
