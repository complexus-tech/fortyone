package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type Rules struct {
	log     *logger.Logger
	stories *stories.Service
	users   *users.Service
}

func NewRules(log *logger.Logger, stories *stories.Service, users *users.Service) *Rules {
	return &Rules{
		log:     log,
		stories: stories,
		users:   users,
	}
}

// ProcessStoryUpdate applies notification rules for story updates
func (r *Rules) ProcessStoryUpdate(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) ([]CoreNewNotification, error) {
	var notifications []CoreNewNotification

	// Get story title for notifications
	story, err := r.stories.Get(ctx, payload.StoryID, payload.WorkspaceID)
	storyTitle := "Story updated"
	if err == nil {
		storyTitle = story.Title
	}

	// Get actor username
	actorUsername := "Someone"
	if actor, err := r.users.GetUser(ctx, actorID); err == nil {
		actorUsername = actor.Username
	}

	// Rule 1: New assignment notification
	if newAssigneeID := getNewAssignee(payload.Updates); newAssigneeID != nil {
		if shouldNotify(*newAssigneeID, actorID) {
			// Check if actor is assigning to themselves
			var description string
			if *newAssigneeID == actorID {
				description = fmt.Sprintf("%s reassigned story to themselves", actorUsername)
			} else {
				description = fmt.Sprintf("%s assigned you a story", actorUsername)
			}

			notification := createAssignmentNotification(*newAssigneeID, payload, actorID, storyTitle, description)
			notifications = append(notifications, notification)
		}
	}

	// Rule 2: Story update notification (status, priority, due_date)
	if hasRelevantUpdates(payload.Updates) && payload.AssigneeID != nil {
		if shouldNotify(*payload.AssigneeID, actorID) {
			description := generateUpdateDescription(actorUsername, payload.Updates, ctx, r)
			notification := createUpdateNotification(*payload.AssigneeID, payload, actorID, storyTitle, description)
			notifications = append(notifications, notification)
		}
	}

	// Rule 3: Unassignment notification
	if isUnassignment(payload.Updates) && payload.AssigneeID != nil {
		if shouldNotify(*payload.AssigneeID, actorID) {
			// Check if someone else is being assigned
			if newAssigneeID := getNewAssignee(payload.Updates); newAssigneeID != nil {
				// Get new assignee's name
				newAssigneeName := "someone else"
				if newAssignee, err := r.users.GetUser(ctx, *newAssigneeID); err == nil {
					newAssigneeName = newAssignee.Username
				}
				description := fmt.Sprintf("%s reassigned story to %s", actorUsername, newAssigneeName)
				notification := createUnassignmentNotification(*payload.AssigneeID, payload, actorID, storyTitle, description)
				notifications = append(notifications, notification)
			}
		}
	}

	return notifications, nil
}

// generateUpdateDescription creates specific messages for different update types
func generateUpdateDescription(actorUsername string, updates map[string]any, ctx context.Context, r *Rules) string {
	// Priority update
	if priorityValue, exists := updates["priority"]; exists {
		priority := "Unknown"
		if priorityStr, ok := priorityValue.(string); ok {
			priority = priorityStr
		}
		return fmt.Sprintf("%s set the priority to %s", actorUsername, priority)
	}

	// Status update
	if statusValue, exists := updates["status_id"]; exists {
		statusName := "Unknown"
		if _, ok := statusValue.(string); ok {
			// Note: You'll need to add a status service to Rules struct to fetch actual status names
			// For now, using a placeholder
			statusName = "In Progress"
		}
		return fmt.Sprintf("%s changed the status to %s", actorUsername, statusName)
	}

	// Due date update
	if dueDateValue, exists := updates["end_date"]; exists {
		if dueDateValue == nil {
			return fmt.Sprintf("%s removed the deadline", actorUsername)
		}

		dateStr := "Unknown date"
		if strDate, ok := dueDateValue.(string); ok {
			// Parse and format the date
			if parsedTime, err := parseDate(strDate); err == nil {
				dateStr = parsedTime.Format("2 Jan")
			}
		}
		return fmt.Sprintf("%s set the deadline to %s", actorUsername, dateStr)
	}

	// Default case
	return fmt.Sprintf("%s updated the story", actorUsername)
}

// parseDate tries to parse date strings in various formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"2006-01-02T15:04:05Z", // ISO8601
		time.RFC3339,
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// Rule 4: Never notify the actor
func shouldNotify(recipientID, actorID uuid.UUID) bool {
	return recipientID != actorID
}

// Check if updates contain relevant fields for notifications
func hasRelevantUpdates(updates map[string]any) bool {
	relevantFields := []string{"status_id", "priority", "end_date"}
	for _, field := range relevantFields {
		if _, exists := updates[field]; exists {
			return true
		}
	}
	return false
}

// Get new assignee ID from updates
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
	if assigneeUUID, ok := assigneeValue.(uuid.UUID); ok {
		return &assigneeUUID
	}

	return nil
}

// Check if this is an unassignment (assignee_id set to null/empty)
func isUnassignment(updates map[string]any) bool {
	assigneeValue, exists := updates["assignee_id"]
	if !exists {
		return false
	}
	return assigneeValue == nil
}

// Create assignment notification
func createAssignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, description string) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_assignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: description,
	}
}

// Create update notification
func createUpdateNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, description string) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_update",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: description,
	}
}

// Create unassignment notification
func createUnassignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, description string) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_unassignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: description,
	}
}
