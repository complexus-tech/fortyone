package desktophttp

import (
	"encoding/json"
	"fmt"

	storieshttp "github.com/complexus-tech/projects-api/internal/modules/stories/http"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func decodeStoryPatch(raw json.RawMessage) (stories.StoryPatch, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || len(fields) == 0 {
		return stories.StoryPatch{}, errInvalidOperation
	}
	patch, err := storieshttp.ParseStoryPatch(fields)
	if err != nil {
		return stories.StoryPatch{}, fmt.Errorf("%w: %v", errInvalidOperation, err)
	}
	return patch, nil
}
