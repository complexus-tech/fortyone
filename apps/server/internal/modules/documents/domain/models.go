package documentdomain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Visibility defines who may discover and read a document inside its
// workspace. Mutation rights are evaluated separately.
type Visibility string

const (
	VisibilityWorkspace  Visibility = "workspace"
	VisibilityRestricted Visibility = "restricted"
	VisibilityPrivate    Visibility = "private"
)

type RelationshipType string

const (
	RelationshipStory     RelationshipType = "story"
	RelationshipObjective RelationshipType = "objective"
)

type Member struct {
	UserID uuid.UUID
	Role   string
}

type RelatedWork struct {
	ReferencedInContent bool
	EntityID            uuid.UUID
	EntityType          RelationshipType
	Title               string
	Reference           *string
	StatusColor         *string
	TeamID              *uuid.UUID
}

type Document struct {
	Revision           int64
	CollaborationState []byte `json:"-"`
	CollaborationEpoch int64
	Collaborative      bool
	PublicToken        *string
	PublicEnabled      bool
	ID                 uuid.UUID
	WorkspaceID        uuid.UUID
	Title              string
	ContentHTML        string
	ContentText        string
	Visibility         Visibility
	CreatedBy          uuid.UUID
	UpdatedBy          uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ArchivedAt         *time.Time
	CanEdit            bool
	SharedWith         []Member
	RelatedWork        []RelatedWork
	RelatedWorkCount   int
}

type Summary struct {
	ID               uuid.UUID
	WorkspaceID      uuid.UUID
	Title            string
	Visibility       Visibility
	CreatedBy        uuid.UUID
	UpdatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ArchivedAt       *time.Time
	PublicEnabled    bool
	CanEdit          bool
	RelatedWork      []RelatedWork
	RelatedWorkCount int
}

type ListInput struct {
	Offset      int
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Search      string
	Scope       string
	Limit       *int
}

type CreateInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Title       string
	Visibility  Visibility
	ContentHTML string
	ContentText string
}

type UpdateInput struct {
	ExpectedRevision int64
	WorkspaceID      uuid.UUID
	UserID           uuid.UUID
	DocumentID       uuid.UUID
	Title            *string
	ContentHTML      *string
	ContentText      *string
}

type AccessInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Visibility  Visibility
	Members     []Member
}

type RelationshipInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	EntityType  RelationshipType
	EntityID    uuid.UUID
}

type MediaInput struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	DocumentID   uuid.UUID
	AttachmentID uuid.UUID
}

type Comment struct {
	ReplyToCommentID *uuid.UUID
	MentionedUserIDs []uuid.UUID
	ID               uuid.UUID
	Body             string
	CreatedBy        uuid.UUID
	AuthorName       string
	AuthorAvatar     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CommentThread struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Quote       string
	AnchorStart int32
	AnchorEnd   int32
	CreatedBy   uuid.UUID
	ResolvedAt  *time.Time
	ResolvedBy  *uuid.UUID
	CreatedAt   time.Time
	Comments    []Comment
}

type CreateCommentInput struct {
	CreationID  *uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Body        string
	Quote       string
	AnchorStart int32
	AnchorEnd   int32
}

type ReplyCommentInput struct {
	CreationID       *uuid.UUID
	ReplyToCommentID *uuid.UUID
	WorkspaceID      uuid.UUID
	UserID           uuid.UUID
	DocumentID       uuid.UUID
	ThreadID         uuid.UUID
	Body             string
}

type ResolveCommentInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	ThreadID    uuid.UUID
	Resolved    bool
}

// Revision is an immutable content snapshot; access always follows the current document.
type Revision struct {
	Revision    int64      `json:"revision"`
	Title       string     `json:"title"`
	ContentHTML string     `json:"contentHtml,omitempty"`
	ContentText string     `json:"contentText,omitempty"`
	EditedBy    *uuid.UUID `json:"editedBy"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type PublicDocument struct {
	ID          uuid.UUID `json:"-"`
	WorkspaceID uuid.UUID `json:"-"`
	Title       string    `json:"title"`
	ContentHTML string    `json:"contentHtml"`
	ContentText string    `json:"contentText"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CollaborationSession struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// CommentContext is scoped to the current reader and document, never a global ID lookup.
type CommentContext struct {
	Comment  Comment
	Thread   CommentThread
	Document Document
}
type CommentMention struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatarUrl"`
}
type AgentReplyInput struct {
	WorkspaceID, DocumentID, RequesterID, AuthorID, ThreadID, CreationID uuid.UUID
	ReplyToCommentID                                                     *uuid.UUID
	Body                                                                 string
}
type AgentEditInput struct {
	InvocationID, WorkspaceID, DocumentID, RequesterID uuid.UUID
	ExpectedRevision, ExpectedEpoch                    int64
	Operations                                         json.RawMessage
}
type AgentEdit struct {
	ID, InvocationID, WorkspaceID, DocumentID, RequesterID uuid.UUID
	ExpectedRevision, ExpectedEpoch                        int64
	Operations                                             json.RawMessage
	Status                                                 string
	ResultingRevision                                      *int64
	Receipt                                                json.RawMessage
	ErrorCode                                              *string
	CreatedAt, UpdatedAt                                   time.Time
	AppliedAt                                              *time.Time
}
