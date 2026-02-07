package teamshttp

import (
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
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
}

func Routes(cfg Config, app *web.App) {
	teamsService := teams.New(cfg.Log, teamsrepository.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(teamsService, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/teams", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/{id}", h.GetByID, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/public", h.ListPublicTeams, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/{id}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/teams/{id}", h.Delete, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/teams/{id}/members", h.AddMember, auth, workspace, gzip)
	app.Delete("/workspaces/{workspaceSlug}/teams/{id}/members/{userId}", h.RemoveMember, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/order", h.UpdateTeamOrdering, auth, workspace)
}
