package developeroauth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type applicationActorRepositoryStub struct {
	record          developeroauthdomain.ClientSecretRecord
	installation    developeroauthdomain.ApplicationInstallation
	authentication  developeroauthdomain.AuthenticateApplicationCredential
	authenticateErr error
	commitErr       error
	activeErr       error
	authCalls       int
	activeCalls     int
	touchCalls      int
}

func (repository *applicationActorRepositoryStub) AuthenticateApplication(
	_ context.Context,
	command developeroauthdomain.AuthenticateApplicationCredential,
	verify func(developeroauthdomain.ClientSecretRecord, developeroauthdomain.ApplicationInstallation) error,
) (developeroauthdomain.ApplicationInstallation, error) {
	repository.authCalls++
	repository.authentication = command
	if repository.authenticateErr != nil || command.LookupPrefix != repository.record.Material.LookupPrefix ||
		command.InstallationID != repository.installation.ID {
		return developeroauthdomain.ApplicationInstallation{}, errors.Join(
			developeroauthdomain.ErrInvalidClient,
			repository.authenticateErr,
		)
	}
	if err := verify(repository.record, repository.installation); err != nil {
		return developeroauthdomain.ApplicationInstallation{}, err
	}
	if repository.commitErr != nil {
		return developeroauthdomain.ApplicationInstallation{}, repository.commitErr
	}
	return repository.installation, nil
}

func (repository *applicationActorRepositoryStub) GetActiveApplicationInstallation(
	_ context.Context,
	installationID uuid.UUID,
	applicationID uuid.UUID,
	resource string,
	_ time.Time,
) (developeroauthdomain.ApplicationInstallation, error) {
	repository.activeCalls++
	if repository.activeErr != nil || repository.installation.ID != installationID ||
		repository.installation.ApplicationID != applicationID || repository.installation.Resource != resource {
		return developeroauthdomain.ApplicationInstallation{}, errors.Join(
			developeroauthdomain.ErrInvalidGrant,
			repository.activeErr,
		)
	}
	return repository.installation, nil
}

func (repository *applicationActorRepositoryStub) TouchApplicationInstallation(
	context.Context,
	uuid.UUID,
	time.Time,
	time.Time,
) error {
	repository.touchCalls++
	return nil
}

func TestClientCredentialsIssueAndVerifyAnInstallationActor(t *testing.T) {
	t.Parallel()

	service, repository, secret := newApplicationActorTestService(t)
	request := ClientCredentialsExchange{
		ClientID: repository.installation.ClientID, ClientSecret: secret,
		InstallationID: repository.installation.ID, Resource: repository.installation.Resource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "request-client-credentials",
	}

	issued, err := service.ExchangeClientCredentials(context.Background(), request)
	require.NoError(t, err)
	require.NotEmpty(t, issued.AccessToken.Reveal())
	require.Equal(t, []string{string(platformauth.ScopeStoriesWrite)}, issued.Scopes)
	require.Equal(t, 1, repository.authCalls)
	require.Equal(t, repository.installation.ID, repository.authentication.InstallationID)
	require.NotEqual(t, repository.installation.ID, repository.authentication.AccessTokenID)
	require.NotEqual(t, uuid.Nil, repository.authentication.AuditID)
	require.Equal(t, "request-client-credentials", repository.authentication.RequestID)

	identity, err := service.VerifyAccessToken(context.Background(), issued.AccessToken.Reveal())
	require.NoError(t, err)
	require.Equal(t, platformauth.PrincipalOAuthApplication, identity.ActorKind)
	require.Equal(t, repository.installation.PrincipalID, identity.PrincipalID)
	require.Equal(t, repository.installation.ID, identity.InstallationID)
	require.Equal(t, repository.installation.WorkspaceID, identity.WorkspaceID)
	require.Equal(t, uuid.Nil, identity.UserID)
	require.Equal(t, uuid.Nil, identity.GrantID)
	require.NotEqual(t, identity.InstallationID, identity.OAuthCredential, "JWT ID is audit-only, not the stable credential identity")
	require.Equal(t, repository.authentication.AccessTokenID, identity.OAuthCredential)
	require.Equal(t, 1, repository.activeCalls)
	require.Equal(t, 1, repository.touchCalls)
}

