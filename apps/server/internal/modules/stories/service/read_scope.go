package stories

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

var ErrStoryReadForbidden = errors.New("story read is not permitted")

// scopedReadRepository is the migration boundary for the SQLC read model.
// The fallback paths in Service exist only for legacy test doubles and for
// mutation-oriented repository capabilities that are migrated separately.
type scopedReadRepository interface {
	GetVisibleStory(context.Context, StoryReadScope, uuid.UUID) (CoreSingleStory, error)
	QueryVisibleStoryByRef(context.Context, StoryReadScope, string, int) (CoreSingleStory, error)
	ListMyVisibleStories(context.Context, StoryReadScope) ([]CoreStoryList, error)
	ListVisibleStoriesByCategory(context.Context, StoryReadScope, uuid.UUID, string, int, int, bool) ([]CoreStoryList, bool, error)
	ListVisibleStories(context.Context, StoryReadScope, CoreStoryFilters) ([]CoreStoryList, error)
	CountVisibleStories(context.Context, StoryReadScope) (int, error)
	ListVisibleGroupedStories(context.Context, StoryReadScope, CoreStoryQuery) ([]CoreStoryGroup, error)
	ListVisibleGroupStories(context.Context, StoryReadScope, string, CoreStoryQuery) ([]CoreStoryList, bool, error)
}

type legacyMyStoriesRepository interface {
	MyStories(context.Context, uuid.UUID) ([]CoreStoryList, error)
}

type legacyStoryReferenceRepository interface {
	QueryByRef(context.Context, uuid.UUID, string, int) (CoreSingleStory, error)
}

type legacyCategoryStoriesRepository interface {
	ListByCategory(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, int, int, bool) ([]CoreStoryList, bool, error)
}

func readScopeFromContext(ctx context.Context, workspaceID uuid.UUID) (StoryReadScope, error) {
	actor, err := auth.GetActor(ctx)
	if err != nil {
		return StoryReadScope{}, fmt.Errorf("%w: %v", ErrStoryReadForbidden, err)
	}
	if !actor.Scopes.Has(auth.ScopeStoriesRead) {
		return StoryReadScope{}, fmt.Errorf("%w: stories:read scope is required", ErrStoryReadForbidden)
	}
	if actor.WorkspaceID != uuid.Nil && actor.WorkspaceID != workspaceID {
		return StoryReadScope{}, fmt.Errorf("%w: actor is bound to another workspace", ErrStoryReadForbidden)
	}

	scope := StoryReadScope{
		ActorID:                actor.PrincipalID,
		WorkspaceID:            workspaceID,
		UnrestrictedTeamAccess: actor.TeamAccess.IsUnrestricted(),
		AllowedTeamIDs:         actor.TeamAccess.RestrictedTeamIDs(),
	}
	if err := scope.Validate(); err != nil {
		return StoryReadScope{}, fmt.Errorf("%w: %v", ErrStoryReadForbidden, err)
	}
	return scope, nil
}

func (s *Service) getVisibleStory(ctx context.Context, storyID, workspaceID uuid.UUID) (CoreSingleStory, error) {
	repository, migrated := s.repo.(scopedReadRepository)
	if !migrated {
		legacy, ok := s.repo.(legacyStoryRepository)
		if !ok {
			return CoreSingleStory{}, errors.New("story repository does not support scoped story reads")
		}
		return legacy.Get(ctx, storyID, workspaceID)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return CoreSingleStory{}, err
	}
	return repository.GetVisibleStory(ctx, scope, storyID)
}
