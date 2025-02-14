package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	redis         *redis.Client
	log           *logger.Logger
	notifications *notifications.Service
}

func NewConsumer(redis *redis.Client, log *logger.Logger, notifications *notifications.Service) *Consumer {
	return &Consumer{
		redis:         redis,
		log:           log,
		notifications: notifications,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	pubsub := c.redis.Subscribe(ctx,
		string(StoryUpdated),
		string(StoryCommented),
		string(ObjectiveUpdated),
		string(KeyResultUpdated),
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var event Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			c.log.Error(ctx, "failed to unmarshal event", "error", err)
			continue
		}

		if err := c.handleEvent(ctx, event); err != nil {
			c.log.Error(ctx, "failed to handle event", "error", err)
		}
	}

	return nil
}

func (c *Consumer) handleEvent(ctx context.Context, event Event) error {
	switch event.Type {
	case StoryUpdated:
		return c.handleStoryUpdated(ctx, event)
	case StoryCommented:
		return c.handleStoryCommented(ctx, event)
	case ObjectiveUpdated:
		return c.handleObjectiveUpdated(ctx, event)
	case KeyResultUpdated:
		return c.handleKeyResultUpdated(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (c *Consumer) handleStoryUpdated(ctx context.Context, event Event) error {
	// TODO: Implement story update notification logic
	return nil
}

func (c *Consumer) handleStoryCommented(ctx context.Context, event Event) error {
	// TODO: Implement story comment notification logic
	return nil
}

func (c *Consumer) handleObjectiveUpdated(ctx context.Context, event Event) error {
	// TODO: Implement objective update notification logic
	return nil
}

func (c *Consumer) handleKeyResultUpdated(ctx context.Context, event Event) error {
	// TODO: Implement key result update notification logic
	return nil
}
