package notificationshttp

import (
	"context"
	"net/http"
	"strings"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (handlers *Handlers) RegisterPushDevice(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	var input AppPushDeviceInput
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, response, err, http.StatusBadRequest)
	}
	device, err := handlers.notifications.RegisterPushDevice(
		ctx,
		actorID,
		strings.TrimSpace(input.Token),
		notificationsdomain.PushPlatform(strings.ToLower(strings.TrimSpace(input.Platform))),
	)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, AppPushDevice{
		ID: device.ID, Platform: string(device.Platform),
		CreatedAt: device.CreatedAt, UpdatedAt: device.UpdatedAt,
	}, http.StatusOK)
}

func (handlers *Handlers) UnregisterPushDevice(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	var input AppPushDeviceInput
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, response, err, http.StatusBadRequest)
	}
	if err := handlers.notifications.UnregisterPushDevice(ctx, actorID, strings.TrimSpace(input.Token)); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}
