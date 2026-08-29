package notifications

import (
	"context"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (service *Service) ListKeyResultUpdateAudience(ctx context.Context, actorID, workspaceID, keyResultID uuid.UUID) ([]notificationsdomain.KeyResultAudienceMember, error) {
	return service.repo.ListKeyResultAudience(ctx, notificationsdomain.KeyResultAudienceQuery{
		ActorID:     actorID,
		WorkspaceID: workspaceID,
		KeyResultID: keyResultID,
	})
}

func (service *Service) GetEmailDelivery(ctx context.Context, query notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error) {
	return service.repo.GetEmailDelivery(ctx, query)
}

func (service *Service) ListEmailDigest(ctx context.Context, scope notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error) {
	return service.repo.ListEmailDigest(ctx, scope)
}

func (service *Service) ListDeliveryTeamIDs(ctx context.Context, scope notificationsdomain.DeliveryScope) ([]uuid.UUID, error) {
	return service.repo.ListDeliveryTeamIDs(ctx, scope)
}

func (service *Service) MarkEmailSent(ctx context.Context, scope notificationsdomain.DeliveryScope, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}
	return service.repo.MarkEmailSent(ctx, notificationsdomain.MarkEmailSent{
		Scope:           scope,
		NotificationIDs: notificationIDs,
		At:              service.clock.Now().UTC(),
	})
}
