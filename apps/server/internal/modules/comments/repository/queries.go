package commentsrepository

import (
	"context"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

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
