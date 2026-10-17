package storytransport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

// ParseStoryPatch is an explicit transport allowlist. It preserves PATCH's
// omitted/null/value states and rejects fields that the API does not own,
// including Maya-managed scheduling state.
func ParseStoryPatch(requestData map[string]json.RawMessage) (storydomain.StoryPatch, error) {
	var patch storydomain.StoryPatch
	for name, raw := range requestData {
		var err error
		switch name {
		case reconcileDescriptionMediaField:
			// Parsed separately by StoryMediaReconciliationRequest.
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
			return storydomain.StoryPatch{}, fmt.Errorf("unknown story update field %q", name)
		}
		if err != nil {
			return storydomain.StoryPatch{}, err
		}
	}
	if err := patch.Validate(); err != nil {
		return storydomain.StoryPatch{}, err
	}
	return patch, nil
}

func decodeRequiredPatchField[T any](name string, raw json.RawMessage) (storydomain.Field[T], error) {
	if isJSONNull(raw) {
		return storydomain.Field[T]{}, fmt.Errorf("%s cannot be null", name)
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return storydomain.Field[T]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return storydomain.SetField(value), nil
}

func decodeNullablePatchField[T any](name string, raw json.RawMessage) (storydomain.Field[T], error) {
	if isJSONNull(raw) {
		return storydomain.ClearField[T](), nil
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return storydomain.Field[T]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return storydomain.SetField(value), nil
}

func decodeRequiredUUIDPatchField(name string, raw json.RawMessage) (storydomain.Field[uuid.UUID], error) {
	field, err := decodeRequiredPatchField[uuid.UUID](name, raw)
	if err != nil {
		return storydomain.Field[uuid.UUID]{}, err
	}
	value, _ := field.Value()
	if value == nil || *value == uuid.Nil {
		return storydomain.Field[uuid.UUID]{}, fmt.Errorf("%s cannot be an empty UUID", name)
	}
	return field, nil
}

func decodeNullableDatePatchField(name string, raw json.RawMessage) (storydomain.Field[time.Time], error) {
	if isJSONNull(raw) {
		return storydomain.ClearField[time.Time](), nil
	}
	var value date.Date
	if err := json.Unmarshal(raw, &value); err != nil {
		return storydomain.Field[time.Time]{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return storydomain.SetField(value.Time().UTC()), nil
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
