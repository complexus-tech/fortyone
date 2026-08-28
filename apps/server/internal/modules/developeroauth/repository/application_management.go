package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) CreateManagedApplication(
	ctx context.Context,
	command developeroauthdomain.CreateManagedApplication,
) (developeroauthdomain.ManagedApplication, developeroauthdomain.ClientSecret, error) {
	if err := validateClientSecretMaterial(command.Secret); err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, err
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, err
	}
	defer rollback(ctx, tx)
	row, err := queries.CreateManagedOAuthApplication(ctx, developeroauthsql.CreateManagedOAuthApplicationParams{
		ApplicationID: command.ApplicationID, ClientID: command.ClientID, Name: command.Name,
		OwnerWorkspaceID: uuidPointer(command.OwnerWorkspaceID), OwnerUserID: uuidPointer(command.OwnerUserID),
		ExpiresAt: command.ExpiresAt, CreatedAt: command.CreatedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrAccessDenied
	}
	if err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, fmt.Errorf("create managed OAuth application: %w", mapApplicationDatabaseError(err))
	}
	for _, redirectURI := range command.RedirectURIs {
		if err := queries.CreateOAuthApplicationRedirectURI(ctx, developeroauthsql.CreateOAuthApplicationRedirectURIParams{
			ApplicationID: row.ApplicationID, RedirectURI: redirectURI, CreatedAt: command.CreatedAt,
		}); err != nil {
			return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, fmt.Errorf("create managed OAuth redirect URI: %w", err)
		}
	}
	secretRow, err := insertClientSecret(ctx, queries, command.ApplicationID, command.Secret,
		command.SecretExpiresAt, nil, command.OwnerUserID, command.CreatedAt)
	if err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, err
	}
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.ManagedApplication{}, developeroauthdomain.ClientSecret{}, fmt.Errorf("commit managed OAuth application: %w", mapApplicationDatabaseError(err))
	}
	application, err := mapManagedApplication(managedApplicationRow{
		applicationID: row.ApplicationID, clientID: row.ClientID, name: row.Name,
		registrationKind: row.RegistrationKind, status: row.Status,
		ownerWorkspaceID: row.OwnerWorkspaceID, ownerUserID: row.OwnerUserID,
		expiresAt: row.ExpiresAt, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		revokedAt: row.RevokedAt, redirectURIs: command.RedirectURIs,
	})
	return application, mapClientSecretRow(secretRow), err
}

