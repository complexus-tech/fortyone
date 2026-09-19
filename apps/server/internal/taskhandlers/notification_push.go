package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/expopush"
	"github.com/complexus-tech/projects-api/pkg/tasks"
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
	invalidTokens := make([]string, 0)
	for start := 0; start < len(delivery.Tokens); start += expoPushBatchSize {
		end := min(start+expoPushBatchSize, len(delivery.Tokens))
		messages := make([]expopush.Message, 0, end-start)
		for _, token := range delivery.Tokens[start:end] {
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
		result, err := h.pushSender.Send(ctx, messages)
		invalidTokens = append(invalidTokens, result.InvalidTokens...)
		if err != nil {
			if disableErr := h.pushDeliveries.DisablePushDevices(ctx, invalidTokens); disableErr != nil {
				return fmt.Errorf("disable invalid push devices after send failure: %w", disableErr)
			}
			return err
		}
	}
	if err := h.pushDeliveries.DisablePushDevices(ctx, invalidTokens); err != nil {
		return err
	}
	return h.pushDeliveries.MarkPushSent(ctx, delivery.NotificationID)
}
