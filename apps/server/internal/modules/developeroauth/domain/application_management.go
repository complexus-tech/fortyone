package developeroauthdomain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type ManagementAccess struct {
	Actor         platformauth.Actor
	WorkspaceID   uuid.UUID
	WorkspaceRole authorization.WorkspaceRole
}

type ManagedApplication struct {
	Application
	OwnerWorkspaceID uuid.UUID
	OwnerUserID      uuid.UUID
	Status           string
	UpdatedAt        time.Time
	RevokedAt        *time.Time
}

type IssuedManagedApplication struct {
	Application ManagedApplication
	Secret      IssuedClientSecret
}

type CreateManagedApplication struct {
	ApplicationID    uuid.UUID
	ClientID         string
	OwnerWorkspaceID uuid.UUID
	OwnerUserID      uuid.UUID
	Name             string
	RedirectURIs     []string
	ExpiresAt        time.Time
	Secret           SecretMaterial
	SecretExpiresAt  time.Time
	CreatedAt        time.Time
	Audit            AuditEvent
}

type RotateClientSecret struct {
	ApplicationID    uuid.UUID
	OwnerWorkspaceID uuid.UUID
	ActorUserID      uuid.UUID
	Secret           SecretMaterial
	ExpiresAt        time.Time
	OverlapExpiresAt time.Time
	RotatedAt        time.Time
	Audit            AuditEvent
}

type RevokeClientSecret struct {
	ApplicationID    uuid.UUID
	SecretID         uuid.UUID
	OwnerWorkspaceID uuid.UUID
	ActorUserID      uuid.UUID
	Reason           string
	RevokedAt        time.Time
	Audit            AuditEvent
}

type InstallApplication struct {
	InstallationID uuid.UUID
	PrincipalID    uuid.UUID
	ClientID       string
	WorkspaceID    uuid.UUID
	InstalledBy    uuid.UUID
	Resource       string
	Scopes         []string
	InstalledAt    time.Time
	Audit          AuditEvent
}

type UpdateApplicationInstallation struct {
	InstallationID uuid.UUID
	WorkspaceID    uuid.UUID
	ActorUserID    uuid.UUID
	Resource       string
	Scopes         []string
	UpdatedAt      time.Time
	Audit          AuditEvent
}

type RevokeApplicationInstallation struct {
	InstallationID uuid.UUID
	WorkspaceID    uuid.UUID
	ActorUserID    uuid.UUID
	Reason         string
	RevokedAt      time.Time
	Audit          AuditEvent
}
