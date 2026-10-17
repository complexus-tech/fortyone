package workerbootstrap

import (
	"context"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/hibiken/asynq"
)

const documentNotificationDeliveryTask = "notifications:document-delivery"

func registerDocumentNotificationDelivery(mux *asynq.ServeMux, scheduler scheduleRegistrar, service *notifications.Service) error {
	mux.HandleFunc(documentNotificationDeliveryTask, func(ctx context.Context, _ *asynq.Task) error {
		_, err := service.DispatchDocumentNotifications(ctx, 25)
		return err
	})
	_, err := scheduler.Register("@every 5s", asynq.NewTask(documentNotificationDeliveryTask, nil), asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(30*time.Second), asynq.Unique(4*time.Second))
	return err
}
