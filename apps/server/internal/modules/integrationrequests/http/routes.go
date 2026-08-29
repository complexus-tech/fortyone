package integrationrequestshttp

import (
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
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
	Service           *integrationrequests.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	memberOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/integration-requests", h.ListTeamRequests, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams/{teamId}/integration-requests/accept-all", h.AcceptAllTeamRequests, auth, workspace, memberOnly)
	app.Post("/workspaces/{workspaceSlug}/teams/{teamId}/integration-requests/decline-all", h.DeclineAllTeamRequests, auth, workspace, memberOnly)
	app.Get("/workspaces/{workspaceSlug}/integration-requests/{requestId}", h.GetRequest, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/integration-requests/{requestId}/thread", h.GetRequestThreadActivity, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/integration-requests/{requestId}", h.UpdateRequest, auth, workspace, memberOnly)
	app.Post("/workspaces/{workspaceSlug}/integration-requests/{requestId}/comments", h.CreateRequestComment, auth, workspace, memberOnly)
	app.Post("/workspaces/{workspaceSlug}/integration-requests/{requestId}/accept", h.AcceptRequest, auth, workspace, memberOnly)
	app.Post("/workspaces/{workspaceSlug}/integration-requests/{requestId}/decline", h.DeclineRequest, auth, workspace, memberOnly)
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/integration-request-links", h.GetStoryProviderThreads, auth, workspace)
}
