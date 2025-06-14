package sprintsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/internal/repo/sprintsrepo"
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

	sprintsService := sprints.New(cfg.Log, sprintsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(sprintsService)

	app.Get("/workspaces/{workspaceId}/sprints", h.List, auth)
	app.Get("/workspaces/{workspaceId}/sprints/running", h.Running, auth)
	app.Get("/workspaces/{workspaceId}/sprints/{sprintId}", h.GetByID, auth)
	app.Get("/workspaces/{workspaceId}/sprints/{sprintId}/analytics", h.GetAnalytics, auth)
	app.Post("/workspaces/{workspaceId}/sprints", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/sprints/{sprintId}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/sprints/{sprintId}", h.Delete, auth)
}
