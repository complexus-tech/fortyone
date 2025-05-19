package notificationsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/repo/notificationsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	Redis     *redis.Client
	SecretKey string
}

func Routes(cfg Config, app *web.App) {
	notificationsService := notifications.New(cfg.Log, notificationsrepo.New(cfg.Log, cfg.DB), cfg.Redis)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(notificationsService)

	// Notifications
	app.Get("/workspaces/{workspaceId}/notifications", h.List, auth)
	app.Put("/workspaces/{workspaceId}/notifications/{id}/read", h.MarkAsRead, auth)
	app.Put("/workspaces/{workspaceId}/notifications/{id}/unread", h.MarkAsUnread, auth)
	app.Put("/workspaces/{workspaceId}/notifications/read-all", h.MarkAllAsRead, auth)
	app.Get("/workspaces/{workspaceId}/notifications/unread-count", h.GetUnreadCount, auth)
	app.Delete("/workspaces/{workspaceId}/notifications/{id}", h.DeleteNotification, auth)
	app.Delete("/workspaces/{workspaceId}/notifications", h.DeleteAllNotifications, auth)
	app.Delete("/workspaces/{workspaceId}/notifications/read", h.DeleteReadNotifications, auth)

	// Notification Preferences
	app.Get("/workspaces/{workspaceId}/notification-preferences", h.GetPreferences, auth)
	app.Put("/workspaces/{workspaceId}/notification-preferences/{type}", h.UpdatePreference, auth)
}
