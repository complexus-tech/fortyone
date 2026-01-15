package comments

import (
	"time"

	"github.com/google/uuid"
)

type CoreComment struct {
	ID          uuid.UUID     `json:"comment_id"`
	StoryID     uuid.UUID     `json:"story_id"`
	Parent      *uuid.UUID    `json:"parent_id"`
	UserID      uuid.UUID     `json:"commenter_id"`
	Comment     string        `json:"content"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	SubComments []CoreComment `json:"sub_comments"`
}
