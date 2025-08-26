package objectivesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/okractivitiesrepo"
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
	okrActivitiesService := okractivities.New(cfg.Log, okractivitiesrepo.New(cfg.Log, cfg.DB))
	objectivesService := objectives.New(cfg.Log, objectivesrepo.New(cfg.Log, cfg.DB))
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepo.New(cfg.Log, cfg.DB), okrActivitiesService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	h := New(objectivesService, keyResultsService, okrActivitiesService, cfg.Cache, cfg.Log)

	app.Get("/workspaces/{workspaceSlug}/objectives", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}", h.Get, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/objectives/{id}", h.Update, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/objectives/{id}", h.Delete, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/key-results", h.GetKeyResults, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/analytics", h.GetAnalytics, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/objectives/{id}/activities", h.GetActivities, auth, workspace, memberAndAdmin)
	app.Post("/workspaces/{workspaceSlug}/objectives", h.Create, auth, workspace, memberAndAdmin)
}
