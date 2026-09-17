package usershttp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOAuthFailureURLUsesTrustedLoginAndPreservesContinuation(t *testing.T) {
	t.Parallel()
	handler := &Handlers{websiteURL: "https://cloud.fortyone.app", cookieDomain: ".fortyone.app"}
	tests := []struct {
		name         string
		callback     string
		wantHost     string
		continuation string
		mobile       string
	}{
		{name: "no state", wantHost: "cloud.fortyone.app"},
		{name: "untrusted state destination", callback: "https://attacker.example/auth-callback?callbackUrl=/private", wantHost: "cloud.fortyone.app"},
		{name: "malformed state destination", callback: "/%", wantHost: "cloud.fortyone.app"},
		{name: "relative callback", callback: "/auth-callback?callbackUrl=%2Fonboarding%2Fjoin", wantHost: "cloud.fortyone.app", continuation: "/onboarding/join"},
		{name: "invitation", callback: "https://app.fortyone.app/auth-callback?callbackUrl=%2Fonboarding%2Fjoin%3Ftoken%3Dinvite", wantHost: "cloud.fortyone.app", continuation: "/onboarding/join?token=invite"},
		{name: "mobile", callback: "https://cloud.fortyone.app/auth-callback?mobileApp=true&callbackUrl=%2Fauth%2Fmobile%3Fstate%3Dmobile-state%26code_challenge%3Dchallenge", wantHost: "cloud.fortyone.app", continuation: "/auth/mobile?state=mobile-state&code_challenge=challenge", mobile: "true"},
		{name: "direct destination", callback: "https://cloud.fortyone.app/onboarding/join?token=invite", wantHost: "cloud.fortyone.app", continuation: "https://cloud.fortyone.app/onboarding/join?token=invite"},
		{name: "workspace destination", callback: "https://acme.fortyone.app/my-work", wantHost: "cloud.fortyone.app", continuation: "https://acme.fortyone.app/my-work"},
		{name: "workspace root destination", callback: "https://acme.fortyone.app/", wantHost: "cloud.fortyone.app", continuation: "https://acme.fortyone.app/"},
		{name: "API destination", callback: "https://api.fortyone.app/users/profile", wantHost: "cloud.fortyone.app", continuation: "https://api.fortyone.app/users/profile"},
		{name: "API root destination", callback: "https://api.fortyone.app/", wantHost: "cloud.fortyone.app", continuation: "https://api.fortyone.app/"},
		{name: "discard provider fields", callback: "https://cloud.fortyone.app/auth-callback?code=secret&state=secret&error=old-error#secret", wantHost: "cloud.fortyone.app"},
		{name: "discard untrusted nested destination", callback: "https://cloud.fortyone.app/auth-callback?callbackUrl=https%3A%2F%2Fattacker.example", wantHost: "cloud.fortyone.app"},
		{name: "discard duplicate continuation", callback: "https://cloud.fortyone.app/auth-callback?callbackUrl=/first&callbackUrl=/second", wantHost: "cloud.fortyone.app"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destination, err := url.Parse(handler.oauthFailureURL(test.callback, oauthFailureUnavailable))
			require.NoError(t, err)
			require.Equal(t, test.wantHost, destination.Host)
			require.Equal(t, "/", destination.Path)
			require.Empty(t, destination.Fragment)
			wantQuery := url.Values{"error": {oauthFailureUnavailable}}
			if test.continuation != "" {
				wantQuery.Set("callbackUrl", test.continuation)
			}
			if test.mobile != "" {
				wantQuery.Set("mobileApp", test.mobile)
			}
			require.Equal(t, wantQuery, destination.Query())
		})
	}
}

