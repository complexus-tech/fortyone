package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"math"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maximumCommentPageSize = 100

func (r *repo) ListVisibleComments(
	ctx context.Context,
	scope storydomain.ReadScope,
	storyID uuid.UUID,
	page, pageSize int,
) ([]storydomain.Comment, bool, error) {
	if err := r.mutationConfigured(); err != nil {
		return nil, false, err
	}
	if err := scope.Validate(); err != nil {
		return nil, false, err
	}
	offset, limit, err := commentPageBounds(page, pageSize)
	if err != nil {
		return nil, false, err
	}
	rows, err := r.reads.ListVisibleStoryCommentRoots(ctx, storyreadsql.ListVisibleStoryCommentRootsParams{
		ActorID: scope.ActorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         append([]uuid.UUID(nil), scope.AllowedTeamIDs...),
		PageOffset:             offset,
		PageLimit:              limit,
	})
	if err != nil {
		return nil, false, fmt.Errorf("list visible story comment roots: %w", err)
	}
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	comments := make([]storydomain.Comment, len(rows))
	parentIDs := make([]uuid.UUID, len(rows))
	byParent := make(map[uuid.UUID]int, len(rows))
	for index, row := range rows {
		comments[index] = commentRootToDomain(row)
		parentIDs[index] = row.CommentID
		byParent[row.CommentID] = index
	}
	if len(parentIDs) == 0 {
		return comments, hasMore, nil
	}
	replies, err := r.reads.ListVisibleStoryCommentReplies(ctx, storyreadsql.ListVisibleStoryCommentRepliesParams{
		ActorID: scope.ActorID, StoryID: storyID, ParentIds: parentIDs, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         append([]uuid.UUID(nil), scope.AllowedTeamIDs...),
	})
	if err != nil {
		return nil, false, fmt.Errorf("list visible story comment replies: %w", err)
	}
	for _, row := range replies {
		if row.ParentID == nil {
			continue
		}
		parentIndex, exists := byParent[*row.ParentID]
		if !exists {
			continue
		}
		comments[parentIndex].SubComments = append(comments[parentIndex].SubComments, commentReplyToDomain(row))
	}
	return comments, hasMore, nil
}

func (r *repo) GetVisibleComment(
	ctx context.Context,
	scope storydomain.ReadScope,
	commentID, storyID uuid.UUID,
) (storydomain.Comment, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.Comment{}, err
	}
	if err := scope.Validate(); err != nil {
		return storydomain.Comment{}, err
	}
	if commentID == uuid.Nil || storyID == uuid.Nil {
		return storydomain.Comment{}, storydomain.ErrNotFound
	}
	row, err := r.reads.GetVisibleStoryComment(ctx, storyreadsql.GetVisibleStoryCommentParams{
		ActorID: scope.ActorID, CommentID: commentID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         append([]uuid.UUID(nil), scope.AllowedTeamIDs...),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.Comment{}, storydomain.ErrNotFound
	}
	if err != nil {
		return storydomain.Comment{}, fmt.Errorf("get visible story comment: %w", err)
	}
	return storydomain.Comment{
		ID: row.CommentID, StoryID: row.StoryID, Parent: row.ParentID, UserID: row.CommenterID,
		Comment: row.Content, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		SubComments: []storydomain.Comment{},
	}, nil
}

func commentPageBounds(page, pageSize int) (int32, int32, error) {
	if page < 1 || pageSize < 1 || pageSize > maximumCommentPageSize {
		return 0, 0, storydomain.ErrInvalidReadQuery
	}
	if int64(page-1) > int64(math.MaxInt32)/int64(pageSize) {
		return 0, 0, storydomain.ErrInvalidReadQuery
	}
	offset, err := safecast.Int64ToInt32(int64(page-1) * int64(pageSize))
	if err != nil {
		return 0, 0, storydomain.ErrInvalidReadQuery
	}
	limit, err := safecast.Int64ToInt32(int64(pageSize) + 1)
	if err != nil {
		return 0, 0, storydomain.ErrInvalidReadQuery
	}
	return offset, limit, nil
}

func commentRootToDomain(row storyreadsql.ListVisibleStoryCommentRootsRow) storydomain.Comment {
	return storydomain.Comment{
		ID: row.CommentID, StoryID: row.StoryID, Parent: row.ParentID, UserID: row.CommenterID,
		Comment: row.Content, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		SubComments: []storydomain.Comment{},
	}
}

func commentReplyToDomain(row storyreadsql.ListVisibleStoryCommentRepliesRow) storydomain.Comment {
	return storydomain.Comment{
		ID: row.CommentID, StoryID: row.StoryID, Parent: row.ParentID, UserID: row.CommenterID,
		Comment: row.Content, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		SubComments: []storydomain.Comment{},
	}
}
