package notificationshttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidNotificationID = errors.New("notification id is not in its proper form")
	ErrInvalidWorkspaceID    = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	notifications *notifications.Service
}

func New(notifications *notifications.Service) *Handlers {
	return &Handlers{
		notifications: notifications,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.List")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notifications.List(ctx, userID, workspace.ID, limit, offset)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("notifications retrieved", trace.WithAttributes(
		attribute.Int("notifications.count", len(notifications)),
	))

	return web.Respond(ctx, w, toAppNotifications(notifications), http.StatusOK)
}

func (h *Handlers) GetUnreadCount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.GetUnreadCount")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	count, err := h.notifications.GetUnreadCount(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("unread count retrieved", trace.WithAttributes(
		attribute.Int("unread_count", count),
	))

	return web.Respond(ctx, w, count, http.StatusOK)
}

func (h *Handlers) MarkAsRead(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.MarkAsRead")
	defer span.End()

	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	notificationIDParam := web.Params(r, "id")
	notificationID, err := uuid.Parse(notificationIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidNotificationID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.notifications.MarkAsRead(ctx, notificationID, userID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("notification marked as read", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.GetPreferences")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	preferences, err := h.notifications.GetPreferences(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("preferences retrieved", trace.WithAttributes(
		attribute.Int("preferences.count", len(preferences.Preferences)),
	))

	return web.Respond(ctx, w, toAppNotificationPreferences(preferences), http.StatusOK)
}

func (h *Handlers) UpdatePreference(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.UpdatePreference")
	defer span.End()

	var input AppUpdatePreference
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	notificationType := web.Params(r, "type")
	if notificationType == "" {
		return web.RespondError(ctx, w, errors.New("notification type is required"), http.StatusBadRequest)
	}

	updates := make(map[string]any)
	if input.EmailEnabled != nil {
		updates["email_enabled"] = *input.EmailEnabled
	}
	if input.InAppEnabled != nil {
		updates["in_app_enabled"] = *input.InAppEnabled
	}

	if err := h.notifications.UpdatePreference(ctx, userID, workspace.ID, notificationType, updates); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("preference updated", trace.WithAttributes(
		attribute.String("notification.type", notificationType),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) MarkAllAsRead(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.MarkAllAsRead")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.notifications.MarkAllAsRead(ctx, userID, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("all notifications marked as read", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) DeleteNotification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.DeleteNotification")
	defer span.End()

	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	notificationIDParam := web.Params(r, "id")
	notificationID, err := uuid.Parse(notificationIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidNotificationID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.notifications.DeleteNotification(ctx, notificationID, userID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("notification deleted", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) DeleteAllNotifications(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.DeleteAllNotifications")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	count, err := h.notifications.DeleteAllNotifications(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("all notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", count),
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, map[string]int64{"deleted_count": count}, http.StatusOK)
}

func (h *Handlers) DeleteReadNotifications(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.DeleteReadNotifications")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	count, err := h.notifications.DeleteReadNotifications(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("read notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", count),
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, map[string]int64{"deleted_count": count}, http.StatusOK)
}

func (h *Handlers) MarkAsUnread(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.notifications.MarkAsUnread")
	defer span.End()

	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	notificationIDParam := web.Params(r, "id")
	notificationID, err := uuid.Parse(notificationIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidNotificationID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.notifications.MarkAsUnread(ctx, notificationID, userID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("notification marked as unread", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
