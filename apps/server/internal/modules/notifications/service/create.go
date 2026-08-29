package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Create persists the notification and its durable email intent in one SQL
// statement. Queue delivery is retryable: an exact dedupe replay skips realtime
// fanout but re-enqueues the unique digest wake-up.
func (service *Service) Create(ctx context.Context, input CoreNewNotification) (CoreNotification, error) {
	ctx, span := otel.Tracer("fortyone.notifications").Start(ctx, "notifications.Create")
	defer span.End()

	if err := input.Validate(); err != nil {
		return CoreNotification{}, err
	}
	notification, inserted, err := service.repo.Create(ctx, input)
	if err != nil {
		span.RecordError(err)
		return CoreNotification{}, err
	}

	if inserted && notification.InAppEnabled && notification.EntityType != notificationsdomain.EntityTypeFeedback {
		if err := service.publishRealtime(ctx, notification.Public()); err != nil {
			span.RecordError(err)
			return notification, fmt.Errorf("publish notification realtime event: %w", err)
		}
	}
	if err := service.enqueueDigest(notification); err != nil {
		span.RecordError(err)
		return notification, err
	}
	span.SetAttributes(
		attribute.String("notification.id", notification.ID.String()),
		attribute.Bool("notification.inserted", inserted),
	)
	return notification, nil
}

func (service *Service) enqueueDigest(notification CoreNotification) error {
	if service.tasksService == nil {
		return fmt.Errorf("enqueue notification email digest: task service is unavailable")
	}
	_, err := service.tasksService.EnqueueNotificationEmailDigest(tasks.NotificationEmailDigestPayload{
		RecipientID: notification.RecipientID,
		WorkspaceID: notification.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("enqueue notification email digest: %w", err)
	}
	return nil
}

func (service *Service) publishNotification(ctx context.Context, notification CoreNotification) error {
	ctx, span := otel.Tracer("fortyone.notifications").Start(ctx, "notifications.publishRealtime")
	defer span.End()
	if service.redisClient == nil {
		return nil
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal realtime notification: %w", err)
	}
	channel := fmt.Sprintf("user-notifications:%s", notification.RecipientID)
	if err := service.redisClient.Publish(ctx, channel, payload).Err(); err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("redis.channel", channel)))
		return fmt.Errorf("publish realtime notification: %w", err)
	}
	return nil
}
