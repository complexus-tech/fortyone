package notificationshttp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (handlers *Handlers) List(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	query, err := parseInboxListQuery(request.URL.Query())
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusBadRequest)
	}
	items, err := handlers.notifications.List(
		ctx,
		actorID,
		workspace.ID,
		query.Search,
		query.Pagination.PageSize+1,
		query.Offset,
	)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	hasMore := len(items) > query.Pagination.PageSize
	if hasMore {
		items = items[:query.Pagination.PageSize]
	}
	handlers.resolveNotificationActors(ctx, items)
	return web.Respond(ctx, response, toAppNotificationsResponse(
		items,
		query.Pagination.Page,
		query.Pagination.PageSize,
		hasMore,
	), http.StatusOK)
}

func (handlers *Handlers) GetUnreadCount(ctx context.Context, response http.ResponseWriter, _ *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	count, err := handlers.notifications.GetUnreadCount(ctx, actorID, workspace.ID)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, count, http.StatusOK)
}

func (handlers *Handlers) MarkAsRead(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	return handlers.mutateNotification(ctx, response, request, func(ctx context.Context, notificationID, actorID, workspaceID uuid.UUID) error {
		observedCreatedAt, err := parseObservedCreatedAt(request.URL.Query())
		if err != nil {
			return err
		}
		return handlers.notifications.MarkAsReadObserved(ctx, notificationID, actorID, workspaceID, observedCreatedAt)
	})
}

func parseObservedCreatedAt(values url.Values) (*time.Time, error) {
	value, present, err := web.OptionalTextQueryParameter(values, "observedCreatedAt", 64, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid observedCreatedAt", notifications.ErrInvalid)
	}
	if !present {
		return nil, nil
	}
	observed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || observed.IsZero() {
		return nil, fmt.Errorf("%w: observedCreatedAt must be a valid timestamp", notifications.ErrInvalid)
	}
	return &observed, nil
}

func (handlers *Handlers) MarkAsUnread(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	return handlers.mutateNotification(ctx, response, request, handlers.notifications.MarkAsUnread)
}

func (handlers *Handlers) DeleteNotification(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	return handlers.mutateNotification(ctx, response, request, handlers.notifications.DeleteNotification)
}

func (handlers *Handlers) mutateNotification(
	ctx context.Context,
	response http.ResponseWriter,
	request *http.Request,
	mutate func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error,
) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	notificationID, err := uuid.Parse(web.Params(request, "id"))
	if err != nil {
		return web.RespondError(ctx, response, ErrInvalidNotificationID, http.StatusBadRequest)
	}
	if err := mutate(ctx, notificationID, actorID, workspace.ID); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}

func (handlers *Handlers) MarkAllAsRead(ctx context.Context, response http.ResponseWriter, _ *http.Request) error {
	actorID, workspaceID, err := notificationWorkspaceAccess(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	if err := handlers.notifications.MarkAllAsRead(ctx, actorID, workspaceID); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}

func (handlers *Handlers) DeleteAllNotifications(ctx context.Context, response http.ResponseWriter, _ *http.Request) error {
	actorID, workspaceID, err := notificationWorkspaceAccess(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	count, err := handlers.notifications.DeleteAllNotifications(ctx, actorID, workspaceID)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, map[string]int64{"deleted_count": count}, http.StatusOK)
}

func (handlers *Handlers) DeleteReadNotifications(ctx context.Context, response http.ResponseWriter, _ *http.Request) error {
	actorID, workspaceID, err := notificationWorkspaceAccess(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	count, err := handlers.notifications.DeleteReadNotifications(ctx, actorID, workspaceID)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, map[string]int64{"deleted_count": count}, http.StatusOK)
}

func notificationWorkspaceAccess(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return actorID, workspace.ID, nil
}
