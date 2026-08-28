package commentsdomain

import (
	"time"

	"github.com/google/uuid"
)

const (
	MaximumContentRunes = 10_000
	MaximumMentions     = 100
)

// Comment is the persistence-neutral representation of one story comment.
// Content is intentionally excluded from outbound integration events; callers
// receive it only through authorized product reads.
type Comment struct {
	ID          uuid.UUID  `json:"comment_id"`
	StoryID     uuid.UUID  `json:"story_id"`
	Parent      *uuid.UUID `json:"parent_id"`
	UserID      uuid.UUID  `json:"commenter_id"`
	Comment     string     `json:"content"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	SubComments []Comment  `json:"sub_comments"`
}
