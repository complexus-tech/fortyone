package commentsrepo

import (
	"context"
	"fmt"

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

	query := `
		DELETE FROM story_comments
		WHERE comment_id = :comment_id
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
