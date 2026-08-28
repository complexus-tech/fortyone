package objectivestatushttp

import (
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *objectivestatus.Service
}

func Routes(cfg Config, app *web.App) {
	objectiveStatusService := cfg.Service
	h := New(objectiveStatusService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Get("/workspaces/{workspaceSlug}/objective-statuses", h.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/objective-statuses", h.Create, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/objective-statuses/{statusId}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/objective-statuses/{statusId}", h.Delete, auth, workspace, adminOnly)
}
