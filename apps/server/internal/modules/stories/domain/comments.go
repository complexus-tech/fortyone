package domain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// Comment is the stories module's read contract for one node in a visible
// story-comment tree. The comments module owns mutation persistence; bootstrap
// maps its result into this caller-owned shape.
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

// CreateCommentCommand is the narrow caller-owned port used by compatibility
// story workflows. It prevents the stories service from importing a sibling
// module's concrete service or command types.
type CreateCommentCommand struct {
	WorkspaceID      uuid.UUID
	StoryID          uuid.UUID
	ParentID         *uuid.UUID
	Actor            platformauth.Actor
	Content          string
	MentionedUserIDs []uuid.UUID
}
