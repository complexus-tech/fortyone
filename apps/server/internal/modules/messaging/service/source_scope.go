package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type SourceSurface string

const (
	SourceSurfaceStoryComment SourceSurface = "story_comment"
	SourceSurfaceChatMessage  SourceSurface = "chat_message"
)

// SourceScope restricts both reads and writes to the story that owns a native
// comment invocation. Only a server adapter may populate it.
type SourceScope struct {
	Surface        SourceSurface `json:"surface"`
	StoryID        uuid.UUID     `json:"story_id"`
	TeamID         uuid.UUID     `json:"team_id"`
	ConversationID uuid.UUID     `json:"conversation_id,omitempty"`
	MessageID      uuid.UUID     `json:"message_id,omitempty"`
}

func normalizedSourceScope(scope *SourceScope) (*SourceScope, error) {
	if scope == nil {
		return nil, nil
	}
	valid := scope.TeamID != uuid.Nil
	switch scope.Surface {
	case SourceSurfaceStoryComment:
		valid = valid && scope.StoryID != uuid.Nil && scope.ConversationID == uuid.Nil && scope.MessageID == uuid.Nil
	case SourceSurfaceChatMessage:
		valid = valid && scope.StoryID == uuid.Nil && scope.ConversationID != uuid.Nil && scope.MessageID != uuid.Nil
	default:
		valid = false
	}
	if !valid {
		return nil, fmt.Errorf("%w: invalid native source scope", ErrInvalidRequest)
	}
	copy := *scope
	return &copy, nil
}

func sourceAllowsTool(scope *SourceScope, name string) bool {
	if scope == nil {
		return true
	}
	if scope.Surface == SourceSurfaceChatMessage {
		switch name {
		case toolListTeams, toolListTeamWork, toolGetStory, toolListStatuses, toolListTeamMembers, toolCreateStories:
			return true
		default:
			return false
		}
	}
	switch name {
	case toolGetStory, toolListStatuses, toolListTeamMembers, toolUpdateStory:
		return scope.Surface == SourceSurfaceStoryComment
	default:
		return false
	}
}

// Native member discovery already has one authoritative team. The model sees
// names and codes, not a team catalog, so it must not invent a routing UUID.
// Explicit legacy arguments remain checked rather than silently overridden.
func sourceTeamArguments(scope *SourceScope, raw json.RawMessage) (json.RawMessage, error) {
	if scope == nil {
		return raw, nil
	}
	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(raw, &arguments); err != nil || arguments == nil {
		return nil, fmt.Errorf("%w: arguments must be an object", ErrInvalidToolArguments)
	}
	if value, exists := arguments["team_id"]; exists {
		var requested *string
		if err := json.Unmarshal(value, &requested); err != nil {
			return nil, fmt.Errorf("%w: team_id must be a UUID or null", ErrInvalidToolArguments)
		}
		if requested != nil {
			id, err := uuid.Parse(strings.TrimSpace(*requested))
			if err != nil || id != scope.TeamID {
				return nil, fmt.Errorf("%w: only the current story's team is available", ErrTeamNotAccessible)
			}
		}
	}
	teamID, err := json.Marshal(scope.TeamID.String())
	if err != nil {
		return nil, err
	}
	arguments["team_id"] = teamID
	return json.Marshal(arguments)
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
