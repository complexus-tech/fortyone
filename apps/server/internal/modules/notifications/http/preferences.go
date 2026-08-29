package notificationshttp

import (
	"context"
	"net/http"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (handlers *Handlers) GetPreferences(ctx context.Context, response http.ResponseWriter, _ *http.Request) error {
	actorID, workspaceID, err := notificationWorkspaceAccess(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	preferences, err := handlers.notifications.GetPreferences(ctx, actorID, workspaceID)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, toAppNotificationPreferences(preferences), http.StatusOK)
}

func (handlers *Handlers) UpdatePreference(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	var input AppUpdatePreference
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, response, err, http.StatusBadRequest)
	}
	actorID, workspaceID, err := notificationWorkspaceAccess(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	preferenceType, err := notificationsdomain.ParsePreferenceType(web.Params(request, "type"))
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	patch := notificationsdomain.ChannelPatch{}
	if input.EmailEnabled != nil {
		patch.Email = platformpatch.Set(*input.EmailEnabled)
	}
	if input.InAppEnabled != nil {
		patch.InApp = platformpatch.Set(*input.InAppEnabled)
	}
	if _, err := handlers.notifications.UpdatePreference(ctx, actorID, workspaceID, preferenceType, patch); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}
