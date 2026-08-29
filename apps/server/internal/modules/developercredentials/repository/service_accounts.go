package developercredentialsrepository

import (
	"context"
	"errors"
	"fmt"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) CreateServiceAccount(
	ctx context.Context,
	command developercredentialsdomain.CreateServiceAccount,
) (developercredentialsdomain.ServiceAccount, error) {
	var account developercredentialsdomain.ServiceAccount
	err := store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		row, err := queries.InsertServiceAccount(ctx, developercredentialssql.InsertServiceAccountParams{
			PrincipalID: command.PrincipalID, WorkspaceID: command.WorkspaceID, Name: command.Name,
			WorkspaceRole: developercredentialssql.UserRole(command.WorkspaceRole),
			ActorUserID:   uuidPointer(command.ActorUserID), CreatedAt: command.CreatedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrAccessDenied
		}
		if err != nil {
			return fmt.Errorf("insert service account: %w", err)
		}
		if err := insertAuditEvent(ctx, queries, command.Audit); err != nil {
			return err
		}
		account, err = mapServiceAccount(row.PrincipalID, row.WorkspaceID, row.Name, row.WorkspaceRole,
			row.Status, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt, row.DisabledAt,
			row.DisabledByUserID, row.DisabledReason)
		return err
	})
	return account, err
}

