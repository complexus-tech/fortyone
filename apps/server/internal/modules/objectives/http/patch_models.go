package objectiveshttp

import (
	"bytes"
	"encoding/json"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

// PatchField preserves the three states a JSON update needs: omitted, a
// concrete value, and explicit null. Plain pointers collapse omitted and null
// into the same Go value and previously made nullable objective fields
// impossible to clear safely.
type PatchField[T any] struct {
	specified bool
	value     *T
}

func (field *PatchField[T]) UnmarshalJSON(data []byte) error {
	field.specified = true
	field.value = nil
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	field.value = &value
	return nil
}

func (field PatchField[T]) Specified() bool { return field.specified }
func (field PatchField[T]) Value() (*T, bool) {
	return field.value, field.specified
}

type AppUpdateObjective struct {
	Name              PatchField[string]    `json:"name"`
	Description       PatchField[string]    `json:"description"`
	ShortSummary      PatchField[string]    `json:"shortSummary"`
	LeadUser          PatchField[uuid.UUID] `json:"leadUser"`
	StartDate         PatchField[date.Date] `json:"startDate"`
	EndDate           PatchField[date.Date] `json:"endDate"`
	IsPrivate         PatchField[bool]      `json:"isPrivate"`
	Status            PatchField[uuid.UUID] `json:"statusId"`
	Priority          PatchField[string]    `json:"priority"`
	Health            PatchField[string]    `json:"health"`
	Color             PatchField[string]    `json:"color"`
	Comment           *string               `json:"comment"`
	ExpectedUpdatedAt *time.Time            `json:"expectedUpdatedAt"`
}

func (request AppUpdateObjective) Validate() error {
	return request.ObjectivePatch().Validate()
}

func (request AppUpdateObjective) ObjectivePatch() objectivesdomain.ObjectivePatch {
	return objectivesdomain.ObjectivePatch{
		Name: patchField(request.Name), Description: patchField(request.Description),
		ShortSummary: patchField(request.ShortSummary), LeadUser: patchField(request.LeadUser),
		StartDate: dateField(request.StartDate), EndDate: dateField(request.EndDate),
		IsPrivate: patchField(request.IsPrivate), Status: patchField(request.Status),
		Priority: patchField(request.Priority), Health: healthField(request.Health),
		Color: patchField(request.Color),
	}
}

func patchField[T any](field PatchField[T]) objectivesdomain.Field[T] {
	value, specified := field.Value()
	if !specified {
		return objectivesdomain.Field[T]{}
	}
	if value == nil {
		return objectivesdomain.ClearField[T]()
	}
	return objectivesdomain.SetField(*value)
}

func dateField(field PatchField[date.Date]) objectivesdomain.Field[time.Time] {
	value, specified := field.Value()
	if !specified {
		return objectivesdomain.Field[time.Time]{}
	}
	if value == nil {
		return objectivesdomain.ClearField[time.Time]()
	}
	return objectivesdomain.SetField(value.Time().UTC())
}

func healthField(field PatchField[string]) objectivesdomain.Field[objectivesdomain.ObjectiveHealth] {
	value, specified := field.Value()
	if !specified {
		return objectivesdomain.Field[objectivesdomain.ObjectiveHealth]{}
	}
	if value == nil {
		return objectivesdomain.ClearField[objectivesdomain.ObjectiveHealth]()
	}
	return objectivesdomain.SetField(objectivesdomain.ObjectiveHealth(*value))
}
