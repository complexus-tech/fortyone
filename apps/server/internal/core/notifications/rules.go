package notifications

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type Rules struct {
	log      *logger.Logger
	stories  *stories.Service
	users    *users.Service
	statuses *states.Service
}

func NewRules(log *logger.Logger, stories *stories.Service, users *users.Service, statuses *states.Service) *Rules {
	return &Rules{
		log:      log,
		stories:  stories,
		users:    users,
		statuses: statuses,
	}
}

// ProcessStoryCreated applies notification rules for story creation
func (r *Rules) ProcessStoryCreated(ctx context.Context, payload events.StoryCreatedPayload, actorID uuid.UUID) ([]CoreNewNotification, error) {
	r.log.Info(ctx, "ProcessStoryCreated", "payload", payload, "actor_id", actorID)

	var notifications []CoreNewNotification

	// Notify assignee if story is created with an assignee
	if payload.AssigneeID != nil && shouldNotify(*payload.AssigneeID, actorID) {
		actorName := r.getUserName(ctx, actorID)

		message := NotificationMessage{
			Template: "{actor} assigned you a new task",
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: *payload.AssigneeID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "story_update",
			EntityType:  "story",
			EntityID:    payload.StoryID,
			ActorID:     actorID,
			Title:       payload.Title,
			Message:     message,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
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
