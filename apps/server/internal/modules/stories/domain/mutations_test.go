package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestStoryPatchPreservesTriStateAndStableFieldOrder(t *testing.T) {
	t.Parallel()

	patch := StoryPatch{
		Title:       SetField("Typed mutations"),
		Description: ClearField[string](),
		AssigneeID:  SetField(uuid.New()),
	}
	if got, want := patch.Fields(), []string{"title", "description", "assignee_id"}; !equalStrings(got, want) {
		t.Fatalf("fields = %#v, want %#v", got, want)
	}
	if value, specified := patch.Description.Value(); !specified || value != nil {
		t.Fatalf("description state = value %#v specified %v, want explicit null", value, specified)
	}
	if value, specified := patch.Priority.Value(); specified || value != nil {
		t.Fatalf("priority state = value %#v specified %v, want omitted", value, specified)
	}
	if err := patch.Validate(); err != nil {
		t.Fatalf("validate patch: %v", err)
	}
}

func TestStoryPatchRejectsInvalidRequiredValues(t *testing.T) {
	t.Parallel()

	for name, patch := range map[string]StoryPatch{
		"empty":         {},
		"blank title":   {Title: SetField("  ")},
		"null title":    {Title: ClearField[string]()},
		"null boolean":  {AutoSchedulingEnabled: ClearField[bool]()},
		"bad priority":  {Priority: SetField("Critical")},
		"null priority": {Priority: ClearField[string]()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := patch.Validate(); !errors.Is(err, ErrInvalidMutation) {
				t.Fatalf("Validate() error = %v, want invalid mutation", err)
			}
		})
	}
}

func TestUpdateCommandRequiresMatchingUpdatedEventAndActivity(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, time.August, 28, 9, 30, 0, 0, time.UTC)
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor: %v", err)
	}
	scope := MutationScope{Actor: actor, WorkspaceID: workspaceID, ActivityUser: &userID}
	activity := MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: userID, Type: "update", Field: "title",
		CurrentValue: "After", OldValue: json.RawMessage(`"Before"`), NewValue: json.RawMessage(`"After"`),
		WorkspaceID: workspaceID, CreatedAt: now,
	}
	command := UpdateStoryCommand{
		Scope: scope, StoryID: storyID, ExpectedUpdatedAt: now.Add(-time.Minute),
		Patch: StoryPatch{Title: SetField("After")},
		Event: MutationEvent{
			ID: uuid.New(), WorkspaceID: workspaceID, StoryID: storyID,
			Type: MutationEventStoryUpdated, Actor: actor, Payload: json.RawMessage(`{"title":"After"}`), OccurredAt: now,
		},
		Activities: []MutationActivity{activity},
	}
	if err := command.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	command.Event.Type = MutationEventStoryCreated
	if err := command.Validate(); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("mismatched event error = %v, want invalid mutation", err)
	}
}

func TestMutationScopeRejectsActivityAttributionSpoofing(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, otherUserID := uuid.New(), uuid.New(), uuid.New()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor workspace: %v", err)
	}
	if err := (MutationScope{
		Actor: actor, WorkspaceID: workspaceID, ActivityUser: &otherUserID,
	}).Validate(); !errors.Is(err, ErrMutationForbidden) {
		t.Fatalf("spoofed activity attribution error = %v, want forbidden", err)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
