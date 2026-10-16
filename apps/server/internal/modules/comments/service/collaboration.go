package comments

import (
	"context"
	"fmt"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type StoryScope = commentsdomain.StoryScope
type StoryState = commentsdomain.StoryState
type CommentState = commentsdomain.CommentState
type ReactionCommand = commentsdomain.ReactionCommand
type ResolutionCommand = commentsdomain.ResolutionCommand
type ReadCommand = commentsdomain.ReadCommand

type collaborationStore interface {
	GetStoryState(context.Context, commentsdomain.StoryScope) (commentsdomain.StoryState, error)
	SetReaction(context.Context, commentsdomain.ReactionCommand) (commentsdomain.CommentState, error)
	SetResolution(context.Context, commentsdomain.ResolutionCommand) (commentsdomain.CommentState, error)
	MarkRead(context.Context, commentsdomain.ReadCommand) (commentsdomain.StoryState, error)
}

func (s *Service) GetStoryState(ctx context.Context, scope StoryScope) (StoryState, error) {
	if err := validateStoryScope(scope, platformauth.ScopeCommentsRead); err != nil {
		return StoryState{}, err
	}
	repository, err := s.collaborationStore()
	if err != nil {
		return StoryState{}, err
	}
	return repository.GetStoryState(ctx, scope)
}

func (s *Service) SetReaction(ctx context.Context, command ReactionCommand) (CommentState, error) {
	if err := validateActorScope(command.Scope); err != nil {
		return CommentState{}, err
	}
	if !commentsdomain.SupportsReaction(command.Emoji) {
		return CommentState{}, fmt.Errorf("%w: unsupported reaction", ErrInvalidComment)
	}
	repository, err := s.collaborationStore()
	if err != nil {
		return CommentState{}, err
	}
	return repository.SetReaction(ctx, command)
}

func (s *Service) SetResolution(ctx context.Context, command ResolutionCommand) (CommentState, error) {
	if err := validateActorScope(command.Scope); err != nil {
		return CommentState{}, err
	}
	repository, err := s.collaborationStore()
	if err != nil {
		return CommentState{}, err
	}
	return repository.SetResolution(ctx, command)
}

func (s *Service) MarkRead(ctx context.Context, command ReadCommand) (StoryState, error) {
	if err := validateStoryScope(command.Scope, platformauth.ScopeCommentsRead); err != nil {
		return StoryState{}, err
	}
	if command.ObservedAt.IsZero() {
		return StoryState{}, fmt.Errorf("%w: observedAt is required", ErrInvalidComment)
	}
	if command.ThreadID != nil && *command.ThreadID == uuid.Nil {
		return StoryState{}, fmt.Errorf("%w: threadId must identify a root comment", ErrInvalidComment)
	}
	repository, err := s.collaborationStore()
	if err != nil {
		return StoryState{}, err
	}
	return repository.MarkRead(ctx, command)
}

func (s *Service) collaborationStore() (collaborationStore, error) {
	repository, ok := s.repo.(collaborationStore)
	if !ok {
		return nil, fmt.Errorf("%w: collaboration store is unavailable", ErrInvalidComment)
	}
	return repository, nil
}

func validateStoryScope(scope StoryScope, required platformauth.Scope) error {
	if scope.StoryID == uuid.Nil || scope.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: story and workspace are required", ErrInvalidComment)
	}
	if err := scope.Actor.Validate(); err != nil || scope.Actor.WorkspaceID != scope.WorkspaceID ||
		!scope.Actor.IsUserActor() || !scope.Actor.Scopes.Has(required) {
		return ErrForbidden
	}
	return nil
}
