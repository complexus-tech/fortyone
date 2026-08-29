package developeroauthrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) AuthenticateApplication(
	ctx context.Context,
	command developeroauthdomain.AuthenticateApplicationCredential,
	validate func(developeroauthdomain.ClientSecretRecord, developeroauthdomain.ApplicationInstallation) error,
) (developeroauthdomain.ApplicationInstallation, error) {
	if validate == nil {
		return developeroauthdomain.ApplicationInstallation{}, errors.New("OAuth application credential validator is required")
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	defer rollback(ctx, tx)
	row, err := queries.GetOAuthApplicationCredentialForUpdate(ctx, developeroauthsql.GetOAuthApplicationCredentialForUpdateParams{
		LookupPrefix: command.LookupPrefix, InstallationID: command.InstallationID,
		ActiveAt: command.AuthenticatedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInvalidClient
	}
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("lock OAuth application credential: %w", err)
	}
	record, installation := mapCredentialRecord(row)
	if err := validate(record, installation); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if _, err := queries.TouchOAuthClientSecret(ctx, developeroauthsql.TouchOAuthClientSecretParams{
		UsedAt: timePointer(command.AuthenticatedAt), SecretID: row.SecretID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInvalidClient
	} else if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("touch OAuth client secret: %w", err)
	}
	if _, err := queries.TouchOAuthApplicationInstallationAuthentication(ctx, developeroauthsql.TouchOAuthApplicationInstallationAuthenticationParams{
		UsedAt: timePointer(command.AuthenticatedAt), InstallationID: row.InstallationID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInstallationRevoked
	} else if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("touch OAuth application installation: %w", err)
	}
	metadata, err := json.Marshal(struct {
		Resource   string `json:"resource"`
		ScopeCount int    `json:"scope_count"`
	}{Resource: installation.Resource, ScopeCount: len(installation.Scopes)})
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("marshal OAuth client-credentials audit metadata: %w", err)
	}
	applicationID, installationID := installation.ApplicationID, installation.ID
	principalID, secretID, workspaceID := installation.PrincipalID, row.SecretID, installation.WorkspaceID
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: command.AuditID, ApplicationID: &applicationID, InstallationID: &installationID,
		PrincipalID: &principalID, SecretID: &secretID, WorkspaceID: &workspaceID,
		ActorKind: platformauth.PrincipalOAuthApplication, ActorID: &principalID,
		ActorCredentialID: &installationID, RequestID: command.RequestID,
		SubjectType: "access_token", SubjectID: &command.AccessTokenID,
		Operation: "client_credentials.exchanged", Result: "succeeded", Metadata: metadata,
		CreatedAt: command.AuthenticatedAt,
	}); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("commit OAuth client-credentials exchange: %w", mapApplicationDatabaseError(err))
	}
	installation.LastUsedAt = timePointer(command.AuthenticatedAt)
	installation.UpdatedAt = command.AuthenticatedAt
	return installation, nil
}

func (store *Store) GetActiveApplicationInstallation(
	ctx context.Context,
	installationID uuid.UUID,
	applicationID uuid.UUID,
	resource string,
	activeAt time.Time,
) (developeroauthdomain.ApplicationInstallation, error) {
	row, err := store.queries.GetActiveOAuthApplicationInstallation(ctx, developeroauthsql.GetActiveOAuthApplicationInstallationParams{
		InstallationID: installationID, ApplicationID: applicationID, Resource: resource, ActiveAt: activeAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInstallationNotFound
	}
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("get active OAuth application installation: %w", err)
	}
	return mapApplicationInstallation(installationRow{
		installationID: row.InstallationID, applicationID: row.ApplicationID, clientID: row.ClientID,
		workspaceID: row.WorkspaceID, principalID: row.PrincipalID, resource: row.Resource,
		status: row.Status, installedBy: row.InstalledByUserID, createdAt: row.CreatedAt,
		updatedAt: row.UpdatedAt, lastUsedAt: row.LastUsedAt, revokedAt: row.RevokedAt,
		revokedBy: row.RevokedByUserID, revokedReason: row.RevokedReason, scopes: row.Scopes,
	}), nil
}

func (store *Store) TouchApplicationInstallation(
	ctx context.Context,
	installationID uuid.UUID,
	usedAt time.Time,
	touchBefore time.Time,
) error {
	_, err := store.queries.TouchActiveOAuthApplicationInstallation(ctx, developeroauthsql.TouchActiveOAuthApplicationInstallationParams{
		TouchBefore: timePointer(touchBefore), UsedAt: timePointer(usedAt), InstallationID: installationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ErrInstallationRevoked
	}
	if err != nil {
		return fmt.Errorf("touch active OAuth application installation: %w", err)
	}
	return nil
}
