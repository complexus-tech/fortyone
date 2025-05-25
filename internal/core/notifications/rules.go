package notifications

import (
	"context"
	"fmt"

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
			notification := createAssignmentNotification(*newAssigneeID, payload, actorID, storyTitle, actorUsername)
			notifications = append(notifications, notification)
		}
	}

	// Rule 2: Story update notification (status, priority, due_date)
	if hasRelevantUpdates(payload.Updates) && payload.AssigneeID != nil {
		if shouldNotify(*payload.AssigneeID, actorID) {
			notification := createUpdateNotification(*payload.AssigneeID, payload, actorID, storyTitle, actorUsername)
			notifications = append(notifications, notification)
		}
	}

	// Rule 3: Unassignment notification
	if isUnassignment(payload.Updates) && payload.AssigneeID != nil {
		if shouldNotify(*payload.AssigneeID, actorID) {
			notification := createUnassignmentNotification(*payload.AssigneeID, payload, actorID, storyTitle, actorUsername)
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
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
func createAssignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, actorUsername string) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_assignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: fmt.Sprintf("%s assigned you a story", actorUsername),
	}
}

// Create update notification
func createUpdateNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, actorUsername string) CoreNewNotification {
	updateType := getUpdateType(payload.Updates)
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_update",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: fmt.Sprintf("%s updated the %s", actorUsername, updateType),
	}
}

// Create unassignment notification
func createUnassignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle, actorUsername string) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_unassignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Description: fmt.Sprintf("%s unassigned your story", actorUsername),
	}
}

// Get human-readable update type
func getUpdateType(updates map[string]any) string {
	if _, exists := updates["status_id"]; exists {
		return "status"
	}
	if _, exists := updates["priority"]; exists {
		return "priority"
	}
	if _, exists := updates["end_date"]; exists {
		return "due date"
	}
	return "story"
}
