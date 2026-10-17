package chatdomain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("chat resource not found")
	ErrForbidden = errors.New("chat operation forbidden")
	ErrInvalid   = errors.New("invalid chat request")
	ErrConflict  = errors.New("chat message changed or creation ID was reused")
)

const MaximumTextRunes = 4000

type ActorScope struct {
	WorkspaceID uuid.UUID
	Actor       platformauth.Actor
}

type Person struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"fullName"`
	AvatarURL *string   `json:"avatarUrl"`
	IsSystem  bool      `json:"isSystem"`
}

type Conversation struct {
	ID               uuid.UUID  `json:"id"`
	Kind             string     `json:"kind"`
	TeamID           *uuid.UUID `json:"teamId"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Color            *string    `json:"color"`
	Members          []Person   `json:"members"`
	MemberCount      int        `json:"memberCount"`
	LastMessageAt    *time.Time `json:"lastMessageAt"`
	LastMessageText  string     `json:"lastMessageText"`
	UnreadCount      int        `json:"unreadCount"`
	UnreadRootCount  int        `json:"unreadRootCount"`
	UnreadReplyCount int        `json:"unreadReplyCount"`
	LastReadSequence int64      `json:"lastReadSequence"`
}

type Attachment struct {
	ID         uuid.UUID `json:"id"`
	FileName   string    `json:"fileName"`
	MediaType  string    `json:"mediaType"`
	SizeBytes  int64     `json:"sizeBytes"`
	Kind       string    `json:"kind,omitempty"`
	DurationMS *int32    `json:"durationMs,omitempty"`
}

type AttachmentMetadata struct {
	Kind       string
	DurationMS *int32
}

type Reaction struct {
	Emoji   string   `json:"emoji"`
	Count   int      `json:"count"`
	Reacted bool     `json:"reacted"`
	Users   []Person `json:"users,omitempty"`
}

type Message struct {
	ID               uuid.UUID       `json:"id"`
	ConversationID   uuid.UUID       `json:"conversationId"`
	ThreadRootID     *uuid.UUID      `json:"threadRootId"`
	CreationID       uuid.UUID       `json:"creationId"`
	Sequence         int64           `json:"sequence"`
	Version          int             `json:"version"`
	Author           Person          `json:"author"`
	Text             string          `json:"text"`
	RichContent      json.RawMessage `json:"richContent"`
	CreatedAt        time.Time       `json:"createdAt"`
	EditedAt         *time.Time      `json:"editedAt"`
	DeletedAt        *time.Time      `json:"deletedAt"`
	MentionedUserIDs []uuid.UUID     `json:"mentionedUserIds"`
	AgentMentions    []string        `json:"agentMentions"`
	AttachmentIDs    []uuid.UUID     `json:"attachmentIds"`
	Attachments      []Attachment    `json:"attachments"`
	Reactions        []Reaction      `json:"reactions"`
	ReplyCount       int             `json:"replyCount"`
	UnreadReplyCount int             `json:"unreadReplyCount"`
	LastReplyAt      *time.Time      `json:"lastReplyAt"`
	ReplyAuthors     []Person        `json:"replyAuthors"`
	Replayed         bool            `json:"-"`
}

type ConversationPage struct {
	Items      []Conversation `json:"items"`
	NextCursor *string        `json:"nextCursor"`
}

type MessagePage struct {
	Items       []Message `json:"items"`
	OlderCursor *string   `json:"olderCursor"`
	NewerCursor *string   `json:"newerCursor"`
	HasOlder    bool      `json:"hasOlder"`
	HasNewer    bool      `json:"hasNewer"`
}

type SearchPage struct {
	Items      []Message `json:"items"`
	NextCursor *string   `json:"nextCursor"`
}

type ResourcePage struct {
	Items      []Resource `json:"items"`
	NextCursor *string    `json:"nextCursor"`
}

