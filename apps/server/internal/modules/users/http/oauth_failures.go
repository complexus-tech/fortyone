package usershttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func oauthStateFailureCode(err error) string {
	if errors.Is(err, cache.ErrNotFound) {
		return oauthFailureExpired
	}
	return oauthFailureGeneric
}

func oauthProviderFailureCode(providerError string) string {
	if providerError == "access_denied" {
		return oauthFailureCancelled
	}
	return oauthFailureGeneric
}

func oauthSignInFailureCode(err error) string {
	_, publicError := publicSignInError(err)
	if errors.Is(publicError, users.ErrInvalidCredentials) {
		return oauthFailureUnavailable
	}
	return oauthFailureGeneric
}

func (h *Handlers) redirectOAuthFailure(ctx context.Context, w http.ResponseWriter, r *http.Request, callbackURL, failureCode string) error {
	// OAuth completions supply only consumed server-side state; starts supply
	// the validated app callback. Provider query fields never enter this URL.
	failureURL := h.oauthFailureURL(callbackURL, failureCode)
	if failureURL == "" {
		return web.RespondError(ctx, w, errors.New("sign-in could not be completed"), http.StatusServiceUnavailable)
	}
	if h.log != nil {
		h.log.Warn(ctx, "browser sign-in could not be completed", "failure_code", failureCode)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, failureURL, http.StatusSeeOther)
	return nil
}

func (h *Handlers) oauthFailureURL(callbackURL, failureCode string) string {
	website, websiteErr := h.trustedOAuthURL(h.websiteURL)
	if websiteErr != nil || website == nil || website.Host == "" {
		return ""
	}
	callback, callbackErr := h.trustedOAuthURL(callbackURL)
	if callbackErr != nil || callback == nil {
		callback = nil
	}
	if callback != nil && callback.Host == "" {
		callback = website.ResolveReference(callback)
	}

	query := url.Values{"error": {failureCode}}
	if callback != nil {
		isWebsiteOrigin := callback.Scheme == website.Scheme && callback.Host == website.Host
		isLoginPath := callback.Path == "" || callback.Path == "/" || callback.Path == "/signup"
		if callback.Path == "/auth-callback" || (isWebsiteOrigin && isLoginPath) {
			continuation, _, err := web.OptionalTextQueryParameter(callback.Query(), "callbackUrl", maxCallbackURLLength, maxCallbackURLLength)
			if err == nil {
				if safe, err := sanitizeCallbackURL(continuation, h.cookieDomain, h.websiteURL); err == nil && safe != "" {
					query.Set("callbackUrl", safe)
				}
			}
		} else {
			query.Set("callbackUrl", callback.String())
		}
		if mobileApp := callback.Query()["mobileApp"]; len(mobileApp) == 1 && mobileApp[0] == "true" {
			query.Set("mobileApp", "true")
		}
	}
	// Workspace and API hosts may also be valid success destinations, but their
	// root paths do not render the login page. Always use the configured app.
	return (&url.URL{Scheme: website.Scheme, Host: website.Host, Path: "/", RawQuery: query.Encode()}).String()
}

func (h *Handlers) trustedOAuthURL(raw string) (*url.URL, error) {
	safe, err := sanitizeCallbackURL(raw, h.cookieDomain, h.websiteURL)
	if err != nil || safe == "" {
		return nil, err
	}
	return url.Parse(safe)
}
