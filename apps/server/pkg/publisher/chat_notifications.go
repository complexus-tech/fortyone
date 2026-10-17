package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ChatNotificationType = "chat.notification"

// ChatNotification is an interruption for one recipient, separate from the
// work inbox. Its durable event and receipt identities let clients deduplicate
// retries without treating reconnecting or loading history as a new message.
type ChatNotification struct {
	Type           string     `json:"type"`
	NotificationID uuid.UUID  `json:"notificationId"`
	EventID        uuid.UUID  `json:"eventId"`
	WorkspaceID    uuid.UUID  `json:"workspaceId"`
	UserID         uuid.UUID  `json:"userId"`
	ActorID        uuid.UUID  `json:"actorId"`
	ConversationID uuid.UUID  `json:"conversationId"`
	MessageID      uuid.UUID  `json:"messageId"`
	ThreadRootID   *uuid.UUID `json:"threadRootId"`
	CreatedAt      time.Time  `json:"createdAt"`
	OccurredAt     time.Time  `json:"occurredAt"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
}

func (p *Publisher) PublishChatNotification(ctx context.Context, notification ChatNotification) error {
	for _, id := range []uuid.UUID{notification.NotificationID, notification.EventID, notification.WorkspaceID, notification.UserID, notification.ActorID, notification.ConversationID, notification.MessageID} {
		if id == uuid.Nil {
			return errors.New("chat notification requires receipt, event, workspace, recipient, actor, conversation and message")
		}
	}
	if notification.UserID == notification.ActorID ||
		(notification.ThreadRootID != nil && *notification.ThreadRootID == uuid.Nil) ||
		notification.CreatedAt.IsZero() || notification.OccurredAt.IsZero() ||
		strings.TrimSpace(notification.Title) == "" || len(notification.Title) > 256 ||
		strings.TrimSpace(notification.Body) == "" || len(notification.Body) > 1024 {
		return errors.New("invalid chat notification recipient, destination, timestamp or content")
	}
	notification.Type = ChatNotificationType
	notification.CreatedAt = notification.CreatedAt.UTC()
	notification.OccurredAt = notification.OccurredAt.UTC()
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshal chat notification: %w", err)
	}
	if err := p.redis.Publish(ctx, UserUpdatesChannel(notification.UserID), payload).Err(); err != nil {
		return fmt.Errorf("publish chat notification: %w", err)
	}
	return nil
}
