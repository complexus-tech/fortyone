package workspacesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
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

// List returns a list of workspaces a user is a member of.
func (r *repo) List(ctx context.Context, userID uuid.UUID) ([]workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.List")
	defer span.End()

	var workspaces []dbWorkspace

	params := map[string]interface{}{
		"user_id": userID,
	}

	query := `
		SELECT DISTINCT
			w.workspace_id,
			w.slug,
			w.name,
			w.description,
			w.created_at,
			w.updated_at
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		WHERE
			wm.user_id = :user_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg, userID)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &workspaces, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve workspaces from the database: %s", err)
		span.RecordError(errors.New("failed to retrieve workspaces"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to retrieve workspaces %w", err)
	}

	return toCoreWorkspaces(workspaces), nil
}
