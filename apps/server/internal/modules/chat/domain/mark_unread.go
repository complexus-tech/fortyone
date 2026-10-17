package chatdomain

import "github.com/google/uuid"

type MarkUnreadCommand struct {
	Scope          ActorScope
	ConversationID uuid.UUID
	FromMessageID  uuid.UUID
	ThreadRootID   *uuid.UUID
}

// MarkUnreadState includes the selected stream's cursor and the workspace-wide
// counts from the same transaction, so clients can reconcile every unread badge.
type MarkUnreadState struct {
	ReadState
	Summary            UnreadSummary `json:"summary"`
	FirstUnreadReplyID *uuid.UUID    `json:"firstUnreadReplyId,omitempty"`
}
