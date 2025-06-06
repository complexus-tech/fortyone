package teamsettingsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teamsettings"
	"github.com/complexus-tech/projects-api/internal/repo/teamsettingsrepo"
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
	teamsettingsService := teamsettings.New(cfg.Log, teamsettingsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(teamsettingsService)

	app.Get("/workspaces/{workspaceId}/teams/{teamId}/settings", h.GetSettings, auth)
	app.Put("/workspaces/{workspaceId}/teams/{teamId}/settings/sprints", h.UpdateSprintSettings, auth)
	app.Put("/workspaces/{workspaceId}/teams/{teamId}/settings/story-automation", h.UpdateStoryAutomationSettings, auth)
}
