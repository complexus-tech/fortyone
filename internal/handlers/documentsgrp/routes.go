package documentsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/complexus-tech/projects-api/internal/repo/documentsrepo"
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

	documentsService := documents.New(cfg.Log, documentsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(documentsService)

	app.Get("/workspaces/{workspaceSlug}/documents", h.List, auth, workspace)

}
