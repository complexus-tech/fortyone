package commentsrepository

import (
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
)

func toCoreComment(row commentsql.GetCommentForWorkspaceRow) commentsdomain.Comment {
	return commentFromValues(
		row.CommentID,
		row.StoryID,
		row.ParentID,
		row.CommenterID,
		row.Content,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func commentFromValues(
	commentID, storyID uuid.UUID,
	parentID *uuid.UUID,
	commenterID uuid.UUID,
	content string,
	createdAt, updatedAt time.Time,
) commentsdomain.Comment {
	return commentsdomain.Comment{
		ID: commentID, StoryID: storyID, Parent: parentID, UserID: commenterID,
		Comment: content, CreatedAt: createdAt.UTC(), UpdatedAt: updatedAt.UTC(),
		SubComments: []commentsdomain.Comment{},
	}
}
