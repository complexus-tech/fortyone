package objectivesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivesrepo"
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
	objectivesService := objectives.New(cfg.Log, objectivesrepo.New(cfg.Log, cfg.DB))
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(objectivesService, keyResultsService, cfg.Cache, cfg.Log)

	app.Get("/workspaces/{workspaceId}/objectives", h.List, auth)
	app.Get("/workspaces/{workspaceId}/objectives/{id}", h.Get, auth)
	app.Put("/workspaces/{workspaceId}/objectives/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/objectives/{id}", h.Delete, auth)
	app.Get("/workspaces/{workspaceId}/objectives/{id}/key-results", h.GetKeyResults, auth)
	app.Get("/workspaces/{workspaceId}/objectives/{id}/analytics", h.GetAnalytics, auth)
	app.Post("/workspaces/{workspaceId}/objectives", h.Create, auth)
}
