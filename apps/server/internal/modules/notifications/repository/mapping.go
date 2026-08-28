package notificationsrepository

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
)

type notificationRecord struct {
	ID           uuid.UUID
	RecipientID  uuid.UUID
	WorkspaceID  uuid.UUID
	Type         notificationssql.NotificationType
	EntityType   notificationssql.EntityType
	EntityID     uuid.UUID
	ActorID      uuid.UUID
	Title        string
	Message      []byte
	InAppEnabled bool
	CreatedAt    *time.Time
	ReadAt       *time.Time
}

func marshalNewNotification(notification notificationsdomain.NewNotification) ([]byte, string, error) {
	message, err := json.Marshal(notification.Message)
	if err != nil {
		return nil, "", fmt.Errorf("marshal notification message: %w", err)
	}
	dedupeKey := strings.TrimSpace(notification.DedupeKey)
	if dedupeKey == "" {
		dedupeKey = "notification:" + uuid.NewString()
	}
	return message, dedupeKey, nil
}

func toNotification(record notificationRecord) (notificationsdomain.Notification, error) {
	if record.CreatedAt == nil {
		return notificationsdomain.Notification{}, fmt.Errorf("notification %s has null created_at", record.ID)
	}
	notificationType, err := notificationsdomain.ParseNotificationType(string(record.Type))
	if err != nil {
		return notificationsdomain.Notification{}, fmt.Errorf("map notification type: %w", err)
	}
	entityType, err := notificationsdomain.ParseEntityType(string(record.EntityType))
	if err != nil {
		return notificationsdomain.Notification{}, fmt.Errorf("map notification entity type: %w", err)
	}
	var message notificationsdomain.NotificationMessage
	if err := json.Unmarshal(record.Message, &message); err != nil {
		return notificationsdomain.Notification{}, fmt.Errorf("decode notification message: %w", err)
	}
	return notificationsdomain.Notification{
		ID:           record.ID,
		RecipientID:  record.RecipientID,
		WorkspaceID:  record.WorkspaceID,
		InAppEnabled: record.InAppEnabled,
		Type:         notificationType,
		EntityType:   entityType,
		EntityID:     record.EntityID,
		ActorID:      record.ActorID,
		Title:        record.Title,
		Message:      message,
		CreatedAt:    *record.CreatedAt,
		ReadAt:       record.ReadAt,
	}, nil
}

func toPreferences(id, userID, workspaceID uuid.UUID, raw []byte, createdAt, updatedAt time.Time) (notificationsdomain.Preferences, error) {
	preferences := notificationsdomain.PreferenceSet{}
	if err := json.Unmarshal(raw, &preferences); err != nil {
		return notificationsdomain.Preferences{}, fmt.Errorf("decode notification preferences: %w", err)
	}
	return notificationsdomain.Preferences{
		ID:          id,
		UserID:      userID,
		WorkspaceID: workspaceID,
		Preferences: preferences.WithDefaults(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func defaultPreferencesJSON() ([]byte, error) {
	encoded, err := json.Marshal(notificationsdomain.DefaultPreferences())
	if err != nil {
		return nil, fmt.Errorf("encode default notification preferences: %w", err)
	}
	return encoded, nil
}
