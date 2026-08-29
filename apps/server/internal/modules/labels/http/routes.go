package labelshttp

import (
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
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
	Service           *labels.Service
}

func Routes(cfg Config, app *web.App) {
	labelsService := cfg.Service
	h := New(labelsService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	app.Get("/workspaces/{workspaceSlug}/labels", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/labels/{id}", h.Get, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/labels", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/labels/{id}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/labels/{id}", h.Delete, auth, workspace)
}
