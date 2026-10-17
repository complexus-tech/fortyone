package chatdomain

import (
	"time"

	"github.com/google/uuid"
)

type TypingCommand struct {
	Scope          ActorScope
	ConversationID uuid.UUID
	ThreadRootID   *uuid.UUID
	IsTyping       bool
}

type TypingActivity struct {
	WorkspaceID    uuid.UUID
	ConversationID uuid.UUID
	ThreadRootID   *uuid.UUID
	Actor          Person
	RecipientIDs   []uuid.UUID
	IsTyping       bool
	ExpiresAt      time.Time
}
