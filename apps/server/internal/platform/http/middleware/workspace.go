package mid

import (
	"context"
	"errors"
	"net/http"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	workspaceKey           ctxKey = "workspace"
	workspaceAccessTimeout        = 250 * time.Millisecond
)

var ErrWorkspaceAccessDenied = errors.New("workspace access denied")

// WorkspaceInfo is the current authoritative workspace membership bound to a
// request after authentication.
type WorkspaceInfo struct {
	ID       uuid.UUID
	Name     string
	Slug     string
	UserRole string
}

// WorkspaceResolver is the narrow tenant capability required by HTTP
// middleware. Implementations must resolve membership from authoritative
// storage on every call and must not serve cached authorization state.
type WorkspaceResolver interface {
	ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (WorkspaceInfo, error)
	RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error
}

// GetWorkspace retrieves the workspace from the context.
func GetWorkspace(ctx context.Context) (WorkspaceInfo, error) {
	workspace, ok := ctx.Value(workspaceKey).(WorkspaceInfo)
	if !ok {
		return WorkspaceInfo{}, errors.New("workspace not found in context")
	}
	return workspace, nil
}

// Workspace resolves live membership and role state, binds the actor tenant,
// then records last access as a bounded best-effort write.
func Workspace(log *logger.Logger, resolver WorkspaceResolver) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			workspaceSlug := web.Params(r, "workspaceSlug")
			if workspaceSlug == "" {
				return web.RespondError(ctx, w, errors.New("workspace slug is required"), http.StatusBadRequest)
			}

			userID, err := GetUserID(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("user not authenticated"), http.StatusUnauthorized)
			}
			if resolver == nil {
				log.Error(ctx, "workspace resolver is unavailable")
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			workspace, err := resolver.ResolveCurrentWorkspace(ctx, workspaceSlug, userID)
			if err != nil {
				if errors.Is(err, ErrWorkspaceAccessDenied) {
					return web.RespondError(ctx, w, errors.New("access denied"), http.StatusNotFound)
				}
				log.Error(ctx, "failed to resolve current workspace membership", "error", err)
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			ctx, err = platformauth.BindWorkspace(ctx, workspace.ID)
			if err != nil {
				log.Error(ctx, "failed to bind actor workspace", "error", err, "workspace_id", workspace.ID)
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}
			ctx = context.WithValue(ctx, workspaceKey, workspace)

			if ctx.Err() == nil {
				accessCtx, cancel := context.WithTimeout(ctx, workspaceAccessTimeout)
				if accessErr := resolver.RecordWorkspaceAccess(accessCtx, workspace.ID, userID); accessErr != nil {
					log.Warn(ctx, "failed to record workspace access", "error", accessErr, "workspace_id", workspace.ID)
				}
				cancel()
			}

			return next(ctx, w, r)
		}
	}
}
