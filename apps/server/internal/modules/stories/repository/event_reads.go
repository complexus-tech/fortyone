package storiesrepository

import (
	"context"
	"errors"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetEventStoryTitle(ctx context.Context, actorID, storyID, workspaceID uuid.UUID) (string, error) {
	if actorID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil {
		return "", storydomain.ErrInvalidReadQuery
	}
	title, err := r.reads.GetEventStoryTitle(ctx, storyreadsql.GetEventStoryTitleParams{
		ActorID: actorID, StoryID: storyID, WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storydomain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get event story title: %w", err)
	}
	return title, nil
}

func (r *repo) GetSystemStatusCategory(ctx context.Context, scope storydomain.MutationScope, statusID uuid.UUID) (string, error) {
	if err := scope.Validate(); err != nil {
		return "", err
	}
	if scope.Actor.Kind != auth.PrincipalSystem || statusID == uuid.Nil {
		return "", storydomain.ErrMutationForbidden
	}
	return r.reads.GetSystemStoryStatusCategory(ctx, storyreadsql.GetSystemStoryStatusCategoryParams{
		ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID, StatusID: statusID,
		AllTeams: scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
}

func (r *repo) GetSystemComment(ctx context.Context, scope storydomain.MutationScope, commentID, storyID uuid.UUID) (storydomain.Comment, error) {
	if err := scope.Validate(); err != nil {
		return storydomain.Comment{}, err
	}
	if scope.Actor.Kind != auth.PrincipalSystem || commentID == uuid.Nil || storyID == uuid.Nil {
		return storydomain.Comment{}, storydomain.ErrMutationForbidden
	}
	row, err := r.reads.GetSystemStoryComment(ctx, storyreadsql.GetSystemStoryCommentParams{
		ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID, StoryID: storyID, CommentID: commentID,
		AllTeams: scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.Comment{}, storydomain.ErrNotFound
	}
	if err != nil {
		return storydomain.Comment{}, fmt.Errorf("get system story comment: %w", err)
	}
	return storydomain.Comment{
		ID: row.CommentID, StoryID: row.StoryID, Parent: row.ParentID, UserID: row.CommenterID,
		Comment: row.Content, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		SubComments: []storydomain.Comment{},
	}, nil
}