func (store *Store) ListServiceAccounts(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorUserID uuid.UUID,
) ([]developercredentialsdomain.ServiceAccount, error) {
	if err := store.requireCurrentWorkspaceAdmin(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := store.queries.ListServiceAccounts(ctx, developercredentialssql.ListServiceAccountsParams{
		WorkspaceID: workspaceID, ActorUserID: actorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list service accounts: %w", mapDatabaseError(err))
	}
	accounts := make([]developercredentialsdomain.ServiceAccount, 0, len(rows))
	for _, row := range rows {
		account, err := mapServiceAccount(row.PrincipalID, row.WorkspaceID, row.Name, row.WorkspaceRole,
			row.Status, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt, row.DisabledAt,
			row.DisabledByUserID, row.DisabledReason)
		if err != nil {
			return nil, fmt.Errorf("map service account: %w", err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (store *Store) DisableServiceAccount(
	ctx context.Context,
	command developercredentialsdomain.DisableServiceAccount,
) error {
	return store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		_, err := queries.DisableServiceAccount(ctx, developercredentialssql.DisableServiceAccountParams{
			DisabledAt: command.DisabledAt, ActorUserID: uuidPointer(command.ActorUserID),
			DisabledReason: stringPointer(command.Reason), PrincipalID: command.PrincipalID,
			WorkspaceID: command.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrPrincipalNotFound
		}
		if err != nil {
			return fmt.Errorf("disable service account: %w", err)
		}
		if err := queries.RevokeServiceAccountCredentials(ctx, developercredentialssql.RevokeServiceAccountCredentialsParams{
			RevokedAt: timePointer(command.DisabledAt), ActorUserID: uuidPointer(command.ActorUserID),
			RevokedReason: stringPointer(command.Reason), PrincipalID: command.PrincipalID,
			WorkspaceID: command.WorkspaceID,
		}); err != nil {
			return fmt.Errorf("revoke disabled service account keys: %w", err)
		}
		command.Audit.ReasonCode = command.Reason
		return insertAuditEvent(ctx, queries, command.Audit)
	})
}

func (store *Store) CreateServiceAccountKey(
	ctx context.Context,
	command developercredentialsdomain.CreateServiceAccountKey,
) (developercredentialsdomain.Credential, error) {
	digestKeyVersion, err := safecast.Uint32ToInt32(command.Credential.DigestKey.Version)
	if err != nil {
		return developercredentialsdomain.Credential{}, fmt.Errorf("validate service-account key digest version: %w", err)
	}
	var credential developercredentialsdomain.Credential
	err = store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		validTeams, err := queries.ValidateServiceAccountTeamRestrictions(ctx, developercredentialssql.ValidateServiceAccountTeamRestrictionsParams{
			TeamIds: command.TeamIDs, WorkspaceID: command.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("validate service account team restrictions: %w", err)
		}
		if !validTeams {
			return developercredentialsdomain.ErrTeamRestrictionNotAllowed
		}
		row, err := queries.InsertServiceAccountKey(ctx, developercredentialssql.InsertServiceAccountKeyParams{
			CredentialID: command.Credential.ID, Name: command.Name,
			LookupPrefix: command.Credential.LookupPrefix,
			SecretDigest: append([]byte(nil), command.Credential.SecretDigest...),
			TokenVersion: command.Credential.TokenVersion, DigestKeyID: command.Credential.DigestKey.ID,
			DigestKeyVersion: digestKeyVersion, ExpiresAt: command.ExpiresAt,
			ActorUserID: uuidPointer(command.ActorUserID), CreatedAt: command.CreatedAt,
			PrincipalID: command.PrincipalID, WorkspaceID: command.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrPrincipalNotFound
		}
		if err != nil {
			return fmt.Errorf("insert service account key: %w", err)
		}
		if err := insertCredentialGrants(ctx, queries, command.Credential.ID, command.WorkspaceID, command.Scopes, command.TeamIDs); err != nil {
			return err
		}
		if err := insertAuditEvent(ctx, queries, command.Audit); err != nil {
			return err
		}
		credential, err = mapCredential(insertedServiceAccountKeyRow(row, command.Scopes, command.TeamIDs))
		return err
	})
	return credential, err
}

func (store *Store) ListServiceAccountKeys(
	ctx context.Context,
	workspaceID uuid.UUID,
	principalID uuid.UUID,
	actorUserID uuid.UUID,
) ([]developercredentialsdomain.Credential, error) {
	if err := store.requireCurrentWorkspaceAdmin(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := store.queries.ListServiceAccountKeys(ctx, developercredentialssql.ListServiceAccountKeysParams{
		WorkspaceID: workspaceID, PrincipalID: principalID, ActorUserID: actorUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list service account keys: %w", mapDatabaseError(err))
	}
	credentials := make([]developercredentialsdomain.Credential, 0, len(rows))
	for _, row := range rows {
		credential, err := mapCredential(listedServiceAccountKeyRow(row))
		if err != nil {
			return nil, fmt.Errorf("map service account key: %w", err)
		}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func (store *Store) RotateServiceAccountKey(
	ctx context.Context,
	command developercredentialsdomain.RotateCredential,
) (developercredentialsdomain.Credential, error) {
	digestKeyVersion, err := safecast.Uint32ToInt32(command.NewCredential.DigestKey.Version)
	if err != nil {
		return developercredentialsdomain.Credential{}, fmt.Errorf("validate rotated service-account key digest version: %w", err)
	}
	var credential developercredentialsdomain.Credential
	err = store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		_, err := queries.LockServiceAccountKeyForRotation(ctx, developercredentialssql.LockServiceAccountKeyForRotationParams{
			CredentialID: command.OldCredentialID, WorkspaceID: command.WorkspaceID,
			PrincipalID: command.PrincipalID, RotatedAt: command.RotatedAt, ActorUserID: command.ActorUserID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrCredentialRotationConflict
		}
		if err != nil {
			return fmt.Errorf("lock service account key for rotation: %w", err)
		}
		row, err := queries.InsertRotatedServiceAccountKey(ctx, developercredentialssql.InsertRotatedServiceAccountKeyParams{
			NewCredentialID: command.NewCredential.ID, LookupPrefix: command.NewCredential.LookupPrefix,
			SecretDigest: append([]byte(nil), command.NewCredential.SecretDigest...),
			TokenVersion: command.NewCredential.TokenVersion, DigestKeyID: command.NewCredential.DigestKey.ID,
			DigestKeyVersion: digestKeyVersion, ExpiresAt: command.ExpiresAt,
			ActorUserID: uuidPointer(command.ActorUserID), CreatedAt: command.RotatedAt,
			OldCredentialID: command.OldCredentialID, WorkspaceID: command.WorkspaceID,
			PrincipalID: command.PrincipalID,
		})
		if err != nil {
			return fmt.Errorf("insert rotated service account key: %w", err)
		}
		if err := copyCredentialGrants(ctx, queries, command.OldCredentialID, command.NewCredential.ID); err != nil {
			return err
		}
		if err := markCredentialRotated(ctx, queries, command); err != nil {
			return err
		}
		if err := insertAuditEvent(ctx, queries, command.Audit); err != nil {
			return err
		}
		grants, err := queries.GetCredentialGrants(ctx, developercredentialssql.GetCredentialGrantsParams{CredentialID: command.NewCredential.ID})
		if err != nil {
			return fmt.Errorf("load rotated service account key grants: %w", err)
		}
		credential, err = mapCredential(rotatedServiceAccountKeyRow(row, grants.Scopes, grants.TeamRestrictions))
		return err
	})
	return credential, err
}

func (store *Store) RevokeServiceAccountKey(ctx context.Context, command developercredentialsdomain.RevokeCredential) error {
	return store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		_, err := queries.RevokeServiceAccountKey(ctx, developercredentialssql.RevokeServiceAccountKeyParams{
			RevokedAt: timePointer(command.RevokedAt), ActorUserID: uuidPointer(command.ActorUserID),
			RevokedReason: stringPointer(command.Reason), CredentialID: command.CredentialID,
			WorkspaceID: command.WorkspaceID, PrincipalID: command.PrincipalID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrCredentialNotFound
		}
		if err != nil {
			return fmt.Errorf("revoke service account key: %w", err)
		}
		command.Audit.ReasonCode = command.Reason
		return insertAuditEvent(ctx, queries, command.Audit)
	})
}

func (store *Store) requireCurrentWorkspaceAdmin(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorUserID uuid.UUID,
) error {
	allowed, err := store.queries.IsCurrentWorkspaceAdmin(ctx, developercredentialssql.IsCurrentWorkspaceAdminParams{
		WorkspaceID: workspaceID,
		UserID:      actorUserID,
	})
	if err != nil {
		return fmt.Errorf("authorize service-account administration: %w", mapDatabaseError(err))
	}
	if !allowed {
		return developercredentialsdomain.ErrAccessDenied
	}
	return nil
}
