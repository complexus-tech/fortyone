package mid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type ctxKey string

const (
	authCookieName = "fortyone_session"
)

// SessionStore is the narrow browser-session cache capability required by
// authentication. It is injected so multiple app/test instances cannot race
// through mutable package-global state.
type SessionStore interface {
	Get(ctx context.Context, key string, dest any) error
}

// SessionAccountResolver is the authoritative PostgreSQL capability used by
// browser authentication. Active is returned separately so unknown and
// inactive accounts can fail closed without exposing account state.
type SessionAccountResolver interface {
	ResolveActiveBrowserSessionVersion(ctx context.Context, userID uuid.UUID) (version int64, active bool, err error)
}

// SessionResolver is the single browser-session path shared by required,
// optional, and OAuth authorization endpoints.
type SessionResolver interface {
	Resolve(ctx context.Context, r *http.Request) (uuid.UUID, bool, error)
}

// BrowserSessionResolver combines the opaque Redis record with authoritative
// account state. Redis is an index, never the source of truth for activation or
// revocation.
type BrowserSessionResolver struct {
	sessions SessionStore
	accounts SessionAccountResolver
}

func NewBrowserSessionResolver(sessions SessionStore, accounts SessionAccountResolver) *BrowserSessionResolver {
	return &BrowserSessionResolver{sessions: sessions, accounts: accounts}
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	return auth.GetUserID(ctx)
}

func GetActor(ctx context.Context) (auth.Actor, error) {
	return auth.GetActor(ctx)
}

// RequireScopes performs the cheap credential-level half of authorization.
// Services must still authorize the current workspace role, team membership,
// resource visibility, and requested mutation after loading the resource.
func RequireScopes(required ...auth.Scope) web.Middleware {
	if _, err := auth.NewScopeSet(required...); err != nil {
		panic(err)
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			actor, err := auth.GetActor(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("authentication is required"), http.StatusUnauthorized)
			}
			if !actor.Scopes.ContainsAll(required...) {
				return web.RespondError(ctx, w, errors.New("credential scope is insufficient"), http.StatusForbidden)
			}
			return next(ctx, w, r)
		}
	}
}

// ResolveSessionUserID resolves the browser session without writing an HTTP
// response. Authorization endpoints use it before issuing scoped credentials.
func ResolveSessionUserID(ctx context.Context, r *http.Request, resolver SessionResolver) (uuid.UUID, bool, error) {
	if resolver == nil {
		return uuid.Nil, false, errors.New("browser session resolver is not configured")
	}
	return resolver.Resolve(ctx, r)
}

// Auth authenticates the first-party browser session. Public API credentials
// deliberately use MachineAuth, whose verifier supplies an attributed actor,
// scopes, tenant restrictions, expiry, and revocation state. Keeping those
// credential families separate prevents an arbitrary HS256 value signed with
// the broad application secret from being confused with a user session.
//
// The logger and secret parameters remain temporarily for route-constructor
// compatibility while modules are migrated. Neither participates in browser
// authentication and may be removed once the route configuration bridge is
// gone.
func Auth(_ *logger.Logger, _ string, resolver SessionResolver) web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if userID, ok, err := ResolveSessionUserID(ctx, r, resolver); err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			} else if ok {
				ctx = auth.SetUserID(ctx, userID)
				return next(ctx, w, r)
			}

			if resolveTokenFromRequest(r) != "" {
				return web.RespondError(ctx, w, errors.New("legacy user bearer tokens are not accepted"), http.StatusUnauthorized)
			}
			return web.RespondError(ctx, w, errors.New("authentication is required"), http.StatusUnauthorized)
		}
		return h
	}
	return m
}

// OptionalAuth attaches a normal FortyOne user when valid credentials are
// present while allowing a genuinely unauthenticated request to continue.
// Invalid supplied bearer credentials still fail closed.
func OptionalAuth(_ *logger.Logger, _ string, resolver SessionResolver) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			cookie, cookieErr := r.Cookie(authCookieName)
			cookieSupplied := cookieErr == nil && strings.TrimSpace(cookie.Value) != ""
			if userID, ok, err := ResolveSessionUserID(ctx, r, resolver); err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			} else if ok {
				return next(auth.SetUserID(ctx, userID), w, r)
			} else if cookieSupplied {
				return web.RespondError(ctx, w, errors.New("session is invalid or expired"), http.StatusUnauthorized)
			}

			if resolveTokenFromRequest(r) == "" {
				return next(ctx, w, r)
			}
			return web.RespondError(ctx, w, errors.New("legacy user bearer tokens are not accepted"), http.StatusUnauthorized)
		}
	}
}

func (resolver *BrowserSessionResolver) Resolve(
	ctx context.Context,
	r *http.Request,
) (uuid.UUID, bool, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return uuid.Nil, false, nil
	}

	if resolver == nil || resolver.sessions == nil || resolver.accounts == nil {
		return uuid.Nil, false, errors.New("browser session resolver is not configured")
	}

	var session auth.BrowserSession
	err = resolver.sessions.Get(ctx, cache.AuthSessionCacheKey(cookie.Value), &session)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}

	// Legacy string-valued records and legacy raw-token cache keys are rejected.
	// Their user-only shape cannot prove a revocation epoch and accepting them
	// would resurrect sessions after account reactivation.
	if err := session.Validate(); err != nil {
		return uuid.Nil, false, err
	}

	version, active, err := resolver.accounts.ResolveActiveBrowserSessionVersion(ctx, session.UserID)
	if err != nil {
		return uuid.Nil, false, err
	}
	if !active || version <= 0 || version != session.Version {
		return uuid.Nil, false, auth.ErrInvalidBrowserSession
	}

	return session.UserID, true, nil
}

func resolveTokenFromRequest(r *http.Request) string {
	parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
