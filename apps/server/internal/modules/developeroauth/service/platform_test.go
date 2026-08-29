package developeroauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPlatformRoutesAuthorizationAndVerificationByExactResource(t *testing.T) {
	t.Parallel()

	mcp, repository := newTestService(t)
	api, err := newService(
		repository,
		mcp.tokens,
		mcp.clock,
		mcp.ids,
		bytes.NewReader(testRandomBytes(1024)),
		Config{
			Issuer: "https://api.fortyone.app", Resource: "https://api.fortyone.app/api/v1",
			ScopePolicy:           PublicAPIResourceScopePolicy(),
			AccessTokenSigningKey: "test-oauth-access-signing-key-001",
		},
	)
	require.NoError(t, err)
	platform, err := NewPlatform(mcp, api)
	require.NoError(t, err)
	require.Equal(t, mcp.Resource(), platform.Resource())
	require.Equal(t, []string{api.Resource(), mcp.Resource()}, platform.Resources())
	require.Contains(t, platform.SupportedScopes(api.Resource()), "stories:read")
	require.NotContains(t, platform.SupportedScopes(api.Resource()), developeroauthdomain.ScopeMCPAccess)

	ctx := context.Background()
	application, err := platform.RegisterPublicApplication(ctx, "API client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	digest := sha256.Sum256([]byte(verifier))
	request := AuthorizationRequest{
		ClientID: application.ClientID, UserID: uuid.New(), RedirectURI: application.RedirectURIs[0],
		Resource: api.Resource(), Scopes: []string{"stories:read"},
		CodeChallenge: base64.RawURLEncoding.EncodeToString(digest[:]),
	}
	code, err := platform.AuthorizeUser(ctx, request)
	require.NoError(t, err)
	pair, err := platform.ExchangeAuthorizationCode(ctx, AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: api.Resource(), CodeVerifier: verifier,
	})
	require.NoError(t, err)
	require.Equal(t, []string{developeroauthdomain.ScopeOfflineAccess, "stories:read"}, pair.Scopes)

	identity, err := api.VerifyAccessToken(ctx, pair.AccessToken.Reveal())
	require.NoError(t, err)
	require.Equal(t, api.Resource(), identity.Resource)
	_, err = platform.VerifyAccessToken(ctx, pair.AccessToken.Reveal())
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidGrant, "API token must not authenticate to default MCP resource")
}

func TestPlatformRejectsUnknownAndDuplicateResources(t *testing.T) {
	t.Parallel()

	mcp, _ := newTestService(t)
	_, err := NewPlatform(mcp, mcp)
	require.Error(t, err)

	platform, err := NewPlatform(mcp)
	require.NoError(t, err)
	application, err := platform.RegisterPublicApplication(context.Background(), "Test client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	preparedApplication, _, err := platform.PrepareAuthorization(context.Background(), AuthorizationRequest{
		ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: "https://api.fortyone.app/api/v1",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidResource)
	require.Equal(t, application.ID, preparedApplication.ID, "a trusted redirect remains available for an invalid_target response")
	preparedApplication, _, err = platform.PrepareAuthorization(context.Background(), AuthorizationRequest{
		ClientID: application.ClientID, RedirectURI: "https://attacker.example/callback",
		Resource: "https://api.fortyone.app/api/v1",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidRedirectURI)
	require.Equal(t, uuid.Nil, preparedApplication.ID, "an unregistered redirect must never receive an OAuth error")
	require.Nil(t, platform.SupportedScopes("https://api.fortyone.app/api/v1"))
}
