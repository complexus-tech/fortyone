package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	eventStreamKey      = "events-stream"
	eventConsumerGroup  = "events-processors"
	streamReadCount     = 10
	pendingClaimTimeout = time.Minute * 5
)

type Consumer struct {
	redis         *redis.Client
	log           *logger.Logger
	notifications *notifications.Service
	emailService  email.Service
	stories       *stories.Service
	objectives    *objectives.Service
	users         *users.Service
	statuses      *states.Service
	websiteURL    string
}

func New(redis *redis.Client, db *sqlx.DB, log *logger.Logger, websiteURL string, notifications *notifications.Service, emailService email.Service, stories *stories.Service, objectives *objectives.Service, users *users.Service, statuses *states.Service) *Consumer {
	return &Consumer{
		redis:         redis,
		log:           log,
		notifications: notifications,
		emailService:  emailService,
		stories:       stories,
		objectives:    objectives,
		users:         users,
		statuses:      statuses,
		websiteURL:    websiteURL,
	}
}

// Start initializes and runs the consumer using Redis Streams
func (c *Consumer) Start(ctx context.Context) error {
	// Generate a unique ID for this consumer instance
	instanceID := uuid.New().String()
	c.log.Info(ctx, "starting redis stream consumer", "instance_id", instanceID)

	// Create consumer group (idempotent operation)
	err := c.redis.XGroupCreateMkStream(ctx, eventStreamKey, eventConsumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		c.log.Error(ctx, "failed to create consumer group", "error", err, "stream", eventStreamKey, "group", eventConsumerGroup)
		// Continue anyway - the group might already exist
	}

	// Start a goroutine to claim pending messages periodically
	go c.claimPendingMessages(ctx, instanceID)

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.processNewMessages(ctx, instanceID); err != nil {
				c.log.Error(ctx, "error processing messages", "error", err)
				// Wait a bit before retrying to avoid tight loop
				time.Sleep(time.Second)
			}
		}
	}
}

// processNewMessages reads and processes new messages from the stream
func (c *Consumer) processNewMessages(ctx context.Context, instanceID string) error {
	// Read new messages from stream
	streams, err := c.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    eventConsumerGroup,
		Consumer: instanceID,
		Streams:  []string{eventStreamKey, ">"}, // ">" means new messages only
		Count:    streamReadCount,
		Block:    time.Second * 2, // Block with timeout for efficiency
	}).Result()

	if err != nil {
		if err == redis.Nil || strings.Contains(err.Error(), "NOGROUP") {
			// No messages or group doesn't exist yet
			return nil
		}
		return err
	}

	// Process messages
	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := c.processStreamMessage(ctx, message, instanceID); err != nil {
				c.log.Error(ctx, "failed to process message", "message_id", message.ID, "error", err)
				// Continue with other messages
			}
		}
	}

	return nil
}

// processStreamMessage processes a single message from the stream
func (c *Consumer) processStreamMessage(ctx context.Context, message redis.XMessage, instanceID string) error {
	// Extract event data from the message
	eventType, ok := message.Values["type"].(string)
	if !ok {
		return fmt.Errorf("invalid event type in message")
	}

	payloadStr, ok := message.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("invalid payload in message")
	}

	// Parse the event
	var event events.Event
	event.Type = events.EventType(eventType)

	// Unmarshal the full event first
	if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Handle the event
	if err := c.handleEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	// Acknowledge the message
	if err := c.redis.XAck(ctx, eventStreamKey, eventConsumerGroup, message.ID).Err(); err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	return nil
}

