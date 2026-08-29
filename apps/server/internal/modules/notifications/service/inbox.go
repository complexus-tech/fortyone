package notifications

import (
	"context"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (service *Service) List(ctx context.Context, actorID, workspaceID uuid.UUID, search string, limit, offset int) ([]CoreNotification, error) {
	notifications, err := service.repo.List(ctx, notificationsdomain.ListQuery{
		Access: notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID},
		Search: search,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	for index := range notifications {
		notifications[index] = notifications[index].Public()
	}
	return notifications, nil
}

func (service *Service) GetUnreadCount(ctx context.Context, actorID, workspaceID uuid.UUID) (int, error) {
	return service.repo.CountUnread(ctx, notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID})
}

func (service *Service) MarkAsRead(ctx context.Context, notificationID, actorID, workspaceID uuid.UUID) error {
	return service.mutate(ctx, notificationID, actorID, workspaceID, notificationsdomain.NotificationMutationRead)
}

func (service *Service) MarkAsUnread(ctx context.Context, notificationID, actorID, workspaceID uuid.UUID) error {
	return service.mutate(ctx, notificationID, actorID, workspaceID, notificationsdomain.NotificationMutationUnread)
}

func (service *Service) DeleteNotification(ctx context.Context, notificationID, actorID, workspaceID uuid.UUID) error {
	return service.mutate(ctx, notificationID, actorID, workspaceID, notificationsdomain.NotificationMutationDelete)
}

func (service *Service) mutate(ctx context.Context, notificationID, actorID, workspaceID uuid.UUID, kind notificationsdomain.NotificationMutationKind) error {
	return service.repo.Mutate(ctx, notificationsdomain.NotificationMutation{
		Access:         notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID},
		NotificationID: notificationID,
		Kind:           kind,
		At:             service.clock.Now().UTC(),
	})
}

func (service *Service) MarkAllAsRead(ctx context.Context, actorID, workspaceID uuid.UUID) error {
	_, err := service.mutateAll(ctx, actorID, workspaceID, notificationsdomain.WorkspaceMutationReadAll)
	return err
}

func (service *Service) DeleteAllNotifications(ctx context.Context, actorID, workspaceID uuid.UUID) (int64, error) {
	count, err := service.mutateAll(ctx, actorID, workspaceID, notificationsdomain.WorkspaceMutationDeleteAll)
	return int64(count), err
}

func (service *Service) DeleteReadNotifications(ctx context.Context, actorID, workspaceID uuid.UUID) (int64, error) {
	count, err := service.mutateAll(ctx, actorID, workspaceID, notificationsdomain.WorkspaceMutationDeleteRead)
	return int64(count), err
}

func (service *Service) mutateAll(ctx context.Context, actorID, workspaceID uuid.UUID, kind notificationsdomain.WorkspaceMutationKind) (int, error) {
	return service.repo.MutateAll(ctx, notificationsdomain.WorkspaceMutation{
		Access: notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID},
		Kind:   kind,
		At:     service.clock.Now().UTC(),
	})
}
