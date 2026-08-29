package epicshttp

import (
	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
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
	Service           *epics.Service
}

func Routes(cfg Config, app *web.App) {
	epicsService := cfg.Service
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	h := New(epicsService)

	app.Get("/workspaces/{workspaceSlug}/epics", h.List, auth, workspace)
}