func TestClientCredentialsAndApplicationTokensFailClosedOnCurrentState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Service, *applicationActorRepositoryStub, *ClientCredentialsExchange)
		issue  bool
		error  error
	}{
		{name: "wrong client", error: developeroauthdomain.ErrInvalidClient, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.ClientID = "another-client"
		}},
		{name: "wrong secret", error: developeroauthdomain.ErrInvalidClient, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.ClientSecret = request.ClientSecret[:len(request.ClientSecret)-1] + "x"
		}},
		{name: "wrong installation", error: developeroauthdomain.ErrInvalidClient, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.InstallationID = uuid.New()
		}},
		{name: "wrong resource", error: developeroauthdomain.ErrInvalidResource, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.Resource = "https://api.fortyone.app/mcp"
		}},
		{name: "offline scope", error: developeroauthdomain.ErrInvalidScope, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.Scopes = []string{developeroauthdomain.ScopeOfflineAccess}
		}},
		{name: "read scope", error: developeroauthdomain.ErrInvalidScope, mutate: func(_ *Service, _ *applicationActorRepositoryStub, request *ClientCredentialsExchange) {
			request.Scopes = []string{string(platformauth.ScopeStoriesRead)}
		}},
		{name: "revoked installation rejects existing token", issue: true, error: developeroauthdomain.ErrInvalidGrant, mutate: func(_ *Service, repository *applicationActorRepositoryStub, _ *ClientCredentialsExchange) {
			repository.activeErr = developeroauthdomain.ErrInstallationRevoked
		}},
		{name: "scope narrowing rejects existing token", issue: true, error: developeroauthdomain.ErrInvalidGrant, mutate: func(_ *Service, repository *applicationActorRepositoryStub, _ *ClientCredentialsExchange) {
			repository.installation.Scopes = []string{string(platformauth.ScopeStoriesRead)}
		}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service, repository, secret := newApplicationActorTestService(t)
			request := ClientCredentialsExchange{
				ClientID: repository.installation.ClientID, ClientSecret: secret,
				InstallationID: repository.installation.ID, Resource: repository.installation.Resource,
				Scopes: []string{string(platformauth.ScopeStoriesWrite)},
			}
			if test.issue {
				issued, err := service.ExchangeClientCredentials(context.Background(), request)
				require.NoError(t, err)
				test.mutate(service, repository, &request)
				_, err = service.VerifyAccessToken(context.Background(), issued.AccessToken.Reveal())
				require.ErrorIs(t, err, test.error)
				return
			}
			test.mutate(service, repository, &request)
			_, err := service.ExchangeClientCredentials(context.Background(), request)
			require.ErrorIs(t, err, test.error)
		})
	}
}

func TestClientCredentialsAreUnavailableOnAnUnconfiguredAudience(t *testing.T) {
	t.Parallel()

	service, _ := newTestService(t)
	_, err := service.ExchangeClientCredentials(context.Background(), ClientCredentialsExchange{})
	require.ErrorIs(t, err, developeroauthdomain.ErrApplicationActorUnavailable)
}

func TestClientCredentialsFailClosedWhenImmutableAuditCannotCommit(t *testing.T) {
	t.Parallel()

	service, repository, secret := newApplicationActorTestService(t)
	repository.commitErr = errors.New("audit store unavailable")
	issued, err := service.ExchangeClientCredentials(context.Background(), ClientCredentialsExchange{
		ClientID: repository.installation.ClientID, ClientSecret: secret,
		InstallationID: repository.installation.ID, Resource: repository.installation.Resource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "request-audit-failure",
	})

	require.Error(t, err)
	require.NotErrorIs(t, err, developeroauthdomain.ErrInvalidClient)
	require.Contains(t, err.Error(), "audit store unavailable")
	require.Empty(t, issued.AccessToken.Reveal(), "a token must never escape when its immutable audit cannot commit")
	require.NotEqual(t, uuid.Nil, repository.authentication.AccessTokenID)
	require.NotEqual(t, uuid.Nil, repository.authentication.AuditID)
}

func newApplicationActorTestService(
	t *testing.T,
) (*Service, *applicationActorRepositoryStub, string) {
	t.Helper()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	applicationID, installationID, workspaceID, principalID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &applicationActorRepositoryStub{installation: developeroauthdomain.ApplicationInstallation{
		ID: installationID, ApplicationID: applicationID, ClientID: "f41_oauth_managed_client_0001",
		WorkspaceID: workspaceID, PrincipalID: principalID, Resource: "https://api.fortyone.app/api/v1",
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, Status: "active",
		InstalledBy: uuid.New(), CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour),
	}}
	tokens, err := newTokenManager(TokenKeyringConfig{
		Active: developeroauthdomain.DigestKeyRef{ID: "test"},
		Keys: []DigestKey{{
			Ref: developeroauthdomain.DigestKeyRef{ID: "test"}, Material: bytes.Repeat([]byte{0x55}, digestKeyBytes),
		}},
	}, bytes.NewReader(testRandomBytes(2048)))
	require.NoError(t, err)
	issuedSecret, err := tokens.Issue(developeroauthdomain.SecretClientSecret, uuid.New())
	require.NoError(t, err)
	repository.record = developeroauthdomain.ClientSecretRecord{
		Secret: developeroauthdomain.ClientSecret{
			ID: issuedSecret.Material.ID, ApplicationID: applicationID,
			LookupPrefix: issuedSecret.Material.LookupPrefix, ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now.Add(-time.Hour),
		},
		ClientID: repository.installation.ClientID,
		Material: issuedSecret.Material,
		Application: developeroauthdomain.Application{
			ID: applicationID, ClientID: repository.installation.ClientID, Name: "Managed app",
			RegistrationKind: "confidential", ExpiresAt: now.Add(365 * 24 * time.Hour), CreatedAt: now.Add(-time.Hour),
		},
	}
	service, err := newService(
		&memoryRepository{},
		tokens,
		fixedClock{now: now},
		&sequentialIDs{},
		bytes.NewReader(testRandomBytes(2048)),
		Config{
			Issuer: "https://api.fortyone.app", Resource: repository.installation.Resource,
			ScopePolicy:           PublicAPIResourceScopePolicy(),
			AccessTokenSigningKey: "test-oauth-access-signing-key-001",
			ApplicationActors:     repository, ApplicationActorScopes: PublicAPIApplicationActorScopePolicy(),
		},
	)
	require.NoError(t, err)
	return service, repository, issuedSecret.Plaintext.Reveal()
}
