package commentsdomain

import (
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// AgentReplyScope separates the displayed system author from the human whose
// live permissions authorize a reply to their source comment.
type AgentReplyScope struct {
	WorkspaceID uuid.UUID
	StoryID     uuid.UUID
	ParentID    uuid.UUID
	Requester   platformauth.Actor
	Author      platformauth.Actor
}

type AgentReplyCommand struct {
	Scope            AgentReplyScope
	CreationID       uuid.UUID
	Content          string
	MentionedUserIDs []uuid.UUID
}

type MentionTarget struct {
	ID   uuid.UUID
	Name string
}

// AgentReply carries persisted recipients, filtered against their current story
// access, so retries never infer notifications from generated text.
type AgentReply struct {
	Comment
	MentionedUserIDs []uuid.UUID
}
