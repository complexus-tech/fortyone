package workautomationshttp

import (
	workautomations "github.com/complexus-tech/projects-api/internal/modules/workautomations/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *workautomations.Service
}

func Routes(cfg Config, app *web.App) {
	handler := &Handlers{service: cfg.Service}
	auth, workspace := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions), mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/team-automations", handler.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/team-automations", handler.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/team-automations/{id}", handler.Pause, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/team-automations/{id}", handler.Archive, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/team-automations/{id}/runs", handler.Runs, auth, workspace)
}
