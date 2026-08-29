package developercredentialsdomain

import (
	"log/slog"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type CredentialKind string

const (
	CredentialPersonalAccessToken CredentialKind = "personal_access_token"
	CredentialServiceAccountKey   CredentialKind = "service_account_key"
)

func (kind CredentialKind) Validate() error {
	switch kind {
	case CredentialPersonalAccessToken, CredentialServiceAccountKey:
		return nil
	default:
		return ErrInvalidCredentialKind
	}
}

type PrincipalStatus string

const (
	PrincipalActive   PrincipalStatus = "active"
	PrincipalDisabled PrincipalStatus = "disabled"
)

type ServiceAccount struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	Name           string
	WorkspaceRole  authorization.WorkspaceRole
	Status         PrincipalStatus
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DisabledAt     *time.Time
	DisabledBy     *uuid.UUID
	DisabledReason *string
}

type Credential struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	PrincipalID   uuid.UUID
	Kind          CredentialKind
	Name          string
	LookupPrefix  string
	TokenVersion  int16
	Scopes        []platformauth.Scope
	TeamIDs       []uuid.UUID
	ExpiresAt     time.Time
	LastUsedAt    *time.Time
	RotatedFromID *uuid.UUID
	RotatedAt     *time.Time
	RevokedAt     *time.Time
	RevokedBy     *uuid.UUID
	RevokedReason *string
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
}

// PlaintextToken is deliberately redacted from formatting and structured
// logging. HTTP adapters must call Reveal only for a create/rotate response.
type PlaintextToken struct {
	value string
}

func NewPlaintextToken(value string) PlaintextToken {
	return PlaintextToken{value: value}
}

func (token PlaintextToken) Reveal() string {
	return token.value
}

func (PlaintextToken) String() string {
	return "[REDACTED]"
}

func (PlaintextToken) GoString() string {
	return "developercredentialsdomain.PlaintextToken{[REDACTED]}"
}

func (PlaintextToken) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

func (PlaintextToken) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

type IssuedCredential struct {
	Credential Credential
	Token      PlaintextToken
}

type Access struct {
	Actor         platformauth.Actor
	WorkspaceID   uuid.UUID
	WorkspaceRole authorization.WorkspaceRole
}

type DigestKeyRef struct {
	ID      string
	Version uint32
}

type CredentialMaterial struct {
	ID           uuid.UUID
	Kind         CredentialKind
	LookupPrefix string
	SecretDigest []byte
	TokenVersion int16
	DigestKey    DigestKeyRef
}

type VerificationRecord struct {
	CredentialID    uuid.UUID
	WorkspaceID     uuid.UUID
	PrincipalRecord uuid.UUID
	PrincipalKind   string
	SubjectUserID   *uuid.UUID
	WorkspaceRole   authorization.WorkspaceRole
	CredentialKind  CredentialKind
	LookupPrefix    string
	SecretDigest    []byte
	TokenVersion    int16
	DigestKey       DigestKeyRef
	Scopes          []platformauth.Scope
	TeamIDs         []uuid.UUID
	ExpiresAt       time.Time
	LastUsedAt      *time.Time
}
