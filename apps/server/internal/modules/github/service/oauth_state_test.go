package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type githubOAuthStateStoreStub struct {
	mu      sync.Mutex
	values  map[string][]byte
	TTLs    map[string]time.Duration
	setKeys []string
}

func newGitHubOAuthStateStoreStub() *githubOAuthStateStoreStub {
	return &githubOAuthStateStoreStub{
		values: make(map[string][]byte),
		TTLs:   make(map[string]time.Duration),
	}
}

func (s *githubOAuthStateStoreStub) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = encoded
	s.TTLs[key] = ttl
	s.setKeys = append(s.setKeys, key)
	return nil
}

func (s *githubOAuthStateStoreStub) Take(_ context.Context, key string, destination any) error {
	s.mu.Lock()
	encoded, ok := s.values[key]
	if ok {
		delete(s.values, key)
	}
	s.mu.Unlock()
	if !ok {
		return cache.ErrNotFound
	}
	return json.Unmarshal(encoded, destination)
}

func (s *githubOAuthStateStoreStub) snapshot(key string, destination any) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.values[key]
	if !ok {
		return false
	}
	return json.Unmarshal(value, destination) == nil
}

func newGitHubOAuthStateService(store OAuthStateStore, now time.Time) *Service {
	return &Service{
		oauthStates: store,
		now:         func() time.Time { return now },
		cfg: Config{
			AppID:       1,
			AppSlug:     "fortyone-test",
			RedirectURL: "https://api.example.test/integrations/github/setup",
			WebsiteURL:  "https://fortyone.example.test",
		},
		privateKey:     &rsa.PrivateKey{},
		workspaceRoles: &githubWorkspaceRoleStub{role: authorization.WorkspaceRoleAdmin},
	}
}

func TestGitHubInstallStateIsOpaqueShortLivedAndIdentityBound(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	store := newGitHubOAuthStateStoreStub()
	service := newGitHubOAuthStateService(store, now)
	workspaceID, userID := uuid.New(), uuid.New()

	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	rawState := installURL.Query().Get("state")
	token, err := oauthstate.Parse(rawState)
	require.NoError(t, err)
	require.Len(t, rawState, 43)
	require.NotContains(t, rawState, ".")
	require.NotContains(t, rawState, "returnTo")
	require.NotContains(t, rawState, "workspace")
	require.NotContains(t, rawState, workspaceID.String())
	require.NotContains(t, rawState, userID.String())

	key := githubOAuthStateCacheKey(githubInstallStatePurpose, token.Digest())
	require.NotContains(t, key, rawState)
	var record githubOAuthStateRecord
	require.True(t, store.snapshot(key, &record))
	require.Equal(t, githubOAuthStateProvider, record.Provider)
	require.Equal(t, githubInstallStatePurpose, record.Purpose)
	require.Equal(t, workspaceID, *record.WorkspaceID)
	require.Equal(t, userID, record.UserID)
	require.Equal(t, now.Add(githubInstallStateTTL), record.ExpiresAt)
	require.Equal(t, githubInstallStateTTL, store.TTLs[key])
}

func TestGitHubUserLinkReturnURLIsBoundToConfiguredOriginFamily(t *testing.T) {
	t.Parallel()
	service := newGitHubOAuthStateService(newGitHubOAuthStateStoreStub(), time.Now())
	service.cfg.WebsiteURL = "https://app.example.co.uk"

	for _, allowed := range []string{
		"/settings/account/profile?tab=integrations",
		"https://app.example.co.uk/settings",
		"https://acme.app.example.co.uk/settings",
	} {
		_, err := service.safeUserLinkReturnTo(allowed)
		require.NoError(t, err, allowed)
	}
	for _, rejected := range []string{
		"https://evil.co.uk/settings",
		"https://example.co.uk/settings",
		"http://app.example.co.uk/settings",
		"https://user@app.example.co.uk/settings",
		"https://localhost/settings",
	} {
		_, err := service.safeUserLinkReturnTo(rejected)
		require.Error(t, err, rejected)
	}
}

