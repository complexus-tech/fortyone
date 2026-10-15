package usershttp

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func authorizeDesktopForTest(t *testing.T, handler *Handlers, userID uuid.UUID) mobileExchangeRequest {
	t.Helper()
	return authorizeDesktopRedirectForTest(t, handler, userID, desktopRedirectURI)
}

func authorizeDesktopRedirectForTest(t *testing.T, handler *Handlers, userID uuid.UUID, redirectURI string) mobileExchangeRequest {
	t.Helper()
	verifier := strings.Repeat("d", 43)
	digest := sha256.Sum256([]byte(verifier))
	request := mobileRequest(t, "/auth/desktop/authorize", mobileAuthorizationRequest{
		State: strings.Repeat("s", 43), CodeChallenge: base64.RawURLEncoding.EncodeToString(digest[:]),
		CodeChallengeMethod: "S256", RedirectURI: redirectURI,
	})
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.AuthorizeDesktop(platformauth.SetUserID(t.Context(), userID), recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var response struct{ Data struct{ Code, State string } }
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return mobileExchangeRequest{Code: response.Data.Code, State: response.Data.State, CodeVerifier: verifier, RedirectURI: redirectURI}
}

func exchangeDesktopForTest(t *testing.T, handler *Handlers, input mobileExchangeRequest) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.ExchangeDesktop(t.Context(), recorder, mobileRequest(t, "/auth/desktop/exchange", input)))
	return recorder
}

func TestDesktopHandoffSeparatesClientNamespacesAndConsumesOnce(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	input := authorizeDesktopForTest(t, handler, repository.userID)
	mobileSubstitution := input
	mobileSubstitution.RedirectURI = mobileRedirectURI
	require.Equal(t, http.StatusBadRequest, exchangeMobileForTest(t, handler, mobileSubstitution).Code)
	wrongVerifier := input
	wrongVerifier.CodeVerifier = strings.Repeat("x", 43)
	require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, wrongVerifier).Code)
	response := exchangeDesktopForTest(t, handler, input)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	require.True(t, cookies[0].HttpOnly)
	require.NotContains(t, response.Body.String(), cookies[0].Value)
	require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, input).Code)
}

func TestDesktopCannotExchangeMobileCodeEvenWithMatchingProof(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	input := authorizeMobileForTest(t, handler, repository.userID)
	forged := input
	forged.RedirectURI = desktopRedirectURI
	require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, forged).Code)
	// Even an accidentally misnamespaced cache record must fail its explicit
	// client binding, preserving the original mobile code for the mobile app.
	var record mobileAuthorizationCode
	require.NoError(t, handler.cache.Get(t.Context(), mobileCodeCacheKey(input.Code), &record))
	require.NoError(t, handler.cache.Set(t.Context(), nativeCodeCacheKey("desktop", input.Code), record, mobileCodeTTL))
	require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, forged).Code)
	require.Equal(t, http.StatusOK, exchangeMobileForTest(t, handler, input).Code)
}

func TestDesktopHandoffRejectsAlternateRedirectAndRevokedSession(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	input := authorizeDesktopForTest(t, handler, repository.userID)
	for _, redirect := range []string{mobileRedirectURI, desktopRedirectURI + "/", desktopRedirectURI + "?next=evil", "https://example.com/login"} {
		invalid := input
		invalid.RedirectURI = redirect
		require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, invalid).Code)
	}
	repository.version++
	response := exchangeDesktopForTest(t, handler, input)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Empty(t, response.Result().Cookies())
}

func TestDesktopHandoffPreservesVerifiedWorkspaceSSOProof(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	var browser platformauth.BrowserSession
	require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey("browser-cookie"), &browser))
	authenticated := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	browser.AuthenticatedAt = authenticated
	browser.WorkspaceSSO = &platformauth.WorkspaceSSOAuthentication{WorkspaceID: uuid.New(), ConnectionID: uuid.New(), Generation: 4, AuthenticatedAt: authenticated}
	require.NoError(t, handler.cache.Set(t.Context(), cache.AuthSessionCacheKey("browser-cookie"), browser, time.Hour))
	response := exchangeDesktopForTest(t, handler, authorizeDesktopForTest(t, handler, repository.userID))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var desktop platformauth.BrowserSession
	require.NoError(t, handler.cache.Get(t.Context(), cache.AuthSessionCacheKey(response.Result().Cookies()[0].Value), &desktop))
	require.Equal(t, browser.WorkspaceSSO, desktop.WorkspaceSSO)
	require.Equal(t, browser.AuthenticatedAt, desktop.AuthenticatedAt, "handoff must not freshen the SSO authentication time")
	require.NotEqual(t, browser.SessionID, desktop.SessionID)
}

