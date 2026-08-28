package developeroauthrepository

import (
	"errors"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/google/uuid"
)

type managedApplicationRow struct {
	applicationID    uuid.UUID
	clientID         string
	name             string
	registrationKind string
	status           string
	ownerWorkspaceID *uuid.UUID
	ownerUserID      *uuid.UUID
	expiresAt        time.Time
	createdAt        time.Time
	updatedAt        time.Time
	revokedAt        *time.Time
	redirectURIs     []string
}

func mapManagedApplication(row managedApplicationRow) (developeroauthdomain.ManagedApplication, error) {
	if row.ownerWorkspaceID == nil || row.ownerUserID == nil {
		return developeroauthdomain.ManagedApplication{}, errors.New("managed OAuth application ownership is incomplete")
	}
	return developeroauthdomain.ManagedApplication{
		Application: developeroauthdomain.Application{
			ID: row.applicationID, ClientID: row.clientID, Name: row.name,
			RegistrationKind: row.registrationKind, RedirectURIs: append([]string(nil), row.redirectURIs...),
			ExpiresAt: row.expiresAt, CreatedAt: row.createdAt,
		},
		OwnerWorkspaceID: *row.ownerWorkspaceID, OwnerUserID: *row.ownerUserID,
		Status: row.status, UpdatedAt: row.updatedAt, RevokedAt: row.revokedAt,
	}, nil
}

type clientSecretRow struct {
	secretID         uuid.UUID
	applicationID    uuid.UUID
	lookupPrefix     string
	expiresAt        time.Time
	lastUsedAt       *time.Time
	rotatedFromID    *uuid.UUID
	overlapExpiresAt *time.Time
	revokedAt        *time.Time
	createdAt        time.Time
}

func mapClientSecret(row clientSecretRow) developeroauthdomain.ClientSecret {
	return developeroauthdomain.ClientSecret{
		ID: row.secretID, ApplicationID: row.applicationID, LookupPrefix: row.lookupPrefix,
		ExpiresAt: row.expiresAt, LastUsedAt: row.lastUsedAt, RotatedFromID: row.rotatedFromID,
		OverlapExpiresAt: row.overlapExpiresAt, RevokedAt: row.revokedAt, CreatedAt: row.createdAt,
	}
}

type installationRow struct {
	installationID uuid.UUID
	applicationID  uuid.UUID
	clientID       string
	workspaceID    uuid.UUID
	principalID    uuid.UUID
	resource       string
	status         string
	installedBy    uuid.UUID
	createdAt      time.Time
	updatedAt      time.Time
	lastUsedAt     *time.Time
	revokedAt      *time.Time
	revokedBy      *uuid.UUID
	revokedReason  *string
	scopes         []string
}

func mapApplicationInstallation(row installationRow) developeroauthdomain.ApplicationInstallation {
	return developeroauthdomain.ApplicationInstallation{
		ID: row.installationID, ApplicationID: row.applicationID, ClientID: row.clientID,
		WorkspaceID: row.workspaceID, PrincipalID: row.principalID, Resource: row.resource,
		Scopes: append([]string(nil), row.scopes...), Status: row.status, InstalledBy: row.installedBy,
		CreatedAt: row.createdAt, UpdatedAt: row.updatedAt, LastUsedAt: row.lastUsedAt,
		RevokedAt: row.revokedAt, RevokedBy: row.revokedBy, RevokedReason: row.revokedReason,
	}
}

func mapCredentialRecord(
	row developeroauthsql.GetOAuthApplicationCredentialForUpdateRow,
) (developeroauthdomain.ClientSecretRecord, developeroauthdomain.ApplicationInstallation) {
	secret := mapClientSecret(clientSecretRow{
		secretID: row.SecretID, applicationID: row.ApplicationID, lookupPrefix: row.LookupPrefix,
		expiresAt: row.SecretExpiresAt, lastUsedAt: row.SecretLastUsedAt,
		rotatedFromID: row.RotatedFromID, overlapExpiresAt: row.OverlapExpiresAt,
		revokedAt: row.SecretRevokedAt, createdAt: row.SecretCreatedAt,
	})
	installation := mapApplicationInstallation(installationRow{
		installationID: row.InstallationID, applicationID: row.ApplicationID, clientID: row.ClientID,
		workspaceID: row.WorkspaceID, principalID: row.PrincipalID, resource: row.Resource,
		status: row.InstallationStatus, installedBy: row.InstalledByUserID,
		createdAt: row.InstallationCreatedAt, updatedAt: row.InstallationUpdatedAt,
		lastUsedAt: row.InstallationLastUsedAt, revokedAt: row.InstallationRevokedAt,
		revokedBy: row.RevokedByUserID, revokedReason: row.RevokedReason, scopes: row.InstallationScopes,
	})
	return developeroauthdomain.ClientSecretRecord{
		Secret: secret, ClientID: row.ClientID,
		Material: developeroauthdomain.SecretMaterial{
			ID: row.SecretID, Kind: developeroauthdomain.SecretClientSecret,
			LookupPrefix: row.LookupPrefix, Digest: append([]byte(nil), row.SecretDigest...),
			DigestKey: developeroauthdomain.DigestKeyRef{ID: row.DigestKeyID},
		},
		Application: developeroauthdomain.Application{
			ID: row.ApplicationID, ClientID: row.ClientID, Name: row.ApplicationName,
			RegistrationKind: row.RegistrationKind, RedirectURIs: append([]string(nil), row.RedirectUris...),
			ExpiresAt: row.ApplicationExpiresAt, CreatedAt: row.ApplicationCreatedAt,
		},
	}, installation
}

func uuidPointer(value uuid.UUID) *uuid.UUID { return &value }

func timePointer(value time.Time) *time.Time { return &value }

func stringPointer(value string) *string { return &value }
