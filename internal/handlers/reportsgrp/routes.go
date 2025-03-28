package reportsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/reports"
	"github.com/complexus-tech/projects-api/internal/repo/reportsrepo"
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
	reportsService := reports.New(cfg.Log, reportsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(cfg.Log, reportsService)

	app.Get("/workspaces/{workspaceId}/analytics/summary", h.GetStoryStats, auth)
	app.Get("/workspaces/{workspaceId}/analytics/contributions", h.GetContributionStats, auth)
	app.Get("/workspaces/{workspaceId}/analytics/users", h.GetUserStats, auth)
	app.Get("/workspaces/{workspaceId}/analytics/status", h.GetStatusStats, auth)
	app.Get("/workspaces/{workspaceId}/analytics/priority", h.GetPriorityStats, auth)
}
