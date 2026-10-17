package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/hibiken/asynq"
)

const TypeChatTranscription = "chat:transcribe-audio"

// PostgreSQL owns requests and claim leases. Redis is only a wake-up; the same
// unique task is scheduled periodically to recover a missed enqueue.
func ChatTranscriptionOptions() []asynq.Option {
	return []asynq.Option{asynq.Queue("default"), asynq.MaxRetry(1), asynq.Timeout(4*time.Minute + 30*time.Second), asynq.Unique(5 * time.Minute)}
}

func (s *Service) EnqueueChatTranscription(ctx context.Context) error {
	_, err := s.asynqClient.EnqueueContext(ctx, asynq.NewTask(TypeChatTranscription, nil), ChatTranscriptionOptions()...)
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	return err
}
