package invitationsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/google/uuid"
)

// Request Models
type AppNewInvitation struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required"`
}

type AppNewInvitationLink struct {
	Role      string     `json:"role" validate:"required"`
	MaxUses   *int       `json:"maxUses"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// Response Models
type AppInvitation struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	InviterID   uuid.UUID  `json:"inviterId"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	UsedAt      *time.Time `json:"usedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AppInvitationLink struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	CreatorID   uuid.UUID  `json:"creatorId"`
	Token       string     `json:"token"`
	Role        string     `json:"role"`
	MaxUses     *int       `json:"maxUses"`
	UsedCount   int        `json:"usedCount"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Conversion functions
func toAppInvitation(i invitations.CoreWorkspaceInvitation) AppInvitation {
	return AppInvitation{
		ID:          i.ID,
		WorkspaceID: i.WorkspaceID,
		InviterID:   i.InviterID,
		Email:       i.Email,
		Role:        i.Role,
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

func toAppInvitationLink(l invitations.CoreWorkspaceInvitationLink) AppInvitationLink {
	return AppInvitationLink{
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

func toAppInvitationLinks(ls []invitations.CoreWorkspaceInvitationLink) []AppInvitationLink {
	result := make([]AppInvitationLink, len(ls))
	for i, link := range ls {
		result[i] = toAppInvitationLink(link)
	}
	return result
}
