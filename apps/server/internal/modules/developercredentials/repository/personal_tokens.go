package developercredentialsrepository

import (
	"context"
	"errors"
	"fmt"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) CreatePersonalAccessToken(
	ctx context.Context,
	command developercredentialsdomain.CreatePersonalToken,
) (developercredentialsdomain.Credential, error) {
	digestKeyVersion, err := safecast.Uint32ToInt32(command.Credential.DigestKey.Version)
	if err != nil {
		return developercredentialsdomain.Credential{}, fmt.Errorf("validate personal token digest key version: %w", err)
	}
	var credential developercredentialsdomain.Credential
	err = store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		validTeams, err := queries.ValidatePersonalTokenTeamRestrictions(ctx, developercredentialssql.ValidatePersonalTokenTeamRestrictionsParams{
			TeamIds:     command.TeamIDs,
			WorkspaceID: command.WorkspaceID,
			UserID:      command.UserID,
		})
		if err != nil {
			return fmt.Errorf("validate personal token team restrictions: %w", err)
		}
		if !validTeams {
			return developercredentialsdomain.ErrTeamRestrictionNotAllowed
		}

		principalID, err := queries.EnsureHumanPrincipal(ctx, developercredentialssql.EnsureHumanPrincipalParams{
			UserID:      command.UserID,
			WorkspaceID: command.WorkspaceID,
			PrincipalID: command.PrincipalCandidateID,
			CreatedAt:   command.CreatedAt,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return developercredentialsdomain.ErrAccessDenied
			}
			return fmt.Errorf("ensure personal token principal: %w", err)
		}
		row, err := queries.InsertPersonalAccessToken(ctx, developercredentialssql.InsertPersonalAccessTokenParams{
			CredentialID:     command.Credential.ID,
			Name:             command.Name,
			LookupPrefix:     command.Credential.LookupPrefix,
			SecretDigest:     append([]byte(nil), command.Credential.SecretDigest...),
			TokenVersion:     command.Credential.TokenVersion,
			DigestKeyID:      command.Credential.DigestKey.ID,
			DigestKeyVersion: digestKeyVersion,
			ExpiresAt:        command.ExpiresAt,
			UserID:           uuidPointer(command.UserID),
			CreatedAt:        command.CreatedAt,
			PrincipalID:      principalID,
			WorkspaceID:      command.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("insert personal access token: %w", err)
		}
		if err := insertCredentialGrants(ctx, queries, command.Credential.ID, command.WorkspaceID, command.Scopes, command.TeamIDs); err != nil {
			return err
		}
		if err := insertAuditEvent(ctx, queries, command.Audit); err != nil {
			return err
		}
		credential, err = mapCredential(insertedPersonalTokenRow(row, command.Scopes, command.TeamIDs))
		return err
	})
	return credential, err
}

