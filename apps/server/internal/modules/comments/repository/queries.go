package commentsrepository

import (
	"context"
	"errors"
	"fmt"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
)

func (r *Repository) GetComment(
	ctx context.Context,
	query commentsdomain.GetQuery,
) (commentsdomain.Comment, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.comments.GetComment")
	defer span.End()

	span.SetAttributes(
		attribute.String("commentId", query.CommentID.String()),
		attribute.String("workspaceId", query.WorkspaceID.String()),
	)

	comment, err := r.queries.GetCommentForWorkspace(ctx, commentsql.GetCommentForWorkspaceParams{
		CommentID:   query.CommentID,
		WorkspaceID: query.WorkspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commentsdomain.Comment{}, commentsdomain.ErrNotFound
		}
		r.log.Error(ctx, "error getting comment", "error", err)
		return commentsdomain.Comment{}, fmt.Errorf("get comment: %w", err)
	}

	return toCoreComment(comment), nil
}
