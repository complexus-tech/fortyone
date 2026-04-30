package integrationrequestshttp

import (
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
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
	Service   *integrationrequests.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/integration-requests", h.ListTeamRequests, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/integration-requests/{requestId}", h.GetRequest, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/integration-requests/{requestId}", h.UpdateRequest, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integration-requests/{requestId}/accept", h.AcceptRequest, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integration-requests/{requestId}/decline", h.DeclineRequest, auth, workspace)
}
