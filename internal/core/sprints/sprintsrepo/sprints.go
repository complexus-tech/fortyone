package sprintsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error) {

	ctx, span := web.AddSpan(ctx, "business.repository.sprints.List")
	defer span.End()

	var sprints []dbSprint
	query := `
		SELECT
			sprint_id,
			name,
			goal,
			team_id,
			objective_id,
			workspace_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM
			sprints
	`

	var setClauses []string
	filters["workspace_id"] = workspaceId

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	query += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY end_date DESC;"
	fmt.Println(query)
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching sprints.")
	if err := stmt.SelectContext(ctx, &sprints, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve sprints from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("sprints not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "sprints retrieved successfully.")
	span.AddEvent("sprints retrieved.", trace.WithAttributes(
		attribute.Int("sprints.count", len(sprints)),
		attribute.String("query", query),
	))

	return toCoreSprints(sprints), nil
}
