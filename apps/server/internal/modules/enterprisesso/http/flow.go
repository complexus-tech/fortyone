package enterprisessohttp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const stateTTL = 10 * time.Minute
const stateCookie = "fortyone_sso_state"

func randomValue() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}
func stateKey(state string) string {
	digest := sha256.Sum256([]byte(state))
	return "workspace-sso.state." + hex.EncodeToString(digest[:])
}
func (h *Handlers) setStateCookie(w http.ResponseWriter, r *http.Request, state string, clear bool) {
	age := int(stateTTL.Seconds())
	expires := time.Now().Add(stateTTL)
	if clear {
		age = -1
		expires = time.Unix(0, 0)
	}
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: state, Path: "/auth/sso", Domain: h.cookieDomain, HttpOnly: true, Secure: h.production || r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"), SameSite: http.SameSiteLaxMode, MaxAge: age, Expires: expires})
}
func (h *Handlers) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	if len(query) > 1 || (len(query) == 1 && (len(query["returnTo"]) != 1 || query.Get("returnTo") != "security")) {
		return respond(ctx, w, nil, domain.ErrInvalid)
	}
	connection, err := h.service.Public(ctx, web.Params(r, "workspaceSlug"))
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	state, err := randomValue()
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	nonce, err := randomValue()
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	actorID, _ := auth.GetUserID(ctx)
	if query.Get("returnTo") == "security" && actorID == uuid.Nil {
		return respond(ctx, w, nil, domain.ErrForbidden)
	}
	attempt := domain.Attempt{WorkspaceID: connection.WorkspaceID, ConnectionID: connection.ID, Generation: connection.Generation, LinkUserID: actorID, Slug: web.Params(r, "workspaceSlug"), Nonce: nonce, Verifier: oauth2.GenerateVerifier(), CreatedAt: time.Now().UTC()}
	attempt.ReturnToSettings = query.Get("returnTo") == "security"
	authorize, err := h.service.AuthorizationURL(ctx, connection, attempt, state)
	if err != nil {
		return respond(ctx, w, nil, domain.ErrProvider)
	}
	if err := h.states.Set(ctx, stateKey(state), attempt, stateTTL); err != nil {
		return respond(ctx, w, nil, err)
	}
	h.setStateCookie(w, r, state, false)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, authorize, http.StatusTemporaryRedirect)
	return nil
}
func (h *Handlers) failure(w http.ResponseWriter, r *http.Request, code string) error {
	h.setStateCookie(w, r, "", true)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, h.websiteURL+"/?error="+url.QueryEscape(code), http.StatusSeeOther)
	return nil
}
func (h *Handlers) Callback(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	callback, err := web.ParseOAuthCallbackQuery(r.URL.Query())
	if err != nil {
		return h.failure(w, r, "sso_failed")
	}
	cookie, err := r.Cookie(stateCookie)
	if err != nil || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(callback.State)) != 1 {
		return h.failure(w, r, "sso_failed")
	}
	var attempt domain.Attempt
	if err := h.states.Take(ctx, stateKey(callback.State), &attempt); err != nil {
		return h.failure(w, r, "sso_expired")
	}
	if !attempt.CreatedAt.Add(stateTTL).After(time.Now()) || attempt.CreatedAt.After(time.Now().Add(30*time.Second)) || attempt.WorkspaceID == uuid.Nil || attempt.ConnectionID == uuid.Nil || callback.ProviderError != "" {
		return h.failure(w, r, "sso_failed")
	}
	if attempt.LinkUserID != uuid.Nil {
		current, err := auth.GetUserID(ctx)
		if err != nil || current != attempt.LinkUserID {
			return h.failure(w, r, "sso_failed")
		}
	}
	userID, identity, err := h.service.Complete(ctx, attempt, callback.Code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkRequired) {
			return h.failure(w, r, "sso_link_required")
		}
		return h.failure(w, r, "sso_failed")
	}
	assertion := auth.WorkspaceSSOAuthentication{WorkspaceID: attempt.WorkspaceID, ConnectionID: attempt.ConnectionID, Generation: attempt.Generation, AuthenticatedAt: identity.AuthenticatedAt}
	if err := h.issuer.Issue(ctx, w, r, userID, assertion); err != nil {
		return h.failure(w, r, "sso_failed")
	}
	h.setStateCookie(w, r, "", true)
	w.Header().Set("Cache-Control", "no-store")
	target := h.websiteURL + "/" + url.PathEscape(attempt.Slug)
	if attempt.ReturnToSettings {
		target += "/settings/workspace/security?tab=sso"
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
	return nil
}
