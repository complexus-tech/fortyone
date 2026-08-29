package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const MaximumSecondaryMutationTargets = 500

type SecondaryMutationAction string

const (
	SecondaryMutationSoftDelete SecondaryMutationAction = "soft_delete"
	SecondaryMutationHardDelete SecondaryMutationAction = "hard_delete"
	SecondaryMutationRestore    SecondaryMutationAction = "restore"
	SecondaryMutationArchive    SecondaryMutationAction = "archive"
	SecondaryMutationUnarchive  SecondaryMutationAction = "unarchive"
)

func (action SecondaryMutationAction) EventType() MutationEventType {
	if action == SecondaryMutationSoftDelete || action == SecondaryMutationHardDelete {
		return MutationEventStoryDeleted
	}
	return MutationEventStoryUpdated
}

func (action SecondaryMutationAction) Validate() error {
	switch action {
	case SecondaryMutationSoftDelete, SecondaryMutationHardDelete, SecondaryMutationRestore,
		SecondaryMutationArchive, SecondaryMutationUnarchive:
		return nil
	default:
		return fmt.Errorf("%w: unsupported secondary mutation action", ErrInvalidMutation)
	}
}

type SecondaryLifecycleCommand struct {
	Scope     MutationScope
	Action    SecondaryMutationAction
	StoryIDs  []uuid.UUID
	ChangedAt time.Time
	Events    []MutationEvent
}

func (command SecondaryLifecycleCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if err := command.Action.Validate(); err != nil {
		return err
	}
	ids, err := normalizeSecondaryIDs(command.StoryIDs)
	if err != nil {
		return err
	}
	if command.ChangedAt.IsZero() || len(command.Events) != len(ids) {
		return fmt.Errorf("%w: lifecycle timestamp and one event per story are required", ErrInvalidMutation)
	}
	eventsByStory := make(map[uuid.UUID]struct{}, len(command.Events))
	for _, event := range command.Events {
		if _, duplicate := eventsByStory[event.StoryID]; duplicate {
			return fmt.Errorf("%w: duplicate story mutation event", ErrInvalidMutation)
		}
		if event.WorkspaceID != command.Scope.WorkspaceID || event.Type != command.Action.EventType() ||
			!sameMutationActor(event, command.Scope) {
			return fmt.Errorf("%w: lifecycle event does not match its command", ErrInvalidMutation)
		}
		if err := event.Validate(); err != nil {
			return err
		}
		eventsByStory[event.StoryID] = struct{}{}
	}
	for _, storyID := range ids {
		if _, exists := eventsByStory[storyID]; !exists {
			return fmt.Errorf("%w: lifecycle event is missing", ErrInvalidMutation)
		}
	}
	return nil
}

type SecondaryLifecycleResult struct {
	StoryIDs                         []uuid.UUID
	ChangedStoryIDs                  []uuid.UUID
	OrphanedAttachmentIDs            []uuid.UUID
	AttachmentObjectDeletionDeferred bool
}

type ReplaceStoryLabelsCommand struct {
	Scope    MutationScope
	StoryID  uuid.UUID
	LabelIDs []uuid.UUID
	Event    MutationEvent
	Activity *MutationActivity
}

func (command ReplaceStoryLabelsCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.StoryID == uuid.Nil || command.Event.StoryID != command.StoryID ||
		command.Event.WorkspaceID != command.Scope.WorkspaceID || command.Event.Type != MutationEventStoryUpdated ||
		!sameMutationActor(command.Event, command.Scope) {
		return fmt.Errorf("%w: label command and event scope must match", ErrInvalidMutation)
	}
	if _, err := normalizeSecondaryIDsAllowEmpty(command.LabelIDs); err != nil {
		return err
	}
	if command.Activity != nil {
		if err := command.Activity.Validate(command.Scope, command.StoryID); err != nil {
			return err
		}
	}
	return command.Event.Validate()
}

type ReplaceStoryCollaboratorsCommand struct {
	Scope           MutationScope
	StoryID         uuid.UUID
	CollaboratorIDs []uuid.UUID
	Event           MutationEvent
	Activity        *MutationActivity
}

func (command ReplaceStoryCollaboratorsCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.StoryID == uuid.Nil || command.Event.StoryID != command.StoryID ||
		command.Event.WorkspaceID != command.Scope.WorkspaceID || command.Event.Type != MutationEventStoryUpdated ||
		!sameMutationActor(command.Event, command.Scope) {
		return fmt.Errorf("%w: collaborator command and event scope must match", ErrInvalidMutation)
	}
	if _, err := normalizeSecondaryIDsAllowEmpty(command.CollaboratorIDs); err != nil {
		return err
	}
	if command.Activity != nil {
		if err := command.Activity.Validate(command.Scope, command.StoryID); err != nil {
			return err
		}
	}
	return command.Event.Validate()
}

type ReplacementResult struct {
	Changed     bool
	PreviousIDs []uuid.UUID
	CurrentIDs  []uuid.UUID
	AssigneeID  *uuid.UUID
}

func NormalizeSecondaryMutationIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	return normalizeSecondaryIDs(ids)
}

func NormalizeSecondaryReplacementIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	return normalizeSecondaryIDsAllowEmpty(ids)
}

func normalizeSecondaryIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: at least one story is required", ErrInvalidMutation)
	}
	return normalizeSecondaryIDsAllowEmpty(ids)
}

func normalizeSecondaryIDsAllowEmpty(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) > MaximumSecondaryMutationTargets {
		return nil, fmt.Errorf("%w: too many mutation targets", ErrInvalidMutation)
	}
	result := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			return nil, fmt.Errorf("%w: mutation ids cannot be empty", ErrInvalidMutation)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func sameMutationActor(event MutationEvent, scope MutationScope) bool {
	return event.Actor.PrincipalID == scope.Actor.PrincipalID &&
		event.Actor.Kind == scope.Actor.Kind &&
		event.Actor.WorkspaceID == scope.Actor.WorkspaceID &&
		event.Actor.CredentialID == scope.Actor.CredentialID
}
