package epicsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/epics"
	"github.com/complexus-tech/projects-api/internal/core/epics/epicsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	epicsService := epics.New(cfg.Log, epicsrepo.New(cfg.Log, cfg.DB))

	h := New(epicsService)

	app.Get("/epics", h.List)

}
