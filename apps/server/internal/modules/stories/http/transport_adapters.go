package storieshttp

import (
	"encoding/json"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// ToCoreNewStory shares the app's validated transport conversion with native
// receipt-backed requests. The actor always comes from the authenticated context.
func ToCoreNewStory(input AppNewStory, actorID uuid.UUID) stories.CoreNewStory {
	return toCoreNewStory(input, actorID)
}

// ParseStoryPatch applies the same explicit allowlist and omitted/null/value
// semantics as the app's PATCH endpoint. Media reconciliation is a separate
// side effect which callers must handle or explicitly reject.
func ParseStoryPatch(fields map[string]json.RawMessage) (stories.StoryPatch, error) {
	return parseStoryPatch(fields)
}

// StoryMediaReconciliationRequest validates the authoritative editor snapshot
// using the same exact story/workspace attachment references as the web app.
func StoryMediaReconciliationRequest(fields map[string]json.RawMessage, workspaceSlug string, storyID uuid.UUID) (bool, []uuid.UUID, error) {
	return storyMediaReconciliationRequest(fields, workspaceSlug, storyID)
}
