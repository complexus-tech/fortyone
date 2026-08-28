package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) InstallApplication(
	ctx context.Context,
	command developeroauthdomain.InstallApplication,
) (developeroauthdomain.ApplicationInstallation, error) {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	defer rollback(ctx, tx)
	if err := store.requireOAuthApplicationAdmin(ctx, queries, command.WorkspaceID, command.InstalledBy); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	application, err := queries.LockManagedOAuthApplicationForInstallation(ctx, developeroauthsql.LockManagedOAuthApplicationForInstallationParams{
		ClientID: command.ClientID, ActiveAt: command.InstalledAt,
		WorkspaceID: command.WorkspaceID, ActorUserID: command.InstalledBy,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrApplicationNotFound
	}
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("lock managed OAuth application for installation: %w", err)
	}
	if _, err := queries.InsertOAuthApplicationPrincipal(ctx, developeroauthsql.InsertOAuthApplicationPrincipalParams{
		PrincipalID: command.PrincipalID, WorkspaceID: command.WorkspaceID,
		Name: application.Name, ActorUserID: uuidPointer(command.InstalledBy), CreatedAt: command.InstalledAt,
	}); errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrAccessDenied
	} else if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("insert OAuth application principal: %w", mapApplicationDatabaseError(err))
	}
	row, err := queries.InsertOAuthApplicationInstallation(ctx, developeroauthsql.InsertOAuthApplicationInstallationParams{
		InstallationID: command.InstallationID, ApplicationID: application.ApplicationID,
		WorkspaceID: command.WorkspaceID, PrincipalID: command.PrincipalID, Resource: command.Resource,
		InstalledByUserID: command.InstalledBy, CreatedAt: command.InstalledAt,
	})
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("insert OAuth application installation: %w", mapApplicationDatabaseError(err))
	}
	if err := insertInstallationScopes(ctx, queries, command.InstallationID, command.Scopes); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	command.Audit.ApplicationID = &application.ApplicationID
	command.Audit.PrincipalID = &command.PrincipalID
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("commit OAuth application installation: %w", mapApplicationDatabaseError(err))
	}
	return mapApplicationInstallation(installationRow{
		installationID: row.InstallationID, applicationID: row.ApplicationID, clientID: application.ClientID,
		workspaceID: row.WorkspaceID, principalID: row.PrincipalID, resource: row.Resource,
		status: row.Status, installedBy: row.InstalledByUserID, createdAt: row.CreatedAt,
		updatedAt: row.UpdatedAt, lastUsedAt: row.LastUsedAt, revokedAt: row.RevokedAt,
		revokedBy: row.RevokedByUserID, revokedReason: row.RevokedReason, scopes: command.Scopes,
	}), nil
}

func (store *Store) ListApplicationInstallations(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorUserID uuid.UUID,
) ([]developeroauthdomain.ApplicationInstallation, error) {
	if err := store.requireOAuthApplicationAdmin(ctx, store.queries, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := store.queries.ListOAuthApplicationInstallations(ctx, developeroauthsql.ListOAuthApplicationInstallationsParams{
		WorkspaceID: workspaceID, ActorUserID: actorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list OAuth application installations: %w", err)
	}
	installations := make([]developeroauthdomain.ApplicationInstallation, 0, len(rows))
	for _, row := range rows {
		installations = append(installations, mapApplicationInstallation(installationRow{
			installationID: row.InstallationID, applicationID: row.ApplicationID, clientID: row.ClientID,
			workspaceID: row.WorkspaceID, principalID: row.PrincipalID, resource: row.Resource,
			status: row.Status, installedBy: row.InstalledByUserID, createdAt: row.CreatedAt,
			updatedAt: row.UpdatedAt, lastUsedAt: row.LastUsedAt, revokedAt: row.RevokedAt,
			revokedBy: row.RevokedByUserID, revokedReason: row.RevokedReason, scopes: row.Scopes,
		}))
	}
	return installations, nil
}

func (store *Store) UpdateApplicationInstallation(
	ctx context.Context,
	command developeroauthdomain.UpdateApplicationInstallation,
) (developeroauthdomain.ApplicationInstallation, error) {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	defer rollback(ctx, tx)
	if err := store.requireOAuthApplicationAdmin(ctx, queries, command.WorkspaceID, command.ActorUserID); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	row, err := queries.GetOAuthApplicationInstallationForUpdate(ctx, developeroauthsql.GetOAuthApplicationInstallationForUpdateParams{
		InstallationID: command.InstallationID, WorkspaceID: command.WorkspaceID,
		ActorUserID: command.ActorUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInstallationNotFound
	}
	if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("lock OAuth application installation: %w", err)
	}
	if row.Status != "active" {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrInstallationRevoked
	}
	if err := queries.DeleteOAuthApplicationInstallationScopes(ctx, developeroauthsql.DeleteOAuthApplicationInstallationScopesParams{
		InstallationID: command.InstallationID,
	}); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("delete OAuth installation scopes: %w", err)
	}
	if err := insertInstallationScopes(ctx, queries, command.InstallationID, command.Scopes); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if _, err := queries.UpdateOAuthApplicationInstallation(ctx, developeroauthsql.UpdateOAuthApplicationInstallationParams{
		Resource: command.Resource, UpdatedAt: command.UpdatedAt,
		InstallationID: command.InstallationID, WorkspaceID: command.WorkspaceID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ApplicationInstallation{}, developeroauthdomain.ErrConcurrentUpdate
	} else if err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("update OAuth application installation: %w", err)
	}
	command.Audit.ApplicationID = &row.ApplicationID
	command.Audit.PrincipalID = &row.PrincipalID
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, fmt.Errorf("commit OAuth installation update: %w", mapApplicationDatabaseError(err))
	}
	return mapApplicationInstallation(installationRow{
		installationID: row.InstallationID, applicationID: row.ApplicationID, clientID: row.ClientID,
		workspaceID: row.WorkspaceID, principalID: row.PrincipalID, resource: command.Resource,
		status: row.Status, installedBy: row.InstalledByUserID, createdAt: row.CreatedAt,
		updatedAt: command.UpdatedAt, lastUsedAt: row.LastUsedAt, revokedAt: row.RevokedAt,
		revokedBy: row.RevokedByUserID, revokedReason: row.RevokedReason, scopes: command.Scopes,
	}), nil
}

