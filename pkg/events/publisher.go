package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	redis *redis.Client
	log   *logger.Logger
}

func NewPublisher(redis *redis.Client, log *logger.Logger) *Publisher {
	return &Publisher{
		redis: redis,
		log:   log,
	}
}

func (p *Publisher) Publish(ctx context.Context, event Event) error {
	p.log.Info(ctx, "events.publisher.Publish", "event", event.Type)
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.redis.Publish(ctx, string(event.Type), payload).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
