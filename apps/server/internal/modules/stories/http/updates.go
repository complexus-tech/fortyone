package storieshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

// parseStoryPatch is an explicit transport allowlist. It preserves PATCH's
// omitted/null/value states and rejects fields that the API does not own,
// including Maya-managed scheduling state.
func parseStoryPatch(requestData map[string]json.RawMessage) (stories.StoryPatch, error) {
	var patch stories.StoryPatch
	for name, raw := range requestData {
		var err error
		switch name {
		case reconcileDescriptionMediaField:
			// Parsed separately by storyMediaReconciliationRequest.
			continue
		case "title":
			patch.Title, err = decodeRequiredPatchField[string](name, raw)
		case "estimateValue":
			patch.EstimateValue, err = decodeNullablePatchField[int16](name, raw)
		case "estimatedDurationMinutes":
			patch.EstimatedDurationMinutes, err = decodeNullablePatchField[int](name, raw)
		case "minimumFocusBlockMinutes":
			patch.MinimumFocusBlockMinutes, err = decodeNullablePatchField[int](name, raw)
		case "autoSchedulingEnabled":
			patch.AutoSchedulingEnabled, err = decodeRequiredPatchField[bool](name, raw)
		case "autoSchedulingLocked":
			patch.AutoSchedulingLocked, err = decodeRequiredPatchField[bool](name, raw)
		case "description":
			patch.Description, err = decodeNullablePatchField[string](name, raw)
		case "descriptionHTML":
			patch.DescriptionHTML, err = decodeNullablePatchField[string](name, raw)
		case "parentId":
			patch.ParentID, err = decodeNullablePatchField[uuid.UUID](name, raw)
		case "objectiveId":
			patch.ObjectiveID, err = decodeNullablePatchField[uuid.UUID](name, raw)
		case "statusId":
			patch.StatusID, err = decodeRequiredUUIDPatchField(name, raw)
		case "assigneeId":
			patch.AssigneeID, err = decodeNullablePatchField[uuid.UUID](name, raw)
		case "priority":
			patch.Priority, err = decodeRequiredPatchField[string](name, raw)
		case "sprintId":
			patch.SprintID, err = decodeNullablePatchField[uuid.UUID](name, raw)
		case "keyResultId":
			patch.KeyResultID, err = decodeNullablePatchField[uuid.UUID](name, raw)
		case "startDate":
			patch.StartDate, err = decodeNullableDatePatchField(name, raw)
		case "endDate":
			patch.EndDate, err = decodeNullableDatePatchField(name, raw)
		default:
			return stories.StoryPatch{}, fmt.Errorf("unknown story update field %q", name)
		}
		if err != nil {
			return stories.StoryPatch{}, err
		}
	}
	if err := patch.Validate(); err != nil {
		return stories.StoryPatch{}, err
	}
	return patch, nil
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

func decodeRequiredPatchField[T any](name string, raw json.RawMessage) (stories.Field[T], error) {
	if isJSONNull(raw) {
		return stories.Field[T]{}, fmt.Errorf("%s cannot be null", name)
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return stories.Field[T]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return stories.SetField(value), nil
}

func decodeNullablePatchField[T any](name string, raw json.RawMessage) (stories.Field[T], error) {
	if isJSONNull(raw) {
		return stories.ClearField[T](), nil
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return stories.Field[T]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return stories.SetField(value), nil
}

func decodeRequiredUUIDPatchField(name string, raw json.RawMessage) (stories.Field[uuid.UUID], error) {
	field, err := decodeRequiredPatchField[uuid.UUID](name, raw)
	if err != nil {
		return stories.Field[uuid.UUID]{}, err
	}
	value, _ := field.Value()
	if value == nil || *value == uuid.Nil {
		return stories.Field[uuid.UUID]{}, fmt.Errorf("%s cannot be an empty UUID", name)
	}
	return field, nil
}

func decodeNullableDatePatchField(name string, raw json.RawMessage) (stories.Field[time.Time], error) {
	if isJSONNull(raw) {
		return stories.ClearField[time.Time](), nil
	}
	var value date.Date
	if err := json.Unmarshal(raw, &value); err != nil {
		return stories.Field[time.Time]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return stories.SetField(value.Time().UTC()), nil
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
