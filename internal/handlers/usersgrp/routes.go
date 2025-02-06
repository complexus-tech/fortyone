package usersgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/users/usersrepo"
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
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(usersService, cfg.SecretKey)

	// Public endpoints
	app.Post("/users/login", h.Login)

	// Protected endpoints
	app.Get("/workspaces/{workspaceId}/members", h.List, auth)
	app.Get("/users/me", h.GetProfile, auth)
	app.Put("/users/me", h.UpdateProfile, auth)
	app.Delete("/users/me", h.DeleteProfile, auth)
	app.Post("/users/me/workspace", h.SwitchWorkspace, auth)
}
