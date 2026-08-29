package domain

import (
	"time"

	"github.com/google/uuid"
)

// Activity is the persistence-neutral compatibility command used by callers
// that append one human-readable story activity outside a full story mutation.
type Activity struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	OldValue     any       `json:"oldValue"`
	NewValue     any       `json:"newValue"`
	Reason       *string   `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
}

// StoryLink is the persistence-neutral projection returned by the story link
// support read. The links module remains an HTTP/service concern.
type StoryLink struct {
	ID        uuid.UUID
	Title     *string
	URL       string
	StoryID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ActivityUser struct {
	ID        uuid.UUID
	Username  string
	FullName  string
	AvatarURL string
	IsActive  bool
	IsSystem  bool
}

type ActivityWithUser struct {
	ID           uuid.UUID
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	OldValue     any
	NewValue     any
	Reason       *string
	CreatedAt    time.Time
	WorkspaceID  uuid.UUID
	User         ActivityUser
}
