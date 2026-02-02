package activitiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/repo/activitiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	Cache          *cache.Service
	SecretKey      string
	StorageConfig  storage.Config
	StorageService storage.StorageService
}

func Routes(cfg Config, app *web.App) {
	activitiesService := activities.New(cfg.Log, activitiesrepo.New(cfg.Log, cfg.DB))
	attachmentsService := attachments.New(cfg.Log, attachmentsrepo.New(cfg.Log, cfg.DB), cfg.StorageService, cfg.StorageConfig)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(cfg.Log, activitiesService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/activities", h.GetActivities, auth, workspace)
}
