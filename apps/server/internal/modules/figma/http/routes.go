package figmahttp

import (
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
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
	Service   *figma.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	admin := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Get("/workspaces/{workspaceSlug}/integrations/figma", h.GetIntegration, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/integrations/figma/handoff-statuses", h.ListStoryHandoffStatuses, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/figma/install-session", h.CreateInstallSession, auth, workspace, admin)
	app.Post("/workspaces/{workspaceSlug}/integrations/figma/resolve-link", h.ResolveLink, auth, workspace, member)
	app.Delete("/workspaces/{workspaceSlug}/integrations/figma", h.Disconnect, auth, workspace, admin)
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/figma-links", h.ListStoryLinks, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/{storyId}/figma-links", h.LinkStory, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/stories/{storyId}/figma-links/{linkId}/refresh", h.RefreshStoryLink, auth, workspace, member)
	app.Delete("/workspaces/{workspaceSlug}/stories/{storyId}/figma-links/{linkId}", h.DeleteStoryLink, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/integrations/figma/stories", h.CreateStoryFromLink, auth, workspace, member)

	app.Get("/integrations/figma/callback", h.CompleteOAuth)
	app.Post("/webhooks/figma", h.HandleWebhook)
}
