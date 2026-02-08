package activitieshttp

import (
	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
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
	Cache          *cache.Service
	SecretKey      string
	StorageConfig  storage.Config
	StorageService storage.StorageService
	Activities     *activities.Service
	Attachments    *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	activitiesService := cfg.Activities
	attachmentsService := cfg.Attachments
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(cfg.Log, activitiesService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/activities", h.GetActivities, auth, workspace)
}
