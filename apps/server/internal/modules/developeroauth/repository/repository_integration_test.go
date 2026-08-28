//go:build integration

package developeroauthrepository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestOAuthRepositoryPersistsOneTimeSecretsAndFencesRefreshReplay(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	userID := insertOAuthUser(t, ctx, postgres.Pool, true)
	service := newIntegrationOAuthService(t, postgres.Pool)
	application, err := service.RegisterPublicApplication(ctx, "Integration client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	verifierDigest := sha256.Sum256([]byte(verifier))
	code, err := service.AuthorizeUser(ctx, developeroauth.AuthorizationRequest{
		ClientID: application.ClientID, UserID: userID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), Scopes: []string{developeroauthdomain.ScopeMCPAccess},
		CodeChallenge: base64.RawURLEncoding.EncodeToString(verifierDigest[:]),
	})
	require.NoError(t, err)

	var codeDigest []byte
	var codeConsumedAt *time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT secret_digest, consumed_at
		FROM oauth_authorization_codes
		WHERE application_id = $1
	`, application.ID).Scan(&codeDigest, &codeConsumedAt))
	require.Len(t, codeDigest, 32)
	require.Nil(t, codeConsumedAt)
	require.NotContains(t, string(codeDigest), code.Reveal())

	_, err = service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: "wrong-client", RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidClient)
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT consumed_at FROM oauth_authorization_codes WHERE application_id = $1
	`, application.ID).Scan(&codeConsumedAt))
	require.Nil(t, codeConsumedAt, "a binding failure must not consume the authorization code")

	pair, err := service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.NoError(t, err)
	_, err = service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrAuthorizationCodeUsed)

	type refreshResult struct {
		pair developeroauthdomain.TokenPair
		err  error
	}
	start := make(chan struct{})
	results := make(chan refreshResult, 2)
	var waitGroup sync.WaitGroup
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			rotated, rotateErr := service.ExchangeRefreshToken(ctx, developeroauth.RefreshExchange{
				RefreshToken: pair.RefreshToken.Reveal(), ClientID: application.ClientID, Resource: service.Resource(),
			})
			results <- refreshResult{pair: rotated, err: rotateErr}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	var succeeded, replayed int
	var replacement developeroauthdomain.TokenPair
	for result := range results {
		switch {
		case result.err == nil:
			succeeded++
			replacement = result.pair
		case errors.Is(result.err, developeroauthdomain.ErrRefreshTokenReuse):
			replayed++
		default:
			t.Fatalf("unexpected concurrent refresh result: %v", result.err)
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, replayed)
	require.NotEmpty(t, replacement.RefreshToken.Reveal())

	var revokedReason *string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT revoked_reason
		FROM oauth_refresh_token_families
		WHERE grant_id = (SELECT grant_id FROM oauth_grants WHERE application_id = $1)
	`, application.ID).Scan(&revokedReason))
	require.NotNil(t, revokedReason)
	require.Equal(t, "refresh_token_reuse", *revokedReason)
	_, err = service.ExchangeRefreshToken(ctx, developeroauth.RefreshExchange{
		RefreshToken: replacement.RefreshToken.Reveal(), ClientID: application.ClientID, Resource: service.Resource(),
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrRefreshToken)

	identity, err := service.VerifyAccessToken(ctx, replacement.AccessToken.Reveal())
	require.NoError(t, err)
	require.Equal(t, userID, identity.UserID)
	_, err = postgres.Pool.Exec(ctx, `UPDATE users SET is_active = FALSE WHERE user_id = $1`, userID)
	require.NoError(t, err)
	_, err = service.VerifyAccessToken(ctx, replacement.AccessToken.Reveal())
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidGrant)

	_, err = postgres.Pool.Exec(ctx, `UPDATE oauth_audit_events SET result = 'failed'`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM oauth_audit_events`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")
}

