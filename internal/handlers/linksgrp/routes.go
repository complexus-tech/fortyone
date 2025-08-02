package linksgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/repo/linksrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
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
	linksService := links.New(cfg.Log, linksrepo.New(cfg.Log, cfg.DB))
	h := New(cfg.Log, linksService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Post("/workspaces/{workspaceSlug}/links", h.CreateLink, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/links/{id}", h.UpdateLink, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/links/{id}", h.DeleteLink, auth, workspace)
}
