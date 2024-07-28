package statesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/states"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]states.CoreState, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.states.List")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
	}

	var statuses []dbState
	q := `
		SELECT
			status_id,
			name,
			color,
			category,
			order_index,
			team_id,
			workspace_id,
			created_at,
			updated_at
		FROM
			statuses
		WHERE workspace_id = :workspace_id
		ORDER BY order_index ASC;
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching statuses.")
	if err := stmt.SelectContext(ctx, &statuses, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve statuses from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("statuses not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Statuses retrieved successfully.")
	span.AddEvent("Statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
		attribute.String("query", q),
	))

	return toCoreStates(statuses), nil
}
