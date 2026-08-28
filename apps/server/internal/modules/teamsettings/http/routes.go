package teamsettingshttp

import (
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
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
	Service           *teamsettings.Service
}

func Routes(cfg Config, app *web.App) {
	teamsettingsService := cfg.Service
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	teamsRead := mid.RequireScopes(platformauth.ScopeTeamsRead)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	h := New(teamsettingsService)

	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/settings", h.GetSettings, auth, workspace, teamsRead, memberAndAdmin)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/sprints", h.UpdateSprintSettings, auth, workspace, teamsRead, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/story-automation", h.UpdateStoryAutomationSettings, auth, workspace, teamsRead, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/estimation", h.UpdateEstimationSettings, auth, workspace, teamsRead, adminOnly)
}
