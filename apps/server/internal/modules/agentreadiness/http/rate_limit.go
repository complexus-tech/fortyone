package agentreadinesshttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

const (
	oauthAuthorizationRequestLimit  int64 = 3_000
	oauthAuthorizationRequestWindow       = time.Minute
	oauthTokenRequestLimit          int64 = 6_000
	oauthTokenRequestWindow               = time.Minute
	oauthRegistrationRequestLimit   int64 = 600
	oauthRegistrationRequestWindow        = time.Hour
	oauthRevocationRequestLimit     int64 = 6_000
	oauthRevocationRequestWindow          = time.Minute
	mcpGrantRequestLimit            int64 = 600
	mcpGrantRequestWindow                 = time.Minute
)

func oauthGlobalRateLimit(cfg Config, scope string, limit int64, window time.Duration) web.Middleware {
	if strings.TrimSpace(scope) == "" || limit <= 0 || window <= 0 {
		panic("valid OAuth rate limit configuration is required")
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
			if cfg.Cache == nil {
				return oauthError(writer, http.StatusServiceUnavailable, "temporarily_unavailable", "OAuth rate limit is unavailable")
			}
			count, err := cfg.Cache.IncrementWithTTL(ctx, "rate-limit:mcp-oauth:global:"+scope, window)
			if err != nil {
				if cfg.Log != nil {
					cfg.Log.Error(ctx, "failed to enforce MCP OAuth rate limit", "scope", scope, "error", err)
				}
				return oauthError(writer, http.StatusServiceUnavailable, "temporarily_unavailable", "OAuth rate limit is unavailable")
			}
			web.SetRateLimitHeaders(writer, "mcp-oauth-"+scope, limit, limit-count, window)
			if count > limit {
				writer.Header().Set("Retry-After", retryAfter(window))
				return oauthError(writer, http.StatusTooManyRequests, "temporarily_unavailable", "OAuth request rate exceeded")
			}
			return next(ctx, writer, request)
		}
	}
}

func (h *Handler) mcpGrantRateLimit(next http.Handler, limit int64, window time.Duration) http.Handler {
	if limit <= 0 || window <= 0 {
		panic("valid MCP grant rate limit configuration is required")
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		tokenInfo := mcpauth.TokenInfoFromContext(request.Context())
		if tokenInfo == nil {
			http.Error(writer, "authenticated OAuth grant is unavailable", http.StatusInternalServerError)
			return
		}
		grantID, _ := tokenInfo.Extra["grant_id"].(string)
		if strings.TrimSpace(grantID) == "" {
			http.Error(writer, "authenticated OAuth grant is unavailable", http.StatusInternalServerError)
			return
		}
		if h.cfg.Cache == nil {
			http.Error(writer, "MCP rate limit is unavailable", http.StatusServiceUnavailable)
			return
		}
		count, err := h.cfg.Cache.IncrementWithTTL(request.Context(), "rate-limit:mcp:grant:"+grantID, window)
		if err != nil {
			if h.cfg.Log != nil {
				h.cfg.Log.Error(request.Context(), "failed to enforce MCP grant rate limit", "error", err)
			}
			http.Error(writer, "MCP rate limit is unavailable", http.StatusServiceUnavailable)
			return
		}
		web.SetRateLimitHeaders(writer, "mcp-oauth-grant", limit, limit-count, window)
		if count > limit {
			writer.Header().Set("Retry-After", retryAfter(window))
			http.Error(writer, "MCP request rate exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func retryAfter(window time.Duration) string {
	seconds := max(int64(window/time.Second), 1)
	return strconv.FormatInt(seconds, 10)
}
