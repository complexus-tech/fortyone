package searchgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/search"
	"github.com/complexus-tech/projects-api/internal/repo/searchrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
}

func Routes(cfg Config, app *web.App) {
	searchRepo := searchrepo.New(cfg.Log, cfg.DB)
	searchService := search.New(cfg.Log, searchRepo)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(searchService)

	app.Get("/workspaces/{workspaceId}/search", h.Search, auth)
}
