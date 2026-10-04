package usershttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestPersistSessionStoresStructuredAccountEpoch(t *testing.T) {
	t.Parallel()
	handler, userID := sessionPersistenceTestHandler(t)
	expiresAt := time.Now().Add(time.Hour)
	request := httptest.NewRequest(http.MethodPost, "/users/verify", nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 Chrome/140.0.0.0 Safari/537.36")
	request.Header.Set("Sec-CH-UA", `"Opera";v="123", "Chromium";v="140"`)

	require.NoError(t, handler.persistSession(t.Context(), request, userID, "opaque-token", expiresAt))
	var stored platformauth.BrowserSession
	require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey("opaque-token"), &stored))
	require.Equal(t, userID, stored.UserID)
	require.Equal(t, int64(12), stored.Version)
	require.NotEqual(t, uuid.Nil, stored.SessionID)
	require.False(t, stored.AuthenticatedAt.IsZero())
	require.Equal(t, expiresAt.UTC(), stored.ExpiresAt)
	require.NotNil(t, stored.BrowserName)
	require.Equal(t, "Opera", *stored.BrowserName)
}

func sessionPersistenceTestHandler(t *testing.T) (*Handlers, uuid.UUID) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "session-test")
	cacheService := cache.New(client, log)
	userID := uuid.New()
	userService := users.New(log, sessionVersionRepository{
		userID: userID, version: 12,
	}, nil)
	handler := &Handlers{users: userService, cache: cacheService}
	return handler, userID
}

func TestSessionRenewalPreservesBrowserAndHistoricalUnknown(t *testing.T) {
	t.Parallel()
	safari := "Safari"
	for _, test := range []struct {
		name    string
		browser *string
	}{{"historical unknown", nil}, {"recorded browser", &safari}} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			handler, userID := sessionPersistenceTestHandler(t)
			previous, err := platformauth.NewBrowserSession(userID, 12)
			require.NoError(t, err)
			previous.BrowserName = test.browser
			ctx := platformauth.SetBrowserSession(t.Context(), previous)
			request := httptest.NewRequest(http.MethodPost, "/users/session", nil)
			request.Header.Set("User-Agent", "Mozilla/5.0 Chrome/140.0.0.0 Safari/537.36")
			require.NoError(t, handler.persistSession(ctx, request, userID, "renewed-token", time.Now().Add(time.Hour)))
			var stored platformauth.BrowserSession
			require.NoError(t, handler.cache.Get(ctx, cache.AuthSessionCacheKey("renewed-token"), &stored))
			require.Equal(t, previous.SessionID, stored.SessionID)
			require.Equal(t, previous.AuthenticatedAt, stored.AuthenticatedAt)
			require.Equal(t, test.browser, stored.BrowserName)
		})
	}
}

func TestWorkspaceSSOSessionCapturesReportedBrowser(t *testing.T) {
	t.Parallel()
	handler, userID := sessionPersistenceTestHandler(t)
	writer := &WorkspaceSSOSessionWriter{handlers: handler}
	request := httptest.NewRequest(http.MethodGet, "/workspaces/security/sso/callback", nil)
	request.Header.Set("Sec-CH-UA", `"Dia";v="1", "Chromium";v="140"`)
	proof := platformauth.WorkspaceSSOAuthentication{WorkspaceID: uuid.New(), ConnectionID: uuid.New(), Generation: 1, AuthenticatedAt: time.Now().UTC()}
	response := httptest.NewRecorder()
	require.NoError(t, writer.Issue(t.Context(), response, request, userID, proof))
	cookie := requireSingleSessionCookie(t, response)
	var stored platformauth.BrowserSession
	require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey(cookie.Value), &stored))
	require.Equal(t, &proof, stored.WorkspaceSSO)
	require.Equal(t, proof.AuthenticatedAt, stored.AuthenticatedAt)
	require.NotNil(t, stored.BrowserName)
	require.Equal(t, "Dia", *stored.BrowserName)
}

type sessionVersionRepository struct {
	users.Repository
	userID  uuid.UUID
	version int64
}

func (repository sessionVersionRepository) ResolveActiveBrowserSessionVersion(
	_ context.Context,
	userID uuid.UUID,
) (int64, bool, error) {
	if userID != repository.userID {
		return 0, false, nil
	}
	return repository.version, true, nil
}

func TestSessionCookieSecurityMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mode           deployment.Mode
		tls            bool
		forwardedProto string
		wantSecure     bool
	}{
		{
			name:           "production forces secure on plain upstream HTTP",
			mode:           deployment.Production,
			forwardedProto: "http",
			wantSecure:     true,
		},
		{
			name:       "development permits local HTTP",
			mode:       deployment.Development,
			wantSecure: false,
		},
		{
			name:       "development secures direct TLS",
			mode:       deployment.Development,
			tls:        true,
			wantSecure: true,
		},
		{
			name:           "development preserves HTTPS reverse proxy support",
			mode:           deployment.Development,
			forwardedProto: "https",
			wantSecure:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handlers{deploymentMode: test.mode}
			request := httptest.NewRequest(http.MethodPost, "http://localhost/users/session", nil)
			if test.tls {
				request.TLS = &tls.ConnectionState{}
			}
			if test.forwardedProto != "" {
				request.Header.Set("X-Forwarded-Proto", test.forwardedProto)
			}

			setRecorder := httptest.NewRecorder()
			handler.setSessionCookie(setRecorder, request, "opaque-session", time.Now().Add(time.Hour))
			setCookie := requireSingleSessionCookie(t, setRecorder)
			require.Equal(t, test.wantSecure, setCookie.Secure)
			require.True(t, setCookie.HttpOnly)
			require.Equal(t, http.SameSiteLaxMode, setCookie.SameSite)
			require.Equal(t, "/", setCookie.Path)
			require.Positive(t, setCookie.MaxAge)

			clearRecorder := httptest.NewRecorder()
			handler.clearSessionCookie(clearRecorder, request)
			clearCookie := requireSingleSessionCookie(t, clearRecorder)
			require.Equal(t, test.wantSecure, clearCookie.Secure)
			require.True(t, clearCookie.HttpOnly)
			require.Equal(t, http.SameSiteLaxMode, clearCookie.SameSite)
			require.Equal(t, "/", clearCookie.Path)
			require.Equal(t, -1, clearCookie.MaxAge)
		})
	}
}

func requireSingleSessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, sessionCookieName, cookies[0].Name)
	return cookies[0]
}
