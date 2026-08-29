package messaging

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

func proposedChangedFields(
	story messagingStory,
	userID uuid.UUID,
	title, priority *string,
	assigneeAction string,
	statusID, sprintID, objectiveID, keyResultID *uuid.UUID,
	startDate, endDate *time.Time,
	timeMutation storyTimeMutation,
	autoSchedulingMutation storyAutoSchedulingMutation,
) ([]string, error) {
	updates, err := desiredStoryUpdates(
		story,
		userID,
		title,
		priority,
		assigneeAction,
		statusID,
		sprintID,
		objectiveID,
		keyResultID,
		startDate,
		endDate,
		timeMutation,
		autoSchedulingMutation,
	)
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(updates))
	for _, field := range []string{"title", "priority", "assignee_id", "status_id", "sprint_id", "objective_id", "key_result_id", "start_date", "end_date", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked"} {
		if _, changed := updates[field]; changed {
			fields = append(fields, field)
		}
	}
	return fields, nil
}

func desiredStoryUpdates(
	story messagingStory,
	userID uuid.UUID,
	title, priority *string,
	assigneeAction string,
	statusID, sprintID, objectiveID, keyResultID *uuid.UUID,
	startDate, endDate *time.Time,
	timeMutation storyTimeMutation,
	autoSchedulingMutation storyAutoSchedulingMutation,
) (map[string]any, error) {
	updates := make(map[string]any, 3)
	if title != nil && story.Title != *title {
		updates["title"] = *title
	}
	if priority != nil && story.Priority != *priority {
		updates["priority"] = *priority
	}
	switch assigneeAction {
	case assigneeActionMe:
		if story.Assignee == nil || *story.Assignee != userID {
			updates["assignee_id"] = userID
		}
	case assigneeActionUnassigned:
		if story.Assignee != nil {
			updates["assignee_id"] = nil
		}
	}
	// Keep optional pointer values typed until after the nil check. Boxing a nil
	// *uuid.UUID or *time.Time in any produces a non-nil interface, which would
	// otherwise turn the tool's "leave unchanged" null into a destructive SQL
	// NULL update.
	for field, value := range map[string]*uuid.UUID{
		"status_id":     statusID,
		"sprint_id":     sprintID,
		"objective_id":  objectiveID,
		"key_result_id": keyResultID,
	} {
		if value != nil && !storyFieldMatches(story, field, value) {
			updates[field] = *value
		}
	}
	for field, value := range map[string]*time.Time{
		"start_date": startDate,
		"end_date":   endDate,
	} {
		if value != nil && !storyFieldMatches(story, field, value) {
			updates[field] = *value
		}
	}
	timeUpdates, err := desiredStoryTimeUpdates(story, timeMutation)
	if err != nil {
		return nil, err
	}
	for field, value := range timeUpdates {
		updates[field] = value
	}
	autoSchedulingUpdates, err := desiredStoryAutoSchedulingUpdates(story, autoSchedulingMutation)
	if err != nil {
		return nil, err
	}
	for field, value := range autoSchedulingUpdates {
		updates[field] = value
	}
	return updates, nil
}

func desiredStoryAutoSchedulingUpdates(story messagingStory, mutation storyAutoSchedulingMutation) (map[string]any, error) {
	enabled := story.AutoSchedulingEnabled
	if mutation.enabled != nil {
		enabled = *mutation.enabled
	}
	locked := story.AutoSchedulingLocked
	disabling := mutation.enabled != nil && !enabled
	if disabling {
		locked = false
	} else if mutation.locked != nil {
		locked = *mutation.locked
	}
	status := story.AutoSchedulingStatus
	if status == "" {
		status = storyAutoSchedulingStatusOff
	}
	if mutation.locked != nil && *mutation.locked && !story.AutoSchedulingLocked &&
		(!story.AutoSchedulingEnabled || (status != storyAutoSchedulingStatusScheduled && status != storyAutoSchedulingStatusAtRisk)) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToolArguments, storyAutoSchedulingLockEmptyError())
	}
	if err := validateStoryAutoSchedulingContract(enabled, locked, status); err != nil {
		return nil, err
	}

	updates := make(map[string]any, 2)
	if mutation.enabled != nil && enabled != story.AutoSchedulingEnabled {
		updates["auto_scheduling_enabled"] = enabled
	}
	if (mutation.locked != nil || disabling) && locked != story.AutoSchedulingLocked {
		updates["auto_scheduling_locked"] = locked
	}
	return updates, nil
}

