package mid

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const workspaceKey ctxKey = "workspace"

// WorkspaceInfo holds workspace information in context
type WorkspaceInfo struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Slug     string    `db:"slug"`
	UserRole string    `db:"user_role"`
}

// GetWorkspace retrieves the workspace from the context.
func GetWorkspace(ctx context.Context) (WorkspaceInfo, error) {
	workspace, ok := ctx.Value(workspaceKey).(WorkspaceInfo)
	if !ok {
		return WorkspaceInfo{}, errors.New("workspace not found in context")
	}
	return workspace, nil
}

// Workspace validates that the user belongs to the workspace and stores workspace data in context.
func Workspace(log *logger.Logger, db *sqlx.DB) web.Middleware {
	m := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			workspaceSlug := web.Params(r, "workspaceSlug")
			if workspaceSlug == "" {
				return web.RespondError(ctx, w, errors.New("workspace slug is required"), http.StatusBadRequest)
			}
			userID, err := GetUserID(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("user not authenticated"), http.StatusUnauthorized)
			}

			query := `
				SELECT 
					w.workspace_id as id,
					w.name,
					w.slug,
					wm.role as user_role
				FROM workspaces w
				INNER JOIN workspace_members wm ON w.workspace_id = wm.workspace_id
				WHERE w.slug = :slug AND wm.user_id = :user_id
			`

			params := map[string]any{
				"slug":    workspaceSlug,
				"user_id": userID,
			}

			stmt, err := db.PrepareNamedContext(ctx, query)
			if err != nil {
				log.Error(ctx, "failed to prepare named statement", "error", err)
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}
			defer stmt.Close()

			var workspace WorkspaceInfo
			if err := stmt.GetContext(ctx, &workspace, params); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return web.RespondError(ctx, w, errors.New("workspace not found or access denied"), http.StatusNotFound)
				}
				log.Error(ctx, "failed to get workspace", "error", err, "slug", workspaceSlug, "userID", userID)
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			ctx = context.WithValue(ctx, workspaceKey, workspace)

			return next(ctx, w, r)
		}
		return h
	}
	return m
}
