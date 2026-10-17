package commentsrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	queries "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) MarkSeen(ctx context.Context, scope domain.StoryScope, commentID uuid.UUID) error {
	return r.withinTransaction(ctx, func(q queries.Querier) error {
		_, err := q.MarkStoryCommentSeen(ctx, queries.MarkStoryCommentSeenParams{
			WorkspaceID: scope.WorkspaceID, StoryID: scope.StoryID, CommentID: commentID,
			ActorID: scope.Actor.PrincipalID, TeamAccessUnrestricted: scope.Actor.TeamAccess.IsUnrestricted(),
			AllowedTeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
		})
		if err != nil {
			return mapCommentWriteError("acknowledge visible comment", err)
		}
		return nil
	})
}
