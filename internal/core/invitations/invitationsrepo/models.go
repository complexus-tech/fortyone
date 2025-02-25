package invitationsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/google/uuid"
)

type dbWorkspaceInvitation struct {
	ID          uuid.UUID   `db:"invitation_id"`
	WorkspaceID uuid.UUID   `db:"workspace_id"`
	InviterID   uuid.UUID   `db:"inviter_id"`
	Email       string      `db:"email"`
	Role        string      `db:"role"`
	Token       string      `db:"token"`
	TeamIDs     []uuid.UUID `db:"team_ids"`
	ExpiresAt   time.Time   `db:"expires_at"`
	UsedAt      *time.Time  `db:"used_at"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

type dbWorkspaceInvitationLink struct {
	ID          uuid.UUID   `db:"invitation_link_id"`
	WorkspaceID uuid.UUID   `db:"workspace_id"`
	CreatorID   uuid.UUID   `db:"creator_id"`
	Token       string      `db:"token"`
	Role        string      `db:"role"`
	TeamIDs     []uuid.UUID `db:"team_ids"`
	MaxUses     *int        `db:"max_uses"`
	UsedCount   int         `db:"used_count"`
	ExpiresAt   *time.Time  `db:"expires_at"`
	IsActive    bool        `db:"is_active"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

// Conversion functions
func toCoreInvitation(i dbWorkspaceInvitation) invitations.CoreWorkspaceInvitation {
	return invitations.CoreWorkspaceInvitation{
		ID:          i.ID,
		WorkspaceID: i.WorkspaceID,
		InviterID:   i.InviterID,
		Email:       i.Email,
		Role:        i.Role,
		Token:       i.Token,
		ExpiresAt:   i.ExpiresAt,
		UsedAt:      i.UsedAt,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func toCoreInvitations(invites []dbWorkspaceInvitation) []invitations.CoreWorkspaceInvitation {
	result := make([]invitations.CoreWorkspaceInvitation, len(invites))
	for i, invite := range invites {
		result[i] = toCoreInvitation(invite)
	}
	return result
}

func toCoreInvitationLink(l dbWorkspaceInvitationLink) invitations.CoreWorkspaceInvitationLink {
	return invitations.CoreWorkspaceInvitationLink{
		ID:          l.ID,
		WorkspaceID: l.WorkspaceID,
		CreatorID:   l.CreatorID,
		Token:       l.Token,
		Role:        l.Role,
		MaxUses:     l.MaxUses,
		UsedCount:   l.UsedCount,
		ExpiresAt:   l.ExpiresAt,
		IsActive:    l.IsActive,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func toCoreInvitationLinks(links []dbWorkspaceInvitationLink) []invitations.CoreWorkspaceInvitationLink {
	result := make([]invitations.CoreWorkspaceInvitationLink, len(links))
	for i, link := range links {
		result[i] = toCoreInvitationLink(link)
	}
	return result
}
