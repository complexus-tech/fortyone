package notifications

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

// isNewAssignment detects when someone is assigned to a story (initial assignment OR reassignment to them)
func (r *Rules) isNewAssignment(payload events.StoryUpdatedPayload) bool {
	return getNewAssignee(payload.Updates) != nil
}

// isReassignment detects when a story is reassigned from one person to another (different people)
func (r *Rules) isReassignment(payload events.StoryUpdatedPayload) bool {
	oldAssigneeID := payload.AssigneeID
	newAssigneeID := getNewAssignee(payload.Updates)

	// Must have both old and new assignee, and they must be different people
	return oldAssigneeID != nil && newAssigneeID != nil && *oldAssigneeID != *newAssigneeID
}

// isPureUnassignment detects when someone is unassigned without reassigning to anyone else
func (r *Rules) isPureUnassignment(payload events.StoryUpdatedPayload) bool {
	return payload.AssigneeID != nil && isUnassignment(payload.Updates)
}

// hasNonAssignmentUpdates detects updates to status, priority, due date (excluding assignment changes)
func (r *Rules) hasNonAssignmentUpdates(payload events.StoryUpdatedPayload) bool {
	if payload.AssigneeID == nil {
		return false // No current assignee to notify
	}

	nonAssignmentFields := []string{"status_id", "priority", "end_date"}
	for _, field := range nonAssignmentFields {
		if _, exists := payload.Updates[field]; exists {
			return true
		}
	}
	return false
}

// handleNewAssignment creates notifications for new assignments (including self-assignment)
func (r *Rules) handleNewAssignment(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) []CoreNewNotification {
	newAssigneeID := getNewAssignee(payload.Updates)
	if newAssigneeID == nil {
		return nil
	}

	actorName := r.getUserName(ctx, actorID)
	storyTitle := r.getStoryTitle(ctx, payload.StoryID, payload.WorkspaceID)

	// Self-assignment (always notify for activity tracking)
	if *newAssigneeID == actorID {
		message := NotificationMessage{
			Template: "{actor} assigned story to themselves",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
			},
		}
		return []CoreNewNotification{
			r.createNotification(*newAssigneeID, payload, actorID, "story_update", storyTitle, message),
		}
	}

	// Assignment to someone else
	if shouldNotify(*newAssigneeID, actorID) {
		message := NotificationMessage{
			Template: "{actor} assigned you a story",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
			},
		}
		return []CoreNewNotification{
			r.createNotification(*newAssigneeID, payload, actorID, "story_update", storyTitle, message),
		}
	}

	return nil
}

// handleReassignment creates notifications for reassignments (notify old assignee about who it went to)
func (r *Rules) handleReassignment(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) []CoreNewNotification {
	oldAssigneeID := payload.AssigneeID
	newAssigneeID := getNewAssignee(payload.Updates)

	if oldAssigneeID == nil || newAssigneeID == nil || !shouldNotify(*oldAssigneeID, actorID) {
		return nil
	}

	actorName := r.getUserName(ctx, actorID)
	newAssigneeName := r.getUserName(ctx, *newAssigneeID)
	storyTitle := r.getStoryTitle(ctx, payload.StoryID, payload.WorkspaceID)

	message := NotificationMessage{
		Template: "{actor} reassigned story to {assignee}",
		Variables: map[string]Variable{
			"actor":    {Value: actorName, Type: "actor"},
			"assignee": {Value: newAssigneeName, Type: "assignee"},
		},
	}

	return []CoreNewNotification{
		r.createNotification(*oldAssigneeID, payload, actorID, "story_update", storyTitle, message),
	}
}

// handlePureUnassignment creates notifications for pure unassignments (no new assignee)
func (r *Rules) handlePureUnassignment(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) []CoreNewNotification {
	oldAssigneeID := payload.AssigneeID
	if oldAssigneeID == nil || !shouldNotify(*oldAssigneeID, actorID) {
		return nil
	}

	actorName := r.getUserName(ctx, actorID)
	storyTitle := r.getStoryTitle(ctx, payload.StoryID, payload.WorkspaceID)

	message := NotificationMessage{
		Template: "{actor} removed your assignment",
		Variables: map[string]Variable{
			"actor": {Value: actorName, Type: "actor"},
		},
	}

	return []CoreNewNotification{
		r.createNotification(*oldAssigneeID, payload, actorID, "story_update", storyTitle, message),
	}
}

// handleStoryUpdates creates notifications for non-assignment updates (status, priority, due date)
func (r *Rules) handleStoryUpdates(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) []CoreNewNotification {
	if payload.AssigneeID == nil || !shouldNotify(*payload.AssigneeID, actorID) {
		return nil
	}

	actorName := r.getUserName(ctx, actorID)
	storyTitle := r.getStoryTitle(ctx, payload.StoryID, payload.WorkspaceID)
	message := r.generateNonAssignmentUpdateMessage(actorName, payload.Updates, ctx)

	return []CoreNewNotification{
		r.createNotification(*payload.AssigneeID, payload, actorID, "story_update", storyTitle, message),
	}
}

// generateNonAssignmentUpdateMessage creates messages for status, priority, due date updates
func (r *Rules) generateNonAssignmentUpdateMessage(actorName string, updates map[string]any, ctx context.Context) NotificationMessage {
	// Priority update
	if priorityValue, exists := updates["priority"]; exists {
		priority := "Unknown"
		if priorityStr, ok := priorityValue.(string); ok {
			priority = priorityStr
		}
		return NotificationMessage{
			Template: "{actor} set the {field} to {value}",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
				"field": {Value: "Priority", Type: "field"},
				"value": {Value: priority, Type: "value"},
			},
		}
	}

	// Status update
	if _, exists := updates["status_id"]; exists {
		return NotificationMessage{
			Template: "{actor} updated {field}",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
				"field": {Value: "Status", Type: "field"},
			},
		}
	}

	// Due date update
	if dueDateValue, exists := updates["end_date"]; exists {
		if dueDateValue == nil {
			return NotificationMessage{
				Template: "{actor} removed the {field}",
				Variables: map[string]Variable{
					"actor": {Value: actorName, Type: "actor"},
					"field": {Value: "deadline", Type: "field"},
				},
			}
		}

		dateStr := "Unknown date"
		if strDate, ok := dueDateValue.(string); ok {
			if parsedTime, err := parseDate(strDate); err == nil {
				dateStr = parsedTime.Format("2 Jan")
			}
		}

		return NotificationMessage{
			Template: "{actor} set the {field} to {value}",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
				"field": {Value: "Deadline", Type: "field"},
				"value": {Value: dateStr, Type: "date"},
			},
		}
	}

	// Default case
	return NotificationMessage{
		Template: "{actor} updated the story",
		Variables: map[string]Variable{
			"actor": {Value: actorName, Type: "actor"},
		},
	}
}

// getNewAssignee extracts the new assignee ID from updates
func getNewAssignee(updates map[string]any) *uuid.UUID {
	assigneeValue, exists := updates["assignee_id"]
	if !exists {
		return nil
	}

	// Handle different types
	if assigneeStr, ok := assigneeValue.(string); ok {
		if parsedUUID, err := uuid.Parse(assigneeStr); err == nil {
			return &parsedUUID
		}
	}
	if assigneeUUID, ok := assigneeValue.(*uuid.UUID); ok {
		return assigneeUUID
	}

	return nil
}

// isUnassignment checks if this is an unassignment (assignee_id set to null)
func isUnassignment(updates map[string]any) bool {
	assigneeValue, exists := updates["assignee_id"]
	if !exists {
		return false
	}
	return assigneeValue == nil
}
