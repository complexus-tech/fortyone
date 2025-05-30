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
			var message NotificationMessage

			// Check if actor is assigning to themselves
			if *newAssigneeID == actorID {
				message = NotificationMessage{
					Template: "{actor} reassigned story to themselves",
					Variables: map[string]Variable{
						"actor": {Value: actorUsername, Type: "actor"},
					},
				}
			} else {
				message = NotificationMessage{
					Template: "{actor} assigned you a story",
					Variables: map[string]Variable{
						"actor": {Value: actorUsername, Type: "actor"},
					},
				}
			}

			notification := createAssignmentNotification(*newAssigneeID, payload, actorID, storyTitle, message)
			notifications = append(notifications, notification)
		}
	}

	// Rule 2: Story update notification (status, priority, due_date)
	if hasRelevantUpdates(payload.Updates) && payload.AssigneeID != nil {
		if shouldNotify(*payload.AssigneeID, actorID) {
			message := generateUpdateMessage(actorUsername, payload.Updates, ctx, r)
			notification := createUpdateNotification(*payload.AssigneeID, payload, actorID, storyTitle, message)
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

				message := NotificationMessage{
					Template: "{actor} reassigned story to {assignee}",
					Variables: map[string]Variable{
						"actor":    {Value: actorUsername, Type: "actor"},
						"assignee": {Value: newAssigneeName, Type: "assignee"},
					},
				}

				notification := createUnassignmentNotification(*payload.AssigneeID, payload, actorID, storyTitle, message)
				notifications = append(notifications, notification)
			}
		}
	}

	return notifications, nil
}

// ProcessCommentCreated applies notification rules for comment creation
func (r *Rules) ProcessCommentCreated(ctx context.Context, payload events.CommentCreatedPayload, actorID uuid.UUID) ([]CoreNewNotification, error) {
	var notifications []CoreNewNotification

	// Get actor username
	actorUsername := "Someone"
	if r.users != nil {
		if actor, err := r.users.GetUser(ctx, actorID); err == nil {
			actorUsername = actor.Username
		}
	}

	// Rule 1: Notify story assignee when someone comments on their assigned story
	if payload.AssigneeID != nil && shouldNotify(*payload.AssigneeID, actorID) {
		message := NotificationMessage{
			Template: fmt.Sprintf("{actor} left a comment: %s", payload.Content),
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: *payload.AssigneeID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "story_comment",
			EntityType:  "story",
			EntityID:    payload.StoryID,
			ActorID:     actorID,
			Title:       payload.StoryTitle,
			Message:     message,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// ProcessCommentReplied applies notification rules for comment replies
func (r *Rules) ProcessCommentReplied(ctx context.Context, payload events.CommentRepliedPayload, actorID uuid.UUID) ([]CoreNewNotification, error) {
	var notifications []CoreNewNotification

	// Get actor username
	actorUsername := "Someone"
	if r.users != nil {
		if actor, err := r.users.GetUser(ctx, actorID); err == nil {
			actorUsername = actor.Username
		}
	}

	// Rule 1: Notify parent comment author when someone replies to their comment
	if shouldNotify(payload.ParentAuthorID, actorID) {
		message := NotificationMessage{
			Template: fmt.Sprintf("{actor} replied: %s", payload.Content),
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: payload.ParentAuthorID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "comment_reply",
			EntityType:  "story",
			EntityID:    payload.StoryID,
			ActorID:     actorID,
			Title:       payload.StoryTitle,
			Message:     message,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// ProcessUserMentioned applies notification rules for user mentions
func (r *Rules) ProcessUserMentioned(ctx context.Context, payload events.UserMentionedPayload, actorID uuid.UUID) ([]CoreNewNotification, error) {
	var notifications []CoreNewNotification

	// Get actor username
	actorUsername := "Someone"
	if r.users != nil {
		if actor, err := r.users.GetUser(ctx, actorID); err == nil {
			actorUsername = actor.Username
		}
	}

	// Rule 1: Notify mentioned user (if not the actor)
	if shouldNotify(payload.MentionedUser, actorID) {
		// Check if this is a comment on a story - if so, get the story to check assignee
		// to avoid duplicate notifications if the mentioned user is also the assignee
		if r.stories != nil {
			if story, err := r.stories.Get(ctx, payload.StoryID, payload.WorkspaceID); err == nil {
				// If the mentioned user is the story assignee, they already got a notification
				// from ProcessCommentCreated, so skip this mention notification
				if story.Assignee != nil && *story.Assignee == payload.MentionedUser {
					return notifications, nil
				}
			}
		}

		message := NotificationMessage{
			Template: fmt.Sprintf("{actor} mentioned you: %s", payload.Content),
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: payload.MentionedUser,
			WorkspaceID: payload.WorkspaceID,
			Type:        "mention",
			EntityType:  "story",
			EntityID:    payload.StoryID,
			ActorID:     actorID,
			Title:       payload.StoryTitle,
			Message:     message,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// generateUpdateMessage creates structured messages for different update types
func generateUpdateMessage(actorUsername string, updates map[string]any, ctx context.Context, r *Rules) NotificationMessage {
	// Priority update
	if priorityValue, exists := updates["priority"]; exists {
		priority := "Unknown"
		if priorityStr, ok := priorityValue.(string); ok {
			priority = priorityStr
		}

		return NotificationMessage{
			Template: "{actor} set the {field} to {value}",
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
				"field": {Value: "Priority", Type: "field"},
				"value": {Value: priority, Type: "value"},
			},
		}
	}

	// Status update
	if _, exists := updates["status_id"]; exists {
		statusName := "In Progress" // Placeholder - can be enhanced with actual status lookup

		return NotificationMessage{
			Template: "{actor} updated {field} to {value}",
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
				"field": {Value: "Status", Type: "field"},
				"value": {Value: statusName, Type: "value"},
			},
		}
	}

	// Due date update
	if dueDateValue, exists := updates["end_date"]; exists {
		if dueDateValue == nil {
			return NotificationMessage{
				Template: "{actor} removed the {field}",
				Variables: map[string]Variable{
					"actor": {Value: actorUsername, Type: "actor"},
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
				"actor": {Value: actorUsername, Type: "actor"},
				"field": {Value: "Deadline", Type: "field"},
				"value": {Value: dateStr, Type: "date"},
			},
		}
	}

	// Default case
	return NotificationMessage{
		Template: "{actor} updated the story",
		Variables: map[string]Variable{
			"actor": {Value: actorUsername, Type: "actor"},
		},
	}
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
func createAssignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle string, message NotificationMessage) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_assignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Message:     message,
	}
}

// Create update notification
func createUpdateNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle string, message NotificationMessage) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_update",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Message:     message,
	}
}

// Create unassignment notification
func createUnassignmentNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, storyTitle string, message NotificationMessage) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "story_unassignment",
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       storyTitle,
		Message:     message,
	}
}