type Resource struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversationId"`
	MessageID      uuid.UUID  `json:"messageId"`
	Kind           string     `json:"kind"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Author         Person     `json:"author"`
	CreatedAt      time.Time  `json:"createdAt"`
	AttachmentID   *uuid.UUID `json:"attachmentId,omitempty"`
	FileName       *string    `json:"fileName,omitempty"`
	MediaType      *string    `json:"mediaType,omitempty"`
	SizeBytes      *int64     `json:"sizeBytes,omitempty"`
	AttachmentKind *string    `json:"attachmentKind,omitempty"`
	DurationMS     *int32     `json:"durationMs,omitempty"`
	URL            *string    `json:"url,omitempty"`
}

type ListQuery struct {
	Search string
	Scope  ActorScope
	Cursor string
	Limit  int
}

type MessageQuery struct {
	Scope          ActorScope
	ConversationID uuid.UUID
	ThreadRootID   *uuid.UUID
	Before         string
	After          string
	Around         *uuid.UUID
	Limit          int
}

type SearchQuery struct {
	Scope          ActorScope
	ConversationID uuid.UUID
	Query          string
	Cursor         string
	Limit          int
	Kind           string
	Sort           string
}

type ContextQuery struct {
	Scope           ActorScope
	ConversationID  uuid.UUID
	ThreadRootID    *uuid.UUID
	BeforeMessageID *uuid.UUID
	Limit           int
}

type CreateMessageCommand struct {
	Scope            ActorScope
	ConversationID   uuid.UUID
	CreationID       uuid.UUID
	ThreadRootID     *uuid.UUID
	Text             string
	RichContent      json.RawMessage
	MentionedUserIDs []uuid.UUID
	AgentMentions    []string
	AttachmentIDs    []uuid.UUID
}

type UpdateMessageCommand struct {
	Scope            ActorScope
	MessageID        uuid.UUID
	Version          int
	Text             string
	RichContent      json.RawMessage
	MentionedUserIDs []uuid.UUID
	AgentMentions    []string
	AttachmentIDs    []uuid.UUID
}

type DeleteMessageCommand struct {
	Scope     ActorScope
	MessageID uuid.UUID
	Version   int
}

type DeletedMessage struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversationId"`
	ThreadRootID   *uuid.UUID `json:"threadRootId"`
	Version        int        `json:"version"`
	DeletedAt      time.Time  `json:"deletedAt"`
}

type ReadCommand struct {
	Scope            ActorScope
	ConversationID   uuid.UUID
	ThroughMessageID uuid.UUID
	ThreadRootID     *uuid.UUID
}

type ReadState struct {
	ConversationID   uuid.UUID  `json:"conversationId"`
	ThreadRootID     *uuid.UUID `json:"threadRootId"`
	LastReadSequence int64      `json:"lastReadSequence"`
	UnreadCount      int        `json:"unreadCount"`
}

type AgentReplyCommand struct {
	WorkspaceID      uuid.UUID
	ConversationID   uuid.UUID
	ThreadRootID     uuid.UUID
	CreationID       uuid.UUID
	Requester        platformauth.Actor
	Author           platformauth.Actor
	Text             string
	RichContent      json.RawMessage
	MentionedUserIDs []uuid.UUID
}

type MessageCreatedHook func(context.Context, CreateMessageCommand, Message) error

type Event struct {
	ID             uuid.UUID  `json:"id"`
	LeaseToken     uuid.UUID  `json:"-"`
	Type           string     `json:"type"`
	WorkspaceID    uuid.UUID  `json:"workspaceId"`
	ConversationID uuid.UUID  `json:"conversationId"`
	MessageID      uuid.UUID  `json:"messageId"`
	ActorID        uuid.UUID  `json:"actorId"`
	ThreadRootID   *uuid.UUID `json:"threadRootId"`
	Sequence       int64      `json:"sequence"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type NotificationRecipient struct {
	UserID uuid.UUID
	Reason string
}

// NotificationContent is resolved from the current message once per outbox
// event. Recipient eligibility is still rechecked before creating each receipt.
type NotificationContent struct {
	ConversationKind string          `json:"conversationKind"`
	ConversationName string          `json:"conversationName"`
	AuthorName       string          `json:"authorName"`
	ActorName        string          `json:"actorName"`
	ReactionEmoji    string          `json:"reactionEmoji"`
	Text             string          `json:"text"`
	RichContent      json.RawMessage `json:"richContent"`
	Attachments      []Attachment    `json:"attachments"`
}

type AttachmentScope struct {
	ActorScope
	ConversationID uuid.UUID
}
