package notificationsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/google/uuid"
)

type AppNotification struct {
	ID          uuid.UUID  `json:"id"`
	RecipientID uuid.UUID  `json:"recipientId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Type        string     `json:"type"`
	EntityType  string     `json:"entityType"`
	EntityID    uuid.UUID  `json:"entityId"`
	ActorID     uuid.UUID  `json:"actorId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	ReadAt      *time.Time `json:"readAt"`
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
	return AppNotification{
		ID:          n.ID,
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        n.Type,
		EntityType:  n.EntityType,
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Title:       n.Title,
		Description: n.Description,
		CreatedAt:   n.CreatedAt,
		ReadAt:      n.ReadAt,
	}
}

func toAppNotifications(ns []notifications.CoreNotification) []AppNotification {
	result := make([]AppNotification, len(ns))
	for i, n := range ns {
		result[i] = toAppNotification(n)
	}
	return result
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

	// Convert between internal and API representations
	for key, channels := range p.Preferences {
		appPrefs.Preferences[key] = NotificationChannel{
			Email: channels.Email,
			InApp: channels.InApp,
		}
	}

	return appPrefs
}
