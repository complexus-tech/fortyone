package mentionsrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"

	mentionssql "github.com/complexus-tech/projects-api/internal/modules/mentions/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maximumMentionsPerComment = 100

var (
	ErrNotConfigured       = errors.New("mentions repository is not configured")
	ErrCommentNotFound     = errors.New("comment was not found in the workspace")
	ErrInvalidMention      = errors.New("comment mention is invalid")
	ErrMentionTargetDenied = errors.New("mentioned user is not an active workspace member")
)

type Repository struct {
	pool       *pgxpool.Pool
	queries    *mentionssql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{
		pool: pool, queries: mentionssql.New(pool), transactor: platformdatabase.NewTransactor(pool),
	}
}

func (repository *Repository) SaveMentions(
	ctx context.Context,
	workspaceID, commentID uuid.UUID,
	userIDs []uuid.UUID,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	targets, err := validatedMentionTargets(workspaceID, commentID, userIDs)
	if err != nil {
		return err
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := mentionssql.New(tx)
		if err := lockComment(ctx, queries, workspaceID, commentID); err != nil {
			return err
		}
		if err := queries.DeleteCommentMentions(ctx, mentionssql.DeleteCommentMentionsParams{CommentID: commentID}); err != nil {
			return fmt.Errorf("delete comment mentions: %w", err)
		}
		if len(targets) == 0 {
			return nil
		}
		inserted, err := queries.InsertActiveWorkspaceCommentMentions(ctx, mentionssql.InsertActiveWorkspaceCommentMentionsParams{
			CommentID: commentID, WorkspaceID: workspaceID, MentionedUserIds: targets,
		})
		if err != nil {
			return fmt.Errorf("insert comment mentions: %w", err)
		}
		if inserted != int64(len(targets)) {
			return ErrMentionTargetDenied
		}
		return nil
	})
}

func (repository *Repository) DeleteMentions(ctx context.Context, workspaceID, commentID uuid.UUID) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if workspaceID == uuid.Nil || commentID == uuid.Nil {
		return ErrInvalidMention
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := mentionssql.New(tx)
		if err := lockComment(ctx, queries, workspaceID, commentID); err != nil {
			return err
		}
		if err := queries.DeleteCommentMentions(ctx, mentionssql.DeleteCommentMentionsParams{CommentID: commentID}); err != nil {
			return fmt.Errorf("delete comment mentions: %w", err)
		}
		return nil
	})
}

func (repository *Repository) GetMentions(ctx context.Context, workspaceID, commentID uuid.UUID) ([]uuid.UUID, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	if workspaceID == uuid.Nil || commentID == uuid.Nil {
		return nil, ErrInvalidMention
	}
	var userIDs []uuid.UUID
	err := repository.transactor.WithinTransaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		queries := mentionssql.New(tx)
		params := mentionssql.GetWorkspaceCommentForMentionsParams{CommentID: commentID, WorkspaceID: workspaceID}
		if _, err := queries.GetWorkspaceCommentForMentions(ctx, params); errors.Is(err, pgx.ErrNoRows) {
			return ErrCommentNotFound
		} else if err != nil {
			return fmt.Errorf("get comment for mentions: %w", err)
		}
		var err error
		userIDs, err = queries.ListWorkspaceCommentMentions(ctx, mentionssql.ListWorkspaceCommentMentionsParams{
			CommentID: commentID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("list comment mentions: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return append([]uuid.UUID(nil), userIDs...), nil
}

func (repository *Repository) configured() error {
	if repository == nil || repository.pool == nil || repository.queries == nil {
		return ErrNotConfigured
	}
	return nil
}

func lockComment(ctx context.Context, queries *mentionssql.Queries, workspaceID, commentID uuid.UUID) error {
	_, err := queries.LockWorkspaceCommentForMentions(ctx, mentionssql.LockWorkspaceCommentForMentionsParams{
		CommentID: commentID, WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCommentNotFound
	}
	if err != nil {
		return fmt.Errorf("lock comment for mentions: %w", err)
	}
	return nil
}

func validatedMentionTargets(workspaceID, commentID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	if workspaceID == uuid.Nil || commentID == uuid.Nil || len(userIDs) > maximumMentionsPerComment {
		return nil, ErrInvalidMention
	}
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	targets := make([]uuid.UUID, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == uuid.Nil {
			return nil, ErrInvalidMention
		}
		if _, duplicate := seen[userID]; duplicate {
			return nil, ErrInvalidMention
		}
		seen[userID] = struct{}{}
		targets = append(targets, userID)
	}
	slices.SortFunc(targets, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})
	return targets, nil
}
