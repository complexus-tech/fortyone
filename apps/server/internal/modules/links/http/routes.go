package linkshttp

import (
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
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
	Service           *links.Service
}

func Routes(cfg Config, app *web.App) {
	linksService := cfg.Service
	h := New(cfg.Log, linksService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Post("/workspaces/{workspaceSlug}/links", h.CreateLink, auth, workspace, memberAndAdmin)
	app.Put("/workspaces/{workspaceSlug}/links/{id}", h.UpdateLink, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/links/{id}", h.DeleteLink, auth, workspace, memberAndAdmin)
}
