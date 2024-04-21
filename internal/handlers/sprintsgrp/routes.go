package sprintsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/internal/core/sprints/sprintsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	sprintsService := sprints.New(cfg.Log, sprintsrepo.New(cfg.Log, cfg.DB))

	h := New(sprintsService)

	app.Get("/sprints", h.List)

}
