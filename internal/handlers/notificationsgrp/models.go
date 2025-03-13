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
	Description *string    `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	ReadAt      *time.Time `json:"readAt"`
}

type AppNotificationPreference struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
	Type         string    `json:"type"`
	EmailEnabled bool      `json:"emailEnabled"`
	InAppEnabled bool      `json:"inAppEnabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
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

func toAppNotificationPreference(p notifications.CoreNotificationPreference) AppNotificationPreference {
	return AppNotificationPreference{
		ID:           p.ID,
		UserID:       p.UserID,
		WorkspaceID:  p.WorkspaceID,
		Type:         p.Type,
		EmailEnabled: p.EmailEnabled,
		InAppEnabled: p.InAppEnabled,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toAppNotificationPreferences(preferences []notifications.CoreNotificationPreference) []AppNotificationPreference {
	result := make([]AppNotificationPreference, len(preferences))
	for i, p := range preferences {
		result[i] = toAppNotificationPreference(p)
	}
	return result
}
