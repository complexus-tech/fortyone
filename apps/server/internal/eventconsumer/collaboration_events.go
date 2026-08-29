package eventconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/events"
)

func (c *Consumer) handleCommentCreated(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleCommentCreated", "event_type", event.Type)

	var payload events.CommentCreatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessCommentCreated(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process comment created rules", "error", err)
		return err
	}

	// Create notifications
	for index, notification := range notifications {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}

func (c *Consumer) handleCommentReplied(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleCommentReplied", "event_type", event.Type)

	var payload events.CommentRepliedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessCommentReplied(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process comment replied rules", "error", err)
		return err
	}

	// Create notifications
	for index, notification := range notifications {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}

func (c *Consumer) handleUserMentioned(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleUserMentioned", "event_type", event.Type)

	var payload events.UserMentionedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessUserMentioned(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process user mentioned rules", "error", err)
		return err
	}

	// Create notifications
	for index, notification := range notifications {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}
