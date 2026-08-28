package commentsdomain

import (
	"encoding/json"
	"fmt"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const MutationPayloadVersion = 1

type MutationEventType string

const (
	EventCreated MutationEventType = "comment.created"
	EventUpdated MutationEventType = "comment.updated"
	EventDeleted MutationEventType = "comment.deleted"
)

// MutationEvent is the comments module's transaction-owned outbox record.
// Payload and DeliveryBody are pre-encoded once so every delivery attempt signs
// exactly the same bytes. They deliberately contain identifiers only: comment
// text, mentions, user profiles, tokens, and provider metadata are not exposed.
type MutationEvent struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	CommentID    uuid.UUID
	Type         MutationEventType
	Actor        platformauth.Actor
	Payload      []byte
	DeliveryBody []byte
	OccurredAt   time.Time
}

type mutationEventData struct {
	CommentID uuid.UUID  `json:"comment_id"`
	StoryID   uuid.UUID  `json:"story_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
}

type mutationEventEnvelope struct {
	ID             uuid.UUID         `json:"id"`
	Type           MutationEventType `json:"type"`
	PayloadVersion int               `json:"payload_version"`
	OccurredAt     time.Time         `json:"occurred_at"`
	Data           json.RawMessage   `json:"data"`
}

func NewMutationEvent(
	eventID uuid.UUID,
	eventType MutationEventType,
	workspaceID uuid.UUID,
	actor platformauth.Actor,
	comment Comment,
) (MutationEvent, error) {
	if eventID == uuid.Nil || workspaceID == uuid.Nil || comment.ID == uuid.Nil || comment.StoryID == uuid.Nil {
		return MutationEvent{}, fmt.Errorf("%w: event identity is required", ErrInvalidComment)
	}
	if eventType != EventCreated && eventType != EventUpdated && eventType != EventDeleted {
		return MutationEvent{}, fmt.Errorf("%w: unsupported event type", ErrInvalidComment)
	}
	if err := actor.Validate(); err != nil || actor.WorkspaceID != workspaceID {
		return MutationEvent{}, fmt.Errorf("%w: invalid event actor", ErrInvalidComment)
	}
	occurredAt := comment.UpdatedAt.UTC()
	if eventType == EventCreated {
		occurredAt = comment.CreatedAt.UTC()
	}
	if occurredAt.IsZero() {
		return MutationEvent{}, fmt.Errorf("%w: event timestamp is required", ErrInvalidComment)
	}

	payload, err := json.Marshal(mutationEventData{
		CommentID: comment.ID,
		StoryID:   comment.StoryID,
		ParentID:  cloneUUIDPointer(comment.Parent),
	})
	if err != nil {
		return MutationEvent{}, fmt.Errorf("encode comment event payload: %w", err)
	}
	body, err := json.Marshal(mutationEventEnvelope{
		ID:             eventID,
		Type:           eventType,
		PayloadVersion: MutationPayloadVersion,
		OccurredAt:     occurredAt,
		Data:           payload,
	})
	if err != nil {
		return MutationEvent{}, fmt.Errorf("encode comment event envelope: %w", err)
	}

	return MutationEvent{
		ID: eventID, WorkspaceID: workspaceID, CommentID: comment.ID,
		Type: eventType, Actor: actor,
		Payload: append([]byte(nil), payload...), DeliveryBody: append([]byte(nil), body...),
		OccurredAt: occurredAt,
	}, nil
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
