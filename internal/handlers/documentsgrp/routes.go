package documentsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/complexus-tech/projects-api/internal/repo/documentsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {

	documentsService := documents.New(cfg.Log, documentsrepo.New(cfg.Log, cfg.DB))

	h := New(documentsService)

	app.Get("/workspaces/{workspaceId}/documents", h.List)

}
