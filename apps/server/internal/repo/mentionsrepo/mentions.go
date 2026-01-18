package mentionsrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

func (r *repo) SaveMentions(ctx context.Context, commentID uuid.UUID, userIDs []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.mentions.SaveMentions")
	defer span.End()

	if len(userIDs) == 0 {
		return nil
	}

	// First delete existing mentions for this comment
	if err := r.DeleteMentions(ctx, commentID); err != nil {
		return fmt.Errorf("failed to delete existing mentions: %w", err)
	}

	// Insert new mentions
	query := `
		INSERT INTO comment_mentions (comment_id, mentioned_user_id)
		VALUES (:comment_id, :mentioned_user_id)
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare statement for mentions", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		mention := DbCommentMention{
			CommentID:       commentID,
			MentionedUserID: userID,
		}
		if _, err := stmt.ExecContext(ctx, mention); err != nil {
			r.log.Error(ctx, "failed to insert mention", "error", err, "commentId", commentID, "userId", userID)
			return fmt.Errorf("failed to insert mention: %w", err)
		}
	}

	r.log.Info(ctx, "mentions saved successfully", "commentId", commentID, "count", len(userIDs))
	return nil
}

func (r *repo) DeleteMentions(ctx context.Context, commentID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.mentions.DeleteMentions")
	defer span.End()

	query := `DELETE FROM comment_mentions WHERE comment_id = :comment_id`

	params := map[string]any{
		"comment_id": commentID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare statement for deleting mentions", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, "failed to delete mentions", "error", err, "commentId", commentID)
		return fmt.Errorf("failed to delete mentions: %w", err)
	}

	r.log.Info(ctx, "mentions deleted successfully", "commentId", commentID)
	return nil
}

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
