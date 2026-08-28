package eventconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	eventStreamKey      = "events-stream"
	eventConsumerGroup  = "events-processors"
	streamReadCount     = 10
	pendingClaimTimeout = time.Minute * 5
)

// GitHubCommentSyncer syncs FortyOne comments to linked GitHub issues.
type GitHubCommentSyncer interface {
	SyncCommentToGitHub(ctx context.Context, workspaceID, storyID, teamID, localCommentID uuid.UUID, authorName, content string) error
}

type FeedbackStatusBridge interface {
	NotifyLinkedStoryStatusTransition(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) error
}

type StoryScheduleReconcileQueue interface {
	EnqueueStoryScheduleReconcile(context.Context, uuid.UUID, uuid.UUID) error
	EnqueueCalendarWorkspaceScheduleBatch(context.Context, uuid.UUID) error
}

type storyScheduleReconcileDispatch uint8

const (
	storyScheduleReconcileNone storyScheduleReconcileDispatch = iota
	storyScheduleReconcileBatch
	storyScheduleReconcileImmediate
)

type notificationCreator interface {
	Create(context.Context, notifications.CoreNewNotification) (notifications.CoreNotification, error)
}

type notificationContextReader interface {
	ListKeyResultUpdateAudience(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]notificationsdomain.KeyResultAudienceMember, error)
}

type Consumer struct {
	redis                *redis.Client
	log                  *logger.Logger
	notifications        notificationCreator
	notificationContexts notificationContextReader
	notificationRules    *notifications.Rules
	mailerService        mailer.Service
	stories              *stories.Service
	objectives           *objectives.Service
	users                *users.Service
	statuses             *states.Service
	githubSyncer         GitHubCommentSyncer
	feedbackStatuses     FeedbackStatusBridge
	scheduleReconcile    StoryScheduleReconcileQueue
	websiteURL           string
	lifecycle            consumerLifecycle
}

func New(redis *redis.Client, log *logger.Logger, websiteURL string, notificationsService *notifications.Service, mailerService mailer.Service, stories *stories.Service, objectives *objectives.Service, users *users.Service, statuses *states.Service, githubSyncer GitHubCommentSyncer, feedbackStatuses FeedbackStatusBridge, scheduleReconcile StoryScheduleReconcileQueue) *Consumer {
	notificationRules := notifications.NewRules(log, stories, users, statuses)

	return &Consumer{
		redis:                redis,
		log:                  log,
		notifications:        notificationsService,
		notificationContexts: notificationsService,
		notificationRules:    notificationRules,
		mailerService:        mailerService,
		stories:              stories,
		objectives:           objectives,
		users:                users,
		statuses:             statuses,
		githubSyncer:         githubSyncer,
		feedbackStatuses:     feedbackStatuses,
		scheduleReconcile:    scheduleReconcile,
		websiteURL:           websiteURL,
	}
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

// handleEvent routes events to the appropriate handler based on the event type
func (c *Consumer) handleEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.StoryCreated:
		return c.handleStoryCreated(ctx, event)
	case events.StoryUpdated:
		return c.handleStoryUpdated(ctx, event)
	case events.CommentCreated:
		return c.handleCommentCreated(ctx, event)
	case events.CommentReplied:
		return c.handleCommentReplied(ctx, event)
	case events.FeedbackCommentCreated:
		return c.handleFeedbackCommentCreated(ctx, event)
	case events.FeedbackStatusUpdated:
		return c.handleFeedbackStatusUpdated(ctx, event)
	case events.FeedbackContributorVerification:
		return c.handleFeedbackContributorVerification(ctx, event)
	case events.FeedbackUpdatePublished:
		return c.handleFeedbackUpdatePublished(ctx, event)
	case events.FeedbackItemMerged:
		return c.handleFeedbackItemMerged(ctx, event)
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
	case events.WorkspaceDeletionScheduledConfirmation:
		return c.handleWorkspaceDeletionScheduledConfirmation(ctx, event)
	case events.WorkspaceDeletionScheduledNotification:
		return c.handleWorkspaceDeletionScheduledNotification(ctx, event)
	case events.WorkspaceRestoredConfirmation:
		return c.handleWorkspaceRestoredConfirmation(ctx, event)
	case events.WorkspaceRestoredNotification:
		return c.handleWorkspaceRestoredNotification(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func withEventDedupeKey(event events.Event, notification notifications.CoreNewNotification, index int) notifications.CoreNewNotification {
	if strings.TrimSpace(notification.DedupeKey) != "" {
		return notification
	}
	notification.DedupeKey = fmt.Sprintf(
		"event:%s:%s:%s:%d:%s",
		event.Type,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		notification.RecipientID,
		index,
		notification.Type,
	)
	return notification
}
