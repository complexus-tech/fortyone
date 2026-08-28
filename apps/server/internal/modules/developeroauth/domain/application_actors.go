package developeroauthdomain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// ClientSecret is non-sensitive lifecycle metadata. The plaintext value is
// represented only by PlaintextSecret and is returned once on create/rotate.
type ClientSecret struct {
	ID               uuid.UUID
	ApplicationID    uuid.UUID
	LookupPrefix     string
	ExpiresAt        time.Time
	LastUsedAt       *time.Time
	RotatedFromID    *uuid.UUID
	OverlapExpiresAt *time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

type IssuedClientSecret struct {
	Secret                         ClientSecret
	Plaintext                      PlaintextSecret
	PreviousSecretOverlapExpiresAt *time.Time
}

type ApplicationAccessToken struct {
	AccessToken PlaintextSecret
	ExpiresIn   time.Duration
	Scopes      []string
}

// AuthenticateApplicationCredential identifies the exact secret candidate and
// workspace installation participating in a client_credentials exchange. The
// repository appends the immutable issuance audit in the same transaction that
// authenticates and touches the installation. AccessTokenID is the JWT jti and
// is audit-only; InstallationID remains the stable credential identity.
type AuthenticateApplicationCredential struct {
	LookupPrefix    string
	InstallationID  uuid.UUID
	AccessTokenID   uuid.UUID
	AuditID         uuid.UUID
	RequestID       string
	AuthenticatedAt time.Time
}

// ClientSecretRecord is the narrow verification projection. Repositories
// return at most one active prefix candidate; the service always performs the
// keyed constant-time digest check before issuing an access token.
type ClientSecretRecord struct {
	Secret      ClientSecret
	ClientID    string
	Material    SecretMaterial
	Application Application
}

// ApplicationInstallation is a workspace-owned authorization grant for one
// confidential application. PrincipalID is the runtime actor. InstalledBy is
// immutable audit metadata and must never become the runtime principal.
type ApplicationInstallation struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	ClientID      string
	WorkspaceID   uuid.UUID
	PrincipalID   uuid.UUID
	Resource      string
	Scopes        []string
	Status        string
	InstalledBy   uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastUsedAt    *time.Time
	RevokedAt     *time.Time
	RevokedBy     *uuid.UUID
	RevokedReason *string
}

func (installation ApplicationInstallation) AccessIdentity(expiresAt time.Time, tokenID uuid.UUID) AccessIdentity {
	return AccessIdentity{
		PrincipalID: installation.PrincipalID, ApplicationID: installation.ApplicationID,
		InstallationID: installation.ID, WorkspaceID: installation.WorkspaceID,
		ActorKind: platformauth.PrincipalOAuthApplication, ClientID: installation.ClientID,
		Resource: installation.Resource, Scopes: append([]string(nil), installation.Scopes...),
		ExpiresAt: expiresAt, OAuthCredential: tokenID,
	}
}
