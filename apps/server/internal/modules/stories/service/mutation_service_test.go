package stories

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type typedStoryMutationRepositoryStub struct {
	Repository
	preconditions storydomain.MutationPreconditions
	story         CoreSingleStory
	createCommand *storydomain.CreateStoryCommand
	updateCommand *storydomain.UpdateStoryCommand
	deleteCommand *storydomain.DeleteStoryCommand
	applyErr      error
	deleteErr     error
}

func (repository *typedStoryMutationRepositoryStub) PrepareStoryMutation(
	context.Context,
	storydomain.MutationScope,
	uuid.UUID,
	*uuid.UUID,
) (storydomain.MutationPreconditions, error) {
	return repository.preconditions, nil
}

func (repository *typedStoryMutationRepositoryStub) GetStoryForMutation(
	context.Context,
	storydomain.MutationScope,
	uuid.UUID,
) (storydomain.Story, error) {
	return repository.story, nil
}

func (repository *typedStoryMutationRepositoryStub) CreateStoryMutation(
	_ context.Context,
	command storydomain.CreateStoryCommand,
) (storydomain.CreateStoryResult, error) {
	repository.createCommand = &command
	created := command.Story
	created.SequenceID = 1
	created.CreatedNow = true
	return storydomain.CreateStoryResult{Story: created, Created: true}, nil
}

func (repository *typedStoryMutationRepositoryStub) ApplyStoryMutation(
	_ context.Context,
	command storydomain.UpdateStoryCommand,
) (storydomain.UpdateStoryResult, error) {
	repository.updateCommand = &command
	return storydomain.UpdateStoryResult{UpdatedAt: command.Event.OccurredAt}, repository.applyErr
}

func (repository *typedStoryMutationRepositoryStub) DeleteStoryMutation(
	_ context.Context,
	command storydomain.DeleteStoryCommand,
) (storydomain.DeleteStoryResult, error) {
	repository.deleteCommand = &command
	return storydomain.DeleteStoryResult{Deleted: repository.deleteErr == nil}, repository.deleteErr
}

