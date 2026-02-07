package links

import (
	"time"

	"github.com/google/uuid"
)

type CoreLink struct {
	LinkID    uuid.UUID
	Title     *string
	URL       string
	StoryID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CoreNewLink struct {
	Title   *string
	URL     string
	StoryID uuid.UUID
}

type CoreUpdateLink struct {
	Title *string
	URL   *string
}
