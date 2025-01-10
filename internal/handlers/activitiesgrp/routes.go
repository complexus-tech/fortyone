package activitiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/internal/core/activities/activitiesrepo"
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
	activitiesService := activities.New(cfg.Log, activitiesrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(cfg.Log, activitiesService)

	app.Get("/workspaces/{workspaceId}/activities", h.GetActivities, auth)
}
