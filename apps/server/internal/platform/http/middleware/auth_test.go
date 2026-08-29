package mid

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestResolveTokenFromRequestOnlyAcceptsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		url    string
		want   string
	}{
		{name: "bearer", header: "Bearer signed-token", url: "/resource", want: "signed-token"},
		{name: "case-insensitive scheme", header: "bearer signed-token", url: "/resource", want: "signed-token"},
		{name: "trimmed bearer", header: "  Bearer signed-token  ", url: "/resource", want: "signed-token"},
		{name: "query bearer rejected", url: "/resource?token=leaked-token"},
		{name: "wrong scheme", header: "Basic credentials", url: "/resource"},
		{name: "empty bearer", header: "Bearer ", url: "/resource"},
		{name: "extra bearer component", header: "Bearer signed-token unexpected", url: "/resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			require.Equal(t, tt.want, resolveTokenFromRequest(request))
		})
	}
}

func TestBrowserSessionResolverRequiresActiveMatchingAccountVersion(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name       string
		version    int64
		active     bool
		accountErr error
		wantOK     bool
		wantErr    error
	}{
		{name: "active matching", version: 7, active: true, wantOK: true},
		{name: "inactive", version: 7, active: false, wantErr: platformauth.ErrInvalidBrowserSession},
		{name: "version mismatch", version: 8, active: true, wantErr: platformauth.ErrInvalidBrowserSession},
		{name: "database failure", version: 7, active: true, accountErr: databaseErr, wantErr: databaseErr},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			resolver := newTestBrowserSessionResolver(t, "opaque", platformauth.BrowserSession{
				UserID: userID, Version: 7,
			}, func(_ context.Context, gotUserID uuid.UUID) (int64, bool, error) {
				require.Equal(t, userID, gotUserID)
				return test.version, test.active, test.accountErr
			})
			request := requestWithSession("opaque")

			gotUserID, ok, err := ResolveSessionUserID(context.Background(), request, resolver)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				require.False(t, ok)
				require.Equal(t, uuid.Nil, gotUserID)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.wantOK, ok)
			require.Equal(t, userID, gotUserID)
		})
	}
}

func TestAuthAcceptsOnlyValidatedBrowserSession(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	resolver := newTestBrowserSessionResolver(t, "opaque", platformauth.BrowserSession{
		UserID: userID, Version: 3,
	}, func(context.Context, uuid.UUID) (int64, bool, error) {
		return 3, true, nil
	})
	called := false
	handler := Auth(nil, "ignored-legacy-secret", resolver)(func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		called = true
		resolvedUserID, err := GetUserID(ctx)
		require.NoError(t, err)
		require.Equal(t, userID, resolvedUserID)
		writer.WriteHeader(http.StatusNoContent)
		return nil
	})
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(context.Background(), recorder, requestWithSession("opaque")))
	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.True(t, called)
}

func TestOptionalAuthDistinguishesAnonymousFromInvalidCredentials(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	valid := newTestBrowserSessionResolver(t, "valid", platformauth.BrowserSession{
		UserID: userID, Version: 2,
	}, func(context.Context, uuid.UUID) (int64, bool, error) {
		return 2, true, nil
	})
	invalid := newTestBrowserSessionResolver(t, "invalid", platformauth.BrowserSession{
		UserID: userID, Version: 1,
	}, func(context.Context, uuid.UUID) (int64, bool, error) {
		return 2, true, nil
	})

	tests := []struct {
		name       string
		resolver   SessionResolver
		request    *http.Request
		wantStatus int
		wantUser   bool
		wantNext   bool
	}{
		{name: "anonymous", resolver: valid, request: httptest.NewRequest(http.MethodGet, "/", nil), wantStatus: http.StatusNoContent, wantNext: true},
		{name: "valid session", resolver: valid, request: requestWithSession("valid"), wantStatus: http.StatusNoContent, wantUser: true, wantNext: true},
		{name: "stale session", resolver: invalid, request: requestWithSession("invalid"), wantStatus: http.StatusUnauthorized},
		{name: "legacy bearer", resolver: valid, request: requestWithBearer("legacy-user-jwt"), wantStatus: http.StatusUnauthorized},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			called := false
			handler := OptionalAuth(nil, "ignored", test.resolver)(func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
				called = true
				_, err := GetUserID(ctx)
				require.Equal(t, test.wantUser, err == nil)
				writer.WriteHeader(http.StatusNoContent)
				return nil
			})
			recorder := httptest.NewRecorder()
			require.NoError(t, handler(context.Background(), recorder, test.request))
			require.Equal(t, test.wantStatus, recorder.Code)
			require.Equal(t, test.wantNext, called)
		})
	}
}

