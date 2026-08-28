package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestSecondaryLifecycleCommandNormalizesTargetsAndRequiresMatchingEvents(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	storyA, storyB := uuid.New(), uuid.New()
	changedAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	scope := secondaryMutationTestScope(t, workspaceID, actorID)
	command := SecondaryLifecycleCommand{
		Scope:     scope,
		Action:    SecondaryMutationArchive,
		StoryIDs:  []uuid.UUID{storyA, storyB, storyA},
		ChangedAt: changedAt,
		Events: []MutationEvent{
			secondaryMutationTestEvent(t, scope, storyA, MutationEventStoryUpdated, changedAt),
			secondaryMutationTestEvent(t, scope, storyB, MutationEventStoryUpdated, changedAt),
		},
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	command.Events[1].Type = MutationEventStoryDeleted
	if err := command.Validate(); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("mismatched event type error = %v, want invalid mutation", err)
	}
}

func TestSecondaryLifecycleCommandRejectsAmbiguousOrOversizedInput(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, storyID := uuid.New(), uuid.New(), uuid.New()
	changedAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	scope := secondaryMutationTestScope(t, workspaceID, actorID)
	validEvent := secondaryMutationTestEvent(t, scope, storyID, MutationEventStoryDeleted, changedAt)

	tests := map[string]SecondaryLifecycleCommand{
		"missing targets": {
			Scope: scope, Action: SecondaryMutationSoftDelete, ChangedAt: changedAt,
		},
		"zero target": {
			Scope: scope, Action: SecondaryMutationSoftDelete, StoryIDs: []uuid.UUID{uuid.Nil},
			ChangedAt: changedAt, Events: []MutationEvent{validEvent},
		},
		"missing event": {
			Scope: scope, Action: SecondaryMutationSoftDelete, StoryIDs: []uuid.UUID{storyID}, ChangedAt: changedAt,
		},
		"duplicate event": {
			Scope: scope, Action: SecondaryMutationSoftDelete, StoryIDs: []uuid.UUID{storyID}, ChangedAt: changedAt,
			Events: []MutationEvent{validEvent, validEvent},
		},
		"unsupported action": {
			Scope: scope, Action: SecondaryMutationAction("purge_everything"), StoryIDs: []uuid.UUID{storyID},
			ChangedAt: changedAt, Events: []MutationEvent{validEvent},
		},
	}

	oversized := make([]uuid.UUID, MaximumSecondaryMutationTargets+1)
	for index := range oversized {
		oversized[index] = uuid.New()
	}
	tests["oversized batch"] = SecondaryLifecycleCommand{
		Scope: scope, Action: SecondaryMutationSoftDelete, StoryIDs: oversized, ChangedAt: changedAt,
	}

	for name, command := range tests {
		command := command
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := command.Validate(); !errors.Is(err, ErrInvalidMutation) {
				t.Fatalf("Validate() error = %v, want invalid mutation", err)
			}
		})
	}
}

func TestSecondaryReplacementCommandsBindEventActorWorkspaceAndSubject(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, storyID := uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	scope := secondaryMutationTestScope(t, workspaceID, actorID)
	event := secondaryMutationTestEvent(t, scope, storyID, MutationEventStoryUpdated, now)

	labels := ReplaceStoryLabelsCommand{
		Scope: scope, StoryID: storyID, LabelIDs: []uuid.UUID{uuid.New(), uuid.New()}, Event: event,
	}
	if err := labels.Validate(); err != nil {
		t.Fatalf("valid label replacement: %v", err)
	}

	collaborators := ReplaceStoryCollaboratorsCommand{
		Scope: scope, StoryID: storyID, CollaboratorIDs: []uuid.UUID{uuid.New()}, Event: event,
	}
	if err := collaborators.Validate(); err != nil {
		t.Fatalf("valid collaborator replacement: %v", err)
	}

	labels.Event.StoryID = uuid.New()
	if err := labels.Validate(); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("cross-subject label event error = %v, want invalid mutation", err)
	}

	collaborators.Event.Actor = secondaryMutationTestScope(t, workspaceID, uuid.New()).Actor
	if err := collaborators.Validate(); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("cross-actor collaborator event error = %v, want invalid mutation", err)
	}
}

func secondaryMutationTestScope(t *testing.T, workspaceID, actorID uuid.UUID) MutationScope {
	t.Helper()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind secondary mutation actor: %v", err)
	}
	return MutationScope{Actor: actor, WorkspaceID: workspaceID, ActivityUser: &actorID}
}

func secondaryMutationTestEvent(
	t *testing.T,
	scope MutationScope,
	storyID uuid.UUID,
	eventType MutationEventType,
	occurredAt time.Time,
) MutationEvent {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"story_id": storyID, "workspace_id": scope.WorkspaceID,
	})
	if err != nil {
		t.Fatalf("marshal secondary mutation event: %v", err)
	}
	return MutationEvent{
		ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: storyID,
		Type: eventType, Actor: scope.Actor, Payload: payload, OccurredAt: occurredAt,
	}
}
