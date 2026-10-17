package tasks

import (
	"context"
	"errors"
	"github.com/hibiken/asynq"
	"time"
)

const TypeChatDelivery = "chat:deliver"

// EnqueueChatDelivery is a wake-up only; PostgreSQL owns the durable event.
func (s *Service) EnqueueChatDelivery(ctx context.Context) error {
	_, err := s.asynqClient.EnqueueContext(ctx, asynq.NewTask(TypeChatDelivery, nil), asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(30*time.Second), asynq.Unique(time.Second))
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	return err
}
