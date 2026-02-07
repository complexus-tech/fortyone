package keyresultshttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	keyresultsrepository "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	okractivitiesrepository "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

// Config contains the required dependencies for key results routes
type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	SecretKey      string
	Cache          *cache.Service
	StorageConfig  storage.Config
	StorageService storage.StorageService
}

// Routes sets up all the key results routes
func Routes(cfg Config, app *web.App) {
	okrActivitiesService := okractivities.New(cfg.Log, okractivitiesrepository.New(cfg.Log, cfg.DB))
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepository.New(cfg.Log, cfg.DB), okrActivitiesService)
	attachmentsService := attachments.New(cfg.Log, attachmentsrepository.New(cfg.Log, cfg.DB), cfg.StorageService, cfg.StorageConfig)
	h := New(keyResultsService, okrActivitiesService, attachmentsService, cfg.Cache, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Put("/workspaces/{workspaceSlug}/key-results/{id}", h.Update, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/key-results/{id}", h.Delete, auth, workspace, memberAndAdmin)
	app.Post("/workspaces/{workspaceSlug}/key-results", h.Create, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/key-results", h.ListPaginated, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/key-results/{id}/activities", h.GetActivities, auth, workspace, memberAndAdmin)
}
