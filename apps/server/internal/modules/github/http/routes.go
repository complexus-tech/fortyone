package githubhttp

import (
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
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
	Service   *github.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/integrations/github", h.GetIntegration, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/github/install-session", h.CreateInstallSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/github/repositories/resync", h.ResyncRepositories, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/integrations/github/settings", h.GetWorkspaceSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/integrations/github/settings", h.UpdateWorkspaceSettings, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/github/issue-sync-links", h.CreateIssueSyncLink, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}", h.UpdateIssueSyncLink, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}", h.DeleteIssueSyncLink, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/settings/github", h.GetTeamSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/github", h.UpdateTeamSettings, auth, workspace)

	app.Get("/integrations/github/setup", h.HandleSetup)
	app.Post("/webhooks/github", h.HandleWebhook)
}
