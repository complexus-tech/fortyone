package mayahttp

import (
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
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
	Service   *maya.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.DB, cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	admin := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Post("/workspaces/{workspaceSlug}/maya/work-plans", h.CreateWorkPlan, auth, workspace, admin)
}
