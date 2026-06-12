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
	BotToken  string
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service, cfg.BotToken)
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

	app.Post("/internal/bot/slack/options/teams", h.RuntimeSearchTeams, h.BotAuth)
	app.Post("/internal/bot/slack/options/statuses", h.RuntimeSearchStatuses, h.BotAuth)
	app.Post("/internal/bot/slack/options/members", h.RuntimeSearchMembers, h.BotAuth)
	app.Post("/internal/bot/slack/options/objectives", h.RuntimeSearchObjectives, h.BotAuth)
	app.Post("/internal/bot/slack/stories", h.RuntimeCreateStory, h.BotAuth)
	app.Post("/internal/bot/slack/thread-comments", h.RuntimeRecordThreadComment, h.BotAuth)
	app.Post("/internal/bot/slack/unfurls/story", h.RuntimeStoryUnfurl, h.BotAuth)
	app.Post("/internal/bot/slack/notifications/mentions", h.RuntimeMentionNotifications, h.BotAuth)
}
