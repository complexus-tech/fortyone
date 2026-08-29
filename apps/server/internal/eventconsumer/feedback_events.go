package eventconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/events"
)

func (c *Consumer) handleFeedbackCommentCreated(ctx context.Context, event events.Event) error {
	var payload events.FeedbackCommentCreatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal feedback comment payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshal feedback comment payload: %w", err)
	}

	for index, notification := range c.notificationRules.ProcessFeedbackCommentCreated(ctx, payload, event.ActorID) {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create feedback comment notification: %w", err)
		}
	}
	return nil
}

func (c *Consumer) handleFeedbackStatusUpdated(ctx context.Context, event events.Event) error {
	var payload events.FeedbackStatusUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal feedback status payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshal feedback status payload: %w", err)
	}

	for index, notification := range c.notificationRules.ProcessFeedbackStatusUpdated(ctx, payload, event.ActorID) {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create feedback status notification: %w", err)
		}
	}
	return nil
}

func (c *Consumer) handleFeedbackUpdatePublished(ctx context.Context, event events.Event) error {
	var payload events.FeedbackUpdatePublishedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal feedback update payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshal feedback update payload: %w", err)
	}
	for index, notification := range c.notificationRules.ProcessFeedbackUpdatePublished(ctx, payload, event.ActorID) {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create feedback update notification: %w", err)
		}
	}
	return nil
}

func (c *Consumer) handleFeedbackItemMerged(ctx context.Context, event events.Event) error {
	var payload events.FeedbackItemMergedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal feedback merge payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshal feedback merge payload: %w", err)
	}
	for index, notification := range c.notificationRules.ProcessFeedbackItemMerged(ctx, payload, event.ActorID) {
		notification = withEventDedupeKey(event, notification, index)
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("create feedback merge notification: %w", err)
		}
	}
	return nil
}

// handleStoryUpdated processes story update events using the new notification rules
