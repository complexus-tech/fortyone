package storieshttp

import (
	"bytes"
	"encoding/json"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	storytransport "github.com/complexus-tech/projects-api/internal/modules/stories/transport"
)

func parseStoryPatch(requestData map[string]json.RawMessage) (stories.StoryPatch, error) {
	return storytransport.ParseStoryPatch(requestData)
}

// getUpdates remains a narrow compatibility helper for focused tests. HTTP
// handlers use parseStoryPatch and pass the typed value directly to service.
func getUpdates(requestData map[string]json.RawMessage) (map[string]any, error) {
	patch, err := parseStoryPatch(requestData)
	if err != nil {
		return nil, err
	}
	return patchUpdatesForCompatibility(patch), nil
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

// patchUpdatesForCompatibility is intentionally local to transport tests. It
// is not used for persistence and does not expose database identifiers from
// untrusted input.
func patchUpdatesForCompatibility(patch stories.StoryPatch) map[string]any {
	updates := make(map[string]any, len(patch.Fields()))
	putRequiredCompatibility(updates, "title", patch.Title)
	putNullableCompatibility(updates, "estimate_unit", patch.EstimateValue)
	putNullableCompatibility(updates, "estimated_duration_minutes", patch.EstimatedDurationMinutes)
	putNullableCompatibility(updates, "minimum_focus_block_minutes", patch.MinimumFocusBlockMinutes)
	putRequiredCompatibility(updates, "auto_scheduling_enabled", patch.AutoSchedulingEnabled)
	putRequiredCompatibility(updates, "auto_scheduling_locked", patch.AutoSchedulingLocked)
	putNullableStringCompatibility(updates, "description", patch.Description)
	putNullableStringCompatibility(updates, "description_html", patch.DescriptionHTML)
	putNullableCompatibility(updates, "parent_id", patch.ParentID)
	putNullableCompatibility(updates, "objective_id", patch.ObjectiveID)
	putNullableCompatibility(updates, "status_id", patch.StatusID)
	putNullableCompatibility(updates, "assignee_id", patch.AssigneeID)
	putRequiredCompatibility(updates, "priority", patch.Priority)
	putNullableCompatibility(updates, "sprint_id", patch.SprintID)
	putNullableCompatibility(updates, "key_result_id", patch.KeyResultID)
	putNullableCompatibility(updates, "start_date", patch.StartDate)
	putNullableCompatibility(updates, "end_date", patch.EndDate)
	return updates
}

func putRequiredCompatibility[T any](updates map[string]any, name string, field stories.Field[T]) {
	value, specified := field.Value()
	if specified && value != nil {
		updates[name] = *value
	}
}

func putNullableCompatibility[T any](updates map[string]any, name string, field stories.Field[T]) {
	value, specified := field.Value()
	if specified {
		updates[name] = value
	}
}

func putNullableStringCompatibility(updates map[string]any, name string, field stories.Field[string]) {
	value, specified := field.Value()
	if !specified {
		return
	}
	if value == nil {
		updates[name] = nil
		return
	}
	updates[name] = *value
}
