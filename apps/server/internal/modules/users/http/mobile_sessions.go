package usershttp

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	mobileRedirectURI = "fortyone://login"
	mobileCodeTTL     = 2 * time.Minute
)

var (
	mobileRandomValuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	mobileVerifierPattern    = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)
	errInvalidMobileCode     = errors.New("mobile sign-in expired or is invalid; start sign-in again")
)

type mobileAuthorizationRequest struct {
	State               string `json:"state"`
	CodeChallenge       string `json:"codeChallenge"`
	CodeChallengeMethod string `json:"codeChallengeMethod"`
	RedirectURI         string `json:"redirectUri"`
}

type mobileExchangeRequest struct {
	Code         string `json:"code"`
	State        string `json:"state"`
	CodeVerifier string `json:"codeVerifier"`
	RedirectURI  string `json:"redirectUri"`
}

type mobileAuthorizationCode struct {
	Session       platformauth.BrowserSession `json:"session"`
	State         string                      `json:"state"`
	CodeChallenge string                      `json:"codeChallenge"`
	ExpiresAt     time.Time                   `json:"expiresAt"`
}

func mobileCodeCacheKey(code string) string {
	digest := sha256.Sum256([]byte(code))
	return fmt.Sprintf("auth:mobile:code:v1:%x", digest)
}

// AuthorizeMobile creates a narrowly scoped browser-to-app handoff. Browser
// authentication and the global origin policy protect this POST; it neither
// returns the browser cookie nor creates a general-purpose login code.
func (h *Handlers) AuthorizeMobile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	var req mobileAuthorizationRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if !mobileRandomValuePattern.MatchString(req.State) ||
		!mobileRandomValuePattern.MatchString(req.CodeChallenge) ||
		req.CodeChallengeMethod != "S256" || req.RedirectURI != mobileRedirectURI {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	if h.cache == nil || h.users == nil {
		return web.RespondError(ctx, w, errors.New("mobile sign-in is unavailable"), http.StatusServiceUnavailable)
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	var browserSession platformauth.BrowserSession
	if err := h.cache.Get(ctx, cache.AuthSessionCacheKey(cookie.Value), &browserSession); err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			return web.RespondError(ctx, w, errors.New("mobile sign-in is unavailable"), http.StatusServiceUnavailable)
		}
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	if browserSession.Validate() != nil || browserSession.UserID != userID {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	version, active, err := h.users.ResolveActiveBrowserSessionVersion(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	if !active || version != browserSession.Version {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	code, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	record := mobileAuthorizationCode{
		Session: browserSession,
		State:   req.State, CodeChallenge: req.CodeChallenge, ExpiresAt: time.Now().Add(mobileCodeTTL),
	}
	if err := h.cache.Set(ctx, mobileCodeCacheKey(code), record, mobileCodeTTL); err != nil {
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	return web.Respond(ctx, w, map[string]string{"code": code, "state": req.State}, http.StatusOK)
}

func (record mobileAuthorizationCode) matches(req mobileExchangeRequest) bool {
	digest := sha256.Sum256([]byte(req.CodeVerifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])
	return record.Session.Validate() == nil && time.Now().Before(record.ExpiresAt) &&
		subtle.ConstantTimeCompare([]byte(record.State), []byte(req.State)) == 1 &&
		subtle.ConstantTimeCompare([]byte(record.CodeChallenge), []byte(challenge)) == 1
}

// ExchangeMobile establishes an independent opaque session in the app's HTTP
// transport. The code is useless without its per-request PKCE verifier.
// Only Set-Cookie carries the credential; the JSON body contains the user.
func (h *Handlers) ExchangeMobile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	var req mobileExchangeRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if !mobileRandomValuePattern.MatchString(req.Code) || !mobileRandomValuePattern.MatchString(req.State) ||
		!mobileVerifierPattern.MatchString(req.CodeVerifier) || req.RedirectURI != mobileRedirectURI {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
	}
	if h.cache == nil || h.users == nil {
		return web.RespondError(ctx, w, errors.New("mobile sign-in is unavailable"), http.StatusServiceUnavailable)
	}
	var record mobileAuthorizationCode
	if err := h.cache.Get(ctx, mobileCodeCacheKey(req.Code), &record); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	if !record.matches(req) {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
	}
	// Validate before Take so an intercepted code cannot be burned by a caller
	// who does not know the verifier. Take still makes successful use atomic.
	if err := h.cache.Take(ctx, mobileCodeCacheKey(req.Code), &record); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	if !record.matches(req) {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusBadRequest)
	}
	version, active, err := h.users.ResolveActiveBrowserSessionVersion(ctx, record.Session.UserID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	if !active || version != record.Session.Version {
		return web.RespondError(ctx, w, errInvalidMobileCode, http.StatusUnauthorized)
	}
	user, err := h.users.GetUser(ctx, record.Session.UserID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	token, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	expiresAt := time.Now().Add(SessionDuration)
	// Preserve the authorization's epoch. A concurrent revocation must not be
	// bypassed by minting a session at a newer epoch after this check.
	if err := h.cache.Set(ctx, cache.AuthSessionCacheKey(token), record.Session, SessionDuration); err != nil {
		return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
	}
	h.setSessionCookie(w, r, token, expiresAt)
	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}