func desiredStoryTimeUpdates(story messagingStory, mutation storyTimeMutation) (map[string]any, error) {
	estimatedDurationMinutes, err := resolvedStoryTimeValue(
		story.EstimatedDurationMinutes,
		mutation.estimatedDurationAction,
		mutation.estimatedDurationMinutes,
		"estimated_duration",
	)
	if err != nil {
		return nil, err
	}
	minimumFocusBlockMinutes, err := resolvedStoryTimeValue(
		story.MinimumFocusBlockMinutes,
		mutation.minimumFocusBlockAction,
		mutation.minimumFocusBlockMinutes,
		"minimum_focus_block",
	)
	if err != nil {
		return nil, err
	}

	// A focus-block constraint cannot survive without a duration. Clearing the
	// duration therefore clears the dependent constraint even when its own
	// action is unchanged.
	if mutation.estimatedDurationAction == storyTimeActionClear && mutation.minimumFocusBlockAction == storyTimeActionUnchanged {
		minimumFocusBlockMinutes = nil
	}
	if err := validateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes); err != nil {
		return nil, err
	}

	updates := make(map[string]any, 2)
	if !sameOptionalInt(story.EstimatedDurationMinutes, estimatedDurationMinutes) {
		updates["estimated_duration_minutes"] = storyTimeUpdateValue(estimatedDurationMinutes)
	}
	if !sameOptionalInt(story.MinimumFocusBlockMinutes, minimumFocusBlockMinutes) {
		updates["minimum_focus_block_minutes"] = storyTimeUpdateValue(minimumFocusBlockMinutes)
	}
	return updates, nil
}

func storyTimeUpdateValue(minutes *int) any {
	if minutes == nil {
		return nil
	}
	return *minutes
}

func resolvedStoryTimeValue(current *int, action string, minutes *int, field string) (*int, error) {
	switch action {
	case storyTimeActionUnchanged:
		if minutes != nil {
			return nil, fmt.Errorf("%s_minutes must be null when %s_action is unchanged", field, field)
		}
		return cloneIntPointer(current), nil
	case storyTimeActionClear:
		if minutes != nil {
			return nil, fmt.Errorf("%s_minutes must be null when %s_action is clear", field, field)
		}
		return nil, nil
	case storyTimeActionSet:
		if minutes == nil {
			return nil, fmt.Errorf("%s_minutes is required when %s_action is set", field, field)
		}
		if *minutes < 1 || *minutes > maximumEstimatedDurationMinutes {
			return nil, fmt.Errorf("%s_minutes must be between 1 and %d", field, maximumEstimatedDurationMinutes)
		}
		return cloneIntPointer(minutes), nil
	default:
		return nil, fmt.Errorf("%s_action must be unchanged, set, or clear", field)
	}
}

func sameOptionalInt(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func storyFieldMatches(story messagingStory, field string, value any) bool {
	var current any
	switch field {
	case "status_id":
		current = story.Status
	case "sprint_id":
		current = story.Sprint
	case "objective_id":
		current = story.Objective
	case "key_result_id":
		current = story.KeyResult
	case "start_date":
		current = story.StartDate
	case "end_date":
		current = story.EndDate
	default:
		return false
	}
	return reflect.DeepEqual(current, value)
}

func parseOptionalUUID(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	value, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be a UUID", ErrInvalidToolArguments, field)
	}
	return &value, nil
}

func parseOptionalDate(raw *string, field, timezone string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	rawValue := strings.TrimSpace(*raw)
	value, err := time.Parse(time.RFC3339, rawValue)
	if err != nil {
		location, locationErr := time.LoadLocation(strings.TrimSpace(timezone))
		if locationErr != nil {
			location = time.UTC
		}
		value, err = time.ParseInLocation("2006-01-02", rawValue, location)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be YYYY-MM-DD or RFC3339", ErrInvalidToolArguments, field)
	}
	value = value.UTC()
	return &value, nil
}

func parseUUIDList(raw []string, field string) ([]uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	values := make([]uuid.UUID, 0, len(raw))
	seen := make(map[uuid.UUID]struct{}, len(raw))
	for _, item := range raw {
		value, err := uuid.Parse(strings.TrimSpace(item))
		if err != nil || value == uuid.Nil {
			return nil, fmt.Errorf("%w: %s contains an invalid UUID", ErrInvalidToolArguments, field)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, nil
}

func storyMutationResult(status string, operation StoryMutationOperation, story messagingStory, teamCode string) StoryMutationResult {
	return StoryMutationResult{
		Status:                   status,
		Operation:                operation,
		StoryID:                  story.ID,
		Reference:                storyReference(strings.ToUpper(teamCode), story.SequenceID),
		TeamID:                   story.Team,
		Title:                    story.Title,
		Priority:                 story.Priority,
		AssigneeID:               story.Assignee,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
	}
}
