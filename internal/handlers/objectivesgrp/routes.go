package objectivesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/objectives/objectivesrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {
	objectivesService := objectives.New(cfg.Log, objectivesrepo.New(cfg.Log, cfg.DB))

	h := New(objectivesService)

	app.Get("/objectives", h.List)

}
