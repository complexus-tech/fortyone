package commentsrepo

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DbComment represents the database model for an dbComment.
type DbComment struct {
	ID          uuid.UUID        `db:"comment_id"`
	StoryID     uuid.UUID        `db:"story_id"`
	Parent      *uuid.UUID       `db:"parent_id"`
	UserID      uuid.UUID        `db:"commenter_id"`
	Comment     string           `db:"content"`
	CreatedAt   time.Time        `db:"created_at"`
	UpdatedAt   time.Time        `db:"updated_at"`
	SubComments *json.RawMessage `db:"sub_comments"`
}

type DbNewComment struct {
	StoryID uuid.UUID  `db:"story_id"`
	Parent  *uuid.UUID `db:"parent_id"`
	UserID  uuid.UUID  `db:"commenter_id"`
	Comment string     `db:"content"`
}
