package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (store *Store) ExchangeAuthorizationCode(
	ctx context.Context,
	command developeroauthdomain.ExchangeAuthorizationCode,
	validate func(developeroauthdomain.AuthorizationCode) error,
) (developeroauthdomain.Grant, error) {
	if validate == nil {
		return developeroauthdomain.Grant{}, errors.New("OAuth authorization code validator is required")
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.Grant{}, err
	}
	defer rollback(ctx, tx)
	row, err := queries.GetOAuthAuthorizationCodeForUpdate(ctx, developeroauthsql.GetOAuthAuthorizationCodeForUpdateParams{
		LookupPrefix: command.LookupPrefix, ActiveAt: command.UsedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrAuthorizationCode
	}
	if err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("lock OAuth authorization code: %w", err)
	}
	record := mapAuthorizationCode(row)
	if err := validate(record); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if record.ConsumedAt != nil {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrAuthorizationCodeUsed
	}
	if _, err := queries.ConsumeOAuthAuthorizationCode(ctx, developeroauthsql.ConsumeOAuthAuthorizationCodeParams{
		ConsumedAt: &command.UsedAt, AuthorizationCodeID: row.AuthorizationCodeID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return developeroauthdomain.Grant{}, developeroauthdomain.ErrAuthorizationCodeUsed
		}
		return developeroauthdomain.Grant{}, fmt.Errorf("consume OAuth authorization code: %w", err)
	}
	if err := queries.CreateOAuthRefreshTokenFamily(ctx, developeroauthsql.CreateOAuthRefreshTokenFamilyParams{
		FamilyID: command.FamilyID, GrantID: row.GrantID, Resource: row.Resource,
		CreatedAt: command.UsedAt, ExpiresAt: command.FamilyExpiry,
	}); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("create OAuth refresh family: %w", err)
	}
	if err := createRefreshToken(ctx, queries, command.Refresh, command.FamilyID, nil, command.UsedAt, command.FamilyExpiry); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: uuidForAudit(), ApplicationID: &row.ApplicationID, GrantID: &row.GrantID, UserID: &row.UserID,
		Operation: "authorization_code.exchanged", Result: "succeeded", CreatedAt: command.UsedAt,
	}); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("commit OAuth authorization code exchange: %w", err)
	}
	return grantFromAuthorizationCode(record), nil
}

func mapAuthorizationCode(row developeroauthsql.GetOAuthAuthorizationCodeForUpdateRow) developeroauthdomain.AuthorizationCode {
	return developeroauthdomain.AuthorizationCode{
		ID: row.AuthorizationCodeID, ApplicationID: row.ApplicationID, ClientID: row.ClientID,
		GrantID: row.GrantID, UserID: row.UserID, ActorKind: platformauth.PrincipalKind(row.ActorKind),
		LookupPrefix: row.LookupPrefix, Digest: append([]byte(nil), row.SecretDigest...),
		DigestKey: developeroauthdomain.DigestKeyRef{ID: row.DigestKeyID}, RedirectURI: row.RedirectURI,
		Resource: row.Resource, CodeChallenge: row.CodeChallenge, Scopes: append([]string(nil), row.Scopes...),
		ExpiresAt: row.ExpiresAt, ConsumedAt: row.ConsumedAt,
	}
}

func grantFromAuthorizationCode(record developeroauthdomain.AuthorizationCode) developeroauthdomain.Grant {
	return developeroauthdomain.Grant{
		ID: record.GrantID, ApplicationID: record.ApplicationID, ClientID: record.ClientID,
		UserID: record.UserID, ActorKind: record.ActorKind, Resource: record.Resource,
		Scopes: append([]string(nil), record.Scopes...),
	}
}

func createRefreshToken(
	ctx context.Context,
	queries *developeroauthsql.Queries,
	material developeroauthdomain.SecretMaterial,
	familyID uuid.UUID,
	parentID *uuid.UUID,
	createdAt time.Time,
	expiresAt time.Time,
) error {
	if err := queries.CreateOAuthRefreshToken(ctx, developeroauthsql.CreateOAuthRefreshTokenParams{
		RefreshTokenID: material.ID, FamilyID: familyID, ParentTokenID: parentID,
		LookupPrefix: material.LookupPrefix, SecretDigest: material.Digest, DigestKeyID: material.DigestKey.ID,
		CreatedAt: createdAt, ExpiresAt: expiresAt,
	}); err != nil {
		return fmt.Errorf("create OAuth refresh token: %w", err)
	}
	return nil
}
