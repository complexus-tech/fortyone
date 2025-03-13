package statesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
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
	statesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))
	h := New(statesService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/states", h.List, auth)
	app.Post("/workspaces/{workspaceId}/states", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/states/{stateId}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/states/{stateId}", h.Delete, auth)
}
