package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	eventStreamKey = "events-stream"
)

type Publisher struct {
	redis *redis.Client
	log   *logger.Logger
}

func New(redis *redis.Client, log *logger.Logger) *Publisher {
	return &Publisher{
		redis: redis,
		log:   log,
	}
}

func (p *Publisher) Publish(ctx context.Context, event events.Event) error {
	p.log.Info(ctx, "publisher.Publish", "event", event.Type)

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Add to stream
	fields := map[string]any{
		"type":      string(event.Type),
		"payload":   string(payload),
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"actor_id":  event.ActorID.String(),
	}

	_, err = p.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: eventStreamKey,
		Values: fields,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to add event to stream: %w", err)
	}

	return nil
}
