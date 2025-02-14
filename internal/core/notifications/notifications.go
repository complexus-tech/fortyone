package notifications

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the notifications storage.
type Repository interface {
	Create(ctx context.Context, n CoreNewNotification) (CoreNotification, error)
	GetUnread(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]CoreNotification, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) error
	GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreNotificationPreference, error)
	UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error
	CreateDefaultPreferences(ctx context.Context, userID, workspaceID uuid.UUID) error
}

// Service provides notification-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new notifications service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Create creates a new notification.
func (s *Service) Create(ctx context.Context, n CoreNewNotification) (CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.Create")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.Create")
	defer span.End()

	notification, err := s.repo.Create(ctx, n)
	if err != nil {
		span.RecordError(err)
		return CoreNotification{}, err
	}

	span.AddEvent("notification created", trace.WithAttributes(
		attribute.String("notification.id", notification.ID.String()),
	))

	return notification, nil
}

// GetUnread returns a list of unread notifications.
func (s *Service) GetUnread(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]CoreNotification, error) {
	s.log.Info(ctx, "business.core.notifications.GetUnread")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.GetUnread")
	defer span.End()

	notifications, err := s.repo.GetUnread(ctx, userID, workspaceID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("notifications retrieved", trace.WithAttributes(
		attribute.Int("notifications.count", len(notifications)),
	))

	return notifications, nil
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

// GetPreferences returns a list of notification preferences.
func (s *Service) GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreNotificationPreference, error) {
	s.log.Info(ctx, "business.core.notifications.GetPreferences")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.GetPreferences")
	defer span.End()

	preferences, err := s.repo.GetPreferences(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("preferences retrieved", trace.WithAttributes(
		attribute.Int("preferences.count", len(preferences)),
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

// CreateDefaultPreferences creates default notification preferences for a user.
func (s *Service) CreateDefaultPreferences(ctx context.Context, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.notifications.CreateDefaultPreferences")
	ctx, span := web.AddSpan(ctx, "business.core.notifications.CreateDefaultPreferences")
	defer span.End()

	if err := s.repo.CreateDefaultPreferences(ctx, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("default preferences created", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))

	return nil
}