// claimPendingMessages periodically claims pending messages from other consumers
func (c *Consumer) claimPendingMessages(ctx context.Context, instanceID string) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get pending messages from all consumers
			pending, err := c.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: eventStreamKey,
				Group:  eventConsumerGroup,
				Start:  "-",
				End:    "+",
				Count:  50,
			}).Result()

			if err != nil {
				c.log.Error(ctx, "failed to get pending messages", "error", err)
				continue
			}

			// Claim messages that are pending for too long
			for _, p := range pending {
				if p.Idle > pendingClaimTimeout {
					claimed, err := c.redis.XClaim(ctx, &redis.XClaimArgs{
						Stream:   eventStreamKey,
						Group:    eventConsumerGroup,
						Consumer: instanceID,
						MinIdle:  pendingClaimTimeout,
						Messages: []string{p.ID},
					}).Result()

					if err != nil {
						c.log.Error(ctx, "failed to claim message", "message_id", p.ID, "error", err)
						continue
					}

					// Process claimed messages
					for _, msg := range claimed {
						if err := c.processStreamMessage(ctx, msg, instanceID); err != nil {
							c.log.Error(ctx, "failed to process claimed message",
								"message_id", msg.ID, "error", err)
						}
					}
				}
			}
		}
	}
}

// handleEvent routes events to the appropriate handler based on the event type
func (c *Consumer) handleEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.StoryUpdated:
		return c.handleStoryUpdated(ctx, event)
	case events.StoryCommented:
		return c.handleStoryCommented(ctx, event)
	case events.ObjectiveUpdated:
		return c.handleObjectiveUpdated(ctx, event)
	case events.KeyResultUpdated:
		return c.handleKeyResultUpdated(ctx, event)
	case events.EmailVerification:
		return c.handleEmailVerification(ctx, event)
	case events.InvitationEmail:
		return c.handleInvitationEmail(ctx, event)
	case events.InvitationAccepted:
		return c.handleInvitationAccepted(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// getStatusName attempts to get the name of a status using the states service
func (c *Consumer) getStatusName(ctx context.Context, workspaceID uuid.UUID, statusID uuid.UUID) string {
	if c.statuses == nil {
		return "new status"
	}
	// Now we can use the direct Get method on the states service
	state, err := c.statuses.Get(ctx, workspaceID, statusID)
	if err != nil {
		c.log.Error(ctx, "failed to get status", "error", err, "status_id", statusID)
		return "new status"
	}

	return state.Name
}

// generateStoryUpdateDescription creates a human-readable description of a story update
func (c *Consumer) generateStoryUpdateDescription(actorUsername string, updates map[string]any, recipientID uuid.UUID, originalAssigneeID *uuid.UUID, newAssigneeID *uuid.UUID, workspaceID uuid.UUID) string {
	// For assignee changes - handle different messages for original and new assignee
	if newAssigneeValue, exists := updates["assignee_id"]; exists {
		newAssigneeUUID, ok := newAssigneeValue.(uuid.UUID)

		if originalAssigneeID != nil && recipientID == *originalAssigneeID {
			// Message for the original assignee
			if newAssigneeValue == nil {
				return fmt.Sprintf("%s unassigned your story", actorUsername)
			} else if ok {
				// Try to get new assignee's username
				newAssignee, err := c.users.GetUser(context.Background(), newAssigneeUUID)
				if err == nil {
					return fmt.Sprintf("%s assigned your story to %s", actorUsername, newAssignee.Username)
				}
				return fmt.Sprintf("%s assigned your story to someone else", actorUsername)
			}
		} else if newAssigneeID != nil && recipientID == *newAssigneeID {
			// Message for the new assignee
			return fmt.Sprintf("%s assigned you a story", actorUsername)
		}
	}

	// Handle other update types
	if statusValue, exists := updates["status_id"]; exists {
		// Try to get status name from states service
		statusUUID, err := uuid.Parse(statusValue.(string))
		if err == nil {
			statusName := c.getStatusName(context.Background(), workspaceID, statusUUID)
			return fmt.Sprintf("%s changed the status to %s", actorUsername, statusName)
		}
		return fmt.Sprintf("%s changed the story status", actorUsername)
	}

	if priorityValue, exists := updates["priority"]; exists {
		// Format priority nicely
		return fmt.Sprintf("%s changed the priority to %v", actorUsername, priorityValue)
	}

	if dueDateValue, exists := updates["end_date"]; exists {
		if dueDateValue == nil {
			return fmt.Sprintf("%s removed due date", actorUsername)
		} else {
			// Format date nicely
			dateStr := "new date"

			// Since dates are always strings, prioritize string handling
			if strDate, ok := dueDateValue.(string); ok {
				// Handle string date format
				c.log.Info(context.Background(), "due date is string", "date_string", strDate)

				// Try to parse the date string in various formats
				var parsedTime time.Time
				var err error

				// Try common formats
				formats := []string{
					"2006-01-02",           // YYYY-MM-DD
					"2006-01-02T15:04:05Z", // ISO8601
					time.RFC3339,
				}

				for _, format := range formats {
					parsedTime, err = time.Parse(format, strDate)
					if err == nil {
						dateStr = parsedTime.Format("2 Jan")
						break
					}
				}

				if err != nil {
					c.log.Error(context.Background(), "failed to parse date string",
						"error", err,
						"date_string", strDate)
				}
			} else if dateTimePtr, ok := dueDateValue.(*time.Time); ok && dateTimePtr != nil {
				// Fallback for *time.Time
				dateStr = dateTimePtr.Format("2 Jan")
			} else if dateTime, ok := dueDateValue.(time.Time); ok {
				// Fallback for direct time.Time
				dateStr = dateTime.Format("2 Jan")
			} else {
				// Log if it's an unexpected type
				c.log.Info(context.Background(), "due date is unexpected type",
					"type", fmt.Sprintf("%T", dueDateValue),
					"raw_value", fmt.Sprintf("%v", dueDateValue))
			}

			return fmt.Sprintf("%s changed due date to %s", actorUsername, dateStr)
		}
	}

	// Default case
	return fmt.Sprintf("%s updated a story", actorUsername)
}

// Include all the necessary handlers for the different event types
func (c *Consumer) handleStoryUpdated(ctx context.Context, event events.Event) error {
	var payload events.StoryUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var title string

	story, err := c.stories.Get(ctx, payload.StoryID, payload.WorkspaceID)
	if err != nil {
		c.log.Error(ctx, "failed to get story", "error", err, "story_id", payload.StoryID)
		title = "Story updated" // Fallback title
	} else {
		title = story.Title
	}

	// Get actor's username
	var actorUsername string
	actorUser, err := c.users.GetUser(ctx, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to get actor user", "error", err, "actor_id", event.ActorID)
		actorUsername = "Someone" // Fallback if we can't get the username
	} else {
		actorUsername = actorUser.Username
	}

	// Handle assignee change - if payload.Updates contains an assignee_id key
	if newAssigneeValue, isAssigneeUpdated := payload.Updates["assignee_id"]; isAssigneeUpdated {
		var newAssigneeID *uuid.UUID

		// Convert the assignee value to UUID if possible
		if newAssigneeValueStr, ok := newAssigneeValue.(string); ok {
			if parsedUUID, err := uuid.Parse(newAssigneeValueStr); err == nil {
				newAssigneeID = &parsedUUID
			}
		} else if newAssigneeUUID, ok := newAssigneeValue.(uuid.UUID); ok {
			newAssigneeID = &newAssigneeUUID
		}

		// Create notification for original assignee if exists
		if payload.AssigneeID != nil && *payload.AssigneeID != event.ActorID {
			// Skip if the original assignee and new assignee are the same
			if newAssigneeID == nil || *payload.AssigneeID != *newAssigneeID {
				originalAssigneeDescription := c.generateStoryUpdateDescription(
					actorUsername,
					payload.Updates,
					*payload.AssigneeID,
					payload.AssigneeID, // old assignee
					newAssigneeID,      // new assignee
					payload.WorkspaceID,
				)

				notification := notifications.CoreNewNotification{
					RecipientID: *payload.AssigneeID,
					WorkspaceID: payload.WorkspaceID,
					Type:        "story_update",
					EntityType:  "story",
					EntityID:    payload.StoryID,
					ActorID:     event.ActorID,
					Description: originalAssigneeDescription,
					Title:       title,
				}

				if _, err := c.notifications.Create(ctx, notification); err != nil {
					c.log.Error(ctx, "failed to create notification for original assignee", "error", err)
				}
			}
		}

		// Create notification for new assignee if exists
		if newAssigneeID != nil && *newAssigneeID != event.ActorID {
			// Skip if the original assignee and new assignee are the same
			if payload.AssigneeID == nil || *payload.AssigneeID != *newAssigneeID {
				newAssigneeDescription := c.generateStoryUpdateDescription(
					actorUsername,
					payload.Updates,
					*newAssigneeID,
					payload.AssigneeID, // old assignee
					newAssigneeID,      // new assignee
					payload.WorkspaceID,
				)

				notification := notifications.CoreNewNotification{
					RecipientID: *newAssigneeID,
					WorkspaceID: payload.WorkspaceID,
					Type:        "story_update",
					EntityType:  "story",
					EntityID:    payload.StoryID,
					ActorID:     event.ActorID,
					Description: newAssigneeDescription,
					Title:       title,
				}

				if _, err := c.notifications.Create(ctx, notification); err != nil {
					c.log.Error(ctx, "failed to create notification for new assignee", "error", err)
				}
			}
		}
	} else {
		// Handle other updates (status, priority, due date)
		// Only create notification if there's an assignee to notify and they're not the actor
		if payload.AssigneeID != nil && *payload.AssigneeID != event.ActorID {
			description := c.generateStoryUpdateDescription(
				actorUsername,
				payload.Updates,
				*payload.AssigneeID,
				payload.AssigneeID, // old assignee (current assignee in this context)
				nil,                // no new assignee for non-assignee updates
				payload.WorkspaceID,
			)

			notification := notifications.CoreNewNotification{
				RecipientID: *payload.AssigneeID,
				WorkspaceID: payload.WorkspaceID,
				Type:        "story_update",
				EntityType:  "story",
				EntityID:    payload.StoryID,
				ActorID:     event.ActorID,
				Description: description,
				Title:       title,
			}

			if _, err := c.notifications.Create(ctx, notification); err != nil {
				c.log.Error(ctx, "failed to create notification for story update", "error", err)
			}
		}
	}

	return nil
}

func (c *Consumer) handleStoryCommented(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleStoryCommented", "event_type", event.Type)
	var payload events.StoryCommentedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// TODO: Get story details to determine who to notify (e.g., story assignee, other participants)
	// For now, we'll just notify the parent comment author if it exists and is not the current actor.
	// This assumes payload.ParentID is the *author* of the parent comment.
	// A more robust solution would be to fetch the parent comment and get its author's ID.

	if payload.ParentID != nil && *payload.ParentID != event.ActorID {
		// Attempt to get parent comment author. This is a placeholder logic.
		// In a real scenario, you would fetch the parent comment and then its author.
		// For now, we directly use ParentID as RecipientID, assuming it's a user ID.
		parentCommentAuthorID := *payload.ParentID

		// Fetch story for context if needed for the title or other details
		story, err := c.stories.Get(ctx, payload.StoryID, payload.WorkspaceID)
		var storyTitle string
		if err != nil {
			c.log.Error(ctx, "failed to get story for comment notification", "error", err, "story_id", payload.StoryID)
			storyTitle = "New reply to your comment"
		} else {
			storyTitle = fmt.Sprintf("Reply in: %s", story.Title)
		}

		actor, err := c.users.GetUser(ctx, event.ActorID)
		actorName := "Someone"
		if err == nil {
			actorName = actor.Username
		}

		notification := notifications.CoreNewNotification{
			RecipientID: parentCommentAuthorID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "story_comment_reply", // More specific type
			EntityType:  "comment",
			EntityID:    payload.CommentID, // The ID of the new comment (the reply)
			ActorID:     event.ActorID,
			Title:       storyTitle,
			Description: fmt.Sprintf("%s replied to your comment", actorName),
		}

		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification for comment reply", "error", err)
			// Decide if this error should be returned or just logged.
			// For now, logging and continuing.
		}
	}

	// Additional notifications could be added here, e.g., for story followers or @mentions

	return nil
}

