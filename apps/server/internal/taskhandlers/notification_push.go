package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/expopush"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const expoPushBatchSize = 100

func (h *handlers) HandleNotificationPush(ctx context.Context, task *asynq.Task) error {
	var payload tasks.NotificationPushPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal notification push payload: %w: %w", err, asynq.SkipRetry)
	}
	if h.pushDeliveries == nil || h.pushSender == nil {
		return fmt.Errorf("notification push delivery is unavailable")
	}
	delivery, err := h.pushDeliveries.GetPushDelivery(ctx, payload.NotificationID)
	if err != nil || delivery == nil {
		return err
	}
	var notificationMessage NotificationMessage
	if err := json.Unmarshal(delivery.Message, &notificationMessage); err != nil {
		return fmt.Errorf("unmarshal notification push message: %w: %w", err, asynq.SkipRetry)
	}
	body := parseNotificationMessage(notificationMessage).Text
	if body == "" {
		body = "Open FortyOne to view this notification."
	}
	messages := make([]expopush.Message, 0, len(delivery.Tokens))
	for _, token := range delivery.Tokens {
		messages = append(messages, expopush.Message{
			To: token, Title: delivery.Title, Body: body, Sound: "default", Priority: "high",
			Data: map[string]any{
				"notificationId": delivery.NotificationID.String(),
				"recipientId":    delivery.RecipientID.String(),
				"workspaceId":    delivery.WorkspaceID.String(),
				"workspaceSlug":  delivery.WorkspaceSlug,
				"entityType":     string(delivery.EntityType),
				"entityId":       delivery.EntityID.String(),
			},
		})
	}
	if err := h.sendPushMessages(ctx, messages); err != nil {
		return err
	}
	return h.pushDeliveries.MarkPushSent(ctx, delivery.NotificationID)
}

func (h *handlers) HandleNotificationPushTest(ctx context.Context, task *asynq.Task) error {
	var payload tasks.NotificationPushTestPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal test notification push payload: %w: %w", err, asynq.SkipRetry)
	}
	if payload.RecipientID == uuid.Nil {
		return fmt.Errorf("test notification push recipient is required: %w", asynq.SkipRetry)
	}
	if h.pushDeliveries == nil || h.pushSender == nil {
		return fmt.Errorf("notification push delivery is unavailable")
	}
	tokens, err := h.pushDeliveries.ListPushTokens(ctx, payload.RecipientID)
	if err != nil {
		return err
	}
	messages := make([]expopush.Message, 0, len(tokens))
	for _, token := range tokens {
		messages = append(messages, expopush.Message{
			To: token, Title: "FortyOne", Body: "Push notifications are working.", Sound: "default", Priority: "high",
			Data: map[string]any{"kind": "test", "recipientId": payload.RecipientID.String()},
		})
	}
	return h.sendPushMessages(ctx, messages)
}

func (h *handlers) sendPushMessages(ctx context.Context, messages []expopush.Message) error {
	invalidTokens := make([]string, 0)
	for start := 0; start < len(messages); start += expoPushBatchSize {
		end := min(start+expoPushBatchSize, len(messages))
		result, err := h.pushSender.Send(ctx, messages[start:end])
		invalidTokens = append(invalidTokens, result.InvalidTokens...)
		if err != nil {
			if disableErr := h.pushDeliveries.DisablePushDevices(ctx, invalidTokens); disableErr != nil {
				return fmt.Errorf("disable invalid push devices after send failure: %w", disableErr)
			}
			return err
		}
	}
	return h.pushDeliveries.DisablePushDevices(ctx, invalidTokens)
}
