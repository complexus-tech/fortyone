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

	// User details (populated when fetching activities)
	User UserDetails `json:"user"`
}

// UserDetails represents basic user information for activities
type UserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
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
