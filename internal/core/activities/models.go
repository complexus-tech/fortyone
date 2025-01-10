package activities

import (
	"time"

	"github.com/google/uuid"
)

// CoreActivity represents an activity in the system
type CoreActivity struct {
	ID           uuid.UUID `db:"activity_id"`
	StoryID      uuid.UUID `db:"story_id"`
	UserID       uuid.UUID `db:"user_id"`
	Type         string    `db:"activity_type"`
	Field        string    `db:"field_changed"`
	CurrentValue string    `db:"current_value"`
	CreatedAt    time.Time `db:"created_at"`
	WorkspaceID  uuid.UUID `db:"workspace_id"`
}

// CoreNewActivity represents the data needed to create a new activity
type CoreNewActivity struct {
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	WorkspaceID  uuid.UUID
}
