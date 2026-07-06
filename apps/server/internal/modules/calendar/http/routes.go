package calendarhttp

import (
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Cache     *cache.Service
	Service   *calendar.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/integrations/calendar", h.GetIntegration, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/calendar/google/connect-session", h.CreateConnectSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/calendar/{connectionId}/sync", h.SyncConnection, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/integrations/calendar/{connectionId}", h.RevokeConnection, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/calendar/schedule", h.GetSchedule, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/calendar/schedule-blocks", h.CreateScheduleBlock, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/calendar/schedule-blocks/{blockId}", h.UpdateScheduleBlock, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/calendar/schedule-blocks/{blockId}", h.DeleteScheduleBlock, auth, workspace)

	app.Get("/integrations/calendar/google/callback", h.HandleGoogleCallback)
}
