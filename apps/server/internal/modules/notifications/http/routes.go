package notificationshttp

import (
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	DB           *sqlx.DB
	Log          *logger.Logger
	Redis        *redis.Client
	SecretKey    string
	TasksService *tasks.Service
	Cache        *cache.Service
}

func Routes(cfg Config, app *web.App) {
	notificationsService := notifications.New(cfg.Log, notificationsrepository.New(cfg.Log, cfg.DB), cfg.Redis, cfg.TasksService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(notificationsService)

	// Notifications
	app.Get("/workspaces/{workspaceSlug}/notifications", h.List, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/notifications/{id}/read", h.MarkAsRead, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/notifications/{id}/unread", h.MarkAsUnread, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/notifications/read-all", h.MarkAllAsRead, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/notifications/unread-count", h.GetUnreadCount, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/notifications/{id}", h.DeleteNotification, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/notifications", h.DeleteAllNotifications, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/notifications/read", h.DeleteReadNotifications, auth, workspace)

	// Notification Preferences
	app.Get("/workspaces/{workspaceSlug}/notification-preferences", h.GetPreferences, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/notification-preferences/{type}", h.UpdatePreference, auth, workspace)
}
