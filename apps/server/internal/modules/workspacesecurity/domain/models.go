package workspacesecuritydomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrForbidden = errors.New("workspace security access denied")
	ErrInvalid   = errors.New("invalid workspace security input")
	ErrConflict  = errors.New("workspace security state changed; refresh and retry")
	ErrNotFound  = errors.New("workspace security resource not found")
	ErrLimit     = errors.New("audit export exceeds 10000 events or 20 MiB; narrow the date range")
)

type Scope struct{ ActorID, WorkspaceID uuid.UUID }
type Policy struct {
	AllowedDomains     []string   `json:"allowedDomains"`
	AllowGuests        bool       `json:"allowGuests"`
	MaxSessionAgeHours int32      `json:"maxSessionAgeHours"`
	Version            int64      `json:"version"`
	UpdatedAt          *time.Time `json:"updatedAt"`
}
type PolicyUpdate struct {
	AllowedDomains     []string `json:"allowedDomains"`
	AllowGuests        bool     `json:"allowGuests"`
	MaxSessionAgeHours int32    `json:"maxSessionAgeHours"`
	ExpectedVersion    int64    `json:"expectedVersion"`
}
type SessionIdentity struct {
	ID                         uuid.UUID
	AuthenticatedAt, ExpiresAt time.Time
}
type AccessState struct {
	Email, Role              string
	Policy                   Policy
	RevokedBefore, RevokedAt *time.Time
}
type Session struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"userId"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	AuthenticatedAt time.Time  `json:"authenticatedAt"`
	LastSeenAt      time.Time  `json:"lastSeenAt"`
	ExpiresAt       time.Time  `json:"expiresAt"`
	RevokedAt       *time.Time `json:"revokedAt"`
	RevokedBefore   *time.Time `json:"revokedBefore"`
	Current         bool       `json:"current"`
}
type SessionList struct {
	Items   []Session `json:"items"`
	HasMore bool      `json:"hasMore"`
}
type RevokeInput struct {
	Reason string `json:"reason"`
}
type AuditEvent struct {
	ID           uuid.UUID       `json:"id"`
	Source       string          `json:"source"`
	ActorID      *uuid.UUID      `json:"actorId"`
	ActorType    string          `json:"actorType"`
	ResourceType string          `json:"resourceType"`
	ResourceID   *uuid.UUID      `json:"resourceId"`
	Operation    string          `json:"operation"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"createdAt"`
}
type Position struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
	Source    string    `json:"source"`
}
type AuditFilter struct {
	ActorID      *uuid.UUID `json:"actorId"`
	ResourceType string     `json:"resourceType"`
	ResourceID   *uuid.UUID `json:"resourceId"`
	From, To     *time.Time
	Before       *Position
	Limit        int32
}