func (c *Consumer) handleObjectiveUpdated(ctx context.Context, event events.Event) error {
	var payload events.ObjectiveUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleObjectiveUpdated", "objective_id", payload.ObjectiveID, "lead_id", payload.LeadID)

	// Create notification for the lead if assigned/changed and is not the actor
	if payload.LeadID != nil && *payload.LeadID != event.ActorID {
		objective, err := c.objectives.Get(ctx, payload.ObjectiveID, payload.WorkspaceID)
		objectiveName := "an objective"
		if err == nil {
			objectiveName = objective.Name
		} else {
			c.log.Error(ctx, "failed to get objective details for notification", "error", err, "objective_id", payload.ObjectiveID)
		}

		actor, err := c.users.GetUser(ctx, event.ActorID)
		actorName := "Someone"
		if err == nil {
			actorName = actor.Username
		}

		notification := notifications.CoreNewNotification{
			RecipientID: *payload.LeadID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "objective_assigned", // Or objective_lead_changed
			EntityType:  "objective",
			EntityID:    payload.ObjectiveID,
			ActorID:     event.ActorID,
			Title:       fmt.Sprintf("You are now leading: %s", objectiveName),
			Description: fmt.Sprintf("%s assigned you as lead for the objective: %s", actorName, objectiveName),
		}

		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification for objective lead assignment", "error", err)
			// Decide if this error should be returned or just logged.
		}
	}

	// TODO: Handle other objective updates if necessary (e.g., status change, due date)

	return nil
}

