package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AssociationMutationAction string

const (
	AssociationMutationAdd    AssociationMutationAction = "add"
	AssociationMutationUpdate AssociationMutationAction = "update"
	AssociationMutationRemove AssociationMutationAction = "remove"
)

type AssociationSnapshot struct {
	ID             uuid.UUID
	FromStoryID    uuid.UUID
	ToStoryID      uuid.UUID
	Type           string
	PreviousType   *string
	FromStoryTitle string
	ToStoryTitle   string
}

type AssociationMutationCommand struct {
	Scope       MutationScope
	Action      AssociationMutationAction
	Association AssociationSnapshot
	Expected    *AssociationSnapshot
	OccurredAt  time.Time
	Events      []MutationEvent
	Activities  []MutationActivity
}

func (command AssociationMutationCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	association := command.Association
	if association.ID == uuid.Nil || association.FromStoryID == uuid.Nil || association.ToStoryID == uuid.Nil ||
		association.FromStoryID == association.ToStoryID || !validAssociationType(association.Type) || command.OccurredAt.IsZero() {
		return fmt.Errorf("%w: association identity, type, and timestamp are required", ErrInvalidMutation)
	}
	switch command.Action {
	case AssociationMutationAdd:
		if command.Expected != nil {
			return fmt.Errorf("%w: add cannot carry an expected association", ErrInvalidMutation)
		}
	case AssociationMutationUpdate, AssociationMutationRemove:
		if command.Expected == nil || command.Expected.ID != association.ID ||
			command.Expected.FromStoryID == uuid.Nil || command.Expected.ToStoryID == uuid.Nil ||
			command.Expected.FromStoryID == command.Expected.ToStoryID || !validAssociationType(command.Expected.Type) {
			return fmt.Errorf("%w: update and remove require an expected association", ErrInvalidMutation)
		}
		if command.Action == AssociationMutationRemove && !sameAssociationIdentity(*command.Expected, association) {
			return fmt.Errorf("%w: remove association must match the expected state", ErrInvalidMutation)
		}
	default:
		return fmt.Errorf("%w: unsupported association action", ErrInvalidMutation)
	}
	storyIDs := associationAffectedStoryIDs(command)
	if len(command.Events) != len(storyIDs) {
		return fmt.Errorf("%w: one durable event per affected story is required", ErrInvalidMutation)
	}
	eventsByStory := make(map[uuid.UUID]MutationEvent, len(command.Events))
	eventIDs := make(map[uuid.UUID]struct{}, len(command.Events))
	for _, event := range command.Events {
		if _, duplicate := eventsByStory[event.StoryID]; duplicate {
			return fmt.Errorf("%w: duplicate association event story", ErrInvalidMutation)
		}
		if _, duplicate := eventIDs[event.ID]; duplicate || !event.OccurredAt.Equal(command.OccurredAt) {
			return fmt.Errorf("%w: association events require unique ids and the mutation timestamp", ErrInvalidMutation)
		}
		eventsByStory[event.StoryID] = event
		eventIDs[event.ID] = struct{}{}
	}
	for _, storyID := range storyIDs {
		event, exists := eventsByStory[storyID]
		if !exists || event.WorkspaceID != command.Scope.WorkspaceID ||
			event.Type != MutationEventStoryUpdated || !sameMutationActor(event, command.Scope) {
			return fmt.Errorf("%w: association event scope mismatch", ErrInvalidMutation)
		}
		if err := event.Validate(); err != nil {
			return err
		}
	}
	if command.Scope.ActivityUser == nil {
		if len(command.Activities) != 0 {
			return fmt.Errorf("%w: machine association mutations cannot invent user activity", ErrInvalidMutation)
		}
		return nil
	}
	if len(command.Activities) != len(storyIDs) {
		return fmt.Errorf("%w: one activity per affected story is required", ErrInvalidMutation)
	}
	activitiesByStory := make(map[uuid.UUID]MutationActivity, len(command.Activities))
	activityIDs := make(map[uuid.UUID]struct{}, len(command.Activities))
	for _, activity := range command.Activities {
		if _, duplicate := activitiesByStory[activity.StoryID]; duplicate {
			return fmt.Errorf("%w: duplicate association activity story", ErrInvalidMutation)
		}
		if _, duplicate := activityIDs[activity.ID]; duplicate || !activity.CreatedAt.Equal(command.OccurredAt) {
			return fmt.Errorf("%w: association activities require unique ids and the mutation timestamp", ErrInvalidMutation)
		}
		activitiesByStory[activity.StoryID] = activity
		activityIDs[activity.ID] = struct{}{}
	}
	for _, storyID := range storyIDs {
		activity, exists := activitiesByStory[storyID]
		if !exists {
			return fmt.Errorf("%w: association activity story mismatch", ErrInvalidMutation)
		}
		if err := activity.Validate(command.Scope, storyID); err != nil {
			return err
		}
	}
	return nil
}

func associationAffectedStoryIDs(command AssociationMutationCommand) []uuid.UUID {
	values := []uuid.UUID{command.Association.FromStoryID, command.Association.ToStoryID}
	if command.Expected != nil {
		values = append(values, command.Expected.FromStoryID, command.Expected.ToStoryID)
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sameAssociationIdentity(left, right AssociationSnapshot) bool {
	return left.ID == right.ID && left.FromStoryID == right.FromStoryID &&
		left.ToStoryID == right.ToStoryID && left.Type == right.Type
}

func validAssociationType(value string) bool {
	return value == "blocking" || value == "related" || value == "duplicate"
}
