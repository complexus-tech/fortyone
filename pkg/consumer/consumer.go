package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/brevo"
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
	redis             *redis.Client
	log               *logger.Logger
	notifications     *notifications.Service
	notificationRules *notifications.Rules
	emailService      email.Service
	brevoService      *brevo.Service
	stories           *stories.Service
	objectives        *objectives.Service
	users             *users.Service
	statuses          *states.Service
	websiteURL        string
}

func New(redis *redis.Client, db *sqlx.DB, log *logger.Logger, websiteURL string, notificationsService *notifications.Service, emailService email.Service, stories *stories.Service, objectives *objectives.Service, users *users.Service, statuses *states.Service, brevoService *brevo.Service) *Consumer {
	notificationRules := notifications.NewRules(log, stories, users, statuses)

	return &Consumer{
		redis:             redis,
		log:               log,
		notifications:     notificationsService,
		notificationRules: notificationRules,
		emailService:      emailService,
		stories:           stories,
		brevoService:      brevoService,
		objectives:        objectives,
		users:             users,
		statuses:          statuses,
		websiteURL:        websiteURL,
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
	case events.CommentCreated:
		return c.handleCommentCreated(ctx, event)
	case events.CommentReplied:
		return c.handleCommentReplied(ctx, event)
	case events.UserMentioned:
		return c.handleUserMentioned(ctx, event)
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

// handleStoryUpdated processes story update events using the new notification rules
func (c *Consumer) handleStoryUpdated(ctx context.Context, event events.Event) error {
	var payload events.StoryUpdatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleStoryUpdated", "story_id", payload.StoryID, "workspace_id", payload.WorkspaceID, "updates", payload.Updates)

	// Use notification rules to process the story update
	notifications, err := c.notificationRules.ProcessStoryUpdate(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process story update notifications", "error", err)
		return err
	}

	// Create all notifications
	for _, notification := range notifications {
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err, "recipient_id", notification.RecipientID)
			// Continue with other notifications even if one fails
		}
	}

	// Workspace broadcasting
	if c.hasSignificantChanges(payload.Updates) {
		c.broadcastToWorkspace(ctx, payload, event.ActorID)
	}

	return nil
}

// NEW: Check if updates contain workspace-worthy changes
func (c *Consumer) hasSignificantChanges(updates map[string]any) bool {
	significantFields := map[string]bool{
		"status_id":   true,
		"assignee_id": true,
		"priority":    true,
		"title":       true,
	}

	for field := range updates {
		if significantFields[field] {
			return true
		}
	}
	return false
}

