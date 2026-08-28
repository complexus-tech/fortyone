package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMutation       = errors.New("invalid story mutation")
	ErrMutationForbidden     = errors.New("story mutation is not permitted")
	ErrMutationConflict      = errors.New("story mutation conflicts with a newer version")
	ErrStoryChanged          = errors.New("story changed before the update was applied")
	ErrMutationEventNotFound = errors.New("story mutation event was not found")
)

// Field preserves the three states required by PATCH semantics: omitted,
// explicitly null, and a concrete value. Its representation is private so a
// caller cannot create a contradictory state.
type Field[T any] struct {
	specified bool
	value     *T
}

func SetField[T any](value T) Field[T] {
	return Field[T]{specified: true, value: &value}
}

func ClearField[T any]() Field[T] {
	return Field[T]{specified: true}
}

func (field Field[T]) Specified() bool {
	return field.specified
}

func (field Field[T]) Value() (*T, bool) {
	if !field.specified {
		return nil, false
	}
	return field.value, true
}

// StoryPatch is the finite mutation surface for a story. Persistence maps each
// field explicitly; request values can never become SQL identifiers.
type StoryPatch struct {
	Title                    Field[string]
	EstimateValue            Field[int16]
	EstimatedDurationMinutes Field[int]
	MinimumFocusBlockMinutes Field[int]
	AutoSchedulingEnabled    Field[bool]
	AutoSchedulingLocked     Field[bool]
	AutoSchedulingStatus     Field[string]
	AutoSchedulingReason     Field[string]
	AutoSchedulingUpdatedAt  Field[time.Time]
	Description              Field[string]
	DescriptionHTML          Field[string]
	ParentID                 Field[uuid.UUID]
	ObjectiveID              Field[uuid.UUID]
	StatusID                 Field[uuid.UUID]
	AssigneeID               Field[uuid.UUID]
	Priority                 Field[string]
	SprintID                 Field[uuid.UUID]
	KeyResultID              Field[uuid.UUID]
	StartDate                Field[time.Time]
	EndDate                  Field[time.Time]
	CompletedAt              Field[time.Time]
}

var storyPatchFieldOrder = []string{
	"title",
	"estimate_unit",
	"estimated_duration_minutes",
	"minimum_focus_block_minutes",
	"auto_scheduling_enabled",
	"auto_scheduling_locked",
	"auto_scheduling_status",
	"auto_scheduling_reason",
	"auto_scheduling_updated_at",
	"description",
	"description_html",
	"parent_id",
	"objective_id",
	"status_id",
	"assignee_id",
	"priority",
	"sprint_id",
	"key_result_id",
	"start_date",
	"end_date",
	"completed_at",
}

func (patch StoryPatch) Empty() bool {
	return len(patch.Fields()) == 0
}

// Fields returns specified database-domain names in a stable order. The names
// are metadata for activities/events only and are never interpolated into SQL.
func (patch StoryPatch) Fields() []string {
	fields := make([]string, 0, len(storyPatchFieldOrder))
	for _, field := range storyPatchFieldOrder {
		if patch.fieldSpecified(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func (patch StoryPatch) fieldSpecified(name string) bool {
	switch name {
	case "title":
		return patch.Title.Specified()
	case "estimate_unit":
		return patch.EstimateValue.Specified()
	case "estimated_duration_minutes":
		return patch.EstimatedDurationMinutes.Specified()
	case "minimum_focus_block_minutes":
		return patch.MinimumFocusBlockMinutes.Specified()
	case "auto_scheduling_enabled":
		return patch.AutoSchedulingEnabled.Specified()
	case "auto_scheduling_locked":
		return patch.AutoSchedulingLocked.Specified()
	case "auto_scheduling_status":
		return patch.AutoSchedulingStatus.Specified()
	case "auto_scheduling_reason":
		return patch.AutoSchedulingReason.Specified()
	case "auto_scheduling_updated_at":
		return patch.AutoSchedulingUpdatedAt.Specified()
	case "description":
		return patch.Description.Specified()
	case "description_html":
		return patch.DescriptionHTML.Specified()
	case "parent_id":
		return patch.ParentID.Specified()
	case "objective_id":
		return patch.ObjectiveID.Specified()
	case "status_id":
		return patch.StatusID.Specified()
	case "assignee_id":
		return patch.AssigneeID.Specified()
	case "priority":
		return patch.Priority.Specified()
	case "sprint_id":
		return patch.SprintID.Specified()
	case "key_result_id":
		return patch.KeyResultID.Specified()
	case "start_date":
		return patch.StartDate.Specified()
	case "end_date":
		return patch.EndDate.Specified()
	case "completed_at":
		return patch.CompletedAt.Specified()
	default:
		return false
	}
}

func (patch StoryPatch) Validate() error {
	if patch.Empty() {
		return fmt.Errorf("%w: at least one field is required", ErrInvalidMutation)
	}
	for _, field := range []struct {
		name  string
		value any
	}{
		{name: "title", value: patch.Title},
		{name: "auto_scheduling_enabled", value: patch.AutoSchedulingEnabled},
		{name: "auto_scheduling_locked", value: patch.AutoSchedulingLocked},
		{name: "auto_scheduling_status", value: patch.AutoSchedulingStatus},
		{name: "priority", value: patch.Priority},
	} {
		switch value := field.value.(type) {
		case Field[string]:
			if value.Specified() && value.value == nil {
				return fmt.Errorf("%w: %s cannot be null", ErrInvalidMutation, field.name)
			}
		case Field[bool]:
			if value.Specified() && value.value == nil {
				return fmt.Errorf("%w: %s cannot be null", ErrInvalidMutation, field.name)
			}
		}
	}
	if value, ok := patch.Title.Value(); ok && value != nil && strings.TrimSpace(*value) == "" {
		return fmt.Errorf("%w: title cannot be blank", ErrInvalidMutation)
	}
	if value, ok := patch.Priority.Value(); ok && value != nil && !slices.Contains(
		[]string{"No Priority", "Low", "Medium", "High", "Urgent"}, *value,
	) {
		return fmt.Errorf("%w: unsupported priority", ErrInvalidMutation)
	}
	return nil
}
