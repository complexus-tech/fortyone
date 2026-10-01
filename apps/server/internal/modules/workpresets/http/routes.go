package workpresetshttp

import (
	workpresets "github.com/complexus-tech/projects-api/internal/modules/workpresets/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *workpresets.Service
}

func Routes(cfg Config, app *web.App) {
	handler, err := New(cfg.Service, cfg.SecretKey)
	if err != nil {
		panic("work preset cursor configuration is invalid")
	}
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/work-presets", handler.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/work-presets", handler.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/work-presets/{id}", handler.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/work-presets/{id}", handler.Archive, auth, workspace)
}
