package notifications

import (
	"context"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
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
