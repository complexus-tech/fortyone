package feedbackhttp

import (
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Cache     *cache.Service
	Service   *feedback.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Get("/portals/{portalSlug}/feedback", h.GetPortal)
	app.Get("/workspaces/{workspaceSlug}/portals/{portalSlug}/feedback", h.GetWorkspacePortal)
	app.Get("/workspaces/{workspaceSlug}/feedback/portals", h.ListPortals, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/portals/{portalId}", h.UpdatePortal, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/boards", h.CreateBoard, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items", h.CreateItem, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/items/{itemId}/status", h.UpdateItemStatus, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/comments", h.CreateComment, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/vote", h.ToggleVote, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/story", h.CreateStoryFromItem, auth, workspace, adminOnly)
}