func TestBrowserOAuthFailuresReturnToLoginWithoutCreatingSession(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"google", "microsoft"} {
		t.Run(provider, func(t *testing.T) {
			tests := []struct {
				name         string
				query        string
				storeState   bool
				exchangeErr  error
				blocked      bool
				wantCode     string
				wantExchange int
			}{
				{name: "malformed callback", query: "code=private-code&callbackURL=https://attacker.example", wantCode: oauthFailureGeneric},
				{name: "expired state", query: "state=old-state&code=private-code&callbackURL=https://attacker.example", wantCode: oauthFailureExpired},
				{name: "duplicate state", query: "state=state&state=other&code=private-code", wantCode: oauthFailureGeneric},
				{name: "cancelled", query: "state=state&error=access_denied&error_description=private-provider-data", storeState: true, wantCode: oauthFailureCancelled},
				{name: "provider error", query: "state=state&error=private-provider-data", storeState: true, wantCode: oauthFailureGeneric},
				{name: "failed exchange", query: "state=state&code=private-code", storeState: true, exchangeErr: errors.New("private-provider-data"), wantCode: oauthFailureGeneric, wantExchange: 1},
				{name: "blocked account", query: "state=state&code=private-code", storeState: true, blocked: true, wantCode: oauthFailureUnavailable, wantExchange: 1},
			}
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					handler, redisServer, repo, exchange := newOAuthCallbackTestHandler(t)
					exchange.err = test.exchangeErr
					if test.blocked {
						repo.reactivationErr = users.ErrInvalidCredentials
					}
					if test.storeState {
						storeOAuthCallbackTestState(t, handler, provider, "https://app.fortyone.app/auth-callback?callbackUrl=%2Fonboarding%2Fjoin")
					}
					request := httptest.NewRequest(http.MethodGet, "/auth/"+provider+"/callback?"+test.query, nil)
					request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "stale-session"})
					recorder := httptest.NewRecorder()
					require.NoError(t, completeOAuthCallback(t, handler, provider, recorder, request))
					require.Equal(t, http.StatusSeeOther, recorder.Code)
					destination, err := url.Parse(recorder.Header().Get("Location"))
					require.NoError(t, err)
					require.Equal(t, test.wantCode, destination.Query().Get("error"))
					require.Equal(t, "/", destination.Path)
					require.Equal(t, "cloud.fortyone.app", destination.Host)
					if test.storeState {
						require.Equal(t, "/onboarding/join", destination.Query().Get("callbackUrl"))
					} else {
						require.Empty(t, destination.Query().Get("callbackUrl"))
					}
					require.Equal(t, test.wantExchange, exchange.calls)
					require.Equal(t, test.blocked, repo.reactivated)
					require.Zero(t, repo.sessionChecks)
					require.Empty(t, recorder.Result().Cookies())
					require.Empty(t, redisServer.Keys(), "state should be consumed and no session created")
					require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
					require.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
					for _, secret := range []string{"private-code", "private-provider-data", "stale-session"} {
						require.NotContains(t, recorder.Body.String()+recorder.Header().Get("Location"), secret)
					}
				})
			}
		})
	}
}

func TestBrowserOAuthReturningAccountGetsFreshSessionAndPreservesCallback(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"google", "microsoft"} {
		t.Run(provider, func(t *testing.T) {
			handler, _, repo, _ := newOAuthCallbackTestHandler(t)
			callback := "https://cloud.fortyone.app/auth-callback?callbackUrl=%2Fonboarding%2Fjoin"
			storeOAuthCallbackTestState(t, handler, provider, callback)
			request := httptest.NewRequest(http.MethodGet, "/auth/"+provider+"/callback?state=state&code=code", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "stale-session"})
			recorder := httptest.NewRecorder()
			require.NoError(t, completeOAuthCallback(t, handler, provider, recorder, request))
			require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
			require.Equal(t, callback, recorder.Header().Get("Location"))
			require.True(t, repo.reactivated)
			require.Equal(t, 1, repo.sessionChecks)
			cookies := recorder.Result().Cookies()
			require.Len(t, cookies, 1)
			require.Equal(t, sessionCookieName, cookies[0].Name)
			require.NotEmpty(t, cookies[0].Value)
			require.NotEqual(t, "stale-session", cookies[0].Value)
			var session platformauth.BrowserSession
			require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey(cookies[0].Value), &session))
			require.Equal(t, repo.user.ID, session.UserID)
			require.Equal(t, int64(7), session.Version)
			require.ErrorIs(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey("stale-session"), &session), cache.ErrNotFound)
		})
	}
}

