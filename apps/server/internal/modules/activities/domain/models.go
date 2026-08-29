package activitiesdomain

import (
	"time"

	"github.com/google/uuid"
)

// Activity is an immutable audit entry describing a story change.
type Activity struct {
	ID           uuid.UUID
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	CreatedAt    time.Time
	WorkspaceID  uuid.UUID
	User         UserDetails
}

// UserDetails contains the active account presentation attached to an activity.
type UserDetails struct {
	ID        uuid.UUID
	Username  string
	FullName  string
	AvatarURL string
	IsActive  bool
}

// NewActivity is the complete immutable payload required to append an activity.
type NewActivity struct {
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	WorkspaceID  uuid.UUID
}

// Filters bounds the activity time window.
type Filters struct {
	StartDate time.Time
	EndDate   time.Time
}
