package notifications

import (
	"time"

	"github.com/google/uuid"
)

type NotificationMessage struct {
	Template  string              `json:"template"`
	Variables map[string]Variable `json:"variables"`
}

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"` // "actor", "assignee", "field", "value", "date"
}

// CoreNewNotification represents a new notification to be created.
type CoreNewNotification struct {
	RecipientID uuid.UUID           `json:"recipient_id"`
	WorkspaceID uuid.UUID           `json:"workspace_id"`
	Type        string              `json:"type"`
	EntityType  string              `json:"entity_type"`
	EntityID    uuid.UUID           `json:"entity_id"`
	ActorID     uuid.UUID           `json:"actor_id"`
	Title       string              `json:"title"`
	Message     NotificationMessage `json:"message"`
}

// CoreNotification represents a notification.
type CoreNotification struct {
	ID          uuid.UUID           `json:"id"`
	RecipientID uuid.UUID           `json:"recipient_id"`
	WorkspaceID uuid.UUID           `json:"workspace_id"`
	Type        string              `json:"type"`
	EntityType  string              `json:"entity_type"`
	EntityID    uuid.UUID           `json:"entity_id"`
	ActorID     uuid.UUID           `json:"actor_id"`
	Title       string              `json:"title"`
	Message     NotificationMessage `json:"message"`
	CreatedAt   time.Time           `json:"created_at"`
	ReadAt      *time.Time          `json:"read_at"`
}

// CoreNotificationPreferences represents a user's notification preferences.
type CoreNotificationPreferences struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Preferences map[string]interface{} `json:"preferences"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
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

// NotificationChannels represents the different delivery channels for a notification type
type NotificationChannels struct {
	Email bool `json:"email"`
	InApp bool `json:"in_app"`
}
