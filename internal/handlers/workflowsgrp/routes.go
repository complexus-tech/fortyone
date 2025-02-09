package workflowsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/workflows"
	"github.com/complexus-tech/projects-api/internal/core/workflows/workflowsrepo"
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
	workflowsService := workflows.New(cfg.Log, workflowsrepo.New(cfg.Log, cfg.DB))
	h := New(workflowsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/workflows", h.List, auth)
	app.Get("/workspaces/{workspaceId}/teams/{teamId}/workflows", h.ListByTeam, auth)
	app.Post("/workspaces/{workspaceId}/workflows", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/workflows/{workflowId}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/workflows/{workflowId}", h.Delete, auth)
}
