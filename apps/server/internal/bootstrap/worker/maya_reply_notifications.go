package workerbootstrap

import (
	"context"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/hibiken/asynq"
)

const mayaReplyNotificationTask = "notifications:maya-reply-delivery"

func registerMayaReplyNotificationDelivery(mux *asynq.ServeMux, scheduler scheduleRegistrar, service *notifications.Service) error {
	mux.HandleFunc(mayaReplyNotificationTask, func(ctx context.Context, _ *asynq.Task) error {
		_, err := service.DispatchMayaReplies(ctx, 25)
		return err
	})
	_, err := scheduler.Register("@every 5s", asynq.NewTask(mayaReplyNotificationTask, nil), asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(30*time.Second), asynq.Unique(4*time.Second))
	return err
}
