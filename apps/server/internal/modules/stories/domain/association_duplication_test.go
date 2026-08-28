package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestAssociationMutationRequiresIntentForEveryOldAndNewEndpoint(t *testing.T) {
	t.Parallel()

	scope := associationTestHumanScope(t)
	fromID, oldToID, newToID := uuid.New(), uuid.New(), uuid.New()
	associationID := uuid.New()
	occurredAt := time.Now().UTC()
	command := AssociationMutationCommand{
		Scope: scope, Action: AssociationMutationUpdate,
		Association: AssociationSnapshot{
			ID: associationID, FromStoryID: fromID, ToStoryID: newToID, Type: "blocking",
		},
		Expected: &AssociationSnapshot{
			ID: associationID, FromStoryID: fromID, ToStoryID: oldToID, Type: "related",
		},
		OccurredAt: occurredAt,
	}
	for _, storyID := range []uuid.UUID{fromID, newToID} {
		command.Events = append(command.Events, MutationEvent{
			ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: storyID,
			Type: MutationEventStoryUpdated, Actor: scope.Actor,
			Payload: json.RawMessage(`{}`), OccurredAt: occurredAt,
		})
		command.Activities = append(command.Activities, MutationActivity{
			ID: uuid.New(), StoryID: storyID, UserID: *scope.ActivityUser,
			Type: "update", Field: "association", CurrentValue: "changed",
			OldValue: json.RawMessage(`null`), NewValue: json.RawMessage(`null`),
			WorkspaceID: scope.WorkspaceID, CreatedAt: occurredAt,
		})
	}
	if err := command.Validate(); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("Validate() error = %v, want missing old endpoint intent", err)
	}
}

func TestDuplicateStoryCommandRejectsMachineReporterAttribution(t *testing.T) {
	t.Parallel()

	workspaceID, principalID, credentialID := uuid.New(), uuid.New(), uuid.New()
	actor, err := platformauth.NewActor(
		principalID,
		platformauth.PrincipalServiceAccount,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("construct machine actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind machine actor: %v", err)
	}
	occurredAt := time.Now().UTC()
	targetID := uuid.New()
	command := DuplicateStoryCommand{
		Scope:         MutationScope{Actor: actor, WorkspaceID: workspaceID},
		SourceStoryID: uuid.New(), TargetStoryID: targetID,
		ExpectedSourceUpdatedAt: occurredAt.Add(-time.Minute),
		ReporterID:              principalID, OccurredAt: occurredAt,
		Event: MutationEvent{
			ID: uuid.New(), WorkspaceID: workspaceID, StoryID: targetID,
			Type: MutationEventStoryCreated, Actor: actor,
			Payload: json.RawMessage(`{}`), OccurredAt: occurredAt,
		},
	}
	if err := command.Validate(); !errors.Is(err, ErrMutationForbidden) {
		t.Fatalf("Validate() error = %v, want machine attribution forbidden", err)
	}
}

func associationTestHumanScope(t *testing.T) MutationScope {
	t.Helper()
	workspaceID, actorID := uuid.New(), uuid.New()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind human actor: %v", err)
	}
	return MutationScope{Actor: actor, WorkspaceID: workspaceID, ActivityUser: &actorID}
}
