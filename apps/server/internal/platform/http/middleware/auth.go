package mid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ctxKey string

const (
	authCookieName = "fortyone_session"
)

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	return auth.GetUserID(ctx)
}

func Auth(log *logger.Logger, secretKey string) web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			tokenString := ""
			if cookie, err := r.Cookie(authCookieName); err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			} else {
				authHeader := r.Header.Get("Authorization")
				if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				} else {
					queryToken := r.URL.Query().Get("token")
					if queryToken != "" {
						tokenString = queryToken
					}
				}
			}

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
