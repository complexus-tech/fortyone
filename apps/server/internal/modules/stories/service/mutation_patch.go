package stories

import (
	"fmt"
	"math"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

// storyPatchFromUpdates is the compatibility boundary for integrations that
// still construct legacy update maps. It accepts a finite allowlist and maps
// every value into the typed patch used by validation and persistence. Unknown
// keys can therefore never become SQL identifiers.
func storyPatchFromUpdates(updates map[string]any) (StoryPatch, error) {
	var patch StoryPatch
	for name, raw := range updates {
		var err error
		switch name {
		case "title":
			patch.Title, err = requiredStringPatchField(name, raw)
		case "estimate_unit":
			patch.EstimateValue, err = nullableInt16PatchField(name, raw)
		case "estimated_duration_minutes":
			patch.EstimatedDurationMinutes, err = nullableIntPatchField(name, raw)
		case "minimum_focus_block_minutes":
			patch.MinimumFocusBlockMinutes, err = nullableIntPatchField(name, raw)
		case "auto_scheduling_enabled":
			patch.AutoSchedulingEnabled, err = requiredBoolPatchField(name, raw)
		case "auto_scheduling_locked":
			patch.AutoSchedulingLocked, err = requiredBoolPatchField(name, raw)
		case "auto_scheduling_status":
			patch.AutoSchedulingStatus, err = requiredStringPatchField(name, raw)
		case "auto_scheduling_reason":
			patch.AutoSchedulingReason, err = nullableStringPatchField(name, raw)
		case "auto_scheduling_updated_at":
			patch.AutoSchedulingUpdatedAt, err = nullableTimePatchField(name, raw)
		case "description":
			patch.Description, err = nullableStringPatchField(name, raw)
		case "description_html":
			patch.DescriptionHTML, err = nullableStringPatchField(name, raw)
		case "parent_id":
			patch.ParentID, err = nullableUUIDPatchField(name, raw)
		case "objective_id":
			patch.ObjectiveID, err = nullableUUIDPatchField(name, raw)
		case "status_id":
			patch.StatusID, err = nullableUUIDPatchField(name, raw)
		case "assignee_id":
			patch.AssigneeID, err = nullableUUIDPatchField(name, raw)
		case "priority":
			patch.Priority, err = requiredStringPatchField(name, raw)
		case "sprint_id":
			patch.SprintID, err = nullableUUIDPatchField(name, raw)
		case "key_result_id":
			patch.KeyResultID, err = nullableUUIDPatchField(name, raw)
		case "start_date":
			patch.StartDate, err = nullableTimePatchField(name, raw)
		case "end_date":
			patch.EndDate, err = nullableTimePatchField(name, raw)
		case "completed_at":
			patch.CompletedAt, err = nullableTimePatchField(name, raw)
		default:
			return StoryPatch{}, fmt.Errorf("%w: unsupported field %q", ErrInvalidStoryMutation, name)
		}
		if err != nil {
			return StoryPatch{}, err
		}
	}
	if err := patch.Validate(); err != nil {
		return StoryPatch{}, err
	}
	return patch, nil
}

func storyPatchToUpdates(patch StoryPatch) map[string]any {
	updates := make(map[string]any, len(patch.Fields()))
	putRequiredField(updates, "title", patch.Title)
	putNullableField(updates, "estimate_unit", patch.EstimateValue)
	putNullableField(updates, "estimated_duration_minutes", patch.EstimatedDurationMinutes)
	putNullableField(updates, "minimum_focus_block_minutes", patch.MinimumFocusBlockMinutes)
	putRequiredField(updates, "auto_scheduling_enabled", patch.AutoSchedulingEnabled)
	putRequiredField(updates, "auto_scheduling_locked", patch.AutoSchedulingLocked)
	putRequiredField(updates, "auto_scheduling_status", patch.AutoSchedulingStatus)
	putNullableField(updates, "auto_scheduling_reason", patch.AutoSchedulingReason)
	putNullableField(updates, "auto_scheduling_updated_at", patch.AutoSchedulingUpdatedAt)
	putNullableStringField(updates, "description", patch.Description)
	putNullableStringField(updates, "description_html", patch.DescriptionHTML)
	putNullableField(updates, "parent_id", patch.ParentID)
	putNullableField(updates, "objective_id", patch.ObjectiveID)
	putNullableValueField(updates, "status_id", patch.StatusID)
	putNullableField(updates, "assignee_id", patch.AssigneeID)
	putRequiredField(updates, "priority", patch.Priority)
	putNullableField(updates, "sprint_id", patch.SprintID)
	putNullableField(updates, "key_result_id", patch.KeyResultID)
	putNullableField(updates, "start_date", patch.StartDate)
	putNullableField(updates, "end_date", patch.EndDate)
	putNullableField(updates, "completed_at", patch.CompletedAt)
	return updates
}

func requiredStringPatchField(name string, raw any) (storydomain.Field[string], error) {
	switch value := raw.(type) {
	case string:
		return SetField(value), nil
	case *string:
		if value != nil {
			return SetField(*value), nil
		}
	}
	return storydomain.Field[string]{}, fmt.Errorf("%w: %s must be a string", ErrInvalidStoryMutation, name)
}

func nullableStringPatchField(name string, raw any) (storydomain.Field[string], error) {
	if raw == nil {
		return ClearField[string](), nil
	}
	switch value := raw.(type) {
	case string:
		return SetField(value), nil
	case *string:
		if value == nil {
			return ClearField[string](), nil
		}
		return SetField(*value), nil
	default:
		return storydomain.Field[string]{}, fmt.Errorf("%w: %s must be a string or null", ErrInvalidStoryMutation, name)
	}
}

func requiredBoolPatchField(name string, raw any) (storydomain.Field[bool], error) {
	switch value := raw.(type) {
	case bool:
		return SetField(value), nil
	case *bool:
		if value != nil {
			return SetField(*value), nil
		}
	}
	return storydomain.Field[bool]{}, fmt.Errorf("%w: %s must be a boolean", ErrInvalidStoryMutation, name)
}

func nullableUUIDPatchField(name string, raw any) (storydomain.Field[uuid.UUID], error) {
	if raw == nil {
		return ClearField[uuid.UUID](), nil
	}
	value, valid := optionalUUIDUpdate(raw)
	if !valid {
		return storydomain.Field[uuid.UUID]{}, fmt.Errorf("%w: %s must be a UUID or null", ErrInvalidStoryMutation, name)
	}
	if value == nil {
		return ClearField[uuid.UUID](), nil
	}
	if *value == uuid.Nil {
		return storydomain.Field[uuid.UUID]{}, fmt.Errorf("%w: %s cannot be an empty UUID", ErrInvalidStoryMutation, name)
	}
	return SetField(*value), nil
}

func nullableTimePatchField(name string, raw any) (storydomain.Field[time.Time], error) {
	if raw == nil {
		return ClearField[time.Time](), nil
	}
	switch value := raw.(type) {
	case time.Time:
		if value.IsZero() {
			return storydomain.Field[time.Time]{}, fmt.Errorf("%w: %s cannot be a zero time", ErrInvalidStoryMutation, name)
		}
		return SetField(value.UTC()), nil
	case *time.Time:
		if value == nil {
			return ClearField[time.Time](), nil
		}
		if value.IsZero() {
			return storydomain.Field[time.Time]{}, fmt.Errorf("%w: %s cannot be a zero time", ErrInvalidStoryMutation, name)
		}
		return SetField(value.UTC()), nil
	default:
		return storydomain.Field[time.Time]{}, fmt.Errorf("%w: %s must be a timestamp or null", ErrInvalidStoryMutation, name)
	}
}

func nullableIntPatchField(name string, raw any) (storydomain.Field[int], error) {
	if raw == nil {
		return ClearField[int](), nil
	}
	var value int64
	switch typed := raw.(type) {
	case int:
		return SetField(typed), nil
	case *int:
		if typed == nil {
			return ClearField[int](), nil
		}
		return SetField(*typed), nil
	case int32:
		return SetField(int(typed)), nil
	case int64:
		value = typed
	case float64:
		value = int64(typed)
		if float64(value) != typed {
			return storydomain.Field[int]{}, fmt.Errorf("%w: %s must be an integer", ErrInvalidStoryMutation, name)
		}
	default:
		return storydomain.Field[int]{}, fmt.Errorf("%w: %s must be an integer or null", ErrInvalidStoryMutation, name)
	}
	if value < int64(math.MinInt) || value > int64(math.MaxInt) {
		return storydomain.Field[int]{}, fmt.Errorf("%w: %s is outside the supported range", ErrInvalidStoryMutation, name)
	}
	return SetField(int(value)), nil
}

func nullableInt16PatchField(name string, raw any) (storydomain.Field[int16], error) {
	value, err := normalizeEstimateUpdateValue(raw)
	if err != nil {
		return storydomain.Field[int16]{}, fmt.Errorf("%w: %s: %v", ErrInvalidStoryMutation, name, err)
	}
	if value == nil {
		return ClearField[int16](), nil
	}
	return SetField(*value), nil
}

func putRequiredField[T any](updates map[string]any, name string, field storydomain.Field[T]) {
	value, specified := field.Value()
	if specified && value != nil {
		updates[name] = *value
	}
}

func putNullableField[T any](updates map[string]any, name string, field storydomain.Field[T]) {
	value, specified := field.Value()
	if specified {
		updates[name] = value
	}
}

func putNullableValueField[T any](updates map[string]any, name string, field storydomain.Field[T]) {
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

func putNullableStringField(updates map[string]any, name string, field storydomain.Field[string]) {
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
