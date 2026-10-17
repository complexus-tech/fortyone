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
	delivery, err := service.repo.GetEmailDelivery(ctx, query)
	if err != nil || delivery == nil {
		return delivery, err
	}
	allowed, err := service.allowsChatDelivery(ctx, delivery.WorkspaceID, delivery.EntityType)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, nil
	}
	return delivery, nil
}

func (service *Service) ListEmailDigest(ctx context.Context, scope notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error) {
	digest, err := service.repo.ListEmailDigest(ctx, scope)
	if err != nil || digest == nil {
		return digest, err
	}
	// Filter before callers snapshot/acknowledge the batch. Suppressed Messages
	// remain pending and may be delivered if that workspace is enabled again.
	filtered := *digest
	filtered.Items = make([]notificationsdomain.EmailDigestItem, 0, len(digest.Items))
	chatAllowed, err := service.allowsChatDelivery(ctx, scope.WorkspaceID, notificationsdomain.EntityTypeChatMessage)
	if err != nil {
		return nil, err
	}
	for _, item := range digest.Items {
		if item.EntityType != notificationsdomain.EntityTypeChatMessage || chatAllowed {
			filtered.Items = append(filtered.Items, item)
		}
	}
	return &filtered, nil
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

func (service *Service) MarkEmailSnapshotsSent(ctx context.Context, scope notificationsdomain.DeliveryScope, snapshots []notificationsdomain.EmailSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return service.repo.MarkEmailSent(ctx, notificationsdomain.MarkEmailSent{Scope: scope, NotificationSnapshots: snapshots, At: service.clock.Now().UTC()})
}
