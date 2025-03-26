package consumer

import (
	"context"
	"encoding/json"
	"fmt"
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

func (c *Consumer) Start(ctx context.Context) error {
	pubsub := c.redis.Subscribe(ctx,
		string(events.StoryUpdated),
		string(events.StoryCommented),
		string(events.ObjectiveUpdated),
		string(events.KeyResultUpdated),
		string(events.EmailVerification),
		string(events.InvitationEmail),
		string(events.InvitationAccepted),
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var event events.Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			c.log.Error(ctx, "failed to unmarshal event", "error", err)
			continue
		}

		if err := c.handleEvent(ctx, event); err != nil {
			c.log.Error(ctx, "failed to handle event", "error", err)
		}
	}

	return nil
}

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
		statusUUID, ok := statusValue.(uuid.UUID)
		if ok && c.statuses != nil {
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
		c.log.Error(ctx, "failed to get story", "error", err)
		title = "Story updated"
	} else {
		title = story.Title
	}

	// Get actor's username
	var actorUsername string
	actorUser, err := c.users.GetUser(ctx, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to get actor user", "error", err)
		actorUsername = "Someone" // Fallback if we can't get the username
	} else {
		actorUsername = actorUser.Username
	}

	// Handle assignee change - if payload.Updates contains an assignee_id key
	if newAssigneeValue, isAssigneeUpdated := payload.Updates["assignee_id"]; isAssigneeUpdated {
		var newAssigneeID *uuid.UUID

		// Convert the assignee value to UUID if possible
		if newAssigneeValue != nil {
			if newAssigneeUUID, ok := newAssigneeValue.(uuid.UUID); ok {
				newAssigneeID = &newAssigneeUUID
			}
		}

		// Create notification for original assignee if exists
		if payload.AssigneeID != nil && *payload.AssigneeID != event.ActorID {
			// Skip if the original assignee and new assignee are the same
			if newAssigneeID != nil && *payload.AssigneeID == *newAssigneeID {
				c.log.Info(ctx, "same assignee, not creating notification")
			} else {
				originalAssigneeDescription := c.generateStoryUpdateDescription(
					actorUsername,
					payload.Updates,
					*payload.AssigneeID,
					payload.AssigneeID,
					newAssigneeID,
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
			if payload.AssigneeID != nil && *payload.AssigneeID == *newAssigneeID {
				c.log.Info(ctx, "same assignee, not creating notification")
			} else {
				newAssigneeDescription := c.generateStoryUpdateDescription(
					actorUsername,
					payload.Updates,
					*newAssigneeID,
					payload.AssigneeID,
					newAssigneeID,
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
				payload.AssigneeID,
				nil,
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
				c.log.Error(ctx, "failed to create notification", "error", err)
			}
		}
	}

	return nil
}

func (c *Consumer) handleStoryCommented(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleStoryCommented", "event", event.Type)
	var payload events.StoryCommentedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// TODO: Get story details to determine who to notify
	// For now, we'll just notify the parent comment author if it exists
	if payload.ParentID != nil {
		notification := notifications.CoreNewNotification{
			RecipientID: *payload.ParentID, // This should be the parent comment author's ID
			WorkspaceID: payload.WorkspaceID,
			Type:        "story_comment",
			EntityType:  "comment",
			EntityID:    payload.CommentID,
			ActorID:     event.ActorID,
			Title:       "Someone replied to your comment",
		}

		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

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

	// Create notification for the lead if changed
	if payload.LeadID != nil {
		notification := notifications.CoreNewNotification{
			RecipientID: *payload.LeadID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "objective_update",
			EntityType:  "objective",
			EntityID:    payload.ObjectiveID,
			ActorID:     event.ActorID,
			Title:       "You have been assigned as the lead for an objective",
		}

		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

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

	// TODO: Get objective lead to notify them of key result updates
	// For now, we'll skip notification creation until we can get the objective lead
	return nil
}

func (c *Consumer) handleEmailVerification(ctx context.Context, event events.Event) error {
	var payload events.EmailVerificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleEmailVerification", "email", payload.Email)

	// Prepare template data
	templateData := map[string]any{
		"VerificationURL": fmt.Sprintf("%s/verify/%s/%s", c.websiteURL, payload.Email, payload.Token),
		"ExpiresIn":       "10 minutes",
		"Subject":         "Login to Complexus",
	}

	// Send templated email
	templateEmail := email.TemplatedEmail{
		To:       []string{payload.Email},
		Template: "auth/verification",
		Data:     templateData,
	}

	if err := c.emailService.SendTemplatedEmail(ctx, templateEmail); err != nil {
		c.log.Error(ctx, "failed to send verification email", "error", err)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (c *Consumer) handleInvitationEmail(ctx context.Context, event events.Event) error {
	var payload events.InvitationEmailPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationEmail",
		"email", payload.Email,
		"workspace_id", payload.WorkspaceID)

	// Calculate expiration duration
	expiresIn := time.Until(payload.ExpiresAt).Round(time.Hour)

	// Prepare template data
	templateData := map[string]any{
		"InviterName":     payload.InviterName,
		"WorkspaceName":   payload.WorkspaceName,
		"ExpiresIn":       fmt.Sprintf("%d hours", int(expiresIn.Hours())),
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
		c.log.Error(ctx, "failed to send invitation email", "error", err)
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	return nil
}

func (c *Consumer) handleInvitationAccepted(ctx context.Context, event events.Event) error {
	var payload events.InvitationAcceptedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationAccepted",
		"inviter_email", payload.InviterEmail,
		"invitee_email", payload.InviteeEmail,
		"workspace_id", payload.WorkspaceID)

	// Prepare template data
	templateData := map[string]any{
		"InviterName":   payload.InviterName,
		"InviteeName":   payload.InviteeName,
		"WorkspaceName": payload.WorkspaceName,
		"Role":          payload.Role,
		"Subject":       fmt.Sprintf("%s has accepted your invitation to %s", payload.InviteeName, payload.WorkspaceName),
		"LoginURL":      fmt.Sprintf("%s/login", c.websiteURL),
	}

	// Send templated email
	templateEmail := email.TemplatedEmail{
		To:       []string{payload.InviterEmail},
		Template: "invites/acceptance",
		Data:     templateData,
	}

	if err := c.emailService.SendTemplatedEmail(ctx, templateEmail); err != nil {
		c.log.Error(ctx, "failed to send invitation accepted email", "error", err)
		return fmt.Errorf("failed to send invitation accepted email: %w", err)
	}

	return nil
}
