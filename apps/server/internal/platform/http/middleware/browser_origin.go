package mid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/web"
)

var ErrUntrustedBrowserOrigin = errors.New("browser origin is not allowed")

// RequireTrustedBrowserOrigin protects unsafe cookie-authenticated requests
// from CSRF. Bearer/API-key/webhook requests without the browser session cookie
// are unaffected and remain governed by their own authentication contracts.
func RequireTrustedBrowserOrigin(policy web.OriginPolicy) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if isSafeMethod(r.Method) || !hasBrowserSessionCookie(r) {
				return next(ctx, w, r)
			}

			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				if policy.AllowedOrigin(r) != "" {
					return next(ctx, w, r)
				}
				return web.RespondError(ctx, w, ErrUntrustedBrowserOrigin, http.StatusForbidden)
			}

			// Fetch Metadata is controlled by the browser. Only same-origin is a
			// safe no-Origin fallback; sibling subdomains report same-site.
			if strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "same-origin") {
				return next(ctx, w, r)
			}
			return web.RespondError(ctx, w, ErrUntrustedBrowserOrigin, http.StatusForbidden)
		}
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func hasBrowserSessionCookie(r *http.Request) bool {
	cookie, err := r.Cookie(authCookieName)
	return err == nil && strings.TrimSpace(cookie.Value) != ""
}
