package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
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
	Service   *slack.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/integrations/slack", h.GetIntegration, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/integrations/slack/logs", h.GetRequestLogs, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/install-session", h.CreateInstallSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/link-account", h.LinkAccount, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/integrations/slack", h.DisconnectWorkspace, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/channels/resync", h.ResyncChannels, auth, workspace)

	app.Get("/integrations/slack/setup", h.HandleSetup)
	app.Post("/integrations/slack/events", h.HandleEvents)
	app.Post("/integrations/slack/interactivity", h.HandleInteractivity)
	app.Post("/integrations/slack/commands", h.HandleCommands)
}
