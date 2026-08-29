package developercredentialsrepository

import (
	"errors"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type credentialRow struct {
	credentialID     uuid.UUID
	workspaceID      uuid.UUID
	principalID      uuid.UUID
	kind             string
	name             string
	lookupPrefix     string
	tokenVersion     int16
	expiresAt        time.Time
	lastUsedAt       *time.Time
	rotatedFromID    *uuid.UUID
	rotatedAt        *time.Time
	revokedAt        *time.Time
	revokedByUserID  *uuid.UUID
	revokedReason    *string
	createdByUserID  *uuid.UUID
	createdAt        time.Time
	scopes           []string
	teamRestrictions []uuid.UUID
}

func mapCredential(row credentialRow) (developercredentialsdomain.Credential, error) {
	kind := developercredentialsdomain.CredentialKind(row.kind)
	if err := kind.Validate(); err != nil {
		return developercredentialsdomain.Credential{}, err
	}
	scopes, err := mapScopes(row.scopes)
	if err != nil {
		return developercredentialsdomain.Credential{}, err
	}
	if row.credentialID == uuid.Nil || row.workspaceID == uuid.Nil || row.principalID == uuid.Nil || row.lookupPrefix == "" {
		return developercredentialsdomain.Credential{}, developercredentialsdomain.ErrCredentialNotFound
	}
	return developercredentialsdomain.Credential{
		ID:            row.credentialID,
		WorkspaceID:   row.workspaceID,
		PrincipalID:   row.principalID,
		Kind:          kind,
		Name:          row.name,
		LookupPrefix:  row.lookupPrefix,
		TokenVersion:  row.tokenVersion,
		Scopes:        scopes,
		TeamIDs:       append([]uuid.UUID(nil), row.teamRestrictions...),
		ExpiresAt:     row.expiresAt,
		LastUsedAt:    row.lastUsedAt,
		RotatedFromID: row.rotatedFromID,
		RotatedAt:     row.rotatedAt,
		RevokedAt:     row.revokedAt,
		RevokedBy:     row.revokedByUserID,
		RevokedReason: row.revokedReason,
		CreatedBy:     row.createdByUserID,
		CreatedAt:     row.createdAt,
	}, nil
}

func mapScopes(values []string) ([]platformauth.Scope, error) {
	scopes := make([]platformauth.Scope, len(values))
	for index, value := range values {
		scopes[index] = platformauth.Scope(value)
	}
	set, err := platformauth.NewScopeSet(scopes...)
	if err != nil || set.Has(platformauth.ScopeFirstParty) || len(set.Values()) == 0 {
		return nil, errors.Join(developercredentialsdomain.ErrInvalidScope, err)
	}
	return set.Values(), nil
}

func mapServiceAccount(
	principalID uuid.UUID,
	workspaceID uuid.UUID,
	name string,
	role string,
	status string,
	createdBy *uuid.UUID,
	createdAt time.Time,
	updatedAt time.Time,
	disabledAt *time.Time,
	disabledBy *uuid.UUID,
	disabledReason *string,
) (developercredentialsdomain.ServiceAccount, error) {
	workspaceRole := authorization.WorkspaceRole(role)
	if err := authorization.ValidateWorkspaceRole(workspaceRole); err != nil ||
		(workspaceRole != authorization.WorkspaceRoleGuest && workspaceRole != authorization.WorkspaceRoleMember) {
		return developercredentialsdomain.ServiceAccount{}, developercredentialsdomain.ErrInvalidServiceAccountRole
	}
	principalStatus := developercredentialsdomain.PrincipalStatus(status)
	if principalStatus != developercredentialsdomain.PrincipalActive && principalStatus != developercredentialsdomain.PrincipalDisabled {
		return developercredentialsdomain.ServiceAccount{}, developercredentialsdomain.ErrPrincipalNotFound
	}
	return developercredentialsdomain.ServiceAccount{
		ID:             principalID,
		WorkspaceID:    workspaceID,
		Name:           name,
		WorkspaceRole:  workspaceRole,
		Status:         principalStatus,
		CreatedBy:      createdBy,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		DisabledAt:     disabledAt,
		DisabledBy:     disabledBy,
		DisabledReason: disabledReason,
	}, nil
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
