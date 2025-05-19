package notifications

import (
	"time"

	"github.com/google/uuid"
)

type CoreNotification struct {
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

// CoreNotificationPreference represents a single notification preference
// Legacy model - kept for backward compatibility
type CoreNotificationPreference struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	WorkspaceID  uuid.UUID
	Type         string
	EmailEnabled bool
	InAppEnabled bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CoreNotificationPreferences represents all notification preferences for a user in a workspace
// using a JSONB structure for better flexibility
type CoreNotificationPreferences struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Preferences map[string]NotificationChannels
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NotificationChannels represents the different delivery channels for a notification type
type NotificationChannels struct {
	Email bool `json:"email"`
	InApp bool `json:"in_app"`
}

type CoreNewNotification struct {
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
	Type        string
	EntityType  string
	EntityID    uuid.UUID
	ActorID     uuid.UUID
	Title       string
	Description string
}
