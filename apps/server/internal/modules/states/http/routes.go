package stateshttp

import (
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
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
	Service   *states.Service
}

func Routes(cfg Config, app *web.App) {
	statesService := cfg.Service
	h := New(statesService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/states", h.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/states", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/states/{stateId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/states/{stateId}", h.Delete, auth, workspace)
}
