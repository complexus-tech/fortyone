package stories

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

type typedAssociationDuplicationRepository struct {
	Repository
	stories            map[uuid.UUID]storydomain.Story
	prepared           storydomain.AssociationSnapshot
	associationCommand *storydomain.AssociationMutationCommand
	duplicateCommand   *storydomain.DuplicateStoryCommand
	applyErr           error
}

func (repository *typedAssociationDuplicationRepository) GetStoryForMutation(
	_ context.Context,
	_ storydomain.MutationScope,
	storyID uuid.UUID,
) (storydomain.Story, error) {
	story, exists := repository.stories[storyID]
	if !exists {
		return storydomain.Story{}, storydomain.ErrNotFound
	}
	return story, nil
}

func (repository *typedAssociationDuplicationRepository) PrepareStoryAssociationMutation(
	_ context.Context,
	_ storydomain.MutationScope,
	_ uuid.UUID,
) (storydomain.AssociationSnapshot, error) {
	return repository.prepared, nil
}

func (repository *typedAssociationDuplicationRepository) ApplyStoryAssociationMutation(
	_ context.Context,
	command storydomain.AssociationMutationCommand,
) (storydomain.AssociationSnapshot, error) {
	repository.associationCommand = &command
	if repository.applyErr != nil {
		return storydomain.AssociationSnapshot{}, repository.applyErr
	}
	result := command.Association
	if command.Expected != nil {
		previousType := command.Expected.Type
		result.PreviousType = &previousType
	}
	return result, nil
}

func (repository *typedAssociationDuplicationRepository) DuplicateStoryMutation(
	_ context.Context,
	command storydomain.DuplicateStoryCommand,
) (storydomain.DuplicateStoryResult, error) {
	repository.duplicateCommand = &command
	if repository.applyErr != nil {
		return storydomain.DuplicateStoryResult{}, repository.applyErr
	}
	source := repository.stories[command.SourceStoryID]
	source.ID = command.TargetStoryID
	source.Title = "Copy of " + source.Title
	source.Reporter = &command.ReporterID
	source.CreatedAt = command.OccurredAt
	source.UpdatedAt = command.OccurredAt
	return storydomain.DuplicateStoryResult{Story: source}, nil
}

func TestTypedAssociationMutationPersistsActivitiesAndEventsInOneIntent(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	oldFromID, oldToID, nextToID := uuid.New(), uuid.New(), uuid.New()
	associationID := uuid.New()
	version := time.Date(2026, time.August, 28, 16, 0, 0, 0, time.UTC)
	repository := &typedAssociationDuplicationRepository{
		stories: map[uuid.UUID]storydomain.Story{
			oldFromID: {ID: oldFromID, Workspace: workspaceID, Team: uuid.New(), Title: "Source", UpdatedAt: version},
			nextToID:  {ID: nextToID, Workspace: workspaceID, Team: uuid.New(), Title: "New target", UpdatedAt: version},
		},
		prepared: storydomain.AssociationSnapshot{
			ID: associationID, FromStoryID: oldFromID, ToStoryID: oldToID, Type: "related",
			FromStoryTitle: "Source", ToStoryTitle: "Old target",
		},
	}
	service := newTypedMutationService(repository)
	association, err := service.UpdateAssociation(
		storyMutationActorContext(t, workspaceID, actorID),
		associationID,
		oldFromID,
		nextToID,
		"blocking",
		workspaceID,
	)
	if err != nil {
		t.Fatalf("UpdateAssociation() error = %v", err)
	}
	if association.Story.ID != nextToID || association.Story.Title != "New target" {
		t.Fatalf("association target story = %#v", association.Story)
	}
	command := repository.associationCommand
	if command == nil || command.Expected == nil {
		t.Fatal("typed association command was not captured")
	}
	if got, want := len(command.Events), 3; got != want {
		t.Fatalf("durable events = %d, want %d affected stories", got, want)
	}
	if got, want := len(command.Activities), 3; got != want {
		t.Fatalf("activities = %d, want %d affected stories", got, want)
	}
	for _, event := range command.Events {
		if event.ID == uuid.Nil || event.Type != storydomain.MutationEventStoryUpdated {
			t.Fatalf("association event = %#v", event)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			t.Fatalf("decode association event payload: %v", err)
		}
		if _, leaked := payload["fromStoryTitle"]; leaked {
			t.Fatalf("association event leaked a title: %s", event.Payload)
		}
	}
}

