package healthgrp

import (
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	h := New(cfg.Log, cfg.DB)

	app.Get("/readiness", h.Readiness)
	app.Get("/liveness", h.Liveness)

}
