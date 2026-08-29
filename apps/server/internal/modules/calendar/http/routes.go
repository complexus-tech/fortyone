package calendarhttp

import (
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *calendar.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	app.Get("/workspaces/{workspaceSlug}/integrations/calendar", h.GetIntegration, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/calendar/{provider}/connect-session", h.CreateConnectSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/calendar/{connectionId}/sync", h.SyncConnection, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/integrations/calendar/{connectionId}/primary", h.SetPrimaryConnection, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/integrations/calendar/{connectionId}", h.RevokeConnection, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/calendar/schedule", h.GetSchedule, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/calendar/events/{eventId}", h.GetCalendarEvent, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/calendar/schedule-blocks", h.CreateScheduleBlock, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/calendar/schedule-blocks/{blockId}", h.UpdateScheduleBlock, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/calendar/schedule-blocks/{blockId}/manual-reschedule", h.ManualRescheduleScheduleBlock, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/calendar/schedule-blocks/{blockId}", h.DeleteScheduleBlock, auth, workspace)

	app.Get("/integrations/calendar/google/callback", h.HandleGoogleCallback, auth)
	app.Get("/integrations/calendar/microsoft/callback", h.HandleGoogleCallback, auth)
	app.Post("/webhooks/google/calendar", h.HandleGoogleNotification)
	app.Post("/webhooks/microsoft/calendar", h.HandleMicrosoftNotification)
}
