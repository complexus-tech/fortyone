package notifications

import (
	"context"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (service *Service) ListPortalFeedback(ctx context.Context, actorID uuid.UUID, portalSlug string, unreadOnly bool, limit, offset int) ([]CorePortalNotification, error) {
	return service.repo.ListPortalFeedback(ctx, notificationsdomain.PortalListQuery{
		Access:     notificationsdomain.PortalAccess{ActorID: actorID, PortalSlug: portalSlug},
		UnreadOnly: unreadOnly,
		Limit:      limit,
		Offset:     offset,
	})
}

func (service *Service) GetPortalFeedbackUnreadCount(ctx context.Context, actorID uuid.UUID, portalSlug string) (int, error) {
	return service.repo.CountUnreadPortalFeedback(ctx, notificationsdomain.PortalAccess{ActorID: actorID, PortalSlug: portalSlug})
}

func (service *Service) MarkPortalFeedbackAsRead(ctx context.Context, notificationID, actorID uuid.UUID, portalSlug string) error {
	return service.repo.MarkPortalFeedbackRead(ctx, notificationsdomain.PortalNotificationMutation{
		Access:         notificationsdomain.PortalAccess{ActorID: actorID, PortalSlug: portalSlug},
		NotificationID: notificationID,
		At:             service.clock.Now().UTC(),
	})
}

func (service *Service) MarkAllPortalFeedbackAsRead(ctx context.Context, actorID uuid.UUID, portalSlug string) error {
	_, err := service.repo.MarkAllPortalFeedbackRead(ctx, notificationsdomain.PortalMutation{
		Access: notificationsdomain.PortalAccess{ActorID: actorID, PortalSlug: portalSlug},
		At:     service.clock.Now().UTC(),
	})
	return err
}
