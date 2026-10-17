package storieshttp

import (
	"encoding/json"

	storytransport "github.com/complexus-tech/projects-api/internal/modules/stories/transport"
	"github.com/google/uuid"
)

const reconcileDescriptionMediaField = "reconcileDescriptionMedia"

func storyMediaReconciliationRequest(fields map[string]json.RawMessage, workspaceSlug string, storyID uuid.UUID) (bool, []uuid.UUID, error) {
	return storytransport.StoryMediaReconciliationRequest(fields, workspaceSlug, storyID)
}

func parseStoryMediaReferences(contentHTML, workspaceSlug string, storyID uuid.UUID) ([]uuid.UUID, error) {
	return storytransport.ParseStoryMediaReferences(contentHTML, workspaceSlug, storyID)
}
