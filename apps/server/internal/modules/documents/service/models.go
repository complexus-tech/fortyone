package documents

import (
	"time"

	"github.com/google/uuid"
)

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

type CoreDocumentMember struct {
	UserID uuid.UUID
	Role   string
}

type CoreRelatedWork struct {
	EntityID   uuid.UUID
	EntityType RelationshipType
	Title      string
	Reference  *string
	TeamID     *uuid.UUID
}

type CoreDocument struct {
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
	SharedWith       []CoreDocumentMember
	RelatedWork      []CoreRelatedWork
	RelatedWorkCount int
}

type CoreDocumentSummary struct {
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

type CoreListInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Search      string
	Scope       string
	Limit       *int
}

type CoreCreateInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Title       string
	Visibility  Visibility
	ContentHTML string
	ContentText string
}

type CoreUpdateInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Title       *string
	ContentHTML *string
	ContentText *string
}

type CoreAccessInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	Visibility  Visibility
	Members     []CoreDocumentMember
}

type CoreRelationshipInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	DocumentID  uuid.UUID
	EntityType  RelationshipType
	EntityID    uuid.UUID
}

type CoreMediaInput struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	DocumentID   uuid.UUID
	AttachmentID uuid.UUID
}
