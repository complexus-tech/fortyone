package agentreadinesshttp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type stubOAuthPlatform struct {
	resource                     string
	apiResource                  string
	application                  developeroauthdomain.Application
	pair                         developeroauthdomain.TokenPair
	applicationToken             developeroauthdomain.ApplicationAccessToken
	identity                     developeroauthdomain.AccessIdentity
	exchangeCodeErr              error
	exchangeClientCredentialsErr error
	mu                           sync.Mutex
	authorized                   developeroauth.AuthorizationRequest
	clientCredentials            developeroauth.ClientCredentialsExchange
}

func (stub *stubOAuthPlatform) Resource() string {
	return stub.resource
}

func (stub *stubOAuthPlatform) Resources() []string {
	return []string{stub.apiResource, stub.resource}
}

func (stub *stubOAuthPlatform) SupportedScopes(resource string) []string {
	switch resource {
	case stub.resource:
		return []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess}
	case stub.apiResource:
		return []string{developeroauthdomain.ScopeOfflineAccess, "stories:read", "stories:write"}
	default:
		return nil
	}
}

func (stub *stubOAuthPlatform) RegisterPublicApplication(
	_ context.Context,
	name string,
	redirectURIs []string,
) (developeroauthdomain.Application, error) {
	application := stub.application
	application.Name = name
	application.RedirectURIs = append([]string(nil), redirectURIs...)
	return application, nil
}

func (stub *stubOAuthPlatform) PrepareAuthorization(
	_ context.Context,
	request developeroauth.AuthorizationRequest,
) (developeroauthdomain.Application, []string, error) {
	if request.ClientID != stub.application.ClientID {
		return developeroauthdomain.Application{}, nil, developeroauthdomain.ErrApplicationNotFound
	}
	if len(stub.application.RedirectURIs) == 0 || request.RedirectURI != stub.application.RedirectURIs[0] {
		return developeroauthdomain.Application{}, nil, developeroauthdomain.ErrInvalidRedirectURI
	}
	if request.Resource != stub.resource && request.Resource != stub.apiResource {
		return stub.application, nil, developeroauthdomain.ErrInvalidResource
	}
	if request.Resource == stub.apiResource {
		return stub.application, []string{developeroauthdomain.ScopeOfflineAccess, "stories:read"}, nil
	}
	return stub.application, []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess}, nil
}

func (stub *stubOAuthPlatform) AuthorizeUser(
	_ context.Context,
	request developeroauth.AuthorizationRequest,
) (developeroauthdomain.PlaintextSecret, error) {
	stub.mu.Lock()
	stub.authorized = request
	stub.mu.Unlock()
	return developeroauthdomain.NewPlaintextSecret("one-time-code"), nil
}

func (stub *stubOAuthPlatform) authorizedRequest() developeroauth.AuthorizationRequest {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return stub.authorized
}

func (stub *stubOAuthPlatform) ExchangeAuthorizationCode(
	_ context.Context,
	request developeroauth.AuthorizationCodeExchange,
) (developeroauthdomain.TokenPair, error) {
	if request.Resource != stub.resource {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	if stub.exchangeCodeErr != nil {
		return developeroauthdomain.TokenPair{}, stub.exchangeCodeErr
	}
	return stub.pair, nil
}

func (stub *stubOAuthPlatform) ExchangeRefreshToken(
	_ context.Context,
	request developeroauth.RefreshExchange,
) (developeroauthdomain.TokenPair, error) {
	if request.Resource != stub.resource {
		return developeroauthdomain.TokenPair{}, developeroauthdomain.ErrInvalidResource
	}
	return stub.pair, nil
}

func (stub *stubOAuthPlatform) ExchangeClientCredentials(
	_ context.Context,
	request developeroauth.ClientCredentialsExchange,
) (developeroauthdomain.ApplicationAccessToken, error) {
	stub.mu.Lock()
	stub.clientCredentials = request
	stub.mu.Unlock()
	if stub.exchangeClientCredentialsErr != nil {
		return developeroauthdomain.ApplicationAccessToken{}, stub.exchangeClientCredentialsErr
	}
	if request.Resource != stub.apiResource {
		return developeroauthdomain.ApplicationAccessToken{}, developeroauthdomain.ErrInvalidResource
	}
	return stub.applicationToken, nil
}

func (stub *stubOAuthPlatform) clientCredentialsRequest() developeroauth.ClientCredentialsExchange {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return stub.clientCredentials
}

func (stub *stubOAuthPlatform) RevokeRefreshToken(context.Context, string) error {
	return nil
}

func (stub *stubOAuthPlatform) VerifyAccessToken(
	_ context.Context,
	raw string,
) (developeroauthdomain.AccessIdentity, error) {
	if raw != stub.pair.AccessToken.Reveal() {
		return developeroauthdomain.AccessIdentity{}, developeroauthdomain.ErrInvalidGrant
	}
	return stub.identity, nil
}

func newStubOAuthPlatform() *stubOAuthPlatform {
	userID := uuid.MustParse("e1e76f7c-2832-43b6-88f7-0af378bde150")
	resource := "https://api.fortyone.app/mcp"
	return &stubOAuthPlatform{
		resource: resource, apiResource: "https://api.fortyone.app/api/v1",
		application: developeroauthdomain.Application{
			ID: uuid.New(), ClientID: "f41_oauth_test-client-identifier", Name: "Test client",
			RedirectURIs: []string{"https://client.example/callback"},
			CreatedAt:    time.Now().Add(-time.Minute), ExpiresAt: time.Now().Add(time.Hour),
		},
		pair: developeroauthdomain.TokenPair{
			AccessToken:  developeroauthdomain.NewPlaintextSecret("valid-access-token"),
			RefreshToken: developeroauthdomain.NewPlaintextSecret("valid-refresh-token"),
			ExpiresIn:    15 * time.Minute,
			Scopes:       []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess},
		},
		applicationToken: developeroauthdomain.ApplicationAccessToken{
			AccessToken: developeroauthdomain.NewPlaintextSecret("valid-application-access-token"),
			ExpiresIn:   15 * time.Minute, Scopes: []string{"stories:write"},
		},
		identity: developeroauthdomain.AccessIdentity{
			PrincipalID: userID, UserID: userID, ApplicationID: uuid.New(), GrantID: uuid.New(),
			OAuthCredential: uuid.New(), ActorKind: platformauth.PrincipalOAuthUser, Resource: resource,
			Scopes:    []string{developeroauthdomain.ScopeMCPAccess, developeroauthdomain.ScopeOfflineAccess},
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
}

type memoryOAuthStore struct {
	mu       sync.Mutex
	items    map[string][]byte
	counters map[string]int64
}

func (s *memoryOAuthStore) Set(_ context.Context, key string, value any, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.items[key] = data
	return nil
}

func (s *memoryOAuthStore) Get(_ context.Context, key string, dest any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.items[key]
	if !ok {
		return errors.New("not found")
	}
	return json.Unmarshal(data, dest)
}

func (s *memoryOAuthStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *memoryOAuthStore) Take(_ context.Context, key string, dest any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.items[key]
	if !ok {
		return errors.New("not found")
	}
	delete(s.items, key)
	return json.Unmarshal(data, dest)
}

func (s *memoryOAuthStore) IncrementWithTTL(_ context.Context, key string, _ time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.counters == nil {
		s.counters = make(map[string]int64)
	}
	s.counters[key]++
	return s.counters[key], nil
}
