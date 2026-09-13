package usershttp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type mobileSessionRepository struct {
	users.Repository
	userID  uuid.UUID
	version int64
}

func (r *mobileSessionRepository) ResolveActiveBrowserSessionVersion(_ context.Context, id uuid.UUID) (int64, bool, error) {
	return r.version, id == r.userID, nil
}

func (r *mobileSessionRepository) GetUser(_ context.Context, id uuid.UUID) (users.CoreUser, error) {
	return users.CoreUser{ID: id, Email: "mobile@example.com", IsActive: true}, nil
}

func mobileSessionTestHandler(t *testing.T) (*Handlers, *miniredis.Miniredis, *mobileSessionRepository) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	log := logger.NewWithText(io.Discard, slog.LevelError, "mobile-session-test")
	repository := &mobileSessionRepository{userID: uuid.New(), version: 7}
	handler := &Handlers{
		users: users.New(log, repository, nil), cache: cache.New(client, log), log: log,
	}
	require.NoError(t, handler.persistSession(t.Context(), repository.userID, "browser-cookie", time.Now().Add(time.Hour)))
	return handler, server, repository
}

func mobileRequest(t *testing.T, path string, input any) *http.Request {
	t.Helper()
	body, err := json.Marshal(input)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func authorizeMobileForTest(t *testing.T, handler *Handlers, userID uuid.UUID) mobileExchangeRequest {
	t.Helper()
	verifier := strings.Repeat("v", 43)
	digest := sha256.Sum256([]byte(verifier))
	input := mobileAuthorizationRequest{
		State: strings.Repeat("s", 43), CodeChallenge: base64.RawURLEncoding.EncodeToString(digest[:]),
		CodeChallengeMethod: "S256", RedirectURI: mobileRedirectURI,
	}
	request := mobileRequest(t, "/auth/mobile/authorize", input)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.AuthorizeMobile(platformauth.SetUserID(t.Context(), userID), recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var response struct{ Data struct{ Code, State string } }
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data.Code, 43)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	return mobileExchangeRequest{Code: response.Data.Code, State: response.Data.State, CodeVerifier: verifier, RedirectURI: mobileRedirectURI}
}

func exchangeMobileForTest(t *testing.T, handler *Handlers, input mobileExchangeRequest) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.ExchangeMobile(t.Context(), recorder, mobileRequest(t, "/auth/mobile/exchange", input)))
	return recorder
}

func TestMobileHandoffIssuesIndependentRevocableSession(t *testing.T) {
	handler, server, repository := mobileSessionTestHandler(t)
	input := authorizeMobileForTest(t, handler, repository.userID)
	for _, key := range server.Keys() {
		require.NotContains(t, key, input.Code)
	}
	response := exchangeMobileForTest(t, handler, input)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	require.True(t, cookies[0].HttpOnly)
	require.NotEqual(t, "browser-cookie", cookies[0].Value)
	require.NotContains(t, response.Body.String(), cookies[0].Value)
	var session platformauth.BrowserSession
	require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey(cookies[0].Value), &session))
	require.Equal(t, repository.version, session.Version)
	require.Equal(t, repository.userID, session.UserID)
	require.Equal(t, http.StatusBadRequest, exchangeMobileForTest(t, handler, input).Code)
}

func TestMobileHandoffRejectsWrongBindingWithoutBurningValidCode(t *testing.T) {
	for _, field := range []string{"verifier", "state", "redirect"} {
		t.Run(field, func(t *testing.T) {
			handler, _, repository := mobileSessionTestHandler(t)
			input := authorizeMobileForTest(t, handler, repository.userID)
			invalid := input
			switch field {
			case "verifier":
				invalid.CodeVerifier = strings.Repeat("x", 43)
			case "state":
				invalid.State = strings.Repeat("x", 43)
			case "redirect":
				invalid.RedirectURI = "attacker://login"
			}
			rejected := exchangeMobileForTest(t, handler, invalid)
			require.Equal(t, http.StatusBadRequest, rejected.Code)
			require.Empty(t, rejected.Result().Cookies())
			require.Equal(t, http.StatusOK, exchangeMobileForTest(t, handler, input).Code)
		})
	}
}

func TestMobileHandoffExpiresAndPreservesRevocationEpoch(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		handler, server, repository := mobileSessionTestHandler(t)
		input := authorizeMobileForTest(t, handler, repository.userID)
		server.FastForward(mobileCodeTTL + time.Second)
		require.Equal(t, http.StatusBadRequest, exchangeMobileForTest(t, handler, input).Code)
	})
	t.Run("revoked after authorization", func(t *testing.T) {
		handler, _, repository := mobileSessionTestHandler(t)
		input := authorizeMobileForTest(t, handler, repository.userID)
		repository.version++
		response := exchangeMobileForTest(t, handler, input)
		require.Equal(t, http.StatusUnauthorized, response.Code)
		require.Empty(t, response.Result().Cookies())
	})
	t.Run("revoked between middleware and issuance", func(t *testing.T) {
		handler, _, repository := mobileSessionTestHandler(t)
		repository.version++
		request := mobileRequest(t, "/auth/mobile/authorize", mobileAuthorizationRequest{
			State: strings.Repeat("s", 43), CodeChallenge: strings.Repeat("c", 43),
			CodeChallengeMethod: "S256", RedirectURI: mobileRedirectURI,
		})
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
		recorder := httptest.NewRecorder()
		require.NoError(t, handler.AuthorizeMobile(platformauth.SetUserID(t.Context(), repository.userID), recorder, request))
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestMobileHandoffConcurrentExchangeSucceedsOnce(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	input := authorizeMobileForTest(t, handler, repository.userID)
	var group sync.WaitGroup
	statuses := make([]int, 2)
	for i := range statuses {
		group.Go(func() {
			statuses[i] = exchangeMobileForTest(t, handler, input).Code
		})
	}
	group.Wait()
	require.ElementsMatch(t, []int{http.StatusOK, http.StatusBadRequest}, statuses)
}

func TestMobileCookieOriginContractKeepsBrowserCSRFProtection(t *testing.T) {
	policy, err := web.NewOriginPolicy("https://*.fortyone.app")
	require.NoError(t, err)
	for _, origin := range []string{"", "https://attacker.example", "https://cloud.fortyone.app"} {
		t.Run(origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPatch, "/workspaces/example/stories/id", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "native-cookie"})
			request.Header.Set("Origin", origin)
			recorder := httptest.NewRecorder()
			protected := mid.RequireTrustedBrowserOrigin(policy)(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return web.Respond(ctx, w, nil, http.StatusNoContent)
			})
			require.NoError(t, protected(t.Context(), recorder, request))
			if origin == "https://cloud.fortyone.app" {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			} else {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			}
		})
	}
}

func TestMobileLogoutReportsRevocationFailure(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		t.Run(fmt.Sprint(unavailable), func(t *testing.T) {
			handler, server, _ := mobileSessionTestHandler(t)
			if unavailable {
				server.SetError("ERR unavailable")
			}
			request := httptest.NewRequest(http.MethodDelete, "/users/session", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
			recorder := httptest.NewRecorder()
			require.NoError(t, handler.ClearSession(t.Context(), recorder, request))
			require.Len(t, recorder.Result().Cookies(), 1)
			require.Equal(t, -1, recorder.Result().Cookies()[0].MaxAge)
			if unavailable {
				require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
			} else {
				require.Equal(t, http.StatusNoContent, recorder.Code)
				require.False(t, server.Exists(cache.AuthSessionCacheKey("browser-cookie")))
			}
		})
	}
}
