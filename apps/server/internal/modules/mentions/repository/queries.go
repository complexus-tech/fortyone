package mentionsrepository

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (r *repo) GetMentions(ctx context.Context, commentID uuid.UUID) ([]uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.mentions.GetMentions")
	defer span.End()

	query := `SELECT mentioned_user_id FROM comment_mentions WHERE comment_id = :comment_id`

	params := map[string]any{
		"comment_id": commentID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare statement for getting mentions", "error", err)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var userIDs []uuid.UUID
	if err := stmt.SelectContext(ctx, &userIDs, params); err != nil {
		r.log.Error(ctx, "failed to get mentions", "error", err, "commentId", commentID)
		return nil, fmt.Errorf("failed to get mentions: %w", err)
	}

	return userIDs, nil
}
