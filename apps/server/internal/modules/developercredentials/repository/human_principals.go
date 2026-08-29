package developercredentialsrepository

import (
	"context"
	"errors"
	"fmt"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) EnsureHumanPrincipal(
	ctx context.Context,
	command developercredentialsdomain.EnsureHumanPrincipal,
) (uuid.UUID, error) {
	var principalID uuid.UUID
	err := store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		var err error
		principalID, err = queries.EnsureHumanPrincipal(ctx, developercredentialssql.EnsureHumanPrincipalParams{
			UserID: command.UserID, WorkspaceID: command.WorkspaceID,
			PrincipalID: command.PrincipalCandidateID, CreatedAt: command.CreatedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrAccessDenied
		}
		if err != nil {
			return fmt.Errorf("ensure human principal: %w", err)
		}
		command.Audit.SubjectID = principalID
		return insertAuditEvent(ctx, queries, command.Audit)
	})
	return principalID, err
}

func (store *Store) ResolveHumanPrincipal(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (uuid.UUID, error) {
	principalID, err := store.queries.ResolveHumanPrincipal(ctx, developercredentialssql.ResolveHumanPrincipalParams{
		WorkspaceID: workspaceID,
		UserID:      uuidPointer(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, developercredentialsdomain.ErrAccessDenied
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve human principal: %w", err)
	}
	return principalID, nil
}
