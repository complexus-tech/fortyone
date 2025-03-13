package keyresultsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

// Config contains the required dependencies for key results routes
type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
}

// Routes sets up all the key results routes
func Routes(cfg Config, app *web.App) {
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepo.New(cfg.Log, cfg.DB))
	h := New(keyResultsService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Put("/workspaces/{workspaceId}/key-results/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/key-results/{id}", h.Delete, auth)
	app.Post("/workspaces/{workspaceId}/key-results", h.Create, auth)
}
