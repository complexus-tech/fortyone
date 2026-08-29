package ssehttp

import (
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	SSEHub            *sse.Hub
	Origins           web.OriginPolicy
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
}

// Routes wires up the SSE routes.
func Routes(cfg Config, app *web.App) {
	handler := Handler{
		Log:             cfg.Log,
		SSEHub:          cfg.SSEHub,
		Origins:         cfg.Origins,
		BrowserSessions: cfg.BrowserSessions,
		WorkspaceAccess: cfg.WorkspaceResolver,
	}

	authMiddleware := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	app.Get("/workspaces/{workspaceSlug}/notifications/subscribe", handler.StreamNotifications, authMiddleware, workspace)
}
