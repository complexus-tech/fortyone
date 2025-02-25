package invitations

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service errors
var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationUsed     = errors.New("invitation has already been used")
	ErrInvitationRevoked  = errors.New("invitation has been revoked")
	ErrInvalidToken       = errors.New("invalid invitation token")
	ErrMaxUsesReached     = errors.New("invitation link has reached maximum uses")
)

// CoreWorkspaceInvitation represents a workspace invitation in the application layer
type CoreWorkspaceInvitation struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	InviterID   uuid.UUID
	Email       string
	Role        string
	Token       string
	ExpiresAt   time.Time
	UsedAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CoreWorkspaceInvitationLink represents a workspace invitation link in the application layer
type CoreWorkspaceInvitationLink struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	CreatorID   uuid.UUID
	Token       string
	Role        string
	MaxUses     *int
	UsedCount   int
	ExpiresAt   *time.Time
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
