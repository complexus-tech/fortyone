package linksdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("link not found")

type Link struct {
	LinkID    uuid.UUID
	Title     *string
	URL       string
	StoryID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateLink struct {
	Title       *string
	URL         string
	StoryID     uuid.UUID
	WorkspaceID uuid.UUID
}

type UpdateLink struct {
	Title *string
	URL   *string
}
