package teamsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/teams/teamsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))

	h := New(teamsService)

	app.Get("/teams", h.List)

}
