package commentshttp

import (
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
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
	mentionsRepo := mentionsrepository.New(cfg.Log, cfg.DB)
	commentsService := comments.New(cfg.Log, commentsrepository.New(cfg.Log, cfg.DB), mentionsRepo)
	h := New(cfg.Log, commentsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Put("/workspaces/{workspaceSlug}/comments/{id}", h.UpdateComment, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/comments/{id}", h.DeleteComment, auth, workspace)
}
