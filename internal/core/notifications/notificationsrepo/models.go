package notificationsrepo

import (
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
	CreatedAt   time.Time        `db:"created_at"`
	ReadAt      *time.Time       `db:"read_at"`
}

type dbNotificationPreference struct {
	ID           uuid.UUID        `db:"preference_id"`
	UserID       uuid.UUID        `db:"user_id"`
	WorkspaceID  uuid.UUID        `db:"workspace_id"`
	Type         NotificationType `db:"notification_type"`
	EmailEnabled bool             `db:"email_enabled"`
	InAppEnabled bool             `db:"in_app_enabled"`
	CreatedAt    time.Time        `db:"created_at"`
	UpdatedAt    time.Time        `db:"updated_at"`
}

type dbNewNotification struct {
	RecipientID uuid.UUID        `db:"recipient_id"`
	WorkspaceID uuid.UUID        `db:"workspace_id"`
	Type        NotificationType `db:"type"`
	EntityType  EntityType       `db:"entity_type"`
	EntityID    uuid.UUID        `db:"entity_id"`
	ActorID     uuid.UUID        `db:"actor_id"`
	Title       string           `db:"title"`
}

// Conversion functions
func toDBNewNotification(n notifications.CoreNewNotification) dbNewNotification {
	return dbNewNotification{
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        NotificationType(n.Type),
		EntityType:  EntityType(n.EntityType),
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Title:       n.Title,
	}
}

func toCoreNotification(n dbNotification) notifications.CoreNotification {
	return notifications.CoreNotification{
		ID:          n.ID,
		RecipientID: n.RecipientID,
		WorkspaceID: n.WorkspaceID,
		Type:        string(n.Type),
		EntityType:  string(n.EntityType),
		EntityID:    n.EntityID,
		ActorID:     n.ActorID,
		Title:       n.Title,
		CreatedAt:   n.CreatedAt,
		ReadAt:      n.ReadAt,
	}
}

func toCoreNotifications(ns []dbNotification) []notifications.CoreNotification {
	result := make([]notifications.CoreNotification, len(ns))
	for i, n := range ns {
		result[i] = toCoreNotification(n)
	}
	return result
}

func toCoreNotificationPreference(p dbNotificationPreference) notifications.CoreNotificationPreference {
	return notifications.CoreNotificationPreference{
		ID:           p.ID,
		UserID:       p.UserID,
		WorkspaceID:  p.WorkspaceID,
		Type:         string(p.Type),
		EmailEnabled: p.EmailEnabled,
		InAppEnabled: p.InAppEnabled,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toCoreNotificationPreferences(preferences []dbNotificationPreference) []notifications.CoreNotificationPreference {
	result := make([]notifications.CoreNotificationPreference, len(preferences))
	for i, p := range preferences {
		result[i] = toCoreNotificationPreference(p)
	}
	return result
}