// Broadcast to workspace with frontend-friendly field names
func (c *Consumer) broadcastToWorkspace(ctx context.Context, payload events.StoryUpdatedPayload, actorID uuid.UUID) {
	// Map database field names to frontend camelCase
	dbToFrontendFields := map[string]string{
		"status_id":   "statusId",
		"assignee_id": "assigneeId",
		"priority":    "priority",
		"title":       "title",
	}

	// Filter and transform field names
	frontendChanges := make(map[string]any)
	for dbField, value := range payload.Updates {
		if frontendField, isSignificant := dbToFrontendFields[dbField]; isSignificant {
			frontendChanges[frontendField] = value
		}
	}

	// Skip if no significant changes
	if len(frontendChanges) == 0 {
		return
	}

	// Get actor name
	actorName := "Someone"
	if actor, err := c.users.GetUser(ctx, actorID); err == nil {
		actorName = actor.Username
	}

	// Create frontend-friendly workspace update
	workspaceUpdate := map[string]any{
		"type":        "story.workspace_update",
		"storyId":     payload.StoryID,     // camelCase
		"workspaceId": payload.WorkspaceID, // camelCase
		"changes":     frontendChanges,     // camelCase field names
		"actorId":     actorID,             // camelCase
		"actorName":   actorName,           // Actor name for display
		"timestamp":   time.Now().Unix(),
	}

	// Publish to workspace channel
	data, err := json.Marshal(workspaceUpdate)
	if err != nil {
		c.log.Error(ctx, "failed to marshal workspace update", "error", err)
		return
	}

	channelName := fmt.Sprintf("workspace-updates:%s", payload.WorkspaceID.String())
	if err := c.redis.Publish(ctx, channelName, data).Err(); err != nil {
		c.log.Error(ctx, "failed to publish workspace update", "error", err)
		return
	}

	c.log.Debug(ctx, "workspace update broadcasted", "storyId", payload.StoryID, "changes", len(frontendChanges), "actorName", actorName)
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
			Message: notifications.NotificationMessage{
				Template: "{actor} assigned you as lead for the objective: {objective}",
				Variables: map[string]notifications.Variable{
					"actor":     {Value: actorName, Type: "actor"},
					"objective": {Value: objectiveName, Type: "value"},
				},
			},
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
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleEmailVerification", "email", payload.Email)

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"VERIFICATION_URL": fmt.Sprintf("%s/verify/%s/%s", c.websiteURL, payload.Email, payload.Token),
		"EXPIRES_IN":       "10 minutes",
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateLogin,
		To: []brevo.EmailRecipient{
			{
				Email: payload.Email,
			},
		},
		Params: brevoParams,
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send verification email via Brevo", "error", err, "email", payload.Email)
		return fmt.Errorf("failed to send verification email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent verification email via Brevo", "email", payload.Email)
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
	expiresIn := time.Until(payload.ExpiresAt)

	var expiresInStr string
	if expiresIn.Hours() >= 24 {
		// Use Ceil to round up to the next day
		days := int(math.Ceil(expiresIn.Hours() / 24))
		expiresInStr = fmt.Sprintf("%d days", days)
	} else {
		expiresInStr = fmt.Sprintf("%d hours", int(expiresIn.Hours()))
	}

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"INVITER_NAME":     payload.InviterName,
		"WORKSPACE_NAME":   payload.WorkspaceName,
		"ROLE":             payload.Role,
		"EXPIRES_IN":       expiresInStr,
		"VERIFICATION_URL": fmt.Sprintf("%s/onboarding/join?token=%s", c.websiteURL, payload.Token),
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateInvitation,
		To: []brevo.EmailRecipient{
			{
				Email: payload.Email,
			},
		},
		Subject: fmt.Sprintf("You're Invited to Join %s on Complexus", payload.WorkspaceName),
		Params:  brevoParams,
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send invitation email via Brevo", "error", err, "email", payload.Email)
		return fmt.Errorf("failed to send invitation email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent invitation email via Brevo", "email", payload.Email)
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

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"INVITER_NAME":   payload.InviterName,
		"INVITEE_NAME":   payload.InviteeName,
		"WORKSPACE_NAME": payload.WorkspaceName,
		"WORKSPACE_URL":  fmt.Sprintf("https://%s.complexus.app", payload.WorkspaceSlug),
		"ROLE":           payload.Role,
		"LOGIN_URL":      fmt.Sprintf("%s/login", c.websiteURL),
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateInvitationAccepted,
		To: []brevo.EmailRecipient{
			{
				Email: payload.InviterEmail,
			},
		},
		Subject: fmt.Sprintf("Great news! %s has joined %s on Complexus", payload.InviteeName, payload.WorkspaceName),
		Params:  brevoParams,
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send invitation accepted email via Brevo", "error", err, "email", payload.InviterEmail)
		return fmt.Errorf("failed to send invitation accepted email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent invitation accepted email via Brevo", "email", payload.InviterEmail)
	return nil
}

func (c *Consumer) handleCommentCreated(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleCommentCreated", "event_type", event.Type)

	var payload events.CommentCreatedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessCommentCreated(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process comment created rules", "error", err)
		return err
	}

	// Create notifications
	for _, notification := range notifications {
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}

func (c *Consumer) handleCommentReplied(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleCommentReplied", "event_type", event.Type)

	var payload events.CommentRepliedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessCommentReplied(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process comment replied rules", "error", err)
		return err
	}

	// Create notifications
	for _, notification := range notifications {
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}

func (c *Consumer) handleUserMentioned(ctx context.Context, event events.Event) error {
	c.log.Info(ctx, "consumer.handleUserMentioned", "event_type", event.Type)

	var payload events.UserMentionedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply notification rules
	notifications, err := c.notificationRules.ProcessUserMentioned(ctx, payload, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to process user mentioned rules", "error", err)
		return err
	}

	// Create notifications
	for _, notification := range notifications {
		if _, err := c.notifications.Create(ctx, notification); err != nil {
			c.log.Error(ctx, "failed to create notification", "error", err)
			// Continue processing other notifications
		}
	}

	return nil
}