func TestOAuthRepositoryInvalidatesSupersededAuthorizationCodes(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	userID := insertOAuthUser(t, ctx, postgres.Pool, true)
	service := newIntegrationOAuthService(t, postgres.Pool)
	application, err := service.RegisterPublicApplication(ctx, "Scope replacement client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	verifierDigest := sha256.Sum256([]byte(verifier))
	request := developeroauth.AuthorizationRequest{
		ClientID: application.ClientID, UserID: userID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), Scopes: []string{developeroauthdomain.ScopeMCPAccess},
		CodeChallenge: base64.RawURLEncoding.EncodeToString(verifierDigest[:]),
	}
	firstCode, err := service.AuthorizeUser(ctx, request)
	require.NoError(t, err)

	secondCode, err := service.AuthorizeUser(ctx, request)
	require.NoError(t, err)

	_, err = service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
		Code: firstCode.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrAuthorizationCodeUsed)
	pair, err := service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
		Code: secondCode.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		developeroauthdomain.ScopeMCPAccess,
		developeroauthdomain.ScopeOfflineAccess,
	}, pair.Scopes)

	var outstandingCodes int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT count(*)
		FROM oauth_authorization_codes
		WHERE application_id = $1 AND consumed_at IS NULL
	`, application.ID).Scan(&outstandingCodes))
	require.Zero(t, outstandingCodes)
}

func TestOAuthRepositoryAuthorizationAndReauthorizationUseOneLockOrder(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	userID := insertOAuthUser(t, ctx, postgres.Pool, true)
	service := newIntegrationOAuthService(t, postgres.Pool)
	application, err := service.RegisterPublicApplication(ctx, "Concurrent consent client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	verifierDigest := sha256.Sum256([]byte(verifier))
	request := developeroauth.AuthorizationRequest{
		ClientID: application.ClientID, UserID: userID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), Scopes: []string{developeroauthdomain.ScopeMCPAccess},
		CodeChallenge: base64.RawURLEncoding.EncodeToString(verifierDigest[:]),
	}
	expectedScopes := []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess}

	for iteration := 0; iteration < 12; iteration++ {
		oldCode, authorizeErr := service.AuthorizeUser(ctx, request)
		require.NoError(t, authorizeErr)
		start := make(chan struct{})
		errorsByOperation := make(chan error, 2)
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			<-start
			pair, exchangeErr := service.ExchangeAuthorizationCode(ctx, developeroauth.AuthorizationCodeExchange{
				Code: oldCode.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
				Resource: service.Resource(), CodeVerifier: verifier,
			})
			if exchangeErr == nil && !slices.Equal(pair.Scopes, expectedScopes) {
				errorsByOperation <- fmt.Errorf("authorization code scopes changed during exchange: got %v, want %v", pair.Scopes, expectedScopes)
				return
			}
			if exchangeErr != nil && !errors.Is(exchangeErr, developeroauthdomain.ErrAuthorizationCodeUsed) {
				errorsByOperation <- exchangeErr
				return
			}
			errorsByOperation <- nil
		}()
		go func() {
			defer waitGroup.Done()
			<-start
			_, reauthorizeErr := service.AuthorizeUser(ctx, request)
			errorsByOperation <- reauthorizeErr
		}()
		close(start)
		waitGroup.Wait()
		for range 2 {
			require.NoError(t, <-errorsByOperation)
		}
	}
}

func TestOAuthRepositoryRejectsInactiveAuthorizationSubject(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	userID := insertOAuthUser(t, ctx, postgres.Pool, false)
	service := newIntegrationOAuthService(t, postgres.Pool)
	application, err := service.RegisterPublicApplication(ctx, "Inactive subject client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	verifierDigest := sha256.Sum256([]byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"))
	_, err = service.AuthorizeUser(ctx, developeroauth.AuthorizationRequest{
		ClientID: application.ClientID, UserID: userID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), Scopes: []string{developeroauthdomain.ScopeMCPAccess},
		CodeChallenge: base64.RawURLEncoding.EncodeToString(verifierDigest[:]),
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrAuthorizationDenied)
}

func newIntegrationOAuthService(t *testing.T, pool *pgxpool.Pool) *developeroauth.Service {
	t.Helper()
	keyRef := developeroauthdomain.DigestKeyRef{ID: "integration"}
	tokens, err := developeroauth.NewTokenManager(developeroauth.TokenKeyringConfig{
		Active: keyRef,
		Keys: []developeroauth.DigestKey{{
			Ref: keyRef, Material: bytes.Repeat([]byte{0x64}, 32),
		}},
	})
	require.NoError(t, err)
	repository, err := New(pool)
	require.NoError(t, err)
	service, err := developeroauth.New(
		repository,
		tokens,
		testkit.NewFixedClock(time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)),
		developeroauth.RandomIDGenerator{},
		developeroauth.Config{
			Issuer: "https://api.fortyone.app", Resource: "https://api.fortyone.app/mcp",
			AccessTokenSigningKey: "integration-oauth-signing-key-001",
		},
	)
	require.NoError(t, err)
	return service
}

func insertOAuthUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, 'OAuth integration user', $4, FALSE)
	`, id, "oauth-"+id.String(), "oauth-"+id.String()+"@example.com", active)
	require.NoError(t, err)
	return id
}
