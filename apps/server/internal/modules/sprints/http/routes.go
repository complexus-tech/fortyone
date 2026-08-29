package sprintshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	StorageConfig     storage.Config
	StorageService    storage.StorageService
	Sprints           *sprints.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	sprintsService := cfg.Sprints
	attachmentsService := cfg.Attachments
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	h := New(sprintsService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/sprints", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/running", h.Running, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.GetByID, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}/analytics", h.GetAnalytics, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/sprints", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Delete, auth, workspace)
}
