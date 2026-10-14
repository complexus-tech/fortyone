package stories

import (
	"context"
	"errors"
	"strings"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

type checklistRepositoryStub struct {
	command *storydomain.ChecklistCommand
	result  storydomain.ChecklistMutationResult
	err     error
	reads   int
}

func (r *checklistRepositoryStub) ListChecklist(context.Context, storydomain.ReadScope, uuid.UUID) ([]storydomain.ChecklistItem, error) {
	r.reads++
	return []storydomain.ChecklistItem{}, r.err
}
func (r *checklistRepositoryStub) ApplyChecklistMutation(_ context.Context, c storydomain.ChecklistCommand) (storydomain.ChecklistMutationResult, error) {
	r.command = &c
	return r.result, r.err
}

func TestChecklistServiceDerivesActorAndPublishesOnlyCommittedChanges(t *testing.T) {
	workspaceID, storyID := uuid.New(), uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := auth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatal(err)
	}
	title := "  Test item  "
	for _, test := range []struct {
		name       string
		changed    bool
		failure    error
		created    bool
		wantEvents int
	}{
		{name: "changed", changed: true, wantEvents: 1},
		{name: "conversion", changed: true, created: true, wantEvents: 2},
		{name: "retry"},
		{name: "rollback", failure: storydomain.ErrMutationConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &checklistRepositoryStub{result: storydomain.ChecklistMutationResult{Changed: test.changed}, err: test.failure}
			if test.created {
				repository.result.CreatedStory = &storydomain.ChecklistStory{ID: uuid.New(), Title: "Test item"}
			}
			publisher := &recordingStoryEventPublisher{}
			s := New(nil, repository, nil, nil)
			s.publisher = publisher
			_, itemID, err := s.MutateChecklist(ctx, workspaceID, storydomain.ChecklistCommand{StoryID: storyID, Action: storydomain.ChecklistAdd, Title: &title})
			if test.failure == nil && err != nil {
				t.Fatal(err)
			}
			if test.failure != nil && !errors.Is(err, ErrStoryChanged) {
				t.Fatalf("error not mapped: %v", err)
			}
			if len(publisher.events) != test.wantEvents {
				t.Fatalf("published %d events, want %d", len(publisher.events), test.wantEvents)
			}
			command := repository.command
			if command == nil || command.ItemID == uuid.Nil || command.Scope.Actor.PrincipalID != actor.PrincipalID || *command.Title != "Test item" {
				t.Fatalf("unsafe command: %+v", command)
			}
			if command.Scope.ActivityUser == nil || *command.Scope.ActivityUser != actor.PrincipalID || command.Event.Actor.PrincipalID != actor.PrincipalID || command.Event.OccurredAt.IsZero() {
				t.Fatalf("checklist audit identity or timestamp missing: %+v", command)
			}
			if test.failure == nil && itemID != command.ItemID {
				t.Fatal("created item identity lost")
			}
			if len(publisher.events) > 0 {
				payload, ok := publisher.events[0].Payload.(events.StoryUpdatedPayload)
				if !ok || payload.Updates["checklist_changed"] != true {
					t.Fatalf("missing realtime marker: %+v", publisher.events)
				}
			}
		})
	}
}

func TestChecklistServiceRejectsMissingActorCrossWorkspaceAndInvalidInput(t *testing.T) {
	workspaceID := uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := auth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, title string
		ctx         context.Context
		workspaceID uuid.UUID
		want        error
	}{
		{"missing actor", "title", context.Background(), workspaceID, ErrStoryMutationForbidden},
		{"cross workspace", "title", ctx, uuid.New(), ErrStoryMutationForbidden},
		{"empty", " \n ", ctx, workspaceID, ErrInvalidStoryMutation},
		{"too long", strings.Repeat("界", 256), ctx, workspaceID, ErrInvalidStoryMutation},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &checklistRepositoryStub{}
			s := New(nil, repository, nil, nil)
			_, _, err := s.MutateChecklist(test.ctx, test.workspaceID, storydomain.ChecklistCommand{StoryID: uuid.New(), Action: storydomain.ChecklistAdd, Title: &test.title})
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if repository.command != nil {
				t.Fatal("invalid command reached repository")
			}
		})
	}
	repository := &checklistRepositoryStub{}
	s := New(nil, repository, nil, nil)
	if _, err := s.ListChecklist(context.Background(), workspaceID, uuid.New()); !errors.Is(err, ErrStoryReadForbidden) || repository.reads != 0 {
		t.Fatalf("unauthenticated read reached repository: %v", err)
	}
}
