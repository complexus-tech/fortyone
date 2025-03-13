package notificationsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/repo/notificationsrepo"
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
	notificationsService := notifications.New(cfg.Log, notificationsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(notificationsService)

	// Notifications
	app.Get("/workspaces/{workspaceId}/notifications", h.GetUnread, auth)
	app.Put("/workspaces/{workspaceId}/notifications/{id}/read", h.MarkAsRead, auth)
	app.Put("/workspaces/{workspaceId}/notifications/read-all", h.MarkAllAsRead, auth)

	// Notification Preferences
	app.Get("/workspaces/{workspaceId}/notification-preferences", h.GetPreferences, auth)
	app.Put("/workspaces/{workspaceId}/notification-preferences/{type}", h.UpdatePreference, auth)
	app.Post("/workspaces/{workspaceId}/notification-preferences/default", h.CreateDefaultPreferences, auth)
}
