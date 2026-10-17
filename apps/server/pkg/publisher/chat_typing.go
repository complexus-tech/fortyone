package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const ChatTypingType = "chat.typing"

type ChatTyping struct {
	Type           string     `json:"type"`
	WorkspaceID    uuid.UUID  `json:"workspaceId"`
	UserID         uuid.UUID  `json:"userId"`
	ConversationID uuid.UUID  `json:"conversationId"`
	ActorID        uuid.UUID  `json:"actorId"`
	FullName       string     `json:"fullName"`
	ThreadRootID   *uuid.UUID `json:"threadRootId"`
	IsTyping       bool       `json:"isTyping"`
	ExpiresAt      time.Time  `json:"expiresAt"`
}

// PublishChatTyping uses the live user channel only. Typing never enters the
// durable event stream and reconnecting clients cannot replay stale activity.
func (p *Publisher) PublishChatTyping(ctx context.Context, activity ChatTyping) error {
	if activity.WorkspaceID == uuid.Nil || activity.UserID == uuid.Nil || activity.ConversationID == uuid.Nil || activity.ActorID == uuid.Nil || activity.ExpiresAt.IsZero() {
		return errors.New("chat typing requires workspace, recipient, conversation, actor and expiry")
	}
	activity.Type = ChatTypingType
	payload, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("marshal chat typing: %w", err)
	}
	if err := p.redis.Publish(ctx, UserUpdatesChannel(activity.UserID), payload).Err(); err != nil {
		return fmt.Errorf("publish chat typing: %w", err)
	}
	return nil
}
