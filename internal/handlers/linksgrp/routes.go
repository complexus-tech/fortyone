package linksgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/repo/linksrepo"
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
	linksService := links.New(cfg.Log, linksrepo.New(cfg.Log, cfg.DB))
	h := New(cfg.Log, linksService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Post("/workspaces/{workspaceId}/links", h.CreateLink, auth)
	app.Put("/workspaces/{workspaceId}/links/{id}", h.UpdateLink, auth)
	app.Delete("/workspaces/{workspaceId}/links/{id}", h.DeleteLink, auth)
}
