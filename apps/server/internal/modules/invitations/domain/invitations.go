package invitationsdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvitationNotFound     = errors.New("invitation not found")
	ErrInvitationExpired      = errors.New("invitation expired")
	ErrInvitationUsed         = errors.New("invitation already accepted")
	ErrInvitationRevoked      = errors.New("invitation has been revoked")
	ErrInvalidToken           = errors.New("invalid invitation token")
	ErrDuplicateInvitation    = errors.New("duplicate invitation")
	ErrInvalidInvitee         = errors.New("user email does not match invitation email")
	ErrAlreadyWorkspaceMember = errors.New("you are already a member of this workspace")
	ErrInvalidInvitationRole  = errors.New("invalid invitation role")
	ErrInvalidInvitationEmail = errors.New("invalid invitation email")
	ErrInvalidInvitationTeam  = errors.New("invitation team does not belong to the workspace")
	ErrTooManyInvitations     = errors.New("too many invitations in one request")
	ErrOutboxClaimLost        = errors.New("invitation outbox claim is no longer current")
)

type Request struct {
	Email   string
	Role    string
	TeamIDs []uuid.UUID
}

type StoredToken struct {
	Digest  []byte
	Nonce   []byte
	KeyID   string
	Version int16
}

type TokenLookup struct {
	Digest      []byte
	KeyID       string
	Version     int16
	LegacyToken string
}

type NewWorkspaceInvitation struct {
	Invitation  WorkspaceInvitation
	Token       StoredToken
	EmailOutbox EmailOutboxPayload
}

type EmailOutboxPayload struct {
	InviterName   string    `json:"inviter_name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	ExpiresAt     time.Time `json:"expires_at"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
}

// EmailDelivery carries raw bearer material only at the one-time delivery
// boundary. It must never be persisted, logged, or placed on a generic event.
type EmailDelivery struct {
	IdempotencyKey string
	InviterName    string
	Email          string
	Token          string
	Role           string
	ExpiresAt      time.Time
	WorkspaceID    uuid.UUID
	WorkspaceName  string
}

type AcceptCommand struct {
	Lookup     TokenLookup
	UserID     uuid.UUID
	AcceptedAt time.Time
}

type OutboxEvent struct {
	ID                  uuid.UUID
	InvitationID        uuid.UUID
	WorkspaceID         uuid.UUID
	ActorID             uuid.UUID
	EventType           string
	EventPayload        []byte
	IdempotencyKey      string
	ClaimToken          uuid.UUID
	AttemptCount        int
	CreatedAt           time.Time
	StoredToken         *StoredToken
	InvitationExpiresAt *time.Time
	InvitationUsedAt    *time.Time
}

type WorkspaceInvitation struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	InviterID      uuid.UUID
	Email          string
	Role           string
	TeamIDs        []uuid.UUID
	ExpiresAt      time.Time
	UsedAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	WorkspaceName  string
	WorkspaceSlug  string
	WorkspaceColor string
}
