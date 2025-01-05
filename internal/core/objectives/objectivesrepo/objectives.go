package objectivesrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.List")
	defer span.End()

	var objectives []dbObjective
	q := `
		SELECT
			objective_id,
			name,
			description,
			lead_user_id,
			team_id,
			workspace_id,
			start_date,
			end_date,
			is_private,
			created_at,
			updated_at
		FROM
			objectives
	`
	var setClauses []string
	filters["workspace_id"] = workspaceId

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	q += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY created_at DESC;"

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objectives.")
	if err := stmt.SelectContext(ctx, &objectives, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objectives from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("objectives not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "objectives retrieved successfully.")
	span.AddEvent("objectives retrieved.", trace.WithAttributes(
		attribute.Int("objectives.count", len(objectives)),
	))

	return toCoreObjectives(objectives), nil
}
