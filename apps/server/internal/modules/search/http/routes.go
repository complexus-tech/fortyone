package searchhttp

import (
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
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
	Service           *search.Service
}

func Routes(cfg Config, app *web.App) {
	searchService := cfg.Service
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	h := New(searchService)

	app.Get("/workspaces/{workspaceSlug}/search", h.Search, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/search/similar-stories", h.FindSimilarStories, auth, workspace)
}
