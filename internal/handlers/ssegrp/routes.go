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

	// Define the authentication middleware instance.
	// This uses the mid.Auth function from your existing middleware package.
	authMiddleware := mid.Auth(cfg.Log, cfg.SecretKey)

	// Register the SSE endpoint. Choose a suitable path, e.g., "/api/v1/notifications/subscribe".
	// The handler.StreamNotifications method will handle the SSE logic.
	app.Get("/notifications/subscribe", handler.StreamNotifications, authMiddleware)
}
