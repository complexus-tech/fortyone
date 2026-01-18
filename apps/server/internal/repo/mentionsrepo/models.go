package mentionsrepo

import (
	"github.com/google/uuid"
)

// DbCommentMention represents the database model for comment mentions
type DbCommentMention struct {
	CommentID       uuid.UUID `db:"comment_id"`
	MentionedUserID uuid.UUID `db:"mentioned_user_id"`
}
