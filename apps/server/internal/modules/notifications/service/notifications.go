package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the notifications storage.
type Repository interface {
	Create(ctx context.Context, n CoreNewNotification) (CoreNotification, bool, error)
	List(ctx context.Context, userID, workspaceID uuid.UUID, search string, limit, offset int) ([]CoreNotification, error)
	GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) error
	GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreNotificationPreferences, error)
	UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error
	DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error
	DeleteAllNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error)
	DeleteReadNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error)
	MarkAsUnread(ctx context.Context, notificationID, userID uuid.UUID) error
	ListPortalFeedback(ctx context.Context, userID uuid.UUID, portalSlug string, limit, offset int) ([]CorePortalNotification, error)
	GetPortalFeedbackUnreadCount(ctx context.Context, userID uuid.UUID, portalSlug string) (int, error)
	MarkPortalFeedbackAsRead(ctx context.Context, notificationID, userID uuid.UUID, portalSlug string) error
}

var ErrNotificationNotFound = errors.New("notification not found")

// TasksService provides access to the task queue for background processing.
type TasksService interface {
	EnqueueNotificationEmailDigest(payload tasks.NotificationEmailDigestPayload, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// Service provides notification-related operations.
type Service struct {
	repo            Repository
	log             *logger.Logger
	redisClient     *redis.Client
	tasksService    TasksService
	publishRealtime func(context.Context, CoreNotification) error
}

// New constructs a new notifications service instance with the provided repository, Redis client, and tasks service.
func New(log *logger.Logger, repo Repository, redisClient *redis.Client, tasksService TasksService) *Service {
	service := &Service{
		repo:         repo,
		log:          log,
		redisClient:  redisClient,
		tasksService: tasksService,
	}
	service.publishRealtime = service.publishNotification
	return service
}

func (s *Service) publishNotification(ctx context.Context, notification CoreNotification) error {
	ctx, span := web.AddSpan(ctx, "business.core.notifications.publishNotification")
	defer span.End()

	if s.redisClient == nil {
		s.log.Warn(ctx, "notifications.Service.Create: Redis client not configured, skipping publish.", "notificationID", notification.ID)
		span.AddEvent("redis publish skipped", trace.WithAttributes(
			attribute.String("reason", "redis client not configured"),
			attribute.String("notification.id", notification.ID.String()),
		))
		return nil
	}
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

	return nil
}

// Create creates a new notification, saves it to the database, publishes it to Redis for SSE, and enqueues an email task.
func (s *Service) Create(ctx context.Context, n CoreNewNotification) (CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.Create")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.Create")
	defer span.End()

	notification, inserted, err := s.repo.Create(ctx, n)
	if err != nil {
		span.RecordError(err)
		return CoreNotification{}, err
	}
	if !inserted {
		span.AddEvent("duplicate notification replay skipped", trace.WithAttributes(
			attribute.String("notification.id", notification.ID.String()),
		))
		return notification, nil
	}

	span.AddEvent("notification created in DB", trace.WithAttributes(
		attribute.String("notification.id", notification.ID.String()),
	))

	// Public feedback notifications are consumed through the portal-scoped query
	// API. Publishing them on the generic user channel would expose them to any
	// workspace SSE connection for the same user.
	if notification.EntityType != "feedback" {
		if err := s.publishRealtime(ctx, notification); err != nil {
			span.RecordError(err)
			s.log.Error(ctx, "notifications.Service.Create: failed to publish notification to Redis", "error", err, "notificationID", notification.ID)
			return notification, err
		}
	}

	emailPayload := tasks.NotificationEmailDigestPayload{
		RecipientID: notification.RecipientID,
		WorkspaceID: notification.WorkspaceID,
	}

	if _, err := s.tasksService.EnqueueNotificationEmailDigest(emailPayload); err != nil {
		// Log error but don't fail notification creation
		s.log.Error(ctx, "Failed to enqueue email notification digest task",
			"error", err,
			"notification_id", notification.ID,
			"recipient_id", notification.RecipientID)
		span.AddEvent("email task enqueue failed", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("notification.id", notification.ID.String()),
		))
	} else {
		s.log.Info(ctx, "Email notification digest task enqueued successfully",
			"notification_id", notification.ID,
			"recipient_id", notification.RecipientID)
		span.AddEvent("email task enqueued", trace.WithAttributes(
			attribute.String("notification.id", notification.ID.String()),
		))
	}

	return notification, nil
}

// List returns a list of notifications.
func (s *Service) List(ctx context.Context, userID, workspaceID uuid.UUID, search string, limit, offset int) ([]CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.List")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.List")
	defer span.End()

	notifications, err := s.repo.List(ctx, userID, workspaceID, search, limit, offset)
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

// ListPortalFeedback returns only feedback notifications belonging to the
// requested public portal and recipient.
func (s *Service) ListPortalFeedback(ctx context.Context, userID uuid.UUID, portalSlug string, limit, offset int) ([]CorePortalNotification, error) {
	portalSlug = strings.TrimSpace(portalSlug)
	if userID == uuid.Nil || portalSlug == "" {
		return nil, errors.New("user id and portal slug are required")
	}
	if limit < 1 {
		limit = 21
	}
	if limit > 101 {
		limit = 101
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPortalFeedback(ctx, userID, portalSlug, limit, offset)
}

func (s *Service) GetPortalFeedbackUnreadCount(ctx context.Context, userID uuid.UUID, portalSlug string) (int, error) {
	portalSlug = strings.TrimSpace(portalSlug)
	if userID == uuid.Nil || portalSlug == "" {
		return 0, errors.New("user id and portal slug are required")
	}
	return s.repo.GetPortalFeedbackUnreadCount(ctx, userID, portalSlug)
}

func (s *Service) MarkPortalFeedbackAsRead(ctx context.Context, notificationID, userID uuid.UUID, portalSlug string) error {
	portalSlug = strings.TrimSpace(portalSlug)
	if notificationID == uuid.Nil || userID == uuid.Nil || portalSlug == "" {
		return errors.New("notification id, user id, and portal slug are required")
	}
	return s.repo.MarkPortalFeedbackAsRead(ctx, notificationID, userID, portalSlug)
}
