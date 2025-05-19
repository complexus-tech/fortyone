package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the notifications storage.
type Repository interface {
	Create(ctx context.Context, n CoreNewNotification) (CoreNotification, error)
	List(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]CoreNotification, error)
	GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) error
	GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreNotificationPreferences, error)
	UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error
	DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error
	DeleteAllNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error)
	DeleteReadNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error)
	MarkAsUnread(ctx context.Context, notificationID, userID uuid.UUID) error
}

// Service provides notification-related operations.
type Service struct {
	repo        Repository
	log         *logger.Logger
	redisClient *redis.Client
}

// New constructs a new notifications service instance with the provided repository and Redis client.
func New(log *logger.Logger, repo Repository, redisClient *redis.Client) *Service {
	return &Service{
		repo:        repo,
		log:         log,
		redisClient: redisClient,
	}
}

// Create creates a new notification, saves it to the database, and publishes it to Redis for SSE.
func (s *Service) Create(ctx context.Context, n CoreNewNotification) (CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.Create")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.Create")
	defer span.End()

	notification, err := s.repo.Create(ctx, n)
	if err != nil {
		span.RecordError(err)
		return CoreNotification{}, err
	}

	span.AddEvent("notification created in DB", trace.WithAttributes(
		attribute.String("notification.id", notification.ID.String()),
	))

	// Publish notification to Redis Pub/Sub
	if s.redisClient != nil {
		jsonData, marshalErr := json.Marshal(notification)
		if marshalErr != nil {
			s.log.Error(ctx, "notifications.Service.Create: failed to marshal notification for Redis", "error", marshalErr, "notificationID", notification.ID)
		} else {
			channelName := fmt.Sprintf("user-notifications:%s", notification.RecipientID.String())
			if pubErr := s.redisClient.Publish(ctx, channelName, jsonData).Err(); pubErr != nil {
				s.log.Error(ctx, "notifications.Service.Create: failed to publish notification to Redis", "error", pubErr, "channel", channelName, "notificationID", notification.ID)
			} else {
				s.log.Info(ctx, "notifications.Service.Create: notification published to Redis", "channel", channelName, "notificationID", notification.ID)
				span.AddEvent("notification published to Redis", trace.WithAttributes(
					attribute.String("redis.channel", channelName),
					attribute.String("notification.id", notification.ID.String()),
				))
			}
		}
	} else {
		s.log.Warn(ctx, "notifications.Service.Create: Redis client not configured, skipping publish.", "notificationID", notification.ID)
		span.AddEvent("redis publish skipped", trace.WithAttributes(
			attribute.String("reason", "redis client not configured"),
			attribute.String("notification.id", notification.ID.String()),
		))
	}

	return notification, nil
}

// List returns a list of notifications.
func (s *Service) List(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.List")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.List")
	defer span.End()

	notifications, err := s.repo.List(ctx, userID, workspaceID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("notifications retrieved", trace.WithAttributes(
		attribute.Int("notifications.count", len(notifications)),
	))

	return notifications, nil
}

// GetUnreadCount returns the number of unread notifications for a user in a workspace.
func (s *Service) GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (int, error) {
	s.log.Info(ctx, "business.core.notifications.GetUnreadCount")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.GetUnreadCount")
	defer span.End()

	count, err := s.repo.GetUnreadCount(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	span.AddEvent("unread count retrieved", trace.WithAttributes(
		attribute.Int("unread_count", count),
	))

	return count, nil
}

// MarkAsRead marks a notification as read.
func (s *Service) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.notifications.MarkAsRead")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.MarkAsRead")
	defer span.End()

	if err := s.repo.MarkAsRead(ctx, notificationID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("notification marked as read", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return nil
}

// MarkAllAsRead marks all notifications as read for a user in a workspace.
func (s *Service) MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.notifications.MarkAllAsRead")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.MarkAllAsRead")
	defer span.End()

	if err := s.repo.MarkAllAsRead(ctx, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("all notifications marked as read", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

// GetPreferences returns a user's notification preferences for a workspace.
func (s *Service) GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreNotificationPreferences, error) {
	s.log.Info(ctx, "business.core.notifications.GetPreferences")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.GetPreferences")
	defer span.End()

	preferences, err := s.repo.GetPreferences(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreNotificationPreferences{}, err
	}

	span.AddEvent("preferences retrieved", trace.WithAttributes(
		attribute.Int("preferences.count", len(preferences.Preferences)),
	))

	return preferences, nil
}

// UpdatePreference updates a notification preference.
func (s *Service) UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error {
	s.log.Info(ctx, "business.core.notifications.UpdatePreference")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.UpdatePreference")
	defer span.End()

	if err := s.repo.UpdatePreference(ctx, userID, workspaceID, notificationType, updates); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("preference updated", trace.WithAttributes(
		attribute.String("notification.type", notificationType),
	))

	return nil
}

// DeleteNotification deletes a notification.
func (s *Service) DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.notifications.DeleteNotification")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.DeleteNotification")
	defer span.End()

	if err := s.repo.DeleteNotification(ctx, notificationID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("notification deleted", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return nil
}

// DeleteAllNotifications deletes all notifications for a user in a workspace.
func (s *Service) DeleteAllNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error) {
	s.log.Info(ctx, "business.core.notifications.DeleteAllNotifications")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.DeleteAllNotifications")
	defer span.End()

	count, err := s.repo.DeleteAllNotifications(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	span.AddEvent("all notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", count),
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return count, nil
}

// DeleteReadNotifications deletes all read notifications for a user in a workspace.
func (s *Service) DeleteReadNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error) {
	s.log.Info(ctx, "business.core.notifications.DeleteReadNotifications")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.DeleteReadNotifications")
	defer span.End()

	count, err := s.repo.DeleteReadNotifications(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	span.AddEvent("read notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", count),
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return count, nil
}

// MarkAsUnread marks a notification as unread.
func (s *Service) MarkAsUnread(ctx context.Context, notificationID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.notifications.MarkAsUnread")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.MarkAsUnread")
	defer span.End()

	if err := s.repo.MarkAsUnread(ctx, notificationID, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("notification marked as unread", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return nil
}
