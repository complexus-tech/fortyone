package notificationsrepo

import (
	"encoding/json"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/google/uuid"
)

type EntityType string
type NotificationType string

const (
	EntityTypeStory     EntityType = "story"
	EntityTypeComment   EntityType = "comment"
	EntityTypeObjective EntityType = "objective"
	EntityTypeKeyResult EntityType = "key_result"
)

const (
	NotificationTypeStoryUpdate     NotificationType = "story_update"
	NotificationTypeStoryComment    NotificationType = "story_comment"
	NotificationTypeCommentReply    NotificationType = "comment_reply"
	NotificationTypeObjectiveUpdate NotificationType = "objective_update"
	NotificationTypeKeyResultUpdate NotificationType = "key_result_update"
	NotificationTypeMention         NotificationType = "mention"
)

type dbNotification struct {
	ID          uuid.UUID        `db:"notification_id"`
	RecipientID uuid.UUID        `db:"recipient_id"`
	WorkspaceID uuid.UUID        `db:"workspace_id"`
	Type        NotificationType `db:"type"`
	EntityType  EntityType       `db:"entity_type"`
	EntityID    uuid.UUID        `db:"entity_id"`
	ActorID     uuid.UUID        `db:"actor_id"`
	Title       string           `db:"title"`
	Message     json.RawMessage  `db:"message"` // JSONB field for structured message
	CreatedAt   time.Time        `db:"created_at"`
	ReadAt      *time.Time       `db:"read_at"`
}

// New model for the JSONB-based notification preferences
type dbNotificationPreferences struct {
	ID          uuid.UUID       `db:"preference_id"`
	UserID      uuid.UUID       `db:"user_id"`
	WorkspaceID uuid.UUID       `db:"workspace_id"`
	Preferences json.RawMessage `db:"preferences"` // JSONB in PostgreSQL
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

type dbNewNotification struct {
	RecipientID uuid.UUID        `db:"recipient_id"`
	WorkspaceID uuid.UUID        `db:"workspace_id"`
	Type        NotificationType `db:"type"`
	EntityType  EntityType       `db:"entity_type"`
	EntityID    uuid.UUID        `db:"entity_id"`
	ActorID     uuid.UUID        `db:"actor_id"`
	Title       string           `db:"title"`
	Message     json.RawMessage  `db:"message"` // JSONB field for structured message
}

// Conversion functions
func toDBNewNotification(n notifications.CoreNewNotification) (dbNewNotification, error) {
	messageBytes, err := json.Marshal(n.Message)
	if err != nil {
		return dbNewNotification{}, err
	}

	return dbNewNotification{
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        NotificationType(n.Type),
		EntityType:  EntityType(n.EntityType),
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Title:       n.Title,
		Message:     messageBytes,
	}, nil
}

func toCoreNotification(n dbNotification) (notifications.CoreNotification, error) {
	var message notifications.NotificationMessage
	if err := json.Unmarshal(n.Message, &message); err != nil {
		return notifications.CoreNotification{}, err
	}

	return notifications.CoreNotification{
		ID:          n.ID,
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        string(n.Type),
		EntityType:  string(n.EntityType),
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Title:       n.Title,
		Message:     message,
		CreatedAt:   n.CreatedAt,
		ReadAt:      n.ReadAt,
	}, nil
}

func toCoreNotifications(ns []dbNotification) ([]notifications.CoreNotification, error) {
	result := make([]notifications.CoreNotification, len(ns))
	for i, n := range ns {
		coreNotif, err := toCoreNotification(n)
		if err != nil {
			return nil, err
		}
		result[i] = coreNotif
	}
	return result, nil
}

// Convert database notification preferences to core model
func toCoreNotificationPreferences(db dbNotificationPreferences) (notifications.CoreNotificationPreferences, error) {
	var prefs map[string]interface{}
	if err := json.Unmarshal(db.Preferences, &prefs); err != nil {
		return notifications.CoreNotificationPreferences{}, err
	}

	return notifications.CoreNotificationPreferences{
		ID:          db.ID,
		UserID:      db.UserID,
		WorkspaceID: db.WorkspaceID,
		Preferences: prefs,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}, nil
}
