package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DuplicateStoryCommand is a finite, replay-safe duplication intent. The
// caller chooses the target and event IDs before the transaction, so a
// serialization retry cannot create a second story or outbound event.
type DuplicateStoryCommand struct {
	Scope                   MutationScope
	SourceStoryID           uuid.UUID
	TargetStoryID           uuid.UUID
	ExpectedSourceUpdatedAt time.Time
	ReporterID              uuid.UUID
	OccurredAt              time.Time
	Event                   MutationEvent
	Activity                MutationActivity
}

func (command DuplicateStoryCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.SourceStoryID == uuid.Nil || command.TargetStoryID == uuid.Nil ||
		command.SourceStoryID == command.TargetStoryID || command.ReporterID == uuid.Nil ||
		command.ExpectedSourceUpdatedAt.IsZero() || command.OccurredAt.IsZero() {
		return fmt.Errorf("%w: duplicate story identity and timestamp are required", ErrInvalidMutation)
	}
	if command.Scope.ActivityUser == nil || *command.Scope.ActivityUser != command.ReporterID {
		return fmt.Errorf("%w: duplication requires an acting user reporter", ErrMutationForbidden)
	}
	if command.Event.StoryID != command.TargetStoryID ||
		command.Event.WorkspaceID != command.Scope.WorkspaceID ||
		command.Event.Type != MutationEventStoryCreated ||
		!sameMutationActor(command.Event, command.Scope) ||
		!command.Event.OccurredAt.Equal(command.OccurredAt) ||
		!command.Activity.CreatedAt.Equal(command.OccurredAt) {
		return fmt.Errorf("%w: duplicate event scope mismatch", ErrInvalidMutation)
	}
	if err := command.Event.Validate(); err != nil {
		return err
	}
	return command.Activity.Validate(command.Scope, command.TargetStoryID)
}

type DuplicateStoryResult struct {
	Story Story
}
