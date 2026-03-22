package githubhttp

import (
	"context"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserLookup provides minimal user lookup for comment authoring.
type UserLookup interface {
	GetUserName(ctx context.Context, userID uuid.UUID) (string, error)
}

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Cache     *cache.Service
	Service   *github.Service
	Users     UserLookup
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Users)
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
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/github-links", h.GetStoryGitHubLinks, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/{storyId}/github-links/{linkId}", h.DeleteStoryGitHubLink, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/github-comments", h.GetStoryGitHubComments, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/{storyId}/github-comments", h.PostStoryGitHubComment, auth, workspace)

	app.Post("/user/integrations/github/link", h.LinkGitHubUser, auth)
	app.Delete("/user/integrations/github/link", h.UnlinkGitHubUser, auth)

	app.Get("/integrations/github/setup", h.HandleSetup)
	app.Post("/webhooks/github", h.HandleWebhook)
}
