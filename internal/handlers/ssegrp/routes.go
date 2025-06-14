package ssegrp

import (
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log        *logger.Logger
	SecretKey  string
	SSEHub     *sse.Hub
	CorsOrigin string
}

// Routes wires up the SSE routes.
func Routes(cfg Config, app *web.App) {
	handler := Handler{
		Log:        cfg.Log,
		SSEHub:     cfg.SSEHub,
		CorsOrigin: cfg.CorsOrigin,
	}

	authMiddleware := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/notifications/subscribe", handler.StreamNotifications, authMiddleware)
}
