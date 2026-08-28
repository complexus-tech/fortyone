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

func (store *Store) AuthorizeUser(
	ctx context.Context,
	command developeroauthdomain.AuthorizeUser,
) (developeroauthdomain.Grant, error) {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.Grant{}, err
	}
	defer rollback(ctx, tx)
	// Authorization codes derive their scopes from the grant. Invalidate every
	// outstanding code before replacing that grant's scopes so an older code can
	// never inherit permissions approved by a later consent decision. Taking the
	// code locks first also keeps the lock order consistent with code exchange.
	if err := queries.InvalidateUnconsumedOAuthAuthorizationCodesForSubject(
		ctx,
		developeroauthsql.InvalidateUnconsumedOAuthAuthorizationCodesForSubjectParams{
			InvalidatedAt: &command.AuthorizedAt,
			ApplicationID: command.Application.ID,
			Resource:      command.Resource,
			UserID:        command.UserID,
		},
	); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("invalidate superseded OAuth authorization codes: %w", err)
	}
	row, err := queries.UpsertOAuthUserGrant(ctx, developeroauthsql.UpsertOAuthUserGrantParams{
		GrantID: command.GrantID, ApplicationID: command.Application.ID, ClientID: command.Application.ClientID,
		UserID: command.UserID, Resource: command.Resource, GrantedAt: command.AuthorizedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrAuthorizationDenied
	}
	if err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("upsert OAuth user grant: %w", err)
	}
	reauthorizationReason := "grant_reauthorized"
	if err := queries.RevokeActiveOAuthRefreshTokenFamiliesForGrant(
		ctx,
		developeroauthsql.RevokeActiveOAuthRefreshTokenFamiliesForGrantParams{
			RevokedAt: &command.AuthorizedAt, RevokedReason: &reauthorizationReason, GrantID: row.GrantID,
		},
	); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("revoke superseded OAuth refresh families: %w", err)
	}
	if err := queries.DeleteOAuthGrantScopes(ctx, developeroauthsql.DeleteOAuthGrantScopesParams{GrantID: row.GrantID}); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("replace OAuth grant scopes: %w", err)
	}
	for _, scope := range command.Scopes {
		if err := queries.CreateOAuthGrantScope(ctx, developeroauthsql.CreateOAuthGrantScopeParams{
			GrantID: row.GrantID, Scope: scope,
		}); err != nil {
			return developeroauthdomain.Grant{}, fmt.Errorf("create OAuth grant scope: %w", err)
		}
	}
	if err := queries.CreateOAuthAuthorizationCode(ctx, developeroauthsql.CreateOAuthAuthorizationCodeParams{
		AuthorizationCodeID: command.Code.ID,
		ApplicationID:       row.ApplicationID,
		GrantID:             row.GrantID,
		LookupPrefix:        command.Code.LookupPrefix,
		SecretDigest:        command.Code.Digest,
		DigestKeyID:         command.Code.DigestKey.ID,
		RedirectURI:         command.RedirectURI,
		Resource:            command.Resource,
		CodeChallenge:       command.CodeChallenge,
		CreatedAt:           command.AuthorizedAt,
		ExpiresAt:           command.CodeExpiresAt,
	}); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("create OAuth authorization code: %w", err)
	}
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: uuidForAudit(), ApplicationID: &row.ApplicationID, GrantID: &row.GrantID, UserID: &row.UserID,
		Operation: "grant.authorized", Result: "succeeded", CreatedAt: command.AuthorizedAt,
	}); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("commit OAuth user authorization: %w", err)
	}
	return developeroauthdomain.Grant{
		ID: row.GrantID, ApplicationID: row.ApplicationID, ClientID: command.Application.ClientID,
		UserID: row.UserID, ActorKind: platformauth.PrincipalOAuthUser, Resource: row.Resource,
		Scopes: append([]string(nil), command.Scopes...), CreatedAt: row.CreatedAt, LastUsedAt: row.LastUsedAt,
	}, nil
}

func (store *Store) GetActiveGrant(
	ctx context.Context,
	grantID uuid.UUID,
	applicationID uuid.UUID,
	resource string,
	activeAt time.Time,
) (developeroauthdomain.Grant, error) {
	row, err := store.queries.GetActiveOAuthGrant(ctx, developeroauthsql.GetActiveOAuthGrantParams{
		GrantID: grantID, ApplicationID: applicationID, Resource: resource, ActiveAt: activeAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrInvalidGrant
	}
	if err != nil {
		return developeroauthdomain.Grant{}, fmt.Errorf("get active OAuth grant: %w", err)
	}
	actorKind := platformauth.PrincipalKind(row.ActorKind)
	if actorKind != platformauth.PrincipalOAuthUser {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrInvalidGrant
	}
	return developeroauthdomain.Grant{
		ID: row.GrantID, ApplicationID: row.ApplicationID, ClientID: row.ClientID,
		UserID: row.UserID, ActorKind: actorKind, Resource: row.Resource,
		Scopes: append([]string(nil), row.Scopes...), CreatedAt: row.CreatedAt, LastUsedAt: row.LastUsedAt,
	}, nil
}

func (store *Store) TouchGrant(ctx context.Context, grantID uuid.UUID, usedAt, touchBefore time.Time) error {
	if err := store.queries.TouchOAuthGrant(ctx, developeroauthsql.TouchOAuthGrantParams{
		UsedAt: &usedAt, GrantID: grantID, TouchBefore: &touchBefore,
	}); err != nil {
		return fmt.Errorf("touch OAuth grant: %w", err)
	}
	return nil
}
