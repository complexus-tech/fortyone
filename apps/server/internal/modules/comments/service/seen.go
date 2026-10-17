package comments

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type commentSeenStore interface {
	MarkSeen(context.Context, StoryScope, uuid.UUID) error
}

// MarkSeen records an exact visible response, independent of thread read cutoffs.
func (s *Service) MarkSeen(ctx context.Context, scope StoryScope, commentID uuid.UUID) error {
	if err := validateStoryScope(scope, auth.ScopeCommentsRead); err != nil {
		return err
	}
	if commentID == uuid.Nil {
		return ErrInvalidComment
	}
	repository, ok := s.repo.(commentSeenStore)
	if !ok {
		return fmt.Errorf("comment visibility store unavailable")
	}
	return repository.MarkSeen(ctx, scope, commentID)
}