func TestCreateUsesTypedAtomicMutationCommand(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, actorID := uuid.New(), uuid.New(), uuid.New()
	repository := &typedStoryMutationRepositoryStub{
		preconditions: storydomain.MutationPreconditions{EstimateScheme: DefaultEstimateScheme},
	}
	service := newTypedMutationService(repository)
	ctx := storyMutationActorContext(t, workspaceID, actorID)
	created, err := service.Create(ctx, CoreNewStory{
		Title: "Typed create", Team: teamID, Reporter: &actorID, Priority: "High",
	}, workspaceID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == uuid.Nil || repository.createCommand == nil {
		t.Fatalf("created story or command missing: story=%#v command=%#v", created, repository.createCommand)
	}
	command := repository.createCommand
	if command.Event.ID == uuid.Nil || command.Event.Type != storydomain.MutationEventStoryCreated || command.Event.StoryID != created.ID {
		t.Fatalf("create event = %#v", command.Event)
	}
	if command.Activity == nil || command.Activity.ID == uuid.Nil || command.Activity.UserID != actorID {
		t.Fatalf("create activity = %#v", command.Activity)
	}
}

func TestCreateExternalMarksInternalOnlyMutationEventDelivery(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, actorID := uuid.New(), uuid.New(), uuid.New()
	repository := &typedStoryMutationRepositoryStub{
		preconditions: storydomain.MutationPreconditions{EstimateScheme: DefaultEstimateScheme},
	}
	service := newTypedMutationService(repository)
	ctx := storyMutationActorContext(t, workspaceID, actorID)
	_, err := service.CreateExternal(ctx, actorID, CoreNewStory{
		Title: "Imported story", Team: teamID, Reporter: &actorID, Priority: "High",
		ExternalDelivery: storydomain.ExternalStoryDeliveryInternalOnly,
	}, workspaceID)
	if err != nil {
		t.Fatalf("CreateExternal() error = %v", err)
	}
	if repository.createCommand == nil {
		t.Fatal("typed create command was not recorded")
	}
	var payload struct {
		Delivery string `json:"_delivery"`
	}
	if err := json.Unmarshal(repository.createCommand.Event.Payload, &payload); err != nil {
		t.Fatalf("decode mutation event payload: %v", err)
	}
	if payload.Delivery != string(mutationEventDeliveryInternalOnly) {
		t.Fatalf("delivery = %q, want %q", payload.Delivery, mutationEventDeliveryInternalOnly)
	}
}

func TestServiceAccountCreatePreservesMachineActorWithoutInventingReporter(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, principalID, credentialID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	actor, err := platformauth.NewActor(
		principalID,
		platformauth.PrincipalServiceAccount,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("construct service-account actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind service-account actor: %v", err)
	}
	ctx, err := platformauth.SetActor(t.Context(), actor)
	if err != nil {
		t.Fatalf("set service-account actor: %v", err)
	}
	repository := &typedStoryMutationRepositoryStub{
		preconditions: storydomain.MutationPreconditions{EstimateScheme: DefaultEstimateScheme},
	}
	service := newTypedMutationService(repository)

	if _, err := service.Create(ctx, CoreNewStory{
		Title: "Machine-created story", Team: teamID, Priority: "High",
	}, workspaceID); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	command := repository.createCommand
	if command == nil {
		t.Fatal("typed create command was not recorded")
	}
	if command.Scope.Actor.Kind != platformauth.PrincipalServiceAccount || command.Scope.Actor.PrincipalID != principalID {
		t.Fatalf("machine actor was not preserved: %#v", command.Scope.Actor)
	}
	if command.Story.Reporter != nil || command.Activity != nil {
		t.Fatalf("machine create invented user attribution: reporter=%v activity=%#v", command.Story.Reporter, command.Activity)
	}
}

func TestUpdatePatchCarriesCASActivitiesAndDurableEvent(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	version := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	repository := &typedStoryMutationRepositoryStub{
		story: CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: teamID,
			Title: "Before", Priority: "High", UpdatedAt: version,
			EstimateScheme: DefaultEstimateScheme,
		},
	}
	service := newTypedMutationService(repository)
	err := service.UpdatePatch(
		storyMutationActorContext(t, workspaceID, actorID),
		storyID,
		workspaceID,
		StoryPatch{Title: SetField("After"), Description: ClearField[string]()},
	)
	if err != nil {
		t.Fatalf("UpdatePatch() error = %v", err)
	}
	command := repository.updateCommand
	if command == nil || !command.ExpectedUpdatedAt.Equal(version) {
		t.Fatalf("update command = %#v, want version %v", command, version)
	}
	if command.Event.ID == uuid.Nil || command.Event.Type != storydomain.MutationEventStoryUpdated {
		t.Fatalf("update event = %#v", command.Event)
	}
	if got, want := len(command.Activities), 1; got != want || command.Activities[0].Field != "title" {
		t.Fatalf("activities = %#v, want one title activity", command.Activities)
	}
	if fields := command.Patch.Fields(); len(fields) != 1 || fields[0] != "title" {
		t.Fatalf("normalized patch fields = %#v, want the no-op description clear removed", fields)
	}
}

func TestUpdatePatchMapsRepositoryConflict(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &typedStoryMutationRepositoryStub{
		story: CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: teamID, Title: "Before",
			UpdatedAt: time.Now().UTC(), EstimateScheme: DefaultEstimateScheme,
		},
		applyErr: storydomain.ErrMutationConflict,
	}
	service := newTypedMutationService(repository)
	err := service.UpdatePatch(
		storyMutationActorContext(t, workspaceID, actorID), storyID, workspaceID,
		StoryPatch{Title: SetField("After")},
	)
	if !errors.Is(err, ErrStoryChanged) {
		t.Fatalf("UpdatePatch() error = %v, want story changed", err)
	}
}

func TestDeleteUsesCurrentVersionAndDurableDeletedEvent(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID, actorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	version := time.Date(2026, time.August, 28, 8, 30, 0, 0, time.UTC)
	repository := &typedStoryMutationRepositoryStub{
		story: CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: teamID,
			Reporter: &actorID, Title: "Delete me", UpdatedAt: version,
		},
	}
	service := newTypedMutationService(repository)
	err := service.Delete(
		storyMutationActorContext(t, workspaceID, actorID), storyID, workspaceID,
		BulkDeleteAuthorization{ActorID: actorID},
	)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	command := repository.deleteCommand
	if command == nil || !command.ExpectedUpdatedAt.Equal(version) || command.Event.Type != storydomain.MutationEventStoryDeleted {
		t.Fatalf("delete command = %#v", command)
	}
}

func newTypedMutationService(repository Repository) *Service {
	return New(
		logger.NewWithText(io.Discard, slog.LevelError, "test"),
		repository,
		nil,
		nil,
	)
}

func storyMutationActorContext(t *testing.T, workspaceID, actorID uuid.UUID) context.Context {
	t.Helper()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor workspace: %v", err)
	}
	ctx, err := platformauth.SetActor(t.Context(), actor)
	if err != nil {
		t.Fatalf("set actor context: %v", err)
	}
	return ctx
}
