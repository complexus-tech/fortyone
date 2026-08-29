package activitieshttp

import (
	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	SecretKey         string
	StorageConfig     storage.Config
	StorageService    storage.StorageService
	Activities        *activities.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	activitiesService := cfg.Activities
	attachmentsService := cfg.Attachments
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	h := New(cfg.Log, activitiesService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/activities", h.GetActivities, auth, workspace)
}
