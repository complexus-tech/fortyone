package objectivestatusgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/repo/objectivestatusrepo"
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
	objectiveStatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))
	h := New(objectiveStatusService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Get("/workspaces/{workspaceSlug}/objective-statuses", h.List, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/objective-statuses", h.Create, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/objective-statuses/{statusId}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/objective-statuses/{statusId}", h.Delete, auth, workspace, adminOnly)
}
