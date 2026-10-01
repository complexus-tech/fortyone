package workpresets

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	"github.com/google/uuid"
)

const maximumConfigurationBytes = 64 << 10

// Configuration is a versioned snapshot of supported UI state, never an
// arbitrary JSON bag. Unknown keys are rejected at every nested boundary.
func ValidateConfiguration(kind domain.Kind, raw json.RawMessage) error {
	if len(raw) == 0 || len(raw) > maximumConfigurationBytes {
		return domain.ErrInvalidInput
	}
	switch kind {
	case domain.View:
		var value domain.ViewConfiguration
		if strictDecode(raw, &value) != nil || validateView(value) != nil {
			return domain.ErrInvalidInput
		}
	case domain.Template:
		var value domain.TemplateConfiguration
		if strictDecode(raw, &value) != nil || validateTemplate(value) != nil {
			return domain.ErrInvalidInput
		}
	default:
		return domain.ErrInvalidInput
	}
	return nil
}

func strictDecode(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.ErrInvalidInput
	}
	return nil
}

func validateView(value domain.ViewConfiguration) error {
	options := value.ViewOptions
	if len(options.SelectedCustomFieldIDs) > 3 || !validIDs(options.SelectedCustomFieldIDs) {
		return domain.ErrInvalidInput
	}
	if value.Version != 1 || !oneOf(value.Layout, "list", "kanban") || !oneOf(options.GroupBy, "status", "assignee", "priority", "none") || !oneOf(options.OrderBy, "priority", "deadline", "created", "updated") || !oneOf(options.OrderDirection, "asc", "desc") {
		return domain.ErrInvalidInput
	}
	if options.DisplayColumnsVersion < 0 || options.DisplayColumnsVersion > 3 || len(options.DisplayColumns) > 14 || len(options.HiddenKanbanGroups) > 3 {
		return domain.ErrInvalidInput
	}
	for _, column := range options.DisplayColumns {
		if !oneOf(column, "ID", "Status", "Assignee", "Estimate", "Time needed", "Priority", "Deadline", "Created", "Updated", "Sprint", "Objective", "Key Result", "Epic", "Labels") {
			return domain.ErrInvalidInput
		}
	}
	for group, ids := range options.HiddenKanbanGroups {
		if !oneOf(group, "status", "assignee", "priority") || len(ids) > 100 {
			return domain.ErrInvalidInput
		}
		for _, id := range ids {
			if len(id) > 64 {
				return domain.ErrInvalidInput
			}
		}
	}
	filters := value.Filters
	for _, ids := range [][]uuid.UUID{filters.StatusIDs, filters.AssigneeIDs, filters.ReporterIDs, filters.TeamIDs, filters.SprintIDs, filters.LabelIDs} {
		if !validIDs(ids) {
			return domain.ErrInvalidInput
		}
	}
	if len(filters.EstimateValues) > 100 || len(filters.Priorities) > 5 || len(filters.Operators) > 14 {
		return domain.ErrInvalidInput
	}
	for _, estimate := range filters.EstimateValues {
		if math.IsNaN(estimate) || math.IsInf(estimate, 0) || estimate < 0 || estimate > 10000 {
			return domain.ErrInvalidInput
		}
	}
	for _, priority := range filters.Priorities {
		if !validPriority(priority) {
			return domain.ErrInvalidInput
		}
	}
	if filters.ContentContains != nil && len(*filters.ContentContains) > 200 {
		return domain.ErrInvalidInput
	}
	for _, date := range []*string{filters.StartDate, filters.EndDate, filters.CompletedAfter, filters.CompletedBefore} {
		if date != nil {
			if _, err := time.Parse("2006-01-02", *date); err != nil {
				return domain.ErrInvalidInput
			}
		}
	}
	for field, operator := range filters.Operators {
		if !oneOf(field, "contentContains", "statusIds", "assigneeIds", "reporterIds", "priorities", "teamIds", "sprintIds", "labelIds", "estimateValues", "objectiveId", "startDate", "endDate", "hasNoAssignee") || !oneOf(operator, "contains", "doesNotContain", "isAnyOf", "isNotAnyOf", "is", "isNot", "isOnOrBefore", "isOnOrAfter", "isEmpty", "isNotEmpty") {
			return domain.ErrInvalidInput
		}
	}
	return nil
}

func validateTemplate(value domain.TemplateConfiguration) error {
	if value.EstimateValue != nil && (*value.EstimateValue < 0 || *value.EstimateValue > 10000) {
		return domain.ErrInvalidInput
	}
	if value.Version != 1 || len(value.Title) > 255 || len(value.Description) > 20000 || len(value.DescriptionHTML) > 40000 || !validPriority(value.Priority) || !validIDs(value.LabelIDs) || len(value.Checklist) > 50 || len(value.CustomFieldValues) > 50 {
		return domain.ErrInvalidInput
	}
	for _, item := range value.Checklist {
		if len(item) == 0 || len(item) > 500 {
			return domain.ErrInvalidInput
		}
	}
	for _, minutes := range []*int{value.EstimatedDurationMinutes, value.MinimumFocusBlockMinutes} {
		if minutes != nil && (*minutes < 5 || *minutes > 10080) {
			return domain.ErrInvalidInput
		}
	}
	if value.EstimateLabel != nil && len(*value.EstimateLabel) > 32 {
		return domain.ErrInvalidInput
	}
	seen := make(map[uuid.UUID]bool, len(value.CustomFieldValues))
	for _, field := range value.CustomFieldValues {
		if field.FieldID == uuid.Nil || seen[field.FieldID] || (field.Value != nil && len(*field.Value) > 2000) {
			return domain.ErrInvalidInput
		}
		seen[field.FieldID] = true
	}
	return nil
}

func validIDs(ids []uuid.UUID) bool {
	if len(ids) > 100 {
		return false
	}
	for _, id := range ids {
		if id == uuid.Nil {
			return false
		}
	}
	return true
}

func validPriority(value string) bool {
	return oneOf(value, "No Priority", "Urgent", "High", "Medium", "Low")
}

func oneOf(value string, allowed ...string) bool {
	for _, option := range allowed {
		if value == option {
			return true
		}
	}
	return false
}
