package labelsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/labels"
	"github.com/complexus-tech/projects-api/internal/core/labels/labelsrepo"
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
	labelsService := labels.New(cfg.Log, labelsrepo.New(cfg.Log, cfg.DB))
	h := New(labelsService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/labels", h.List, auth)
	app.Get("/workspaces/{workspaceId}/labels/{id}", h.Get, auth)
	app.Post("/workspaces/{workspaceId}/labels", h.Create, auth)
}
