package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
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
	Service           *slack.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	admin := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Get("/workspaces/{workspaceSlug}/integrations/slack", h.GetIntegration, auth, workspace, member)
	app.Get("/workspaces/{workspaceSlug}/integrations/slack/logs", h.GetRequestLogs, auth, workspace, admin)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/install-session", h.CreateInstallSession, auth, workspace, admin)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/account-link-session", h.CreateAccountLinkSession, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/link-account", h.LinkAccount, auth, workspace, member)
	app.Delete("/workspaces/{workspaceSlug}/integrations/slack/account-link", h.DisconnectAccount, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/integrations/slack/channels/resync", h.ResyncChannels, auth, workspace, admin)
	app.Get("/workspaces/{workspaceSlug}/integrations/slack/channel-audiences", h.ListChannelAudiences, auth, workspace, admin)
	app.Put("/workspaces/{workspaceSlug}/integrations/slack/channel-audiences/{channelId}", h.UpdateChannelAudience, auth, workspace, admin)
	app.Get("/workspaces/{workspaceSlug}/integrations/slack/agent-settings", h.GetAgentSettings, auth, workspace, admin)
	app.Put("/workspaces/{workspaceSlug}/integrations/slack/agent-settings", h.UpdateAgentSettings, auth, workspace, admin)
	app.Delete("/workspaces/{workspaceSlug}/integrations/slack", h.DisconnectWorkspace, auth, workspace, admin)

	app.Get("/integrations/slack/setup", h.HandleSetup)
	app.Post("/integrations/slack/events", h.HandleEvents)
	app.Post("/integrations/slack/interactivity", h.HandleInteractivity)
	app.Post("/integrations/slack/commands", h.HandleCommands)
}
