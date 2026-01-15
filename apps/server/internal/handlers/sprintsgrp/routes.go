package sprintsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/internal/repo/sprintsrepo"
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

	sprintsService := sprints.New(cfg.Log, sprintsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(sprintsService)

	app.Get("/workspaces/{workspaceSlug}/sprints", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/running", h.Running, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.GetByID, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/sprints/{sprintId}/analytics", h.GetAnalytics, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/sprints", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/sprints/{sprintId}", h.Delete, auth, workspace)
}
