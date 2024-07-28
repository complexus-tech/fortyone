package statesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/states/statesrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	statesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))

	h := New(statesService)

	app.Get("/workspaces/{workspaceId}/states", h.List)

}
