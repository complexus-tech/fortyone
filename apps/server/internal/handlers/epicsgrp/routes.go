package epicsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/epics"
	"github.com/complexus-tech/projects-api/internal/repo/epicsrepo"
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

	epicsService := epics.New(cfg.Log, epicsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(epicsService)

	app.Get("/workspaces/{workspaceSlug}/epics", h.List, auth, workspace)

}
