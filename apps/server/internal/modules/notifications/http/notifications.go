package notificationshttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidNotificationID = errors.New("notification id is not in its proper form")
	ErrInvalidWorkspaceID    = errors.New("workspace id is not in its proper form")
)

const portalNotificationAvatarExpiry = 24 * time.Hour

type profileImageResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

type Handlers struct {
	notifications *notifications.Service
	profileImages profileImageResolver
	log           *logger.Logger
}

func New(notifications *notifications.Service, profileImages profileImageResolver, log *logger.Logger) *Handlers {
	return &Handlers{
		notifications: notifications,
		profileImages: profileImages,
		log:           log,
	}
}

func (h *Handlers) ListPortalFeedback(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	page, pageSize := portalNotificationPagination(r)
	portalNotifications, err := h.notifications.ListPortalFeedback(
		ctx,
		userID,
		web.Params(r, "portalSlug"),
		pageSize+1,
		(page-1)*pageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	hasMore := len(portalNotifications) > pageSize
	if hasMore {
		portalNotifications = portalNotifications[:pageSize]
	}
	h.resolvePortalNotificationAvatars(ctx, portalNotifications)
	return web.Respond(ctx, w, toAppPortalNotificationsResponse(portalNotifications, page, pageSize, hasMore), http.StatusOK)
}

func (h *Handlers) GetPortalFeedbackUnreadCount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	count, err := h.notifications.GetPortalFeedbackUnreadCount(ctx, userID, web.Params(r, "portalSlug"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, AppUnreadCount{Count: count}, http.StatusOK)
}

func (h *Handlers) MarkPortalFeedbackAsRead(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	notificationID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidNotificationID, http.StatusBadRequest)
	}
	if err := h.notifications.MarkPortalFeedbackAsRead(ctx, notificationID, userID, web.Params(r, "portalSlug")); err != nil {
		if errors.Is(err, notifications.ErrNotificationNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func portalNotificationPagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 20
	if parsed, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && parsed > 0 {
		page = parsed
	}
	if parsed, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && parsed > 0 {
		pageSize = min(parsed, 100)
	}
	return page, pageSize
}

func (h *Handlers) resolvePortalNotificationAvatars(ctx context.Context, portalNotifications []notifications.CorePortalNotification) {
	resolved := make(map[string]*string)
	for index := range portalNotifications {
		avatar := portalNotifications[index].ActorAvatar
		if avatar == nil || strings.TrimSpace(*avatar) == "" {
			portalNotifications[index].ActorAvatar = nil
			continue
		}
		avatarKey := strings.TrimSpace(*avatar)
		if cached, ok := resolved[avatarKey]; ok {
			portalNotifications[index].ActorAvatar = cached
			continue
		}
		if h.profileImages == nil {
			resolved[avatarKey] = nil
			portalNotifications[index].ActorAvatar = nil
			continue
		}
		avatarURL, err := h.profileImages.ResolveProfileImageURL(ctx, avatarKey, portalNotificationAvatarExpiry)
		if err != nil || strings.TrimSpace(avatarURL) == "" {
			if err != nil && h.log != nil {
				h.log.Warn(ctx, "failed to resolve portal notification actor avatar", "error", err)
			}
			resolved[avatarKey] = nil
			portalNotifications[index].ActorAvatar = nil
			continue
		}
		resolved[avatarKey] = &avatarURL
		portalNotifications[index].ActorAvatar = &avatarURL
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
	page := 1
	pageSize := 25

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit = pageSize + 1
	offset = (page - 1) * pageSize
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed + 1
			pageSize = parsed
			page = 1
			offset = 0
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
			if pageSize > 0 {
				page = (offset / pageSize) + 1
			}
		}
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	notifications, err := h.notifications.List(ctx, userID, workspace.ID, search, limit, offset)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	hasMore := len(notifications) > pageSize
	if hasMore {
		notifications = notifications[:pageSize]
	}

	span.AddEvent("notifications retrieved", trace.WithAttributes(
		attribute.Int("notifications.count", len(notifications)),
	))

	return web.Respond(ctx, w, toAppNotificationsResponse(notifications, page, pageSize, hasMore), http.StatusOK)
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