func TestGitHubUserLinkStateRejectsMismatchReplayExpiryAndUnsafeReturn(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	store := newGitHubOAuthStateStoreStub()
	service := newGitHubOAuthStateService(store, now)
	userID := uuid.New()
	ctx := context.Background()

	state, err := service.createUserLinkState(ctx, userID, "/settings/account/profile?tab=integrations")
	require.NoError(t, err)
	_, err = service.consumeUserLinkState(ctx, state, uuid.New())
	require.ErrorIs(t, err, ErrGitHubOAuthStateBinding)
	_, err = service.consumeUserLinkState(ctx, state, userID)
	require.ErrorIs(t, err, ErrGitHubOAuthStateInvalid)

	expired, err := service.createUserLinkState(ctx, userID, "/settings/account/profile")
	require.NoError(t, err)
	service.now = func() time.Time { return now.Add(userLinkStateTTL) }
	_, err = service.consumeUserLinkState(ctx, expired, userID)
	require.ErrorIs(t, err, ErrGitHubOAuthStateInvalid)

	_, err = service.createUserLinkState(ctx, userID, "https://evil.example/callback")
	require.Error(t, err)
}

func TestGitHubOAuthStateIsPurposeBound(t *testing.T) {
	t.Parallel()

	store := newGitHubOAuthStateStoreStub()
	service := newGitHubOAuthStateService(store, time.Now())
	workspaceID, userID := uuid.New(), uuid.New()
	ctx := context.Background()
	state, err := service.createInstallState(ctx, workspaceID, userID, "acme")
	require.NoError(t, err)

	_, err = service.consumeUserLinkState(ctx, state, userID)
	require.ErrorIs(t, err, ErrGitHubOAuthStateInvalid)

	record, payload, err := service.consumeInstallState(ctx, state)
	require.NoError(t, err)
	require.Equal(t, workspaceID, *record.WorkspaceID)
	require.Equal(t, "acme", payload.WorkspaceSlug)
}

func TestGitHubOAuthStateHasExactlyOneConcurrentConsumer(t *testing.T) {
	t.Parallel()

	store := newGitHubOAuthStateStoreStub()
	service := newGitHubOAuthStateService(store, time.Now())
	userID := uuid.New()
	state, err := service.createUserLinkState(context.Background(), userID, "/settings/account/profile")
	require.NoError(t, err)

	const consumers = 32
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(consumers)
	for range consumers {
		go func() {
			defer wait.Done()
			if _, consumeErr := service.consumeUserLinkState(context.Background(), state, userID); consumeErr == nil {
				successes.Add(1)
			} else if !errors.Is(consumeErr, ErrGitHubOAuthStateInvalid) {
				t.Errorf("consumeUserLinkState() error = %v", consumeErr)
			}
		}()
	}
	wait.Wait()
	require.EqualValues(t, 1, successes.Load())
}

func TestGitHubOAuthStateErrorsNeverEchoRawState(t *testing.T) {
	t.Parallel()

	service := newGitHubOAuthStateService(newGitHubOAuthStateStoreStub(), time.Now())
	rawState := strings.Repeat("sensitive-state", 4)
	_, err := service.consumeUserLinkState(context.Background(), rawState, uuid.New())
	require.Error(t, err)
	require.NotContains(t, err.Error(), rawState)
}

func FuzzGitHubOAuthStateCanonicalEncoding(f *testing.F) {
	f.Add(strings.Repeat("A", 43))
	f.Add("legacy.payload.signature")
	f.Add("state=with-padding==")
	f.Fuzz(func(t *testing.T, raw string) {
		token, err := oauthstate.Parse(raw)
		if err != nil {
			return
		}
		require.Len(t, raw, 43)
		require.NotContains(t, raw, ".")
		require.NotContains(t, raw, "=")
		require.Equal(t, raw, token.String())
	})
}
