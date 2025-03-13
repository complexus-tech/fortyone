package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	redis         *redis.Client
	log           *logger.Logger
	notifications *notifications.Service
	emailService  email.Service
	websiteURL    string
}

func NewConsumer(redis *redis.Client, log *logger.Logger, notifications *notifications.Service, emailService email.Service, websiteURL string) *Consumer {
	return &Consumer{
		redis:         redis,
		log:           log,
		notifications: notifications,
		emailService:  emailService,
		websiteURL:    websiteURL,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	pubsub := c.redis.Subscribe(ctx,
		string(StoryUpdated),
		string(StoryCommented),
		string(ObjectiveUpdated),
		string(KeyResultUpdated),
		string(EmailVerification),
		string(InvitationEmail),
		string(InvitationAccepted),
	)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var event Event
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

func (c *Consumer) handleEvent(ctx context.Context, event Event) error {
	switch event.Type {
	case StoryUpdated:
		return c.handleStoryUpdated(ctx, event)
	case StoryCommented:
		return c.handleStoryCommented(ctx, event)
	case ObjectiveUpdated:
		return c.handleObjectiveUpdated(ctx, event)
	case KeyResultUpdated:
		return c.handleKeyResultUpdated(ctx, event)
	case EmailVerification:
		return c.handleEmailVerification(ctx, event)
	case InvitationEmail:
		return c.handleInvitationEmail(ctx, event)
	case InvitationAccepted:
		return c.handleInvitationAccepted(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (c *Consumer) handleStoryUpdated(ctx context.Context, event Event) error {
	var payload StoryUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "events.consumer.handleStoryUpdated", "payloadS", payload)

	// Create notification for the assignee if changed
	if payload.AssigneeID != nil {
		notification := notifications.CoreNewNotification{
			RecipientID: *payload.AssigneeID,
			WorkspaceID: payload.WorkspaceID,
			Type:        "story_update",
			EntityType:  "story",
			EntityID:    payload.StoryID,
			ActorID:     event.ActorID,
			Title:       "You have been assigned to a story",
		}

		if _, err := c.notifications.Create(ctx, notification); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	return nil
}

func (c *Consumer) handleStoryCommented(ctx context.Context, event Event) error {
	c.log.Info(ctx, "events.consumer.handleStoryCommented", "event", event.Type)
	var payload StoryCommentedPayload
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

func (c *Consumer) handleObjectiveUpdated(ctx context.Context, event Event) error {
	var payload ObjectiveUpdatedPayload
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

func (c *Consumer) handleKeyResultUpdated(ctx context.Context, event Event) error {
	var payload KeyResultUpdatedPayload
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

func (c *Consumer) handleEmailVerification(ctx context.Context, event Event) error {
	var payload EmailVerificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "events.consumer.handleEmailVerification", "email", payload.Email)

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

func (c *Consumer) handleInvitationEmail(ctx context.Context, event Event) error {
	var payload InvitationEmailPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "events.consumer.handleInvitationEmail",
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

func (c *Consumer) handleInvitationAccepted(ctx context.Context, event Event) error {
	var payload InvitationAcceptedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "events.consumer.handleInvitationAccepted",
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

func (c *Consumer) generateStoryUpdateTitle(updates map[string]any) string {
	// Check for assignee_id first (highest priority)
	if _, exists := updates["assignee_id"]; exists {
		return "Jack assigned a story to you"
	}

	// Check for status_id (second priority)
	if _, exists := updates["status_id"]; exists {
		return "A story's status was updated"
	}

	// Check for priority (third priority)
	if _, exists := updates["priority"]; exists {
		return "A story's priority was changed"
	}

	// Check for due_date (lowest priority)
	if _, exists := updates["due_date"]; exists {
		return "A story's due date was modified"
	}

	// Default case if none of the above
	return "A story was updated"
}