func (c *Consumer) handleKeyResultUpdated(ctx context.Context, event events.Event) error {
	var payload events.KeyResultUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleKeyResultUpdated", "key_result_id", payload.KeyResultID)

	// TODO: Get objective lead to notify them of key result updates.
	// This requires fetching the Objective to which this Key Result belongs, then fetching the Lead of that Objective.
	// Example steps:
	// 1. Fetch KeyResult: kr, err := c.keyResults.Get(ctx, payload.KeyResultID, payload.WorkspaceID) (assuming a keyResults service)
	// 2. Fetch Objective: obj, err := c.objectives.Get(ctx, kr.ObjectiveID, payload.WorkspaceID)
	// 3. If obj.LeadID is not nil and not the event.ActorID, then create and send notification.

	c.log.Warn(ctx, "Key result update notification not yet implemented", "key_result_id", payload.KeyResultID)
	return nil
}

func (c *Consumer) handleEmailVerification(ctx context.Context, event events.Event) error {
	var payload events.EmailVerificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event.Payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal EmailVerificationPayload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleEmailVerification", "email", payload.Email, "token_type", payload.TokenType)

	subject := "Verify your email for Complexus"
	if payload.TokenType == string(users.TokenTypeLogin) {
		subject = "Confirm your Complexus Login"
	} else if payload.TokenType == string(users.TokenTypeRegistration) {
		subject = "Complete your Complexus Registration"
	}

	// Prepare template data
	templateData := map[string]any{
		"VerificationURL": fmt.Sprintf("%s/verify/%s/%s", c.websiteURL, payload.Email, payload.Token),
		"ExpiresIn":       "10 minutes", // Consider making this dynamic or aligning with actual token expiry
		"Subject":         subject,
		"Email":           payload.Email, // Make email available to the template
		"TokenType":       payload.TokenType,
	}

	// Send templated email
	templateEmail := email.TemplatedEmail{
		To:       []string{payload.Email},
		Template: "auth/verification", // Assuming template exists at templates/auth/verification.html
		Data:     templateData,
	}

	if err := c.emailService.SendTemplatedEmail(ctx, templateEmail); err != nil {
		c.log.Error(ctx, "failed to send verification email", "error", err, "recipient", payload.Email)
		return fmt.Errorf("failed to send verification email: %w", err)
	}
	c.log.Info(ctx, "Verification email sent", "recipient", payload.Email)
	return nil
}

