package developeroauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

type sequentialIDs struct {
	mu   sync.Mutex
	next uint64
}

func (ids *sequentialIDs) NewID() (uuid.UUID, error) {
	ids.mu.Lock()
	defer ids.mu.Unlock()
	ids.next++
	return uuid.MustParse(fmt.Sprintf("00000000-0000-4000-8000-%012d", ids.next)), nil
}

type memoryRepository struct {
	application developeroauthdomain.Application
	grant       developeroauthdomain.Grant
	code        developeroauthdomain.AuthorizationCode
	refresh     developeroauthdomain.RefreshToken
	revoked     bool
}

func (repository *memoryRepository) CreateApplication(
	_ context.Context,
	command developeroauthdomain.RegisterApplication,
) (developeroauthdomain.Application, error) {
	repository.application = developeroauthdomain.Application{
		ID: command.ID, ClientID: command.ClientID, Name: command.Name,
		RegistrationKind: command.RegistrationKind, RedirectURIs: append([]string(nil), command.RedirectURIs...),
		ExpiresAt: command.ExpiresAt, CreatedAt: command.CreatedAt,
	}
	return repository.application, nil
}

func (repository *memoryRepository) GetActiveApplication(
	_ context.Context,
	clientID string,
	activeAt time.Time,
) (developeroauthdomain.Application, error) {
	if repository.application.ClientID != clientID || !repository.application.ExpiresAt.After(activeAt) {
		return developeroauthdomain.Application{}, developeroauthdomain.ErrApplicationNotFound
	}
	return repository.application, nil
}

func (repository *memoryRepository) AuthorizeUser(
	_ context.Context,
	command developeroauthdomain.AuthorizeUser,
) (developeroauthdomain.Grant, error) {
	repository.grant = developeroauthdomain.Grant{
		ID: command.GrantID, ApplicationID: command.Application.ID, ClientID: command.Application.ClientID,
		UserID: command.UserID, ActorKind: platformauth.PrincipalOAuthUser, Resource: command.Resource,
		Scopes: append([]string(nil), command.Scopes...), CreatedAt: command.AuthorizedAt,
	}
	repository.code = developeroauthdomain.AuthorizationCode{
		ID: command.Code.ID, ApplicationID: command.Application.ID, ClientID: command.Application.ClientID,
		GrantID: command.GrantID, UserID: command.UserID, ActorKind: platformauth.PrincipalOAuthUser,
		LookupPrefix: command.Code.LookupPrefix, Digest: append([]byte(nil), command.Code.Digest...),
		DigestKey: command.Code.DigestKey, RedirectURI: command.RedirectURI, Resource: command.Resource,
		CodeChallenge: command.CodeChallenge, Scopes: append([]string(nil), command.Scopes...),
		ExpiresAt: command.CodeExpiresAt,
	}
	return repository.grant, nil
}

func (repository *memoryRepository) ExchangeAuthorizationCode(
	_ context.Context,
	command developeroauthdomain.ExchangeAuthorizationCode,
	validate func(developeroauthdomain.AuthorizationCode) error,
) (developeroauthdomain.Grant, error) {
	if repository.code.LookupPrefix != command.LookupPrefix || repository.code.ConsumedAt != nil {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrAuthorizationCode
	}
	if err := validate(repository.code); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	repository.code.ConsumedAt = &command.UsedAt
	repository.refresh = developeroauthdomain.RefreshToken{
		ID: command.Refresh.ID, FamilyID: command.FamilyID, LookupPrefix: command.Refresh.LookupPrefix,
		Digest: append([]byte(nil), command.Refresh.Digest...), DigestKey: command.Refresh.DigestKey,
		ExpiresAt: command.FamilyExpiry, FamilyExpiresAt: command.FamilyExpiry, Grant: repository.grant,
	}
	return repository.grant, nil
}

func (repository *memoryRepository) RotateRefreshToken(
	_ context.Context,
	command developeroauthdomain.RotateRefreshToken,
	validate func(developeroauthdomain.RefreshToken) error,
) (developeroauthdomain.Grant, error) {
	if repository.revoked || repository.refresh.LookupPrefix != command.LookupPrefix {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrRefreshToken
	}
	if err := validate(repository.refresh); err != nil {
		return developeroauthdomain.Grant{}, err
	}
	parent := repository.refresh.ID
	repository.refresh = developeroauthdomain.RefreshToken{
		ID: command.Replacement.ID, FamilyID: repository.refresh.FamilyID, ParentTokenID: &parent,
		LookupPrefix: command.Replacement.LookupPrefix, Digest: append([]byte(nil), command.Replacement.Digest...),
		DigestKey: command.Replacement.DigestKey, ExpiresAt: repository.refresh.FamilyExpiresAt,
		FamilyExpiresAt: repository.refresh.FamilyExpiresAt, Grant: repository.grant,
	}
	return repository.grant, nil
}

func (repository *memoryRepository) RevokeRefreshToken(
	_ context.Context,
	lookupPrefix string,
	_ time.Time,
	validate func(developeroauthdomain.RefreshToken) error,
) error {
	if repository.refresh.LookupPrefix != lookupPrefix {
		return developeroauthdomain.ErrRefreshToken
	}
	if err := validate(repository.refresh); err != nil {
		return err
	}
	repository.revoked = true
	return nil
}

