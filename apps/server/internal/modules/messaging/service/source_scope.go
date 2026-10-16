package messaging

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type SourceSurface string

const SourceSurfaceStoryComment SourceSurface = "story_comment"

// SourceScope restricts both reads and writes to the story that owns a native
// comment invocation. Only a server adapter may populate it.
type SourceScope struct {
	Surface SourceSurface `json:"surface"`
	StoryID uuid.UUID     `json:"story_id"`
	TeamID  uuid.UUID     `json:"team_id"`
}

func normalizedSourceScope(scope *SourceScope) (*SourceScope, error) {
	if scope == nil {
		return nil, nil
	}
	if scope.Surface != SourceSurfaceStoryComment || scope.StoryID == uuid.Nil || scope.TeamID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid native story source scope", ErrInvalidRequest)
	}
	copy := *scope
	return &copy, nil
}

func sourceAllowsTool(scope *SourceScope, name string) bool {
	if scope == nil {
		return true
	}
	switch name {
	case toolGetStory, toolListStatuses, toolListTeamMembers, toolUpdateStory:
		return scope.Surface == SourceSurfaceStoryComment
	default:
		return false
	}
}

func validateMutationSource(scope ToolScope, claims storyMutationClaims) error {
	if scope.SourceScope == nil && claims.SourceScope == nil {
		return nil
	}
	if scope.SourceScope == nil || claims.SourceScope == nil || *scope.SourceScope != *claims.SourceScope ||
		claims.Operation != StoryMutationUpdate || claims.StoryID == nil || *claims.StoryID != scope.SourceScope.StoryID || claims.TeamID != scope.SourceScope.TeamID {
		return fmt.Errorf("%w: proposal is outside the source story", ErrInvalidConfirmation)
	}
	return nil
}

// sourceStory reads the authorized entity before examining an untrusted
// reference. A reference to another story must never cause that story to load.
func (e *FortyOneToolExecutor) sourceStory(ctx context.Context, scope ToolScope, joined map[uuid.UUID]messagingTeam, reference *string, storyID *string) (messagingStory, error) {
	source := scope.SourceScope
	if source == nil || e.mutations == nil {
		return messagingStory{}, fmt.Errorf("%w: native story tools are unavailable", ErrInvalidRequest)
	}
	team, allowed := joined[source.TeamID]
	if !allowed {
		return messagingStory{}, ErrTeamNotAccessible
	}
	if storyID != nil && strings.TrimSpace(*storyID) != source.StoryID.String() {
		return messagingStory{}, fmt.Errorf("%w: only the current story is available", ErrTeamNotAccessible)
	}
	story, err := e.mutations.stories.Get(ctx, source.StoryID, scope.WorkspaceID)
	if err != nil {
		return messagingStory{}, fmt.Errorf("load source story: %w", err)
	}
	if story.ID != source.StoryID || story.Team != source.TeamID || story.Workspace != scope.WorkspaceID || story.DeletedAt != nil || story.ArchivedAt != nil {
		return messagingStory{}, fmt.Errorf("%w: source story is unavailable", ErrTeamNotAccessible)
	}
	if reference != nil {
		canonical, _, err := accessibleStoryReference(*reference, joined)
		if err != nil || canonical != storyReference(team.Code, story.SequenceID) {
			return messagingStory{}, fmt.Errorf("%w: only the current story is available", ErrTeamNotAccessible)
		}
	}
	return story, nil
}
