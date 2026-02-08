package mid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/pkg/cache"
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
func Workspace(log *logger.Logger, db *sqlx.DB, cache *cache.Service) web.Middleware {
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

			// Define cache key for workspace data
			cacheKey := fmt.Sprintf("workspace:%s:user:%s", workspaceSlug, userID)

			// Try to get workspace from cache first
			var workspace WorkspaceInfo
			if err := cache.Get(ctx, cacheKey, &workspace); err == nil {
				// Cache hit - store in context and continue
				ctx = context.WithValue(ctx, workspaceKey, workspace)
				return next(ctx, w, r)
			}

			// Cache miss or error - get from database
			workspace, err = getWorkspace(ctx, db, workspaceSlug, userID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return web.RespondError(ctx, w, errors.New("access denied"), http.StatusNotFound)
				}
				log.Error(ctx, "failed to get workspace", "error", err, "slug", workspaceSlug, "userID", userID)
				return web.RespondError(ctx, w, errors.New("internal server error"), http.StatusInternalServerError)
			}

			// Store in cache for future requests (e.g., 15 minute TTL)
			if err := cache.Set(ctx, cacheKey, workspace, 15*time.Minute); err != nil {
				log.Error(ctx, "failed to set workspace cache", "error", err)
			}

			// Update last_login_at after successful workspace access and caching
			if err := updateUserLastLogin(ctx, db, userID); err != nil {
				log.Error(ctx, "failed to update last login", "error", err, "userID", userID)
				// Don't fail the request - this is not critical
			}

			// Update workspace last_accessed_at
			if err := updateWorkspaceLastAccessed(ctx, db, workspace.ID); err != nil {
				log.Error(ctx, "failed to update workspace last accessed", "error", err, "workspaceID", workspace.ID)
				// Don't fail the request - this is not critical
			}

			ctx = context.WithValue(ctx, workspaceKey, workspace)

			return next(ctx, w, r)
		}
		return h
	}
	return m
}

// getWorkspace retrieves workspace information from the database
func getWorkspace(ctx context.Context, db *sqlx.DB, workspaceSlug string, userID uuid.UUID) (WorkspaceInfo, error) {
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
		return WorkspaceInfo{}, err
	}
	defer stmt.Close()

	var workspace WorkspaceInfo
	if err := stmt.GetContext(ctx, &workspace, params); err != nil {
		return WorkspaceInfo{}, err
	}

	return workspace, nil
}

// updateUserLastLogin updates the user's last login timestamp
func updateUserLastLogin(ctx context.Context, db *sqlx.DB, userID uuid.UUID) error {
	query := `UPDATE users 
	SET last_login_at = NOW(),
		inactivity_warning_sent_at = NULL
		WHERE user_id = :user_id AND is_active = true`
	params := map[string]any{"user_id": userID}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, params)
	return err
}

// updateWorkspaceLastAccessed updates the workspace's last accessed timestamp
func updateWorkspaceLastAccessed(ctx context.Context, db *sqlx.DB, workspaceID uuid.UUID) error {
	query := `UPDATE workspaces 
						SET last_accessed_at = NOW(),
						inactivity_warning_sent_at = NULL
						WHERE workspace_id = :workspace_id`
	params := map[string]any{"workspace_id": workspaceID}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, params)
	return err
}