func (c *Consumer) handleInvitationEmail(ctx context.Context, event events.Event) error {
	var payload events.InvitationEmailPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event.Payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal InvitationEmailPayload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationEmail",
		"email", payload.Email,
		"workspace_id", payload.WorkspaceID,
		"inviter_name", payload.InviterName)

	// Calculate expiration duration
	expiresIn := "soon" // Default fallback
	if !payload.ExpiresAt.IsZero() {
		duration := time.Until(payload.ExpiresAt)
		if duration > 0 {
			hours := int(duration.Hours())
			if hours > 0 {
				expiresIn = fmt.Sprintf("%d hour(s)", hours)
			} else {
				minutes := int(duration.Minutes())
				expiresIn = fmt.Sprintf("%d minute(s)", minutes)
			}
		} else {
			expiresIn = "now expired"
		}
	}

	// Prepare template data
	templateData := map[string]any{
		"InviterName":     payload.InviterName,
		"WorkspaceName":   payload.WorkspaceName,
		"Role":            payload.Role,
		"ExpiresAt":       payload.ExpiresAt,
		"ExpiresIn":       expiresIn,
		"Subject":         fmt.Sprintf("%s has invited you to join %s on Complexus", payload.InviterName, payload.WorkspaceName),
		"VerificationURL": fmt.Sprintf("%s/onboarding/join?token=%s", c.websiteURL, payload.Token),
	}

	// Send templated email
	templateEmail := email.TemplatedEmail{
		To:       []string{payload.Email},
		Template: "invites/invitation",
		Data:     templateData,
	}

	if err := c.emailService.SendTemplatedEmail(ctx, templateEmail); err != nil {
		c.log.Error(ctx, "failed to send invitation email", "error", err, "recipient", payload.Email)
		return fmt.Errorf("failed to send invitation email: %w", err)
	}
	c.log.Info(ctx, "Invitation email sent", "recipient", payload.Email)
	return nil
}

