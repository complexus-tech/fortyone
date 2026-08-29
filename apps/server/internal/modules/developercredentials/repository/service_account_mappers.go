package developercredentialsrepository

import (
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func insertedServiceAccountKeyRow(
	row developercredentialssql.InsertServiceAccountKeyRow,
	scopes []platformauth.Scope,
	teamIDs []uuid.UUID,
) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name,
		row.LookupPrefix, row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID,
		row.RotatedAt, row.RevokedAt, row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID,
		row.CreatedAt, scopeStrings(scopes), teamIDs)
}

func listedServiceAccountKeyRow(row developercredentialssql.ListServiceAccountKeysRow) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name,
		row.LookupPrefix, row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID,
		row.RotatedAt, row.RevokedAt, row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID,
		row.CreatedAt, row.Scopes, row.TeamRestrictions)
}

func rotatedServiceAccountKeyRow(
	row developercredentialssql.InsertRotatedServiceAccountKeyRow,
	scopes []string,
	teamIDs []uuid.UUID,
) credentialRow {
	return baseCredentialRow(row.CredentialID, row.WorkspaceID, row.PrincipalID, row.Kind, row.Name,
		row.LookupPrefix, row.TokenVersion, row.ExpiresAt, row.LastUsedAt, row.RotatedFromID,
		row.RotatedAt, row.RevokedAt, row.RevokedByUserID, row.RevokedReason, row.CreatedByUserID,
		row.CreatedAt, scopes, teamIDs)
}
