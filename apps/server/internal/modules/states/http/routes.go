package stateshttp

import (
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
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
	Service           *states.Service
}

func Routes(cfg Config, app *web.App) {
	statesService := cfg.Service
	h := New(statesService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	app.Get("/workspaces/{workspaceSlug}/states", h.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/states", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/states/{stateId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/states/{stateId}", h.Delete, auth, workspace)
}
