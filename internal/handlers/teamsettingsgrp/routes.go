package teamsettingsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teamsettings"
	"github.com/complexus-tech/projects-api/internal/repo/teamsettingsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB           *sqlx.DB
	Log          *logger.Logger
	SecretKey    string
	TasksService *tasks.Service
	Cache        *cache.Service
}

func Routes(cfg Config, app *web.App) {
	teamsettingsService := teamsettings.New(cfg.Log, teamsettingsrepo.New(cfg.Log, cfg.DB), cfg.TasksService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(teamsettingsService)

	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/settings", h.GetSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/sprints", h.UpdateSprintSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/teams/{teamId}/settings/story-automation", h.UpdateStoryAutomationSettings, auth, workspace)
}
