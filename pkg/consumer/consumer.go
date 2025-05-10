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

	// In a production implementation, you would use these variables
	// to create notifications, send emails, etc.
	_, err = c.stories.Get(ctx, payload.StoryID, payload.WorkspaceID)
	if err != nil {
		c.log.Error(ctx, "failed to get story", "error", err)
	}

	// Get actor's user info
	_, err = c.users.GetUser(ctx, event.ActorID)
	if err != nil {
		c.log.Error(ctx, "failed to get actor user", "error", err)
	}

	// Implement the rest of your story updated handler
	// This would typically include creating notifications, sending emails, etc.

	return nil
}

func (c *Consumer) handleStoryCommented(ctx context.Context, event events.Event) error {
	// Implement story comment handling
	return nil
}

func (c *Consumer) handleObjectiveUpdated(ctx context.Context, event events.Event) error {
	// Implement objective update handling
	return nil
}

func (c *Consumer) handleKeyResultUpdated(ctx context.Context, event events.Event) error {
	// Implement key result update handling
	return nil
}

func (c *Consumer) handleEmailVerification(ctx context.Context, event events.Event) error {
	// Implement email verification handling
	return nil
}

func (c *Consumer) handleInvitationEmail(ctx context.Context, event events.Event) error {
	// Implement invitation email handling
	return nil
}

func (c *Consumer) handleInvitationAccepted(ctx context.Context, event events.Event) error {
	// Implement invitation accepted handling
	return nil
}
