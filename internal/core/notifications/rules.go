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
	r.log.Info(ctx, "ProcessStoryUpdate", "payload", payload, "actor_id", actorID)

	var notifications []CoreNewNotification

	// Handle assignment scenarios
	if r.isNewAssignment(payload) {
		notifications = append(notifications, r.handleNewAssignment(ctx, payload, actorID)...)
	}

	if r.isReassignment(payload) {
		notifications = append(notifications, r.handleReassignment(ctx, payload, actorID)...)
	}

	if r.isPureUnassignment(payload) {
		notifications = append(notifications, r.handlePureUnassignment(ctx, payload, actorID)...)
	}

	// Handle other story updates (status, priority, due date)
	if r.hasNonAssignmentUpdates(payload) {
		notifications = append(notifications, r.handleStoryUpdates(ctx, payload, actorID)...)
	}

	return notifications, nil
}

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

// =============================================================================
// NOTIFICATION HANDLERS
// =============================================================================

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

// getUserName gets a user's name with fallback
func (r *Rules) getUserName(ctx context.Context, userID uuid.UUID) string {
	if user, err := r.users.GetUser(ctx, userID); err == nil {
		return user.Username
	}
	return "Someone"
}

// getStoryTitle gets a story's title with fallback
func (r *Rules) getStoryTitle(ctx context.Context, storyID, workspaceID uuid.UUID) string {
	if story, err := r.stories.Get(ctx, storyID, workspaceID); err == nil {
		return story.Title
	}
	return "Story updated"
}

// createNotification creates a notification with consistent structure
func (r *Rules) createNotification(recipientID uuid.UUID, payload events.StoryUpdatedPayload, actorID uuid.UUID, notifType, title string, message NotificationMessage) CoreNewNotification {
	return CoreNewNotification{
		RecipientID: recipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        notifType,
		EntityType:  "story",
		EntityID:    payload.StoryID,
		ActorID:     actorID,
		Title:       title,
		Message:     message,
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

// =============================================================================
// UTILITY FUNCTIONS (used by handlers)
// =============================================================================

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

// shouldNotify checks if a recipient should be notified (never notify the actor)
func shouldNotify(recipientID, actorID uuid.UUID) bool {
	return recipientID != actorID
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
