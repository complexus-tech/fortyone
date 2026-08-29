package developercredentialsrepository

import (
	"context"
	"fmt"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func insertCredentialGrants(
	ctx context.Context,
	queries developercredentialssql.Querier,
	credentialID uuid.UUID,
	workspaceID uuid.UUID,
	scopes []platformauth.Scope,
	teamIDs []uuid.UUID,
) error {
	if err := queries.InsertCredentialScopes(ctx, developercredentialssql.InsertCredentialScopesParams{
		CredentialID: credentialID,
		Scopes:       scopeStrings(scopes),
	}); err != nil {
		return fmt.Errorf("insert credential scopes: %w", err)
	}
	if len(teamIDs) == 0 {
		return nil
	}
	if err := queries.InsertCredentialTeamRestrictions(ctx, developercredentialssql.InsertCredentialTeamRestrictionsParams{
		CredentialID: credentialID,
		WorkspaceID:  workspaceID,
		TeamIds:      teamIDs,
	}); err != nil {
		return fmt.Errorf("insert credential team restrictions: %w", err)
	}
	return nil
}

func copyCredentialGrants(
	ctx context.Context,
	queries developercredentialssql.Querier,
	oldCredentialID uuid.UUID,
	newCredentialID uuid.UUID,
) error {
	if err := queries.CopyCredentialScopes(ctx, developercredentialssql.CopyCredentialScopesParams{
		NewCredentialID: newCredentialID,
		OldCredentialID: oldCredentialID,
	}); err != nil {
		return fmt.Errorf("copy credential scopes: %w", err)
	}
	if err := queries.CopyCredentialTeamRestrictions(ctx, developercredentialssql.CopyCredentialTeamRestrictionsParams{
		NewCredentialID: newCredentialID,
		OldCredentialID: oldCredentialID,
	}); err != nil {
		return fmt.Errorf("copy credential team restrictions: %w", err)
	}
	return nil
}

func markCredentialRotated(
	ctx context.Context,
	queries developercredentialssql.Querier,
	command developercredentialsdomain.RotateCredential,
) error {
	revokeImmediately := !command.OverlapUntil.After(command.RotatedAt)
	rows, err := queries.MarkCredentialRotated(ctx, developercredentialssql.MarkCredentialRotatedParams{
		RotatedAt:         command.RotatedAt,
		OverlapExpiresAt:  command.OverlapUntil,
		RevokeImmediately: revokeImmediately,
		UserID:            command.ActorUserID,
		CredentialID:      command.OldCredentialID,
	})
	if err != nil {
		return fmt.Errorf("mark credential rotated: %w", err)
	}
	if rows != 1 {
		return developercredentialsdomain.ErrCredentialRotationConflict
	}
	return nil
}

func scopeStrings(scopes []platformauth.Scope) []string {
	values := make([]string, len(scopes))
	for index, scope := range scopes {
		values[index] = string(scope)
	}
	return values
}

func baseCredentialRow(
	credentialID uuid.UUID,
	workspaceID uuid.UUID,
	principalID uuid.UUID,
	kind string,
	name string,
	lookupPrefix string,
	tokenVersion int16,
	expiresAt time.Time,
	lastUsedAt *time.Time,
	rotatedFromID *uuid.UUID,
	rotatedAt *time.Time,
	revokedAt *time.Time,
	revokedBy *uuid.UUID,
	revokedReason *string,
	createdBy *uuid.UUID,
	createdAt time.Time,
	scopes []string,
	teamIDs []uuid.UUID,
) credentialRow {
	return credentialRow{
		credentialID: credentialID, workspaceID: workspaceID, principalID: principalID, kind: kind,
		name: name, lookupPrefix: lookupPrefix, tokenVersion: tokenVersion, expiresAt: expiresAt,
		lastUsedAt: lastUsedAt, rotatedFromID: rotatedFromID, rotatedAt: rotatedAt, revokedAt: revokedAt,
		revokedByUserID: revokedBy, revokedReason: revokedReason, createdByUserID: createdBy, createdAt: createdAt,
		scopes: scopes, teamRestrictions: teamIDs,
	}
}