func (c *Consumer) handleInvitationAccepted(ctx context.Context, event events.Event) error {
	var payload events.InvitationAcceptedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event.Payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal InvitationAcceptedPayload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationAccepted",
		"inviter_email", payload.InviterEmail,
		"invitee_email", payload.InviteeEmail,
		"workspace_id", payload.WorkspaceID)

	// Prepare template data for notifying the inviter
	templateData := map[string]any{
		"InviterName":   payload.InviterName,
		"InviteeName":   payload.InviteeName,
		"InviteeEmail":  payload.InviteeEmail,
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceSlug": payload.WorkspaceSlug,
		"Role":          payload.Role,
		"Subject":       fmt.Sprintf("%s (%s) has accepted your invitation to %s", payload.InviteeName, payload.InviteeEmail, payload.WorkspaceName),
		"LoginURL":      fmt.Sprintf("%s/login", c.websiteURL), // Or a direct link to the workspace: fmt.Sprintf("%s/workspaces/%s", c.websiteURL, payload.WorkspaceSlug)
		"WorkspaceURL":  fmt.Sprintf("%s/workspaces/%s", c.websiteURL, payload.WorkspaceSlug),
	}

	// Send templated email
	templateEmail := email.TemplatedEmail{
		To:       []string{payload.InviterEmail}, // Send to the person who made the invitation
		Template: "invites/acceptance",           // Assuming template exists at templates/invites/acceptance.html
		Data:     templateData,
	}

	if err := c.emailService.SendTemplatedEmail(ctx, templateEmail); err != nil {
		c.log.Error(ctx, "failed to send invitation accepted email", "error", err, "recipient", payload.InviterEmail)
		return fmt.Errorf("failed to send invitation accepted email: %w", err)
	}
	c.log.Info(ctx, "Invitation accepted email sent", "recipient", payload.InviterEmail)
	return nil
}
