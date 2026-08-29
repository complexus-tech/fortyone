package developercredentialsdomain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type AuditEvent struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	Actor           platformauth.Actor
	Operation       string
	SubjectType     string
	SubjectID       uuid.UUID
	Result          string
	ReasonCode      string
	RequestID       string
	ScopeCount      int
	TeamCount       int
	WorkspaceRole   authorization.WorkspaceRole
	RotatedFromID   *uuid.UUID
	RotationOverlap time.Duration
	CreatedAt       time.Time
}

type CreatePersonalToken struct {
	PrincipalCandidateID uuid.UUID
	Credential           CredentialMaterial
	WorkspaceID          uuid.UUID
	UserID               uuid.UUID
	Name                 string
	Scopes               []platformauth.Scope
	TeamIDs              []uuid.UUID
	ExpiresAt            time.Time
	CreatedAt            time.Time
	Audit                AuditEvent
}

type EnsureHumanPrincipal struct {
	PrincipalCandidateID uuid.UUID
	WorkspaceID          uuid.UUID
	UserID               uuid.UUID
	CreatedAt            time.Time
	Audit                AuditEvent
}

type CreateServiceAccount struct {
	PrincipalID   uuid.UUID
	WorkspaceID   uuid.UUID
	ActorUserID   uuid.UUID
	Name          string
	WorkspaceRole authorization.WorkspaceRole
	CreatedAt     time.Time
	Audit         AuditEvent
}

type CreateServiceAccountKey struct {
	Credential  CredentialMaterial
	WorkspaceID uuid.UUID
	PrincipalID uuid.UUID
	ActorUserID uuid.UUID
	Name        string
	Scopes      []platformauth.Scope
	TeamIDs     []uuid.UUID
	ExpiresAt   time.Time
	CreatedAt   time.Time
	Audit       AuditEvent
}

type RotateCredential struct {
	OldCredentialID uuid.UUID
	NewCredential   CredentialMaterial
	WorkspaceID     uuid.UUID
	PrincipalID     uuid.UUID
	ActorUserID     uuid.UUID
	ExpiresAt       time.Time
	OverlapUntil    time.Time
	RotatedAt       time.Time
	Audit           AuditEvent
}

type RevokeCredential struct {
	CredentialID uuid.UUID
	WorkspaceID  uuid.UUID
	PrincipalID  uuid.UUID
	ActorUserID  uuid.UUID
	Reason       string
	RevokedAt    time.Time
	Audit        AuditEvent
}

type DisableServiceAccount struct {
	PrincipalID uuid.UUID
	WorkspaceID uuid.UUID
	ActorUserID uuid.UUID
	Reason      string
	DisabledAt  time.Time
	Audit       AuditEvent
}
