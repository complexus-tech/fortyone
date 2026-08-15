package notifications

import (
	"context"
	"fmt"
	"strings"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type Rules struct {
	log      *logger.Logger
	stories  storyRulesService
	users    *users.Service
	statuses *states.Service
}

type storyRulesService interface {
	Get(ctx context.Context, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	RecordActivity(ctx context.Context, activity stories.CoreActivity) error
}

func NewRules(log *logger.Logger, stories storyRulesService, users *users.Service, statuses *states.Service) *Rules {
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
	directRecipients := make(map[uuid.UUID]struct{})

	// Handle assignment scenarios
	if r.isNewAssignment(payload) {
		directNotifications := r.handleNewAssignment(ctx, payload, actorID)
		notifications = append(notifications, directNotifications...)
		addNotificationRecipients(directRecipients, directNotifications)
	}

	if r.isReassignment(payload) {
		directNotifications := r.handleReassignment(ctx, payload, actorID)
		notifications = append(notifications, directNotifications...)
		addNotificationRecipients(directRecipients, directNotifications)
	}

	if r.isPureUnassignment(payload) {
		directNotifications := r.handlePureUnassignment(ctx, payload, actorID)
		notifications = append(notifications, directNotifications...)
		addNotificationRecipients(directRecipients, directNotifications)
	}

	collaboratorNotifications, collaboratorRecipients := r.handleCollaboratorUpdates(ctx, payload, actorID)
	notifications = append(notifications, collaboratorNotifications...)
	for recipientID := range collaboratorRecipients {
		directRecipients[recipientID] = struct{}{}
	}

	// Handle other story updates for collaborators, watchers, and the assignee.
	if r.hasNonAssignmentUpdates(payload) {
		notifications = append(notifications, r.handleStoryUpdates(ctx, payload, actorID, directRecipients)...)
	}
	notifications = append(notifications, r.handleScheduleTransition(ctx, payload, actorID, directRecipients)...)

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

	excluded := uuidSet(payload.Mentions)
	for _, recipientID := range storyAudience(payload.AudienceIDs, payload.AudienceResolved, payload.AssigneeID) {
		if !shouldNotify(recipientID, actorID) {
			continue
		}
		if _, mentioned := excluded[recipientID]; mentioned {
			continue
		}
		message := NotificationMessage{
			Template: fmt.Sprintf("{actor} left a comment: %s", payload.Content),
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: recipientID,
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

	audienceIDs := storyAudience(payload.AudienceIDs, payload.AudienceResolved, nil)
	recipients := append([]uuid.UUID{payload.ParentAuthorID}, audienceIDs...)
	excluded := uuidSet(payload.Mentions)
	seen := make(map[uuid.UUID]struct{}, len(recipients))
	for _, recipientID := range recipients {
		if !shouldNotify(recipientID, actorID) {
			continue
		}
		if _, mentioned := excluded[recipientID]; mentioned {
			continue
		}
		if _, exists := seen[recipientID]; exists {
			continue
		}
		seen[recipientID] = struct{}{}
		message := NotificationMessage{
			Template: fmt.Sprintf("{actor} replied: %s", payload.Content),
			Variables: map[string]Variable{
				"actor": {Value: actorUsername, Type: "actor"},
			},
		}

		notification := CoreNewNotification{
			RecipientID: recipientID,
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

// ProcessFeedbackCommentCreated notifies the feedback author when another
// person contributes to the public discussion.
func (r *Rules) ProcessFeedbackCommentCreated(ctx context.Context, payload events.FeedbackCommentCreatedPayload, actorID uuid.UUID) []CoreNewNotification {
	if !shouldNotify(payload.RecipientID, actorID) {
		return nil
	}

	actorName := strings.TrimSpace(payload.ActorName)
	if actorName == "" {
		actorName = r.getUserName(ctx, actorID)
	}
	if actorName == "" {
		actorName = "Someone"
	}

	template := "{actor} commented on your feedback"
	if payload.IsReply {
		template = "{actor} replied to your comment"
	}

	return []CoreNewNotification{{
		DedupeKey:   fmt.Sprintf("feedback-comment:%s:%s", payload.CommentID, payload.RecipientID),
		RecipientID: payload.RecipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "feedback_comment",
		EntityType:  "feedback",
		EntityID:    payload.FeedbackID,
		ActorID:     actorID,
		Title:       payload.FeedbackTitle,
		Message: NotificationMessage{
			Template: template,
			Variables: map[string]Variable{
				"actor": {Value: actorName, Type: "actor"},
			},
		},
	}}
}

// ProcessFeedbackStatusUpdated notifies the feedback author when the team
// advances or closes the item on the public roadmap.
func (r *Rules) ProcessFeedbackStatusUpdated(ctx context.Context, payload events.FeedbackStatusUpdatedPayload, actorID uuid.UUID) []CoreNewNotification {
	if !shouldNotify(payload.RecipientID, actorID) {
		return nil
	}

	actorName := r.getUserName(ctx, actorID)
	if actorName == "" {
		actorName = "Someone"
	}
	status := strings.ReplaceAll(payload.Status, "_", " ")

	return []CoreNewNotification{{
		DedupeKey:   fmt.Sprintf("feedback-status:%s:%s", payload.EventID, payload.RecipientID),
		RecipientID: payload.RecipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "feedback_status_update",
		EntityType:  "feedback",
		EntityID:    payload.FeedbackID,
		ActorID:     actorID,
		Title:       payload.FeedbackTitle,
		Message: NotificationMessage{
			Template: "{actor} marked your feedback as {status}",
			Variables: map[string]Variable{
				"actor":  {Value: actorName, Type: "actor"},
				"status": {Value: status, Type: "value"},
			},
		},
	}}
}

func (r *Rules) ProcessFeedbackUpdatePublished(ctx context.Context, payload events.FeedbackUpdatePublishedPayload, actorID uuid.UUID) []CoreNewNotification {
	if payload.LinkedItemID == uuid.Nil || !shouldNotify(payload.RecipientID, actorID) {
		return nil
	}
	actorName := r.getUserName(ctx, actorID)
	if actorName == "" {
		actorName = "Someone"
	}
	publicationID := payload.PublicationEventID
	if publicationID == uuid.Nil {
		// Backward compatibility for events emitted before publication outbox
		// identities were introduced.
		publicationID = payload.UpdateID
	}
	return []CoreNewNotification{{
		DedupeKey:   fmt.Sprintf("feedback-update:%s:%s", publicationID, payload.RecipientID),
		RecipientID: payload.RecipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "feedback_update_published",
		EntityType:  "feedback",
		EntityID:    payload.LinkedItemID,
		ActorID:     actorID,
		Title:       payload.UpdateTitle,
		Message: NotificationMessage{
			Template: "{actor} published a feedback update: {update}",
			Variables: map[string]Variable{
				"actor":  {Value: actorName, Type: "actor"},
				"update": {Value: payload.UpdateTitle, Type: "value"},
			},
		},
	}}
}

func (r *Rules) ProcessFeedbackItemMerged(ctx context.Context, payload events.FeedbackItemMergedPayload, actorID uuid.UUID) []CoreNewNotification {
	if payload.MergeEventID == uuid.Nil || payload.TargetItemID == uuid.Nil ||
		strings.TrimSpace(payload.TargetItemTitle) == "" || !shouldNotify(payload.RecipientID, actorID) {
		return nil
	}
	actorName := r.getUserName(ctx, actorID)
	if actorName == "" {
		actorName = "Someone"
	}
	return []CoreNewNotification{{
		DedupeKey:   fmt.Sprintf("feedback-merge:%s:%s", payload.MergeEventID, payload.RecipientID),
		RecipientID: payload.RecipientID,
		WorkspaceID: payload.WorkspaceID,
		Type:        "feedback_item_merged",
		EntityType:  "feedback",
		EntityID:    payload.TargetItemID,
		ActorID:     actorID,
		Title:       payload.TargetItemTitle,
		Message: NotificationMessage{
			Template: "{actor} merged feedback into {feedback}",
			Variables: map[string]Variable{
				"actor":    {Value: actorName, Type: "actor"},
				"feedback": {Value: payload.TargetItemTitle, Type: "value"},
			},
		},
	}}
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

	if shouldNotify(payload.MentionedUser, actorID) {
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