func TestTypedAssociationMapsCompareAndSwapConflict(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	fromID, toID, associationID := uuid.New(), uuid.New(), uuid.New()
	repository := &typedAssociationDuplicationRepository{
		stories: map[uuid.UUID]storydomain.Story{
			fromID: {ID: fromID, Workspace: workspaceID, Team: uuid.New(), Title: "Source"},
			toID:   {ID: toID, Workspace: workspaceID, Team: uuid.New(), Title: "Target"},
		},
		prepared: storydomain.AssociationSnapshot{
			ID: associationID, FromStoryID: fromID, ToStoryID: toID, Type: "related",
		},
		applyErr: storydomain.ErrMutationConflict,
	}
	service := newTypedMutationService(repository)
	_, err := service.UpdateAssociation(
		storyMutationActorContext(t, workspaceID, actorID),
		associationID,
		fromID,
		toID,
		"blocking",
		workspaceID,
	)
	if !errors.Is(err, ErrStoryChanged) {
		t.Fatalf("UpdateAssociation() error = %v, want story changed", err)
	}
}

func TestTypedDuplicationUsesStableIDsCASAndActorAttribution(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, actorID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	version := time.Date(2026, time.August, 28, 17, 0, 0, 0, time.UTC)
	repository := &typedAssociationDuplicationRepository{
		stories: map[uuid.UUID]storydomain.Story{
			sourceID: {
				ID: sourceID, Workspace: workspaceID, Team: teamID,
				Title: "Original", Priority: "High", UpdatedAt: version,
			},
		},
	}
	service := newTypedMutationService(repository)
	created, err := service.DuplicateStory(
		storyMutationActorContext(t, workspaceID, actorID), sourceID, workspaceID, actorID,
	)
	if err != nil {
		t.Fatalf("DuplicateStory() error = %v", err)
	}
	command := repository.duplicateCommand
	if command == nil || command.TargetStoryID == uuid.Nil || created.ID != command.TargetStoryID {
		t.Fatalf("duplicate result=%#v command=%#v", created, command)
	}
	if command.ReporterID != actorID || !command.ExpectedSourceUpdatedAt.Equal(version) {
		t.Fatalf("duplicate attribution/version = %#v", command)
	}
	if command.Event.ID == uuid.Nil || command.Event.StoryID != command.TargetStoryID ||
		command.Event.Type != storydomain.MutationEventStoryCreated || command.Activity.ID == uuid.Nil {
		t.Fatalf("duplicate event/activity = %#v %#v", command.Event, command.Activity)
	}
	if !strings.Contains(string(command.Event.Payload), "Copy of Original") {
		t.Fatalf("duplicate event payload = %s", command.Event.Payload)
	}
}

func TestTypedDuplicationRejectsSuppliedActorMismatchBeforePersistence(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, sourceID := uuid.New(), uuid.New(), uuid.New()
	repository := &typedAssociationDuplicationRepository{
		stories: map[uuid.UUID]storydomain.Story{
			sourceID: {ID: sourceID, Workspace: workspaceID, Team: uuid.New(), Title: "Original", UpdatedAt: time.Now().UTC()},
		},
	}
	service := newTypedMutationService(repository)
	_, err := service.DuplicateStory(
		storyMutationActorContext(t, workspaceID, actorID), sourceID, workspaceID, uuid.New(),
	)
	if !errors.Is(err, ErrStoryMutationForbidden) || repository.duplicateCommand != nil {
		t.Fatalf("DuplicateStory() error = %v command=%#v", err, repository.duplicateCommand)
	}
}
