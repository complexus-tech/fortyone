package notifications

import (
	"time"

	"github.com/google/uuid"
)

type CoreNotification struct {
	ID          uuid.UUID
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
	Type        string
	EntityType  string
	EntityID    uuid.UUID
	ActorID     uuid.UUID
	Title       string
	Description *string
	IsRead      bool
	CreatedAt   time.Time
	ReadAt      *time.Time
}

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

type CoreNewNotification struct {
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
	Type        string
	EntityType  string
	EntityID    uuid.UUID
	ActorID     uuid.UUID
	Title       string
	Description *string
}
