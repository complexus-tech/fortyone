package sprintshttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	sprintsrepository "github.com/complexus-tech/projects-api/internal/modules/sprints/repository"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	SecretKey      string
	Cache          *cache.Service
	StorageConfig  storage.Config
	StorageService storage.StorageService
}

func Routes(cfg Config, app *web.App) {

	sprintsService := sprints.New(cfg.Log, sprintsrepository.New(cfg.Log, cfg.DB))
	attachmentsService := attachments.New(cfg.Log, attachmentsrepository.New(cfg.Log, cfg.DB), cfg.StorageService, cfg.StorageConfig)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(sprintsService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/sprints", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/running", h.Running, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.GetByID, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}/analytics", h.GetAnalytics, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/sprints", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Delete, auth, workspace)
}
