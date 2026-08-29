package invitationsrepository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func lockInvitationWorkspaceAdmins(
	ctx context.Context,
	queries invitationsql.Querier,
	actorID uuid.UUID,
	commands []invitationsdomain.NewWorkspaceInvitation,
) error {
	workspaces := make(map[uuid.UUID]struct{}, len(commands))
	for _, command := range commands {
		invitation := command.Invitation
		if invitation.InviterID != actorID {
			return authorization.ErrWorkspaceAdminRequired
		}
		workspaces[invitation.WorkspaceID] = struct{}{}
	}

	workspaceIDs := make([]uuid.UUID, 0, len(workspaces))
	for workspaceID := range workspaces {
		workspaceIDs = append(workspaceIDs, workspaceID)
	}
	sort.Slice(workspaceIDs, func(i, j int) bool {
		return workspaceIDs[i].String() < workspaceIDs[j].String()
	})
	for _, workspaceID := range workspaceIDs {
		if err := lockActiveWorkspaceAdmin(ctx, queries, workspaceID, actorID); err != nil {
			return err
		}
	}
	return nil
}

func lockActiveWorkspaceAdmin(
	ctx context.Context,
	queries invitationsql.Querier,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) error {
	_, err := queries.LockActiveWorkspaceAdmin(ctx, invitationsql.LockActiveWorkspaceAdminParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return authorization.ErrWorkspaceAdminRequired
	}
	if err != nil {
		return fmt.Errorf("lock active workspace administrator: %w", err)
	}
	return nil
}
