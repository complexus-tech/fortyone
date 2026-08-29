package notificationshttp

import (
	"strings"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/google/uuid"
)

type AppNotification struct {
	ID          uuid.UUID                         `json:"id"`
	RecipientID uuid.UUID                         `json:"recipientId"`
	WorkspaceID uuid.UUID                         `json:"workspaceId"`
	Type        string                            `json:"type"`
	EntityType  string                            `json:"entityType"`
	EntityID    uuid.UUID                         `json:"entityId"`
	ActorID     uuid.UUID                         `json:"actorId"`
	Actor       *AppNotificationActor             `json:"actor,omitempty"`
	Title       string                            `json:"title"`
	Message     notifications.NotificationMessage `json:"message"`
	CreatedAt   time.Time                         `json:"createdAt"`
	ReadAt      *time.Time                        `json:"readAt"`
}

type AppNotificationActor struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
	IsSystem  bool      `json:"isSystem"`
}

type AppPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

type AppNotificationsResponse struct {
	Notifications []AppNotification `json:"notifications"`
	Pagination    AppPagination     `json:"pagination"`
}

type AppPortalNotificationActor struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl"`
}

type AppPortalNotificationFeedback struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Slug  string    `json:"slug"`
	Path  string    `json:"path"`
}

type AppPortalNotification struct {
	ID        uuid.UUID                         `json:"id"`
	Type      string                            `json:"type"`
	Title     string                            `json:"title"`
	Message   notifications.NotificationMessage `json:"message"`
	Actor     AppPortalNotificationActor        `json:"actor"`
	Feedback  AppPortalNotificationFeedback     `json:"feedback"`
	CreatedAt time.Time                         `json:"createdAt"`
	ReadAt    *time.Time                        `json:"readAt"`
}

type AppPortalNotificationsResponse struct {
	Notifications []AppPortalNotification `json:"notifications"`
	Pagination    AppPagination           `json:"pagination"`
}

type AppUnreadCount struct {
	Count int `json:"count"`
}

type NotificationChannel struct {
	Email bool `json:"email"`
	InApp bool `json:"inApp"`
}

// AppNotificationPreferences represents the notification preferences for a user in a workspace
type AppNotificationPreferences struct {
	ID          uuid.UUID                      `json:"id"`
	UserID      uuid.UUID                      `json:"userId"`
	WorkspaceID uuid.UUID                      `json:"workspaceId"`
	Preferences map[string]NotificationChannel `json:"preferences"`
	CreatedAt   time.Time                      `json:"createdAt"`
	UpdatedAt   time.Time                      `json:"updatedAt"`
}

type AppUpdatePreference struct {
	EmailEnabled *bool `json:"emailEnabled,omitempty"`
	InAppEnabled *bool `json:"inAppEnabled,omitempty"`
}

func toAppNotification(n notifications.CoreNotification) AppNotification {
	n = n.Public()
	n.Message = normalizeSystemActorMessage(n.Message, n.Actor)
	return AppNotification{
		ID:          n.ID,
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        string(n.Type),
		EntityType:  string(n.EntityType),
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Actor:       toAppNotificationActor(n.Actor),
		Title:       n.Title,
		Message:     n.Message,
		CreatedAt:   n.CreatedAt,
		ReadAt:      n.ReadAt,
	}
}

func normalizeSystemActorMessage(message notifications.NotificationMessage, actor *notifications.CoreNotificationActor) notifications.NotificationMessage {
	if actor == nil || !actor.IsSystem {
		return message
	}

	for _, prefix := range []string{actor.FullName, actor.Username, "Maya"} {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(message.Template, prefix+" ") {
			message.Template = strings.TrimSpace(strings.TrimPrefix(message.Template, prefix))
			break
		}
	}

	return message
}

func toAppNotificationActor(actor *notifications.CoreNotificationActor) *AppNotificationActor {
	if actor == nil {
		return nil
	}

	return &AppNotificationActor{
		ID:        actor.ID,
		Username:  actor.Username,
		FullName:  actor.FullName,
		AvatarURL: actor.AvatarURL,
		IsActive:  actor.IsActive,
		IsSystem:  actor.IsSystem,
	}
}

func toAppNotifications(ns []notifications.CoreNotification) []AppNotification {
	result := make([]AppNotification, len(ns))
	for i, n := range ns {
		result[i] = toAppNotification(n)
	}
	return result
}

func toAppNotificationsResponse(ns []notifications.CoreNotification, page, pageSize int, hasMore bool) AppNotificationsResponse {
	return AppNotificationsResponse{
		Notifications: toAppNotifications(ns),
		Pagination: AppPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: page + 1,
		},
	}
}

func toAppPortalNotificationsResponse(ns []notifications.CorePortalNotification, page, pageSize int, hasMore bool) AppPortalNotificationsResponse {
	result := make([]AppPortalNotification, 0, len(ns))
	for _, portalNotification := range ns {
		notification := portalNotification.Notification.Public()
		result = append(result, AppPortalNotification{
			ID:      notification.ID,
			Type:    string(notification.Type),
			Title:   notification.Title,
			Message: notification.Message,
			Actor: AppPortalNotificationActor{
				ID:        notification.ActorID,
				Name:      portalNotification.ActorName,
				AvatarURL: portalNotification.ActorAvatar,
			},
			Feedback: AppPortalNotificationFeedback{
				ID:    notification.EntityID,
				Title: portalNotification.FeedbackTitle,
				Slug:  portalNotification.FeedbackSlug,
				Path:  "/feedback/" + portalNotification.FeedbackSlug,
			},
			CreatedAt: notification.CreatedAt,
			ReadAt:    notification.ReadAt,
		})
	}

	return AppPortalNotificationsResponse{
		Notifications: result,
		Pagination: AppPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: page + 1,
		},
	}
}

// Convert core notification preferences to API model
func toAppNotificationPreferences(p notifications.CoreNotificationPreferences) AppNotificationPreferences {
	appPrefs := AppNotificationPreferences{
		ID:          p.ID,
		UserID:      p.UserID,
		WorkspaceID: p.WorkspaceID,
		Preferences: make(map[string]NotificationChannel),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	for key, channels := range p.Preferences {
		appPrefs.Preferences[string(key)] = NotificationChannel{
			Email: channels.Email,
			InApp: channels.InApp,
		}
	}

	return appPrefs
}
