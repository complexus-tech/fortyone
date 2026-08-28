package invitationsrepository

import (
	"time"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/google/uuid"
)

func toCoreInvitation(
	id uuid.UUID,
	workspaceID uuid.UUID,
	inviterID uuid.UUID,
	email string,
	role invitationsql.UserRole,
	teamIDs []uuid.UUID,
	expiresAt, createdAt, updatedAt time.Time,
	usedAt *time.Time,
	workspaceName, workspaceSlug, workspaceColor string,
) invitationsdomain.WorkspaceInvitation {
	return invitationsdomain.WorkspaceInvitation{
		ID:             id,
		WorkspaceID:    workspaceID,
		InviterID:      inviterID,
		Email:          email,
		Role:           string(role),
		TeamIDs:        append([]uuid.UUID(nil), teamIDs...),
		ExpiresAt:      expiresAt,
		UsedAt:         usedAt,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		WorkspaceName:  workspaceName,
		WorkspaceSlug:  workspaceSlug,
		WorkspaceColor: workspaceColor,
	}
}

func invitationFromGet(row invitationsql.GetInvitationByTokenRow) invitationsdomain.WorkspaceInvitation {
	return toCoreInvitation(
		row.InvitationID, row.WorkspaceID, row.InviterID, row.Email, row.Role, row.TeamIds,
		row.ExpiresAt, row.CreatedAt, row.UpdatedAt, row.UsedAt,
		row.WorkspaceName, row.WorkspaceSlug, row.WorkspaceColor,
	)
}

func invitationFromLock(row invitationsql.LockInvitationByTokenRow) invitationsdomain.WorkspaceInvitation {
	return toCoreInvitation(
		row.InvitationID, row.WorkspaceID, row.InviterID, row.Email, row.Role, row.TeamIds,
		row.ExpiresAt, row.CreatedAt, row.UpdatedAt, row.UsedAt,
		row.WorkspaceName, row.WorkspaceSlug, row.WorkspaceColor,
	)
}

func invitationFromWorkspaceList(row invitationsql.ListWorkspaceInvitationsRow) invitationsdomain.WorkspaceInvitation {
	return toCoreInvitation(
		row.InvitationID, row.WorkspaceID, row.InviterID, row.Email, row.Role, row.TeamIds,
		row.ExpiresAt, row.CreatedAt, row.UpdatedAt, row.UsedAt,
		row.WorkspaceName, row.WorkspaceSlug, row.WorkspaceColor,
	)
}

func invitationFromEmailList(row invitationsql.ListInvitationsByEmailRow) invitationsdomain.WorkspaceInvitation {
	return toCoreInvitation(
		row.InvitationID, row.WorkspaceID, row.InviterID, row.Email, row.Role, row.TeamIds,
		row.ExpiresAt, row.CreatedAt, row.UpdatedAt, row.UsedAt,
		row.WorkspaceName, row.WorkspaceSlug, row.WorkspaceColor,
	)
}
