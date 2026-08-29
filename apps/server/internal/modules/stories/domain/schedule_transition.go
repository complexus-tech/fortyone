package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ScheduleTransitionOutboxInput is the immutable scheduler-event snapshot
// committed in the same transaction as its story state transition.
type ScheduleTransitionOutboxInput struct {
	EventID             uuid.UUID
	StoryID             uuid.UUID
	WorkspaceID         uuid.UUID
	ActorID             uuid.UUID
	SemanticFingerprint string
	EventPayload        json.RawMessage
	ClaimImmediately    bool
}

// ScheduleTransitionOutboxEvent is a claimed immutable scheduler event. The
// original envelope is retained so retries never rebuild tenant data from
// mutable story state.
type ScheduleTransitionOutboxEvent struct {
	EventID             uuid.UUID
	StoryID             uuid.UUID
	WorkspaceID         uuid.UUID
	ActorID             uuid.UUID
	SemanticFingerprint string
	TransitionSequence  int64
	ClaimToken          uuid.UUID
	AttemptCount        int
	EventPayload        json.RawMessage
}
