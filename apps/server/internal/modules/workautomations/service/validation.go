package workautomations

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/google/uuid"
)

func decode(raw json.RawMessage, target any) error {
	if len(raw) == 0 || len(raw) > 64<<10 {
		return domain.ErrInvalidInput
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domain.ErrInvalidInput
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.ErrInvalidInput
	}
	return nil
}

func validateRule(raw json.RawMessage) (domain.RuleConfiguration, error) {
	var config domain.RuleConfiguration
	if decode(raw, &config) != nil || config.Version != 1 || (config.Trigger != "story.created" && config.Trigger != "story.updated") {
		return config, domain.ErrInvalidInput
	}
	conditions, actions := config.Conditions, config.Actions
	if !validIDs(conditions.StatusIDs) || !validIDs(conditions.AssigneeIDs) || len(conditions.Priorities) > 5 || (conditions.Unassigned && len(conditions.AssigneeIDs) > 0) {
		return config, domain.ErrInvalidInput
	}
	for _, priority := range conditions.Priorities {
		if !validPriority(priority) {
			return config, domain.ErrInvalidInput
		}
	}
	if !validID(actions.StatusID) || !validID(actions.AssigneeID) || (actions.ClearAssignee && actions.AssigneeID != nil) || (actions.Priority != nil && !validPriority(*actions.Priority)) {
		return config, domain.ErrInvalidInput
	}
	if actions.StatusID == nil && actions.Priority == nil && actions.AssigneeID == nil && !actions.ClearAssignee {
		return config, domain.ErrInvalidInput
	}
	return config, nil
}

func validateRecurrence(raw json.RawMessage) (domain.RecurrenceConfiguration, error) {
	var config domain.RecurrenceConfiguration
	if decode(raw, &config) != nil || config.Version != 1 || validateDraft(config.Draft) != nil || validateSchedule(config.Schedule) != nil {
		return config, domain.ErrInvalidInput
	}
	return config, nil
}

func validateDraft(draft domain.Draft) error {
	if strings.TrimSpace(draft.Title) == "" || len(draft.Title) > 255 || len(draft.Description) > 20000 || len(draft.DescriptionHTML) > 40000 || !validPriority(draft.Priority) || !validID(draft.StatusID) || !validID(draft.AssigneeID) || !validIDs(draft.LabelIDs) || len(draft.Checklist) > 50 || len(draft.CustomFieldValues) > 50 {
		return domain.ErrInvalidInput
	}
	if draft.EstimateValue != nil && (*draft.EstimateValue < 0 || *draft.EstimateValue > 10000) {
		return domain.ErrInvalidInput
	}
	for _, minutes := range []*int{draft.EstimatedDurationMinutes, draft.MinimumFocusBlockMinutes} {
		if minutes != nil && (*minutes < 5 || *minutes > 10080) {
			return domain.ErrInvalidInput
		}
	}
	if draft.EstimatedDurationMinutes != nil && draft.MinimumFocusBlockMinutes != nil && *draft.MinimumFocusBlockMinutes > *draft.EstimatedDurationMinutes {
		return domain.ErrInvalidInput
	}
	for _, item := range draft.Checklist {
		if len(item) > 500 {
			return domain.ErrInvalidInput
		}
	}
	seen := make(map[uuid.UUID]bool)
	for _, field := range draft.CustomFieldValues {
		if field.FieldID == uuid.Nil || seen[field.FieldID] || (field.Value != nil && len(*field.Value) > 2000) {
			return domain.ErrInvalidInput
		}
		seen[field.FieldID] = true
	}
	return nil
}

func validPriority(value string) bool {
	return value == "No Priority" || value == "Low" || value == "Medium" || value == "High" || value == "Urgent"
}
func validID(id *uuid.UUID) bool { return id == nil || *id != uuid.Nil }
func validIDs(ids []uuid.UUID) bool {
	if len(ids) > 100 {
		return false
	}
	seen := make(map[uuid.UUID]bool)
	for _, id := range ids {
		if id == uuid.Nil || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
