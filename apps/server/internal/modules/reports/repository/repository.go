package reportsrepository

import (
	"context"
	"errors"
	"fmt"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	platformauthorization "github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRepositoryNotConfigured = errors.New("reports repository is not configured")
	ErrInvalidProjection       = errors.New("invalid reports database projection")
)

type repo struct {
	queries reportssql.Querier
	log     *logger.Logger
}

type reportsActorAccess struct {
	isAdmin bool
}

func New(log *logger.Logger, pool *pgxpool.Pool) *repo {
	if pool == nil {
		return &repo{log: log}
	}

	return &repo{
		queries: reportssql.New(pool),
		log:     log,
	}
}

func (r *repo) authorize(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID) error {
	_, err := r.actorAccess(ctx, actorID, workspaceID)
	return err
}

func (r *repo) actorAccess(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID) (reportsActorAccess, error) {
	if r == nil || r.queries == nil {
		return reportsActorAccess{}, ErrRepositoryNotConfigured
	}
	if actorID == uuid.Nil || workspaceID == uuid.Nil {
		return reportsActorAccess{}, reports.ErrReportsAccessDenied
	}

	role, err := r.queries.GetReportsActorAccess(ctx, reportssql.GetReportsActorAccessParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return reportsActorAccess{}, reports.ErrReportsAccessDenied
		}
		return reportsActorAccess{}, fmt.Errorf("authorizing reports actor: %w", err)
	}

	switch platformauthorization.WorkspaceRole(role) {
	case platformauthorization.WorkspaceRoleAdmin:
		return reportsActorAccess{isAdmin: true}, nil
	case platformauthorization.WorkspaceRoleMember:
		return reportsActorAccess{}, nil
	default:
		return reportsActorAccess{}, reports.ErrReportsAccessDenied
	}
}
