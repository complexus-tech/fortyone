package objectiveshttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	keyresultsrepository "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	okractivitiesrepository "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
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
	okrActivitiesService := okractivities.New(cfg.Log, okractivitiesrepository.New(cfg.Log, cfg.DB))
	objectivesService := objectives.New(cfg.Log, objectivesrepository.New(cfg.Log, cfg.DB), okrActivitiesService)
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepository.New(cfg.Log, cfg.DB), okrActivitiesService)
	attachmentsService := attachments.New(cfg.Log, attachmentsrepository.New(cfg.Log, cfg.DB), cfg.StorageService, cfg.StorageConfig)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	h := New(objectivesService, keyResultsService, okrActivitiesService, attachmentsService, cfg.Cache, cfg.Log)

	app.Get("/workspaces/{workspaceSlug}/objectives", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}", h.Get, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/objectives/{id}", h.Update, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/objectives/{id}", h.Delete, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/key-results", h.GetKeyResults, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/analytics", h.GetAnalytics, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/activities", h.GetActivities, auth, workspace, memberAndAdmin)
	app.Post("/workspaces/{workspaceSlug}/objectives", h.Create, auth, workspace, memberAndAdmin)
}
