package objectivestatusgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus/objectivestatusrepo"
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

	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))

	h := New(objectivestatusService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/objective-statuses", h.List, auth)

}
