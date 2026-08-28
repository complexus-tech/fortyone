package commentsdomain

import (
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// ActorScope is the complete security boundary for an existing comment
// mutation. Tenant identity and caller identity must never be inferred from a
// globally unique comment ID.
type ActorScope struct {
	CommentID   uuid.UUID
	WorkspaceID uuid.UUID
	Actor       platformauth.Actor
}

type CreateCommand struct {
	WorkspaceID      uuid.UUID
	StoryID          uuid.UUID
	ParentID         *uuid.UUID
	Actor            platformauth.Actor
	Content          string
	MentionedUserIDs []uuid.UUID
}

type UpdateCommand struct {
	Scope            ActorScope
	Content          string
	MentionedUserIDs []uuid.UUID
}

type GetQuery struct {
	CommentID   uuid.UUID
	WorkspaceID uuid.UUID
}
