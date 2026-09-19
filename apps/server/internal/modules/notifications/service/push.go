package notifications

import (
	"context"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

func (service *Service) RegisterPushDevice(ctx context.Context, userID uuid.UUID, token string, platform notificationsdomain.PushPlatform) (notificationsdomain.PushDevice, error) {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return notificationsdomain.PushDevice{}, fmt.Errorf("push device repository is unavailable")
	}
	return repository.RegisterPushDevice(ctx, notificationsdomain.RegisterPushDevice{
		UserID: userID, Token: token, Platform: platform,
	})
}

func (service *Service) UnregisterPushDevice(ctx context.Context, userID uuid.UUID, token string) error {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return fmt.Errorf("push device repository is unavailable")
	}
	return repository.UnregisterPushDevice(ctx, userID, token)
}

func (service *Service) SendTestPush(ctx context.Context, userID uuid.UUID) error {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return fmt.Errorf("push device repository is unavailable")
	}
	tokens, err := repository.ListPushTokens(ctx, userID)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return fmt.Errorf("%w: enable push notifications on this device first", notificationsdomain.ErrConflict)
	}
	pushTasks, ok := service.tasksService.(PushTasksService)
	if !ok {
		return fmt.Errorf("enqueue test notification: task service is unavailable")
	}
	if _, err := pushTasks.EnqueueNotificationPushTest(tasks.NotificationPushTestPayload{RecipientID: userID}); err != nil {
		return fmt.Errorf("enqueue test notification: %w", err)
	}
	return nil
}

func (service *Service) ListPushTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return nil, fmt.Errorf("push device repository is unavailable")
	}
	return repository.ListPushTokens(ctx, userID)
}

func (service *Service) GetPushDelivery(ctx context.Context, notificationID uuid.UUID) (*notificationsdomain.PushDelivery, error) {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return nil, fmt.Errorf("push delivery repository is unavailable")
	}
	return repository.GetPushDelivery(ctx, notificationID)
}

func (service *Service) MarkPushSent(ctx context.Context, notificationID uuid.UUID) error {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return fmt.Errorf("push delivery repository is unavailable")
	}
	return repository.MarkPushSent(ctx, notificationID, service.clock.Now().UTC())
}

func (service *Service) DisablePushDevices(ctx context.Context, tokens []string) error {
	repository, ok := service.repo.(PushRepository)
	if !ok {
		return fmt.Errorf("push delivery repository is unavailable")
	}
	return repository.DisablePushDevices(ctx, tokens, service.clock.Now().UTC())
}
