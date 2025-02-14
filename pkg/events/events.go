package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	StoryUpdated     EventType = "story.updated"
	StoryCommented   EventType = "story.commented"
	ObjectiveUpdated EventType = "objective.updated"
	KeyResultUpdated EventType = "keyresult.updated"
)

// Event is the base event structure
type Event struct {
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	ActorID   uuid.UUID `json:"actor_id"`
}

// StoryUpdatedPayload contains data for story update events
type StoryUpdatedPayload struct {
	StoryID     uuid.UUID      `json:"story_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
	AssigneeID  *uuid.UUID     `json:"assignee_id,omitempty"`
}

// StoryCommentedPayload contains data for story comment events
type StoryCommentedPayload struct {
	StoryID     uuid.UUID  `json:"story_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	CommentID   uuid.UUID  `json:"comment_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// ObjectiveUpdatedPayload contains data for objective update events
type ObjectiveUpdatedPayload struct {
	ObjectiveID uuid.UUID      `json:"objective_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
	LeadID      *uuid.UUID     `json:"lead_id,omitempty"`
}

// KeyResultUpdatedPayload contains data for key result update events
type KeyResultUpdatedPayload struct {
	KeyResultID uuid.UUID      `json:"key_result_id"`
	ObjectiveID uuid.UUID      `json:"objective_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
}