func TestRevokedSessionDoesNotResurrectAfterReactivation(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	active := false
	resolver := newTestBrowserSessionResolver(t, "old", platformauth.BrowserSession{
		UserID: userID, Version: 4,
	}, func(context.Context, uuid.UUID) (int64, bool, error) {
		return 5, active, nil
	})
	request := requestWithSession("old")

	_, ok, err := ResolveSessionUserID(context.Background(), request, resolver)
	require.ErrorIs(t, err, platformauth.ErrInvalidBrowserSession)
	require.False(t, ok)

	active = true
	_, ok, err = ResolveSessionUserID(context.Background(), request, resolver)
	require.ErrorIs(t, err, platformauth.ErrInvalidBrowserSession)
	require.False(t, ok)
}

func TestLegacyBrowserSessionRecordsFailClosed(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := cache.New(client, logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "auth-test"))
	accounts := sessionAccountResolverFunc(func(context.Context, uuid.UUID) (int64, bool, error) {
		return 1, true, nil
	})
	ctx := context.Background()
	userID := uuid.New()

	require.NoError(t, service.Set(ctx, cache.AuthSessionCacheKey("string-record"), userID.String(), time.Minute))
	require.NoError(t, service.Set(ctx, cache.LegacyAuthSessionCacheKey("legacy-key"), platformauth.BrowserSession{UserID: userID, Version: 1}, time.Minute))

	resolver := NewBrowserSessionResolver(service, accounts)
	_, ok, err := ResolveSessionUserID(ctx, requestWithSession("string-record"), resolver)
	require.Error(t, err)
	require.False(t, ok)

	_, ok, err = ResolveSessionUserID(ctx, requestWithSession("legacy-key"), resolver)
	require.NoError(t, err)
	require.False(t, ok)
}

type sessionStoreFunc func(context.Context, string, any) error

func (store sessionStoreFunc) Get(ctx context.Context, key string, destination any) error {
	return store(ctx, key, destination)
}

type sessionAccountResolverFunc func(context.Context, uuid.UUID) (int64, bool, error)

func (resolver sessionAccountResolverFunc) ResolveActiveBrowserSessionVersion(ctx context.Context, userID uuid.UUID) (int64, bool, error) {
	return resolver(ctx, userID)
}

func newTestBrowserSessionResolver(
	t *testing.T,
	token string,
	session platformauth.BrowserSession,
	accounts sessionAccountResolverFunc,
) *BrowserSessionResolver {
	t.Helper()
	store := sessionStoreFunc(func(_ context.Context, key string, destination any) error {
		if key != cache.AuthSessionCacheKey(token) {
			return cache.ErrNotFound
		}
		value, ok := destination.(*platformauth.BrowserSession)
		require.True(t, ok, "session destination type = %T", destination)
		*value = session
		return nil
	})
	return NewBrowserSessionResolver(store, accounts)
}

func requestWithSession(token string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	request.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return request
}

func requestWithBearer(token string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func TestRequireScopesUsesActorCredentialScopes(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalServiceAccount,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)

	tests := []struct {
		name       string
		required   platformauth.Scope
		wantStatus int
		wantNext   bool
	}{
		{name: "allowed", required: platformauth.ScopeStoriesRead, wantStatus: http.StatusNoContent, wantNext: true},
		{name: "denied", required: platformauth.ScopeStoriesWrite, wantStatus: http.StatusForbidden},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			called := false
			handler := RequireScopes(test.required)(func(_ context.Context, writer http.ResponseWriter, _ *http.Request) error {
				called = true
				writer.WriteHeader(http.StatusNoContent)
				return nil
			})
			ctx, setErr := platformauth.SetActor(context.Background(), actor)
			require.NoError(t, setErr)
			recorder := httptest.NewRecorder()
			require.NoError(t, handler(ctx, recorder, httptest.NewRequest(http.MethodGet, "/", nil)))
			require.Equal(t, test.wantStatus, recorder.Code)
			require.Equal(t, test.wantNext, called)
		})
	}
}