func (store *Store) ListPersonalAccessTokens(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) ([]developercredentialsdomain.Credential, error) {
	allowed, err := store.queries.IsCurrentWorkspaceMember(ctx, developercredentialssql.IsCurrentWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return nil, fmt.Errorf("authorize personal token listing: %w", mapDatabaseError(err))
	}
	if !allowed {
		return nil, developercredentialsdomain.ErrAccessDenied
	}
	rows, err := store.queries.ListPersonalAccessTokens(ctx, developercredentialssql.ListPersonalAccessTokensParams{
		WorkspaceID: workspaceID,
		UserID:      uuidPointer(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("list personal access tokens: %w", mapDatabaseError(err))
	}
	credentials := make([]developercredentialsdomain.Credential, 0, len(rows))
	for _, row := range rows {
		credential, err := mapCredential(listedPersonalTokenRow(row))
		if err != nil {
			return nil, fmt.Errorf("map personal access token: %w", err)
		}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func (store *Store) RotatePersonalAccessToken(
	ctx context.Context,
	command developercredentialsdomain.RotateCredential,
) (developercredentialsdomain.Credential, error) {
	digestKeyVersion, err := safecast.Uint32ToInt32(command.NewCredential.DigestKey.Version)
	if err != nil {
		return developercredentialsdomain.Credential{}, fmt.Errorf("validate rotated personal token digest key version: %w", err)
	}
	var credential developercredentialsdomain.Credential
	err = store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		_, err := queries.LockPersonalAccessTokenForRotation(ctx, developercredentialssql.LockPersonalAccessTokenForRotationParams{
			CredentialID: command.OldCredentialID,
			WorkspaceID:  command.WorkspaceID,
			RotatedAt:    command.RotatedAt,
			UserID:       uuidPointer(command.ActorUserID),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return developercredentialsdomain.ErrCredentialRotationConflict
		}
		if err != nil {
			return fmt.Errorf("lock personal access token for rotation: %w", err)
		}
		row, err := queries.InsertRotatedPersonalAccessToken(ctx, developercredentialssql.InsertRotatedPersonalAccessTokenParams{
			NewCredentialID:  command.NewCredential.ID,
			LookupPrefix:     command.NewCredential.LookupPrefix,
			SecretDigest:     append([]byte(nil), command.NewCredential.SecretDigest...),
			TokenVersion:     command.NewCredential.TokenVersion,
			DigestKeyID:      command.NewCredential.DigestKey.ID,
			DigestKeyVersion: digestKeyVersion,
			ExpiresAt:        command.ExpiresAt,
			UserID:           uuidPointer(command.ActorUserID),
			CreatedAt:        command.RotatedAt,
			OldCredentialID:  command.OldCredentialID,
			WorkspaceID:      command.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("insert rotated personal access token: %w", err)
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
		grants, err := queries.GetCredentialGrants(ctx, developercredentialssql.GetCredentialGrantsParams{
			CredentialID: command.NewCredential.ID,
		})
		if err != nil {
			return fmt.Errorf("load rotated personal token grants: %w", err)
		}
		credential, err = mapCredential(rotatedPersonalTokenRow(row, grants.Scopes, grants.TeamRestrictions))
		return err
	})
	return credential, err
}

func (store *Store) RevokePersonalAccessToken(
	ctx context.Context,
	command developercredentialsdomain.RevokeCredential,
) error {
	return store.withinTransaction(ctx, func(queries developercredentialssql.Querier) error {
		if _, err := queries.RevokePersonalAccessToken(ctx, developercredentialssql.RevokePersonalAccessTokenParams{
			RevokedAt:     timePointer(command.RevokedAt),
			UserID:        uuidPointer(command.ActorUserID),
			RevokedReason: stringPointer(command.Reason),
			CredentialID:  command.CredentialID,
			WorkspaceID:   command.WorkspaceID,
		}); err != nil {
			return fmt.Errorf("revoke personal access token: %w", err)
		}
		command.Audit.ReasonCode = command.Reason
		return insertAuditEvent(ctx, queries, command.Audit)
	})
}

func insertedPersonalTokenRow(
	row developercredentialssql.InsertPersonalAccessTokenRow,
	scopes []platformauth.Scope,
	teamIDs []uuid.UUID,
) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name, row.LookupPrefix,
		row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID, row.RotatedAt, row.RevokedAt,
		row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID, row.CreatedAt, scopeStrings(scopes), teamIDs)
}

func listedPersonalTokenRow(row developercredentialssql.ListPersonalAccessTokensRow) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name, row.LookupPrefix,
		row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID, row.RotatedAt, row.RevokedAt,
		row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID, row.CreatedAt, row.Scopes, row.TeamRestrictions)
}

func rotatedPersonalTokenRow(
	row developercredentialssql.InsertRotatedPersonalAccessTokenRow,
	scopes []string,
	teamIDs []uuid.UUID,
) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name, row.LookupPrefix,
		row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID, row.RotatedAt, row.RevokedAt,
		row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID, row.CreatedAt, scopes, teamIDs)
}