func TestGoogleTokenVerificationKeepsJSONForBlockedAccount(t *testing.T) {
	t.Parallel()
	handler, _, repo, _ := newOAuthCallbackTestHandler(t)
	repo.reactivationErr = users.ErrInvalidCredentials
	request := httptest.NewRequest(http.MethodPost, "/auth/google", strings.NewReader(`{"token":"id-token"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.GoogleAuth(t.Context(), recorder, request))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Empty(t, recorder.Header().Get("Location"))
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
	require.Contains(t, recorder.Body.String(), users.ErrInvalidCredentials.Error())
	require.Empty(t, recorder.Result().Cookies())
}

func TestBrowserOAuthStartFailuresReturnToLogin(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"google", "microsoft"} {
		t.Run(provider, func(t *testing.T) {
			for _, failure := range []string{"invalid callback", "missing cache", "missing provider", "provider URL"} {
				t.Run(failure, func(t *testing.T) {
					handler, redisServer, _, exchange := newOAuthCallbackTestHandler(t)
					callback := "https://app.fortyone.app/auth-callback?callbackUrl=%2Fonboarding%2Fjoin"
					switch failure {
					case "invalid callback":
						callback = "https://attacker.example/auth-callback"
					case "missing cache":
						handler.cache = nil
					case "missing provider":
						handler.googleService = nil
						handler.microsoftService = nil
					case "provider URL":
						exchange.err = errors.New("private-provider-configuration")
					}
					request := httptest.NewRequest(http.MethodGet, "/auth/"+provider+"?callbackURL="+url.QueryEscape(callback), nil)
					recorder := httptest.NewRecorder()
					if provider == "google" {
						require.NoError(t, handler.StartGoogleAuth(t.Context(), recorder, request))
					} else {
						require.NoError(t, handler.StartMicrosoftAuth(t.Context(), recorder, request))
					}
					require.Equal(t, http.StatusSeeOther, recorder.Code)
					destination, err := url.Parse(recorder.Header().Get("Location"))
					require.NoError(t, err)
					require.Equal(t, oauthFailureGeneric, destination.Query().Get("error"))
					require.Equal(t, "cloud.fortyone.app", destination.Host)
					if failure == "invalid callback" {
						require.Empty(t, destination.Query().Get("callbackUrl"))
					} else {
						require.Equal(t, "/onboarding/join", destination.Query().Get("callbackUrl"))
					}
					require.Empty(t, redisServer.Keys(), "failed OAuth starts must remove pending state")
					require.Empty(t, recorder.Result().Cookies())
					require.NotContains(t, recorder.Header().Get("Location"), "private-provider-configuration")
				})
			}
		})
	}
}

type oauthCallbackTestRepository struct {
	users.Repository
	user            users.CoreUser
	reactivationErr error
	reactivated     bool
	sessionChecks   int
}

func (repo *oauthCallbackTestRepository) GetUserByEmailAnyStatus(context.Context, string) (users.CoreUser, error) {
	return repo.user, nil
}

func (repo *oauthCallbackTestRepository) ResolveExternalIdentity(context.Context, users.CoreExternalIdentityInput) (users.CoreExternalIdentityResult, error) {
	return users.CoreExternalIdentityResult{User: repo.user}, nil
}

func (repo *oauthCallbackTestRepository) ReactivateUserForVerifiedSignIn(_ context.Context, input users.VerifiedSignInReactivation) (users.CoreUser, error) {
	repo.reactivated = true
	if repo.reactivationErr != nil {
		return users.CoreUser{}, repo.reactivationErr
	}
	repo.user.IsActive = true
	repo.user.LastLoginAt = input.SignedInAt
	return repo.user, nil
}

func (repo *oauthCallbackTestRepository) ResolveActiveBrowserSessionVersion(context.Context, uuid.UUID) (int64, bool, error) {
	repo.sessionChecks++
	return 7, repo.user.IsActive, nil
}

type oauthCallbackTestExchange struct {
	err   error
	calls int
}

type googleCallbackTestProvider struct{ exchange *oauthCallbackTestExchange }

func (provider googleCallbackTestProvider) VerifyToken(context.Context, string) (google.Identity, error) {
	return google.Identity{Email: "returning@example.com", EmailVerified: true}, provider.exchange.err
}

func (provider googleCallbackTestProvider) AuthCodeURL(string) (string, error) {
	return "https://accounts.google.com/auth", provider.exchange.err
}

func (provider googleCallbackTestProvider) ExchangeCode(ctx context.Context, code string) (google.Identity, error) {
	provider.exchange.calls++
	return provider.VerifyToken(ctx, code)
}

type microsoftCallbackTestProvider struct{ exchange *oauthCallbackTestExchange }

func (provider microsoftCallbackTestProvider) AuthCodeURL(string, string, string) (string, error) {
	return "https://login.microsoftonline.com/auth", provider.exchange.err
}

func (provider microsoftCallbackTestProvider) ExchangeCode(context.Context, string, string, string) (microsoft.Identity, error) {
	provider.exchange.calls++
	return microsoft.Identity{Email: "returning@example.com", Issuer: "issuer", ObjectID: "subject"}, provider.exchange.err
}

func newOAuthCallbackTestHandler(t *testing.T) (*Handlers, *miniredis.Miniredis, *oauthCallbackTestRepository, *oauthCallbackTestExchange) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	log := logger.NewWithText(io.Discard, slog.LevelError, "oauth-callback-test")
	repo := &oauthCallbackTestRepository{user: users.CoreUser{ID: uuid.New(), Email: "returning@example.com", IsActive: false}}
	exchange := &oauthCallbackTestExchange{}
	return &Handlers{
		users:            users.New(log, repo, nil),
		cache:            cache.New(client, log),
		log:              log,
		websiteURL:       "https://cloud.fortyone.app",
		cookieDomain:     ".fortyone.app",
		googleService:    googleCallbackTestProvider{exchange: exchange},
		microsoftService: microsoftCallbackTestProvider{exchange: exchange},
	}, server, repo, exchange
}

func storeOAuthCallbackTestState(t *testing.T, handler *Handlers, provider, callbackURL string) {
	t.Helper()
	if provider == "google" {
		require.NoError(t, handler.cache.Set(t.Context(), cache.AuthGoogleStateCacheKey("state"), googleAuthState{CallbackURL: callbackURL}, oauthStateTTL))
		return
	}
	require.NoError(t, handler.cache.Set(t.Context(), cache.AuthMicrosoftStateCacheKey("state"), microsoftAuthState{CallbackURL: callbackURL, Nonce: "nonce", Verifier: "verifier"}, oauthStateTTL))
}

func completeOAuthCallback(t *testing.T, handler *Handlers, provider string, recorder *httptest.ResponseRecorder, request *http.Request) error {
	t.Helper()
	if provider == "google" {
		return handler.CompleteGoogleAuth(t.Context(), recorder, request)
	}
	return handler.CompleteMicrosoftAuth(t.Context(), recorder, request)
}
