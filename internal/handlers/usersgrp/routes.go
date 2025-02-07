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

	app.Post("/users/login", h.Login)
	app.Get("/workspaces/{workspaceId}/members", h.List, auth)
	app.Get("/workspaces/{workspaceId}/profile", h.GetProfile, auth)
	app.Put("/workspaces/{workspaceId}/profile", h.UpdateProfile, auth)
	app.Delete("/workspaces/{workspaceId}/profile", h.DeleteProfile, auth)
	app.Post("/workspaces/{workspaceId}/switch", h.SwitchWorkspace, auth)
	app.Post("/workspaces/{workspaceId}/reset-password", h.ResetPassword, auth)
}
