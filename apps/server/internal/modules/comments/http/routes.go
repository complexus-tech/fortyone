package commentshttp

import (
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
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
	Service   *comments.Service
}

func Routes(cfg Config, app *web.App) {
	commentsService := cfg.Service
	h := New(cfg.Log, commentsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Put("/workspaces/{workspaceSlug}/comments/{id}", h.UpdateComment, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/comments/{id}", h.DeleteComment, auth, workspace)
}
