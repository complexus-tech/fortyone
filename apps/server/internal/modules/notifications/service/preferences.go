package notifications

import (
	"context"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (service *Service) GetPreferences(ctx context.Context, actorID, workspaceID uuid.UUID) (CoreNotificationPreferences, error) {
	return service.repo.GetPreferences(ctx, notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID})
}

func (service *Service) UpdatePreference(ctx context.Context, actorID, workspaceID uuid.UUID, preferenceType PreferenceType, patch NotificationChannelPatch) (CoreNotificationPreferences, error) {
	return service.repo.UpdatePreference(ctx, notificationsdomain.UpdatePreference{
		Access: notificationsdomain.WorkspaceAccess{ActorID: actorID, WorkspaceID: workspaceID},
		Type:   preferenceType,
		Patch:  patch,
		At:     service.clock.Now().UTC(),
	})
}
