package invitationsrepository

import (
	"encoding/json"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/google/uuid"
)

type dbWorkspaceInvitation struct {
	ID             uuid.UUID       `db:"invitation_id"`
	WorkspaceID    uuid.UUID       `db:"workspace_id"`
	InviterID      uuid.UUID       `db:"inviter_id"`
	Email          string          `db:"email"`
	Role           string          `db:"role"`
	Token          string          `db:"token"`
	TeamIDs        json.RawMessage `db:"team_ids"`
	ExpiresAt      time.Time       `db:"expires_at"`
	UsedAt         *time.Time      `db:"used_at"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
	WorkspaceName  string          `db:"workspace_name"`
	WorkspaceSlug  string          `db:"workspace_slug"`
	WorkspaceColor string          `db:"workspace_color"`
}

// Conversion functions
func toCoreInvitation(i dbWorkspaceInvitation) invitations.CoreWorkspaceInvitation {
	var teamIDs []uuid.UUID
	if i.TeamIDs != nil {
		_ = json.Unmarshal(i.TeamIDs, &teamIDs)
	}

	return invitations.CoreWorkspaceInvitation{
		ID:             i.ID,
		WorkspaceID:    i.WorkspaceID,
		InviterID:      i.InviterID,
		Email:          i.Email,
		Role:           i.Role,
		Token:          i.Token,
		TeamIDs:        teamIDs,
		ExpiresAt:      i.ExpiresAt,
		UsedAt:         i.UsedAt,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
		WorkspaceName:  i.WorkspaceName,
		WorkspaceSlug:  i.WorkspaceSlug,
		WorkspaceColor: i.WorkspaceColor,
	}
}

func toCoreInvitations(invites []dbWorkspaceInvitation) []invitations.CoreWorkspaceInvitation {
	result := make([]invitations.CoreWorkspaceInvitation, len(invites))
	for i, invite := range invites {
		result[i] = toCoreInvitation(invite)
	}
	return result
}