func (repository *memoryRepository) GetActiveGrant(
	_ context.Context,
	grantID uuid.UUID,
	applicationID uuid.UUID,
	resource string,
	_ time.Time,
) (developeroauthdomain.Grant, error) {
	if repository.revoked || repository.grant.ID != grantID || repository.grant.ApplicationID != applicationID ||
		repository.grant.Resource != resource {
		return developeroauthdomain.Grant{}, developeroauthdomain.ErrInvalidGrant
	}
	return repository.grant, nil
}

func (*memoryRepository) TouchGrant(context.Context, uuid.UUID, time.Time, time.Time) error {
	return nil
}

func TestOAuthAuthorizationCodePKCERotationAndAccessIdentity(t *testing.T) {
	t.Parallel()
	service, repository := newTestService(t)
	ctx := context.Background()
	application, err := service.RegisterPublicApplication(ctx, " Test integration ", []string{"https://client.example/callback"})
	require.NoError(t, err)
	userID := uuid.New()
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	digest := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])

	code, err := service.AuthorizeUser(ctx, AuthorizationRequest{
		ClientID: application.ClientID, UserID: userID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), Scopes: []string{developeroauthdomain.ScopeMCPAccess}, CodeChallenge: challenge,
	})
	require.NoError(t, err)
	require.Equal(t, "[REDACTED]", code.String())

	_, err = service.ExchangeAuthorizationCode(ctx, AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier + "wrong",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrAuthorizationCode)
	require.Nil(t, repository.code.ConsumedAt, "failed PKCE must not consume the code")

	pair, err := service.ExchangeAuthorizationCode(ctx, AuthorizationCodeExchange{
		Code: code.Reveal(), ClientID: application.ClientID, RedirectURI: application.RedirectURIs[0],
		Resource: service.Resource(), CodeVerifier: verifier,
	})
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken.Reveal())
	require.NotEmpty(t, pair.RefreshToken.Reveal())

	identity, err := service.VerifyAccessToken(ctx, pair.AccessToken.Reveal())
	require.NoError(t, err)
	require.Equal(t, userID, identity.UserID)
	require.Equal(t, application.ID, identity.ApplicationID)
	require.Equal(t, platformauth.PrincipalOAuthUser, identity.ActorKind)

	rotated, err := service.ExchangeRefreshToken(ctx, RefreshExchange{
		RefreshToken: pair.RefreshToken.Reveal(), ClientID: application.ClientID, Resource: service.Resource(),
	})
	require.NoError(t, err)
	require.NotEqual(t, pair.RefreshToken.Reveal(), rotated.RefreshToken.Reveal())

	require.NoError(t, service.RevokeRefreshToken(ctx, rotated.RefreshToken.Reveal()))
	_, err = service.VerifyAccessToken(ctx, rotated.AccessToken.Reveal())
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidGrant)
}

func TestPrepareAuthorizationRejectsUnknownScopeAndRedirect(t *testing.T) {
	t.Parallel()
	service, _ := newTestService(t)
	application, err := service.RegisterPublicApplication(context.Background(), "Client", []string{"https://client.example/callback"})
	require.NoError(t, err)
	request := AuthorizationRequest{
		ClientID: application.ClientID, RedirectURI: "https://attacker.example/callback",
		Resource: service.Resource(), CodeChallenge: base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
	}
	_, _, err = service.PrepareAuthorization(context.Background(), request)
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidRedirectURI)

	request.Resource = "https://wrong.example/mcp"
	applicationOnError, _, err := service.PrepareAuthorization(context.Background(), request)
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidRedirectURI)
	require.Equal(t, uuid.Nil, applicationOnError.ID, "an untrusted redirect must never be marked safe")

	request.RedirectURI = application.RedirectURIs[0]
	applicationOnError, _, err = service.PrepareAuthorization(context.Background(), request)
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidResource)
	require.Equal(t, application.ID, applicationOnError.ID, "post-redirect errors may use the exact registered callback")

	request.Resource = service.Resource()
	request.Scopes = []string{"admin:*"}
	_, _, err = service.PrepareAuthorization(context.Background(), request)
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidScope)
}

func newTestService(t *testing.T) (*Service, *memoryRepository) {
	t.Helper()
	repository := &memoryRepository{}
	tokens, err := newTokenManager(TokenKeyringConfig{
		Active: developeroauthdomain.DigestKeyRef{ID: "test"},
		Keys: []DigestKey{{
			Ref: developeroauthdomain.DigestKeyRef{ID: "test"}, Material: bytes.Repeat([]byte{0x7a}, digestKeyBytes),
		}},
	}, bytes.NewReader(testRandomBytes(1024)))
	require.NoError(t, err)
	service, err := newService(
		repository,
		tokens,
		fixedClock{now: time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)},
		&sequentialIDs{},
		bytes.NewReader(testRandomBytes(1024)),
		Config{
			Issuer: "https://api.fortyone.app", Resource: "https://api.fortyone.app/mcp",
			AccessTokenSigningKey: "test-oauth-access-signing-key-001",
		},
	)
	require.NoError(t, err)
	return service, repository
}

func testRandomBytes(size int) []byte {
	result := make([]byte, size)
	for index := range result {
		result[index] = byte(index%251 + 1)
	}
	return result
}
