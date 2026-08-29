package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/jackc/pgx/v5"
)

func (store *Store) RotateRefreshToken(
	ctx context.Context,
	command developeroauthdomain.RotateRefreshToken,
	validate func(developeroauthdomain.RefreshToken) error,
) (developeroauthdomain.Grant, error) {
	if validate == nil {
		return developeroauthdomain.Grant{}, errors.New("OAuth refresh token validator is required")
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.Grant{}, err
	}
	defer rollback(ctx, tx)
	row, err := queries.GetOAuthRefreshTokenForUpdate(ctx, developeroauthsql.GetOAuthRefreshTokenForUpdateParams{
		LookupPrefix: command.LookupPrefix, ActiveAt: command.UsedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrRefreshToken
	}
	if err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("lock OAuth refresh token: %w", err)
	}
	record := mapRefreshToken(row)
	if err := validate(record); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if record.ConsumedAt != nil {
		reason := "refresh_token_reuse"
		if _, err := queries.RevokeOAuthRefreshTokenFamily(ctx, developeroauthsql.RevokeOAuthRefreshTokenFamilyParams{
			RevokedAt: &command.UsedAt, RevokedReason: &reason, FamilyID: row.FamilyID,
		}); err != nil {
			return developeroauthdomain.Grant{}, fmt.Errorf("revoke replayed OAuth refresh family: %w", err)
		}
		if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
			ID: uuidForAudit(), ApplicationID: &row.ApplicationID, GrantID: &row.GrantID, UserID: &row.UserID,
			Operation: "refresh_token.reuse_detected", Result: "denied", ReasonCode: &reason, CreatedAt: command.UsedAt,
		}); err != nil {
			return developeroauthdomain.Grant{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return developeroauthdomain.Grant{}, fmt.Errorf("commit OAuth refresh reuse revocation: %w", err)
		}
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrRefreshTokenReuse
	}
	if _, err := queries.ConsumeOAuthRefreshToken(ctx, developeroauthsql.ConsumeOAuthRefreshTokenParams{
		ConsumedAt: &command.UsedAt, RefreshTokenID: row.RefreshTokenID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return developeroauthdomain.Grant{}, developeroauthdomain.ErrRefreshTokenReuse
		}
		return developeroauthdomain.Grant{}, fmt.Errorf("consume OAuth refresh token: %w", err)
	}
	if err := createRefreshToken(
		ctx,
		queries,
		command.Replacement,
		row.FamilyID,
		&row.RefreshTokenID,
		command.UsedAt,
		row.FamilyExpiresAt,
	); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: uuidForAudit(), ApplicationID: &row.ApplicationID, GrantID: &row.GrantID, UserID: &row.UserID,
		Operation: "refresh_token.rotated", Result: "succeeded", CreatedAt: command.UsedAt,
	}); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("commit OAuth refresh rotation: %w", err)
	}
	return record.Grant, nil
}

func (store *Store) RevokeRefreshToken(
	ctx context.Context,
	lookupPrefix string,
	revokedAt time.Time,
	validate func(developeroauthdomain.RefreshToken) error,
) error {
	if validate == nil {
		return errors.New("OAuth refresh token validator is required")
	}
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	row, err := queries.GetOAuthRefreshTokenForUpdate(ctx, developeroauthsql.GetOAuthRefreshTokenForUpdateParams{
		LookupPrefix: lookupPrefix, ActiveAt: revokedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.ErrRefreshToken
	}
	if err != nil {
		return fmt.Errorf("lock OAuth refresh token for revocation: %w", err)
	}
	record := mapRefreshToken(row)
	if err := validate(record); err != nil {
		return err
	}
	reason := "client_revoked"
	if _, err := queries.RevokeOAuthRefreshTokenFamily(ctx, developeroauthsql.RevokeOAuthRefreshTokenFamilyParams{
		RevokedAt: &revokedAt, RevokedReason: &reason, FamilyID: row.FamilyID,
	}); err != nil {
		return fmt.Errorf("revoke OAuth refresh family: %w", err)
	}
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: uuidForAudit(), ApplicationID: &row.ApplicationID, GrantID: &row.GrantID, UserID: &row.UserID,
		Operation: "refresh_token.revoked", Result: "succeeded", CreatedAt: revokedAt,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit OAuth refresh revocation: %w", err)
	}
	return nil
}

func mapRefreshToken(row developeroauthsql.GetOAuthRefreshTokenForUpdateRow) developeroauthdomain.RefreshToken {
	return developeroauthdomain.RefreshToken{
		ID: row.RefreshTokenID, FamilyID: row.FamilyID, ParentTokenID: row.ParentTokenID,
		LookupPrefix: row.LookupPrefix, Digest: append([]byte(nil), row.SecretDigest...),
		DigestKey: developeroauthdomain.DigestKeyRef{ID: row.DigestKeyID}, ExpiresAt: row.ExpiresAt,
		ConsumedAt: row.ConsumedAt, FamilyExpiresAt: row.FamilyExpiresAt, FamilyRevokedAt: row.FamilyRevokedAt,
		Grant: developeroauthdomain.Grant{
			ID: row.GrantID, ApplicationID: row.ApplicationID, ClientID: row.ClientID,
			UserID: row.UserID, ActorKind: platformauth.PrincipalKind(row.ActorKind), Resource: row.Resource,
			Scopes: append([]string(nil), row.Scopes...),
		},
	}
}
