package commentsdomain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type StoryScope struct {
	StoryID     uuid.UUID
	WorkspaceID uuid.UUID
	Actor       platformauth.Actor
}

type Reaction struct {
	Emoji   string `json:"emoji"`
	Count   int64  `json:"count"`
	Reacted bool   `json:"reacted"`
}

type CommentState struct {
	CommentID  uuid.UUID  `json:"commentId"`
	Reactions  []Reaction `json:"reactions"`
	ResolvedAt *time.Time `json:"resolvedAt"`
	ResolvedBy *uuid.UUID `json:"resolvedBy"`
}

type StoryState struct {
	LastReadAt      *time.Time              `json:"lastReadAt"`
	ObservedAt      time.Time               `json:"observedAt"`
	UnreadCount     int64                   `json:"unreadCount"`
	LatestCommentID *uuid.UUID              `json:"latestCommentId"`
	Comments        []CommentState          `json:"comments"`
	ThreadReads     map[uuid.UUID]time.Time `json:"threadReads"`
}

type ReactionCommand struct {
	Scope  ActorScope
	Emoji  string
	Active bool
}

type ResolutionCommand struct {
	Scope    ActorScope
	Resolved bool
}

type ReadCommand struct {
	Scope      StoryScope
	ObservedAt time.Time
	ThreadID   *uuid.UUID
}
