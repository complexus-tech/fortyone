package notificationshttp

import (
	"context"
	"net/http"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (handlers *Handlers) ListPortalFeedback(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	query, err := parsePortalListQuery(request.URL.Query())
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusBadRequest)
	}
	items, err := handlers.notifications.ListPortalFeedback(
		ctx,
		actorID,
		web.Params(request, "portalSlug"),
		query.UnreadOnly,
		query.Pagination.PageSize+1,
		query.Pagination.Offset(),
	)
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	hasMore := len(items) > query.Pagination.PageSize
	if hasMore {
		items = items[:query.Pagination.PageSize]
	}
	handlers.resolvePortalNotificationAvatars(ctx, items)
	return web.Respond(ctx, response, toAppPortalNotificationsResponse(
		items,
		query.Pagination.Page,
		query.Pagination.PageSize,
		hasMore,
	), http.StatusOK)
}

func (handlers *Handlers) GetPortalFeedbackUnreadCount(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	count, err := handlers.notifications.GetPortalFeedbackUnreadCount(ctx, actorID, web.Params(request, "portalSlug"))
	if err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, AppUnreadCount{Count: count}, http.StatusOK)
}

func (handlers *Handlers) MarkPortalFeedbackAsRead(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	notificationID, err := uuid.Parse(web.Params(request, "id"))
	if err != nil {
		return web.RespondError(ctx, response, ErrInvalidNotificationID, http.StatusBadRequest)
	}
	if err := handlers.notifications.MarkPortalFeedbackAsRead(ctx, notificationID, actorID, web.Params(request, "portalSlug")); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}

func (handlers *Handlers) MarkAllPortalFeedbackAsRead(ctx context.Context, response http.ResponseWriter, request *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, response, err, http.StatusUnauthorized)
	}
	if err := handlers.notifications.MarkAllPortalFeedbackAsRead(ctx, actorID, web.Params(request, "portalSlug")); err != nil {
		return respondNotificationError(ctx, response, err)
	}
	return web.Respond(ctx, response, nil, http.StatusNoContent)
}
