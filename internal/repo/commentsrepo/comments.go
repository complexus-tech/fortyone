package commentsrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		log: log,
		db:  db,
	}
}

func (r *repo) UpdateComment(ctx context.Context, commentID uuid.UUID, comment string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.comments.UpdateComment")
	defer span.End()

	span.SetAttributes(attribute.String("commentId", commentID.String()))

	query := `
		UPDATE story_comments
		SET
			content = :comment,
			updated_at = NOW()
		WHERE comment_id = :comment_id
	`

	params := map[string]interface{}{
		"comment_id": commentID,
		"comment":    comment,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, "error updating comment", err)
		return err
	}

	r.log.Info(ctx, "comment updated successfully", "commentId", commentID)

	return nil
}

func (r *repo) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.comments.DeleteComment")
	defer span.End()

	span.SetAttributes(attribute.String("commentId", commentID.String()))

	// This will delete the comment and all child comments because of the cascade delete on the parent_id column
	query := `
		DELETE FROM story_comments WHERE comment_id = :comment_id
	`
	params := map[string]interface{}{
		"comment_id": commentID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting comment #%s", commentID), "commentId", commentID)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete comment: %s", err), "commentId", commentID)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Comment #%s deleted successfully", commentID), "commentId", commentID)
	span.AddEvent("Comment deleted.", trace.WithAttributes(attribute.String("comment.id", commentID.String())))

	return nil
}

func (r *repo) GetComment(ctx context.Context, commentID uuid.UUID) (comments.CoreComment, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.comments.GetComment")
	defer span.End()

	span.SetAttributes(attribute.String("commentId", commentID.String()))

	query := `
		SELECT comment_id, story_id, commenter_id,
		content, parent_id, created_at, updated_at
		FROM story_comments 
		WHERE comment_id = :comment_id
	`

	params := map[string]any{
		"comment_id": commentID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return comments.CoreComment{}, err
	}
	defer stmt.Close()

	var dbComment DbComment
	if err := stmt.GetContext(ctx, &dbComment, params); err != nil {
		r.log.Error(ctx, "error getting comment", err)
		return comments.CoreComment{}, err
	}

	return toCoreComment(dbComment), nil
}

// toCoreComment converts a DbComment to a CoreComment
func toCoreComment(dbComment DbComment) comments.CoreComment {
	return comments.CoreComment{
		ID:          dbComment.ID,
		StoryID:     dbComment.StoryID,
		Parent:      dbComment.Parent,
		UserID:      dbComment.UserID,
		Comment:     dbComment.Comment,
		CreatedAt:   dbComment.CreatedAt,
		UpdatedAt:   dbComment.UpdatedAt,
		SubComments: []comments.CoreComment{}, // Empty for single comment fetch
	}
}
