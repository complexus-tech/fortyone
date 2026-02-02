package reportsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/reports"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/reportsrepo"
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
	SecretKey      string
	Cache          *cache.Service
	StorageConfig  storage.Config
	StorageService storage.StorageService
}

func Routes(cfg Config, app *web.App) {
	reportsService := reports.New(cfg.Log, reportsrepo.New(cfg.Log, cfg.DB))
	attachmentsService := attachments.New(cfg.Log, attachmentsrepo.New(cfg.Log, cfg.DB), cfg.StorageService, cfg.StorageConfig)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(cfg.Log, reportsService, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}/analytics/summary", h.GetStoryStats, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/contributions", h.GetContributionStats, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/users", h.GetUserStats, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/status", h.GetStatusStats, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/priority", h.GetPriorityStats, auth, workspace)

	// Workspace Analytics
	app.Get("/workspaces/{workspaceSlug}/analytics/overview", h.GetWorkspaceOverview, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/story-analytics", h.GetStoryAnalytics, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/objective-progress", h.GetObjectiveProgress, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/team-performance", h.GetTeamPerformance, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/sprint-analytics", h.GetSprintAnalytics, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/analytics/timeline-trends", h.GetTimelineTrends, auth, workspace)
}
