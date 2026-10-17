package desktophttp

import (
	"encoding/json"
	"fmt"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	storytransport "github.com/complexus-tech/projects-api/internal/modules/stories/transport"
)

func decodeStoryPatch(raw json.RawMessage) (stories.StoryPatch, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || len(fields) == 0 {
		return stories.StoryPatch{}, errInvalidOperation
	}
	patch, err := storytransport.ParseStoryPatch(fields)
	if err != nil {
		return stories.StoryPatch{}, fmt.Errorf("%w: %v", errInvalidOperation, err)
	}
	return patch, nil
}