func TestDesktopDevelopmentHandoffIsBoundToItsExactRedirect(t *testing.T) {
	for _, redirect := range []string{desktopRedirectURI, platformauth.DesktopDevelopmentRedirectURI} {
		t.Run(redirect, func(t *testing.T) {
			handler, _, repository := mobileSessionTestHandler(t)
			handler.deploymentMode = deployment.Development
			input := authorizeDesktopRedirectForTest(t, handler, repository.userID, redirect)
			substitution := input
			if redirect == desktopRedirectURI {
				substitution.RedirectURI = platformauth.DesktopDevelopmentRedirectURI
			} else {
				substitution.RedirectURI = desktopRedirectURI
			}
			require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, substitution).Code)
			response := exchangeDesktopForTest(t, handler, input)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Len(t, response.Result().Cookies(), 1)
			require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, input).Code)
		})
	}
}

func TestDesktopDevelopmentHandoffRequiresDevelopmentAtAuthorizationAndExchange(t *testing.T) {
	for _, mode := range []deployment.Mode{deployment.Production, deployment.Staging, deployment.Test, ""} {
		t.Run(string(mode), func(t *testing.T) {
			handler, _, repository := mobileSessionTestHandler(t)
			handler.deploymentMode = mode
			request := mobileRequest(t, "/auth/desktop/authorize", mobileAuthorizationRequest{
				State: strings.Repeat("s", 43), CodeChallenge: strings.Repeat("c", 43),
				CodeChallengeMethod: "S256", RedirectURI: platformauth.DesktopDevelopmentRedirectURI,
			})
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
			response := httptest.NewRecorder()
			require.NoError(t, handler.AuthorizeDesktop(platformauth.SetUserID(t.Context(), repository.userID), response, request))
			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Empty(t, response.Result().Cookies())

			handler.deploymentMode = deployment.Development
			input := authorizeDesktopRedirectForTest(t, handler, repository.userID, platformauth.DesktopDevelopmentRedirectURI)
			handler.deploymentMode = mode
			rejected := exchangeDesktopForTest(t, handler, input)
			require.Equal(t, http.StatusBadRequest, rejected.Code)
			require.Empty(t, rejected.Result().Cookies())
			// A rejected redirect cannot consume a valid authorization code.
			handler.deploymentMode = deployment.Development
			require.Equal(t, http.StatusOK, exchangeDesktopForTest(t, handler, input).Code)
			handler.deploymentMode = mode
			require.Equal(t, http.StatusOK, exchangeDesktopForTest(t, handler, authorizeDesktopForTest(t, handler, repository.userID)).Code)
		})
	}
}

func TestDesktopDevelopmentRedirectRejectsURLVariants(t *testing.T) {
	handler, _, repository := mobileSessionTestHandler(t)
	handler.deploymentMode = deployment.Development
	input := authorizeDesktopRedirectForTest(t, handler, repository.userID, platformauth.DesktopDevelopmentRedirectURI)
	for _, redirect := range []string{
		"fortyone-desktop-dev://login/", "fortyone-desktop-dev://login?next=evil",
		"fortyone-desktop-dev://login#fragment", "fortyone-desktop-dev://user@login",
		"fortyone-desktop-dev://other", "fortyone-desktop-dev://%6cogin", "FORTYONE-DESKTOP-DEV://login",
	} {
		t.Run(redirect, func(t *testing.T) {
			request := mobileRequest(t, "/auth/desktop/authorize", mobileAuthorizationRequest{
				State: input.State, CodeChallenge: strings.Repeat("c", 43),
				CodeChallengeMethod: "S256", RedirectURI: redirect,
			})
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "browser-cookie"})
			response := httptest.NewRecorder()
			require.NoError(t, handler.AuthorizeDesktop(platformauth.SetUserID(t.Context(), repository.userID), response, request))
			require.Equal(t, http.StatusBadRequest, response.Code)
			invalid := input
			invalid.RedirectURI = redirect
			require.Equal(t, http.StatusBadRequest, exchangeDesktopForTest(t, handler, invalid).Code)
		})
	}
	require.Equal(t, http.StatusOK, exchangeDesktopForTest(t, handler, input).Code)
}
