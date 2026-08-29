package developercredentialsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) LookupCredential(
	ctx context.Context,
	lookupPrefix string,
	kind developercredentialsdomain.CredentialKind,
	tokenVersion int16,
	authenticatedAt time.Time,
) (developercredentialsdomain.VerificationRecord, error) {
	row, err := store.queries.LookupCredentialForAuthentication(ctx, developercredentialssql.LookupCredentialForAuthenticationParams{
		LookupPrefix: lookupPrefix, CredentialKind: string(kind), TokenVersion: tokenVersion,
		AuthenticatedAt: authenticatedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	if err != nil {
		return developercredentialsdomain.VerificationRecord{}, fmt.Errorf("lookup developer credential: %w", err)
	}
	return mapVerificationRecord(row)
}

func (store *Store) ConfirmCredentialActiveAndTouch(
	ctx context.Context,
	credentialID uuid.UUID,
	usedAt time.Time,
	touchBefore time.Time,
) error {
	_, err := store.queries.ConfirmCredentialActiveAndTouch(ctx, developercredentialssql.ConfirmCredentialActiveAndTouchParams{
		CredentialID: credentialID, UsedAt: usedAt, TouchBefore: timePointer(touchBefore),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developercredentialsdomain.ErrAuthenticationFailed
	}
	if err != nil {
		return fmt.Errorf("confirm developer credential active state: %w", err)
	}
	return nil
}

func mapVerificationRecord(
	row developercredentialssql.LookupCredentialForAuthenticationRow,
) (developercredentialsdomain.VerificationRecord, error) {
	kind := developercredentialsdomain.CredentialKind(row.CredentialKind)
	if err := kind.Validate(); err != nil {
		return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	role := authorization.WorkspaceRole(row.WorkspaceRole)
	if err := authorization.ValidateWorkspaceRole(role); err != nil {
		return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	scopes, err := mapScopes(row.Scopes)
	if err != nil {
		return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	if row.DigestKeyVersion <= 0 || len(row.SecretDigest) != 32 {
		return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	if kind == developercredentialsdomain.CredentialPersonalAccessToken {
		if row.PrincipalKind != "human_user" || row.SubjectUserID == nil {
			return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
		}
	} else {
		if row.PrincipalKind != "service_account" || row.SubjectUserID != nil ||
			(role != authorization.WorkspaceRoleGuest && role != authorization.WorkspaceRoleMember) ||
			containsScope(scopes, platformauth.ScopeServiceAccountsManage) {
			return developercredentialsdomain.VerificationRecord{}, developercredentialsdomain.ErrAuthenticationFailed
		}
	}
	return developercredentialsdomain.VerificationRecord{
		CredentialID: row.CredentialID, WorkspaceID: row.WorkspaceID,
		PrincipalRecord: row.PrincipalRecordID, PrincipalKind: row.PrincipalKind,
		SubjectUserID: row.SubjectUserID, WorkspaceRole: role, CredentialKind: kind,
		LookupPrefix: row.LookupPrefix, SecretDigest: append([]byte(nil), row.SecretDigest...),
		TokenVersion: row.TokenVersion,
		DigestKey:    developercredentialsdomain.DigestKeyRef{ID: row.DigestKeyID, Version: uint32(row.DigestKeyVersion)},
		Scopes:       append([]platformauth.Scope(nil), scopes...), TeamIDs: append([]uuid.UUID(nil), row.TeamRestrictions...),
		ExpiresAt: row.ExpiresAt, LastUsedAt: row.LastUsedAt,
	}, nil
}

func containsScope(scopes []platformauth.Scope, expected platformauth.Scope) bool {
	for _, scope := range scopes {
		if scope == expected {
			return true
		}
	}
	return false
}
