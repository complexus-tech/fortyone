package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type RateLimitStore interface {
	IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type AuthenticatedUserRateLimitConfig struct {
	Scope  string
	Limit  int64
	Window time.Duration
}

func AuthenticatedUserRateLimit(
	log *logger.Logger,
	store RateLimitStore,
	config AuthenticatedUserRateLimitConfig,
) web.Middleware {
	if strings.TrimSpace(config.Scope) == "" {
		panic("authenticated user rate limit scope is required")
	}
	if config.Limit <= 0 {
		panic("authenticated user rate limit must be positive")
	}
	if config.Window <= 0 {
		panic("authenticated user rate limit window must be positive")
	}

	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			userID, err := GetUserID(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("user not authenticated"), http.StatusUnauthorized)
			}
			if store == nil {
				if log != nil {
					log.Error(ctx, "rate limit store is not configured", "scope", config.Scope)
				}
				return web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}

			portalSlug := strings.ToLower(strings.TrimSpace(web.Params(r, "portalSlug")))
			key := fmt.Sprintf("rate-limit:%s:portal:%s:user:%s", config.Scope, portalSlug, userID)
			count, err := store.IncrementWithTTL(ctx, key, config.Window)
			if err != nil {
				if log != nil {
					log.Error(ctx, "failed to enforce request rate limit", "scope", config.Scope, "error", err)
				}
				return web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}
			if count > config.Limit {
				retryAfterSeconds := int64(config.Window / time.Second)
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
				return web.RespondError(ctx, w, errors.New("too many requests"), http.StatusTooManyRequests)
			}

			return next(ctx, w, r)
		}
	}
}
