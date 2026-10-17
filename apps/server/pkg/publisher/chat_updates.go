package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const ChatUpdatedType = "chat.updated"

// ChatUpdate is a recipient-scoped invalidation hint. Message content is fetched
// through the authorized chat API, including after a lost or repeated hint.
type ChatUpdate struct {
	Type           string     `json:"type"`
	EventID        uuid.UUID  `json:"eventId"`
	WorkspaceID    uuid.UUID  `json:"workspaceId"`
	UserID         uuid.UUID  `json:"userId"`
	ConversationID uuid.UUID  `json:"conversationId"`
	MessageID      *uuid.UUID `json:"messageId,omitempty"`
	Change         string     `json:"change"`
	Version        int64      `json:"version,omitempty"`
}

func (p *Publisher) PublishChatUpdated(ctx context.Context, update ChatUpdate) error {
	if update.EventID == uuid.Nil || update.WorkspaceID == uuid.Nil || update.UserID == uuid.Nil || update.ConversationID == uuid.Nil || update.Change == "" {
		return errors.New("chat update requires event, workspace, recipient, conversation and change")
	}
	update.Type = ChatUpdatedType
	payload, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("marshal chat update: %w", err)
	}
	if err := p.redis.Publish(ctx, UserUpdatesChannel(update.UserID), payload).Err(); err != nil {
		return fmt.Errorf("publish chat update: %w", err)
	}
	return nil
}
