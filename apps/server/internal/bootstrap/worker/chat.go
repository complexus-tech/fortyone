package workerbootstrap

import (
	"context"
	"github.com/complexus-tech/projects-api/internal/bootstrap/chatadapter"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"time"
)

func registerChatDelivery(mux *asynq.ServeMux, scheduler scheduleRegistrar, dispatcher chatadapter.Dispatcher) error {
	mux.HandleFunc(tasks.TypeChatDelivery, func(ctx context.Context, _ *asynq.Task) error { _, err := dispatcher.Dispatch(ctx, 50); return err })
	_, err := scheduler.Register("@every 5s", asynq.NewTask(tasks.TypeChatDelivery, nil), asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(30*time.Second), asynq.Unique(time.Second))
	return err
}
