package mid

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ctxKey string

const (
	authCookieName = "fortyone_session"
)

var authCache *cache.Service

func SetAuthCache(service *cache.Service) {
	authCache = service
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	return auth.GetUserID(ctx)
}

func Auth(log *logger.Logger, secretKey string) web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if userID, ok, err := resolveUserIDFromSessionCookie(ctx, r); err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			} else if ok {
				ctx = auth.SetUserID(ctx, userID)
				return next(ctx, w, r)
			}

			tokenString := resolveTokenFromRequest(r)
			if tokenString == "" {
				return web.RespondError(ctx, w, errors.New("unauthorized: token not found"), http.StatusUnauthorized)
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			})
			if err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			}
			if !token.Valid {
				return web.RespondError(ctx, w, errors.New("token is invalid"), http.StatusUnauthorized)
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			}

			ctx = auth.SetUserID(ctx, userID)

			return next(ctx, w, r)
		}
		return h
	}
	return m
}

// OptionalAuth attaches a normal FortyOne user when valid credentials are
// present while allowing a genuinely unauthenticated request to continue.
// Invalid supplied bearer credentials still fail closed.
func OptionalAuth(log *logger.Logger, secretKey string) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			cookie, cookieErr := r.Cookie(authCookieName)
			cookieSupplied := cookieErr == nil && strings.TrimSpace(cookie.Value) != ""
			if userID, ok, err := resolveUserIDFromSessionCookie(ctx, r); err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			} else if ok {
				return next(auth.SetUserID(ctx, userID), w, r)
			} else if cookieSupplied {
				return web.RespondError(ctx, w, errors.New("session is invalid or expired"), http.StatusUnauthorized)
			}

			tokenString := resolveTokenFromRequest(r)
			if tokenString == "" {
				return next(ctx, w, r)
			}
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				if err == nil {
					err = errors.New("token is invalid")
				}
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			}
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return web.RespondError(ctx, w, err, http.StatusUnauthorized)
			}
			return next(auth.SetUserID(ctx, userID), w, r)
		}
	}
}

func resolveUserIDFromSessionCookie(
	ctx context.Context,
	r *http.Request,
) (uuid.UUID, bool, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return uuid.Nil, false, nil
	}

	if authCache == nil {
		return uuid.Nil, false, nil
	}

	var userID string
	if err := authCache.Get(ctx, cache.AuthSessionCacheKey(cookie.Value), &userID); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, false, err
	}

	return parsedUserID, true, nil
}

func resolveTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	queryToken := r.URL.Query().Get("token")
	if subtle.ConstantTimeEq(int32(len(queryToken)), 0) == 0 {
		return queryToken
	}

	return ""
}
