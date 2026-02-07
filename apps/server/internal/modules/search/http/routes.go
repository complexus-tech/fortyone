package searchhttp

import (
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
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
	searchRepo := searchrepository.New(cfg.Log, cfg.DB)
	searchService := search.New(cfg.Log, searchRepo)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(searchService)

	app.Get("/workspaces/{workspaceSlug}/search", h.Search, auth, workspace)
}
