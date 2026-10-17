package mid

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/pkg/web"
)

// RequireChatEnabled runs after authentication and workspace membership checks.
func RequireChatEnabled() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			workspace, err := GetWorkspace(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("workspace authentication required"), http.StatusUnauthorized)
			}
			if !workspace.ChatEnabled {
				w.Header().Set("Cache-Control", "private, no-store")
				return web.RespondError(ctx, w, errors.New("Messages is not enabled for this workspace"), http.StatusForbidden)
			}
			return next(ctx, w, r)
		}
	}
}
