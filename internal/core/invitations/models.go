package invitations

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service errors
var (
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationExpired   = errors.New("invitation expired")
	ErrInvitationUsed      = errors.New("invitation already used")
	ErrInvitationRevoked   = errors.New("invitation has been revoked")
	ErrInvalidToken        = errors.New("invalid invitation token")
	ErrDuplicateInvitation = errors.New("duplicate invitation")
	ErrInvalidInvitee      = errors.New("user email does not match invitation email")
)

// InvitationRequest represents a request to create an invitation
type InvitationRequest struct {
	Email   string
	Role    string
	TeamIDs []uuid.UUID
}

// CoreWorkspaceInvitation represents a workspace invitation in the application layer
type CoreWorkspaceInvitation struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	InviterID      uuid.UUID
	Email          string
	Role           string
	Token          string
	TeamIDs        []uuid.UUID
	ExpiresAt      time.Time
	UsedAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	WorkspaceName  string // Workspace details
	WorkspaceSlug  string
	WorkspaceColor string
}
