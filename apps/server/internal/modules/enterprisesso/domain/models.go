package enterprisessodomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrForbidden    = errors.New("workspace SSO access denied")
	ErrInvalid      = errors.New("invalid workspace SSO input")
	ErrNotFound     = errors.New("workspace SSO is not configured")
	ErrConflict     = errors.New("workspace SSO changed; refresh and retry")
	ErrProvider     = errors.New("identity provider sign-in failed")
	ErrLinkRequired = errors.New("sign in to your existing FortyOne account and connect workspace SSO first")
)

type Scope struct{ ActorID, WorkspaceID uuid.UUID }
type Connection struct {
	ID             uuid.UUID `json:"id"`
	WorkspaceID    uuid.UUID `json:"workspaceId"`
	Issuer         string    `json:"issuer"`
	ClientID       string    `json:"clientId"`
	SecretEnvelope string    `json:"-"`
	Enabled        bool      `json:"enabled"`
	RequireSSO     bool      `json:"requireSSO"`
	Generation     int64     `json:"generation"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type Create struct {
	Issuer       string `json:"issuer"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}
type Update struct {
	Enabled         bool    `json:"enabled"`
	RequireSSO      bool    `json:"requireSSO"`
	ClientSecret    *string `json:"clientSecret"`
	ExpectedVersion int64   `json:"expectedVersion"`
}
type SessionProof struct {
	ConnectionID    uuid.UUID
	Generation      int64
	AuthenticatedAt time.Time
}
type Attempt struct {
	WorkspaceID, ConnectionID, LinkUserID uuid.UUID
	Generation                            int64
	Slug, Nonce, Verifier                 string
	CreatedAt                             time.Time
	ReturnToSettings                      bool
}
type Identity struct {
	Subject, Email  string
	AuthenticatedAt time.Time
}
