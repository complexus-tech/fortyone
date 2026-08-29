package developeroauthdomain

import (
	"log/slog"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	ScopeMCPAccess     = "mcp:access"
	ScopeOfflineAccess = "offline_access"
)

type Application struct {
	ID               uuid.UUID
	ClientID         string
	Name             string
	RegistrationKind string
	RedirectURIs     []string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type Grant struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	ClientID      string
	UserID        uuid.UUID
	ActorKind     platformauth.PrincipalKind
	Resource      string
	Scopes        []string
	CreatedAt     time.Time
	LastUsedAt    *time.Time
}

type SecretKind string

const (
	SecretAuthorizationCode SecretKind = "authorization_code"
	SecretRefreshToken      SecretKind = "refresh_token"
	SecretClientSecret      SecretKind = "client_secret"
)

type DigestKeyRef struct {
	ID string
}

type SecretMaterial struct {
	ID           uuid.UUID
	Kind         SecretKind
	LookupPrefix string
	Digest       []byte
	DigestKey    DigestKeyRef
}

// PlaintextSecret deliberately redacts itself from formatting, JSON, and
// structured logs. HTTP adapters may reveal it only in the protocol response
// that created or rotated the secret.
type PlaintextSecret struct {
	value string
}

func NewPlaintextSecret(value string) PlaintextSecret {
	return PlaintextSecret{value: value}
}

func (secret PlaintextSecret) Reveal() string {
	return secret.value
}

func (PlaintextSecret) String() string {
	return "[REDACTED]"
}

func (PlaintextSecret) GoString() string {
	return "developeroauthdomain.PlaintextSecret{[REDACTED]}"
}

func (PlaintextSecret) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

func (PlaintextSecret) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

type IssuedSecret struct {
	Plaintext PlaintextSecret
	Material  SecretMaterial
}

type AuthorizationCode struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	ClientID      string
	GrantID       uuid.UUID
	UserID        uuid.UUID
	ActorKind     platformauth.PrincipalKind
	LookupPrefix  string
	Digest        []byte
	DigestKey     DigestKeyRef
	RedirectURI   string
	Resource      string
	CodeChallenge string
	Scopes        []string
	ExpiresAt     time.Time
	ConsumedAt    *time.Time
}

type RefreshToken struct {
	ID              uuid.UUID
	FamilyID        uuid.UUID
	ParentTokenID   *uuid.UUID
	LookupPrefix    string
	Digest          []byte
	DigestKey       DigestKeyRef
	ExpiresAt       time.Time
	ConsumedAt      *time.Time
	FamilyExpiresAt time.Time
	FamilyRevokedAt *time.Time
	Grant           Grant
}

type AccessIdentity struct {
	PrincipalID     uuid.UUID
	UserID          uuid.UUID
	ApplicationID   uuid.UUID
	GrantID         uuid.UUID
	InstallationID  uuid.UUID
	WorkspaceID     uuid.UUID
	ActorKind       platformauth.PrincipalKind
	ClientID        string
	Resource        string
	Scopes          []string
	ExpiresAt       time.Time
	OAuthCredential uuid.UUID
}

type TokenPair struct {
	AccessToken  PlaintextSecret
	RefreshToken PlaintextSecret
	ExpiresIn    time.Duration
	Scopes       []string
}

type AuditEvent struct {
	ID                uuid.UUID
	ApplicationID     *uuid.UUID
	GrantID           *uuid.UUID
	InstallationID    *uuid.UUID
	PrincipalID       *uuid.UUID
	SecretID          *uuid.UUID
	WorkspaceID       *uuid.UUID
	UserID            *uuid.UUID
	ActorKind         platformauth.PrincipalKind
	ActorID           *uuid.UUID
	ActorCredentialID *uuid.UUID
	RequestID         string
	SubjectType       string
	SubjectID         *uuid.UUID
	Operation         string
	Result            string
	ReasonCode        *string
	Metadata          []byte
	CreatedAt         time.Time
}
