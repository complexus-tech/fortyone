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

type Event struct {
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	ActorID   uuid.UUID `json:"actor_id"`
}
