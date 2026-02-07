package labelshttp

import (
	labelsrepository "github.com/complexus-tech/projects-api/internal/modules/labels/repository"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
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
	labelsService := labels.New(cfg.Log, labelsrepository.New(cfg.Log, cfg.DB))
	h := New(labelsService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Get("/workspaces/{workspaceSlug}/labels", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/labels/{id}", h.Get, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/labels", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/labels/{id}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/labels/{id}", h.Delete, auth, workspace)
}
