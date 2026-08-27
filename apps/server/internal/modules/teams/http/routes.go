package teamshttp

import (
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
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
	Service   *teams.Service
}

func Routes(cfg Config, app *web.App) {
	teamsService := cfg.Service
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	h := New(teamsService, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/teams", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/{id}", h.GetByID, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/public", h.ListPublicTeams, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams", h.Create, auth, workspace, memberAndAdmin)
	app.Put("/workspaces/{workspaceSlug}/teams/{id}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/teams/{id}", h.Delete, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/teams/{id}/join", h.JoinPublicTeam, auth, workspace, gzip)
	app.Delete("/workspaces/{workspaceSlug}/teams/{id}/membership", h.LeaveTeam, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams/{id}/members", h.AddMember, auth, workspace, gzip, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/teams/{id}/members/{userId}/ai-context", h.UpdateMemberAIContext, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/teams/{id}/members/{userId}", h.RemoveMember, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/teams/order", h.UpdateTeamOrdering, auth, workspace)
}
