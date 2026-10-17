package chatdomain

import "github.com/google/uuid"

// UnreadSummary covers every accessible conversation, independently of the
// paginated conversation list. Zero-count conversations are omitted.
type UnreadSummary struct {
	TotalUnread   int                  `json:"totalUnread"`
	UnreadThreads int                  `json:"unreadThreads"`
	Conversations []ConversationUnread `json:"conversations"`
}

type ConversationUnread struct {
	ConversationID   uuid.UUID `json:"conversationId"`
	UnreadCount      int       `json:"unreadCount"`
	UnreadRootCount  int       `json:"unreadRootCount"`
	UnreadReplyCount int       `json:"unreadReplyCount"`
}

type UnreadThread struct {
	Message
	ConversationName   string    `json:"conversationName"`
	ConversationKind   string    `json:"conversationKind"`
	FirstUnreadReplyID uuid.UUID `json:"firstUnreadReplyId"`
}

type UnreadThreadPage struct {
	Items      []UnreadThread `json:"items"`
	NextCursor *string        `json:"nextCursor"`
}

type UnreadThreadQuery struct {
	Scope          ActorScope
	ConversationID *uuid.UUID
	Cursor         string
	Limit          int
}
