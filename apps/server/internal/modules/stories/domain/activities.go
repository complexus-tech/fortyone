package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type ActivityWrite struct {
	Activity MutationActivity
	Compact  bool
}

type RecordActivitiesCommand struct {
	Scope      MutationScope
	Activities []ActivityWrite
}

func (command RecordActivitiesCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if len(command.Activities) == 0 || len(command.Activities) > MaximumSecondaryMutationTargets {
		return fmt.Errorf("%w: a bounded activity batch is required", ErrInvalidMutation)
	}
	for _, write := range command.Activities {
		if err := write.Activity.Validate(command.Scope, write.Activity.StoryID); err != nil {
			return err
		}
		if write.Activity.WorkspaceID != command.Scope.WorkspaceID || write.Activity.StoryID == uuid.Nil {
			return fmt.Errorf("%w: activity scope mismatch", ErrInvalidMutation)
		}
	}
	return nil
}