func (store *Store) RevokeApplicationInstallation(
	ctx context.Context,
	command developeroauthdomain.RevokeApplicationInstallation,
) error {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	if err := store.requireOAuthApplicationAdmin(ctx, queries, command.WorkspaceID, command.ActorUserID); err != nil {
		return err
	}
	row, err := queries.GetOAuthApplicationInstallationForUpdate(ctx, developeroauthsql.GetOAuthApplicationInstallationForUpdateParams{
		InstallationID: command.InstallationID, WorkspaceID: command.WorkspaceID,
		ActorUserID: command.ActorUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ErrInstallationNotFound
	}
	if err != nil {
		return fmt.Errorf("lock OAuth application installation for revocation: %w", err)
	}
	if row.Status == "active" {
		principalID, err := queries.RevokeOAuthApplicationInstallation(ctx, developeroauthsql.RevokeOAuthApplicationInstallationParams{
			RevokedAt: command.RevokedAt, ActorUserID: uuidPointer(command.ActorUserID),
			RevokedReason: stringPointer(command.Reason), InstallationID: command.InstallationID,
			WorkspaceID: command.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developeroauthdomain.ErrConcurrentUpdate
		}
		if err != nil {
			return fmt.Errorf("revoke OAuth application installation: %w", err)
		}
		if principalID != row.PrincipalID {
			return developeroauthdomain.ErrConcurrentUpdate
		}
		if _, err := queries.DisableOAuthApplicationPrincipal(ctx, developeroauthsql.DisableOAuthApplicationPrincipalParams{
			DisabledAt: command.RevokedAt, ActorUserID: uuidPointer(command.ActorUserID),
			DisabledReason: stringPointer(command.Reason), PrincipalID: row.PrincipalID,
			WorkspaceID: command.WorkspaceID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return developeroauthdomain.ErrConcurrentUpdate
		} else if err != nil {
			return fmt.Errorf("disable OAuth application principal: %w", err)
		}
	}
	command.Audit.ApplicationID = &row.ApplicationID
	command.Audit.PrincipalID = &row.PrincipalID
	command.Audit.ReasonCode = stringPointer(command.Reason)
	if err := createAuditEvent(ctx, queries, command.Audit); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit OAuth installation revocation: %w", mapApplicationDatabaseError(err))
	}
	return nil
}

func insertInstallationScopes(
	ctx context.Context,
	queries *developeroauthsql.Queries,
	installationID uuid.UUID,
	scopes []string,
) error {
	if err := queries.InsertOAuthApplicationInstallationScopes(ctx, developeroauthsql.InsertOAuthApplicationInstallationScopesParams{
		InstallationID: installationID, Scopes: append([]string(nil), scopes...),
	}); err != nil {
		return fmt.Errorf("insert OAuth installation scopes: %w", mapApplicationDatabaseError(err))
	}
	return nil
}
