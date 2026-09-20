package documentdomain

import (
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
	EntityID   uuid.UUID
	EntityType RelationshipType
	Title      string
	Reference  *string
	TeamID     *uuid.UUID
}

type Document struct {
	Revision           int64
	CollaborationEpoch int64
	Collaborative      bool
	PublicToken        *string
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
	CanEdit          bool
	RelatedWorkCount int
}

type ListInput struct {
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
	ID           uuid.UUID
	Body         string
	CreatedBy    uuid.UUID
	AuthorName   string
	AuthorAvatar *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Body        string
	Quote       string
	AnchorStart int32
	AnchorEnd   int32
}

type ReplyCommentInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	ThreadID    uuid.UUID
	Body        string
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
