package outboundwebhooksdomain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const PayloadVersion = 1

type EventType string

const (
	EventStoryCreated   EventType = "story.created"
	EventStoryUpdated   EventType = "story.updated"
	EventStoryDeleted   EventType = "story.deleted"
	EventCommentCreated EventType = "comment.created"
	EventCommentUpdated EventType = "comment.updated"
	EventCommentDeleted EventType = "comment.deleted"
)

var eventCatalog = []EventType{
	EventStoryCreated,
	EventStoryUpdated,
	EventStoryDeleted,
	EventCommentCreated,
	EventCommentUpdated,
	EventCommentDeleted,
}

func EventCatalog() []EventType {
	return append([]EventType(nil), eventCatalog...)
}

func (eventType EventType) Validate() error {
	if !slices.Contains(eventCatalog, eventType) {
		return fmt.Errorf("%w: %q", ErrInvalidEventType, eventType)
	}
	return nil
}

func (eventType EventType) SubjectType() SubjectType {
	if eventType == EventStoryCreated || eventType == EventStoryUpdated || eventType == EventStoryDeleted {
		return SubjectStory
	}
	return SubjectComment
}

type SubjectType string

const (
	SubjectStory   SubjectType = "story"
	SubjectComment SubjectType = "comment"
)

type Event struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	Type           EventType
	PayloadVersion int
	SubjectType    SubjectType
	SubjectID      uuid.UUID
	Actor          platformauth.Actor
	Payload        json.RawMessage
	OccurredAt     time.Time
	CreatedAt      time.Time
}

type PublishEvent struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Type        EventType
	SubjectID   uuid.UUID
	Actor       platformauth.Actor
	Payload     json.RawMessage
	OccurredAt  time.Time
}

func (input PublishEvent) Validate() error {
	if input.ID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.SubjectID == uuid.Nil || input.OccurredAt.IsZero() {
		return ErrInvalidSubject
	}
	if err := input.Type.Validate(); err != nil {
		return err
	}
	if err := input.Actor.Validate(); err != nil {
		return fmt.Errorf("%w: actor: %v", ErrInvalidPayload, err)
	}
	if input.Actor.WorkspaceID != input.WorkspaceID {
		return fmt.Errorf("%w: actor workspace mismatch", ErrInvalidPayload)
	}
	if len(input.Payload) < 2 || len(input.Payload) > 256<<10 || !json.Valid(input.Payload) {
		return ErrInvalidPayload
	}
	decoder := json.NewDecoder(bytes.NewReader(input.Payload))
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil || object == nil {
		return ErrInvalidPayload
	}
	return nil
}

type Envelope struct {
	ID             uuid.UUID       `json:"id"`
	Type           EventType       `json:"type"`
	PayloadVersion int             `json:"payload_version"`
	OccurredAt     time.Time       `json:"occurred_at"`
	Data           json.RawMessage `json:"data"`
}

func NewEnvelope(event Event) ([]byte, error) {
	if event.ID == uuid.Nil || event.PayloadVersion != PayloadVersion || event.OccurredAt.IsZero() {
		return nil, ErrInvalidPayload
	}
	body, err := json.Marshal(Envelope{
		ID:             event.ID,
		Type:           event.Type,
		PayloadVersion: event.PayloadVersion,
		OccurredAt:     event.OccurredAt.UTC(),
		Data:           event.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("encode outbound webhook envelope: %w", err)
	}
	if len(body) > 256<<10 {
		return nil, ErrInvalidPayload
	}
	return body, nil
}
