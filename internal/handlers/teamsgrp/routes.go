package teamsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/teams/teamsrepo"
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

	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	t := New(teamsService)
	app.Get("/workspaces/{workspaceId}/teams", t.List, auth)

}
