package usersgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB            *sqlx.DB
	Log           *logger.Logger
	SecretKey     string
	GoogleService *google.Service
	Publisher     *publisher.Publisher
}

func Routes(cfg Config, app *web.App) {
	usersRepo := usersrepo.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersRepo)
	h := New(usersService, cfg.SecretKey, cfg.GoogleService, cfg.Publisher)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	// Public endpoints
	app.Post("/users/google/verify", h.GoogleAuth)
	app.Post("/users/verify/email", h.SendEmailVerification)
	app.Post("/users/verify/email/confirm", h.VerifyEmail)

	// Protected endpoints
	app.Get("/workspaces/{workspaceId}/members", h.List, auth, gzip)
	app.Get("/users/profile", h.GetProfile, auth)
	app.Put("/users/profile", h.UpdateProfile, auth)
	app.Delete("/users/profile", h.DeleteProfile, auth)
	app.Post("/workspaces/switch", h.SwitchWorkspace, auth)

	// Automation preferences endpoints
	app.Get("/workspaces/{workspaceId}/automation/preferences", h.GetAutomationPreferences, auth)
	app.Put("/workspaces/{workspaceId}/automation/preferences", h.UpdateAutomationPreferences, auth)
}
