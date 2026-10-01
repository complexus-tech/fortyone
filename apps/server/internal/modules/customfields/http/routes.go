package customfieldshttp

import (
	customfields "github.com/complexus-tech/projects-api/internal/modules/customfields/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *customfields.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/custom-fields", h.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams/{teamId}/custom-fields", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/custom-fields/{fieldId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/teams/{teamId}/custom-fields/{fieldId}", h.Archive, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/custom-fields", h.Snapshot, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/custom-fields", h.Patch, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/analytics/custom-field-report", h.Report, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/custom-fields/story-values", h.Batch, auth, workspace)
}
