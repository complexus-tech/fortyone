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
	ID               uuid.UUID
	WorkspaceID      uuid.UUID
	Title            string
	ContentHTML      string
	ContentText      string
	Visibility       Visibility
	CreatedBy        uuid.UUID
	UpdatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ArchivedAt       *time.Time
	CanEdit          bool
	SharedWith       []Member
	RelatedWork      []RelatedWork
	RelatedWorkCount int
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
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Title       *string
	ContentHTML *string
	ContentText *string
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
