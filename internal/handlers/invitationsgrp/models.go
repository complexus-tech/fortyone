package invitationsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/google/uuid"
)

// Request Models
type AppNewInvitationBulk struct {
	Invitations []AppNewInvitation `json:"invitations" validate:"required,dive"`
}

type AppNewInvitation struct {
	Email   string      `json:"email" validate:"required,email"`
	Role    string      `json:"role" validate:"required"`
	TeamIDs []uuid.UUID `json:"teamIds,omitempty"`
}

// Response Models
type AppInvitation struct {
	ID          uuid.UUID   `json:"id"`
	WorkspaceID uuid.UUID   `json:"workspaceId"`
	InviterID   uuid.UUID   `json:"inviterId"`
	Email       string      `json:"email"`
	Role        string      `json:"role"`
	TeamIDs     []uuid.UUID `json:"teamIds,omitempty"`
	ExpiresAt   time.Time   `json:"expiresAt"`
	UsedAt      *time.Time  `json:"usedAt,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// Conversion functions
func toAppInvitation(i invitations.CoreWorkspaceInvitation) AppInvitation {
	return AppInvitation{
		ID:          i.ID,
		WorkspaceID: i.WorkspaceID,
		InviterID:   i.InviterID,
		Email:       i.Email,
		Role:        i.Role,
		TeamIDs:     i.TeamIDs,
		ExpiresAt:   i.ExpiresAt,
		UsedAt:      i.UsedAt,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func toAppInvitations(is []invitations.CoreWorkspaceInvitation) []AppInvitation {
	result := make([]AppInvitation, len(is))
	for i, inv := range is {
		result[i] = toAppInvitation(inv)
	}
	return result
}
