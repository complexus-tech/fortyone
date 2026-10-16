package stories

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// AuthorizeAttachmentRead uses current story visibility before issuing an
// attachment signature, without hydrating the story's related records.
func (s *Service) AuthorizeAttachmentRead(ctx context.Context, storyID, workspaceID uuid.UUID) error {
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return err
	}
	repository, ok := s.repo.(interface {
		AuthorizeVisibleStoryAttachmentRead(context.Context, StoryReadScope, uuid.UUID) error
	})
	if !ok {
		return errors.New("story repository does not support scoped attachment reads")
	}
	return repository.AuthorizeVisibleStoryAttachmentRead(ctx, scope, storyID)
}
