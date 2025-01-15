package workspacesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/core/workspaces/workspacesrepo"
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

	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(workspacesService, cfg.SecretKey)

	app.Get("/workspaces", h.List, auth)
	app.Post("/workspaces", h.Create, auth)
	app.Patch("/workspaces/{id}", h.Update, auth)
	app.Delete("/workspaces/{id}", h.Delete, auth)
	app.Post("/workspaces/{id}/members", h.AddMember, auth)

}
