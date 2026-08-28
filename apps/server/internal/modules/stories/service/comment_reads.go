package stories

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// scopedCommentReadRepository keeps story-comment reads on the same live
// workspace, membership, team, and credential fence as the owning story.
type scopedCommentReadRepository interface {
	ListVisibleComments(context.Context, StoryReadScope, uuid.UUID, int, int) ([]CoreComment, bool, error)
	GetVisibleComment(context.Context, StoryReadScope, uuid.UUID, uuid.UUID) (CoreComment, error)
}

type legacyCommentReadRepository interface {
	GetComments(context.Context, uuid.UUID, uuid.UUID, int, int) ([]CoreComment, bool, error)
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreComment, error)
}

func (s *Service) listVisibleComments(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	page, pageSize int,
) ([]CoreComment, bool, error) {
	repository, migrated := s.repo.(scopedCommentReadRepository)
	if !migrated {
		legacy, ok := s.repo.(legacyCommentReadRepository)
		if !ok {
			return nil, false, errors.New("story repository does not support comment reads")
		}
		return legacy.GetComments(ctx, storyID, workspaceID, page, pageSize)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return nil, false, err
	}
	return repository.ListVisibleComments(ctx, scope, storyID, page, pageSize)
}

func (s *Service) getVisibleComment(
	ctx context.Context,
	commentID, storyID, workspaceID uuid.UUID,
) (CoreComment, error) {
	repository, migrated := s.repo.(scopedCommentReadRepository)
	if !migrated {
		legacy, ok := s.repo.(legacyCommentReadRepository)
		if !ok {
			return CoreComment{}, errors.New("story repository does not support comment reads")
		}
		return legacy.GetComment(ctx, commentID, storyID, workspaceID)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return CoreComment{}, err
	}
	return repository.GetVisibleComment(ctx, scope, commentID, storyID)
}