func (store *Store) ListManagedApplications(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorUserID uuid.UUID,
) ([]developeroauthdomain.ManagedApplication, error) {
	if err := store.requireOAuthApplicationAdmin(ctx, store.queries, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := store.queries.ListManagedOAuthApplications(ctx, developeroauthsql.ListManagedOAuthApplicationsParams{
		OwnerWorkspaceID: uuidPointer(workspaceID), ActorUserID: actorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list managed OAuth applications: %w", err)
	}
	applications := make([]developeroauthdomain.ManagedApplication, 0, len(rows))
	for _, row := range rows {
		application, err := mapManagedApplication(managedApplicationRow{
			applicationID: row.ApplicationID, clientID: row.ClientID, name: row.Name,
			registrationKind: row.RegistrationKind, status: row.Status,
			ownerWorkspaceID: row.OwnerWorkspaceID, ownerUserID: row.OwnerUserID,
			expiresAt: row.ExpiresAt, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
			revokedAt: row.RevokedAt, redirectURIs: row.RedirectUris,
		})
		if err != nil {
			return nil, fmt.Errorf("map managed OAuth application: %w", err)
		}
		applications = append(applications, application)
	}
	return applications, nil
}

func (store *Store) ListClientSecrets(
	ctx context.Context,
	workspaceID uuid.UUID,
	applicationID uuid.UUID,
	actorUserID uuid.UUID,
) ([]developeroauthdomain.ClientSecret, error) {
	if err := store.requireOAuthApplicationAdmin(ctx, store.queries, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := store.queries.ListOAuthClientSecrets(ctx, developeroauthsql.ListOAuthClientSecretsParams{
		ApplicationID: applicationID, OwnerWorkspaceID: uuidPointer(workspaceID), ActorUserID: actorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list OAuth client secrets: %w", err)
	}
	secrets := make([]developeroauthdomain.ClientSecret, 0, len(rows))
	for _, row := range rows {
		secrets = append(secrets, mapClientSecret(clientSecretRow{
			secretID: row.SecretID, applicationID: row.ApplicationID, lookupPrefix: row.LookupPrefix,
			expiresAt: row.ExpiresAt, lastUsedAt: row.LastUsedAt, rotatedFromID: row.RotatedFromID,
			overlapExpiresAt: row.OverlapExpiresAt, revokedAt: row.RevokedAt, createdAt: row.CreatedAt,
		}))
	}
	return secrets, nil
}

func (store *Store) RotateClientSecret(
	ctx context.Context,
	command developeroauthdomain.RotateClientSecret,
) (developeroauthdomain.ClientSecret, error) {
	if err := validateClientSecretMaterial(command.Secret); err != nil {
		return developeroauthdomain.ClientSecret{}, err
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.ClientSecret{}, err
	}
	defer rollback(ctx, tx)
	if err := store.requireOAuthApplicationAdmin(ctx, queries, command.OwnerWorkspaceID, command.ActorUserID); err != nil {
		return developeroauthdomain.ClientSecret{}, err
	}
	application, err := queries.LockManagedOAuthApplication(ctx, developeroauthsql.LockManagedOAuthApplicationParams{
		ApplicationID: command.ApplicationID, OwnerWorkspaceID: uuidPointer(command.OwnerWorkspaceID),
		ActiveAt: command.RotatedAt, ActorUserID: command.ActorUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrApplicationNotFound
	}
	if err != nil {
		return developeroauthdomain.ClientSecret{}, fmt.Errorf("lock managed OAuth application: %w", err)
	}
	if command.ExpiresAt.After(application.ExpiresAt) {
		return developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrInvalidExpiry
	}
	head, err := queries.GetOAuthClientSecretRotationHeadForUpdate(ctx, developeroauthsql.GetOAuthClientSecretRotationHeadForUpdateParams{
		ApplicationID: command.ApplicationID,
	})
	var rotatedFromID *uuid.UUID
	if err == nil {
		rotatedFromID = uuidPointer(head.SecretID)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ClientSecret{}, fmt.Errorf("lock OAuth client-secret rotation head: %w", err)
	}
	secretRow, err := insertClientSecret(ctx, queries, command.ApplicationID, command.Secret,
		command.ExpiresAt, rotatedFromID, command.ActorUserID, command.RotatedAt)
	if err != nil {
		return developeroauthdomain.ClientSecret{}, err
	}
	if rotatedFromID != nil {
		if _, err := queries.SetOAuthClientSecretOverlap(ctx, developeroauthsql.SetOAuthClientSecretOverlapParams{
			OverlapExpiresAt: timePointer(command.OverlapExpiresAt), SecretID: *rotatedFromID,
			ApplicationID: command.ApplicationID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return developeroauthdomain.ClientSecret{}, developeroauthdomain.ErrConcurrentUpdate
		} else if err != nil {
			return developeroauthdomain.ClientSecret{}, fmt.Errorf("set OAuth client-secret overlap cutoff: %w", err)
		}
	}
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return developeroauthdomain.ClientSecret{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.ClientSecret{}, fmt.Errorf("commit OAuth client-secret rotation: %w", mapApplicationDatabaseError(err))
	}
	return mapClientSecretRow(secretRow), nil
}

func (store *Store) RevokeClientSecret(ctx context.Context, command developeroauthdomain.RevokeClientSecret) error {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	if err := store.requireOAuthApplicationAdmin(ctx, queries, command.OwnerWorkspaceID, command.ActorUserID); err != nil {
		return err
	}
	secret, err := queries.GetManagedOAuthClientSecretForUpdate(ctx, developeroauthsql.GetManagedOAuthClientSecretForUpdateParams{
		SecretID: command.SecretID, ApplicationID: command.ApplicationID,
		OwnerWorkspaceID: uuidPointer(command.OwnerWorkspaceID), ActorUserID: command.ActorUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ErrSecretNotFound
	}
	if err != nil {
		return fmt.Errorf("lock OAuth client secret for revocation: %w", err)
	}
	if secret.RevokedAt == nil {
		if _, err := queries.SetOAuthClientSecretRevoked(ctx, developeroauthsql.SetOAuthClientSecretRevokedParams{
			RevokedAt: timePointer(command.RevokedAt), ActorUserID: uuidPointer(command.ActorUserID),
			RevokedReason: stringPointer(command.Reason), SecretID: command.SecretID,
			ApplicationID: command.ApplicationID,
		}); err != nil {
			return fmt.Errorf("revoke OAuth client secret: %w", err)
		}
	}
	command.Audit.ReasonCode = stringPointer(command.Reason)
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit OAuth client-secret revocation: %w", mapApplicationDatabaseError(err))
	}
	return nil
}

func insertClientSecret(
	ctx context.Context,
	queries *developeroauthsql.Queries,
	applicationID uuid.UUID,
	material developeroauthdomain.SecretMaterial,
	expiresAt time.Time,
	rotatedFromID *uuid.UUID,
	createdBy uuid.UUID,
	createdAt time.Time,
) (developeroauthsql.InsertOAuthClientSecretRow, error) {
	row, err := queries.InsertOAuthClientSecret(ctx, developeroauthsql.InsertOAuthClientSecretParams{
		SecretID: material.ID, ApplicationID: applicationID, LookupPrefix: material.LookupPrefix,
		SecretDigest: append([]byte(nil), material.Digest...), DigestKeyID: material.DigestKey.ID,
		ExpiresAt: expiresAt, RotatedFromID: rotatedFromID, CreatedByUserID: createdBy, CreatedAt: createdAt,
	})
	if err != nil {
		return developeroauthsql.InsertOAuthClientSecretRow{}, fmt.Errorf("insert OAuth client secret: %w", mapApplicationDatabaseError(err))
	}
	return row, nil
}

func mapClientSecretRow(row developeroauthsql.InsertOAuthClientSecretRow) developeroauthdomain.ClientSecret {
	return mapClientSecret(clientSecretRow{
		secretID: row.SecretID, applicationID: row.ApplicationID, lookupPrefix: row.LookupPrefix,
		expiresAt: row.ExpiresAt, lastUsedAt: row.LastUsedAt, rotatedFromID: row.RotatedFromID,
		overlapExpiresAt: row.OverlapExpiresAt, revokedAt: row.RevokedAt, createdAt: row.CreatedAt,
	})
}

func validateClientSecretMaterial(material developeroauthdomain.SecretMaterial) error {
	if material.Kind != developeroauthdomain.SecretClientSecret || material.ID == uuid.Nil ||
		material.LookupPrefix == "" || len(material.Digest) != 32 || material.DigestKey.ID == "" {
		return developeroauthdomain.ErrClientSecret
	}
	return nil
}

func (store *Store) requireOAuthApplicationAdmin(
	ctx context.Context,
	queries *developeroauthsql.Queries,
	workspaceID uuid.UUID,
	actorUserID uuid.UUID,
) error {
	allowed, err := queries.IsOAuthApplicationWorkspaceAdmin(ctx, developeroauthsql.IsOAuthApplicationWorkspaceAdminParams{
		WorkspaceID: workspaceID, ActorUserID: actorUserID,
	})
	if err != nil {
		return fmt.Errorf("authorize OAuth application administration: %w", err)
	}
	if !allowed {
		return developeroauthdomain.ErrAccessDenied
	}
	return nil
}
