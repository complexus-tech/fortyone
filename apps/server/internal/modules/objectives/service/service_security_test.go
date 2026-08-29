package objectives

import (
	"context"
	"errors"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func TestListIntentBindsTheAuthenticatedActor(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	repository := &objectiveRepositoryStub{}
	service := New(nil, repository)
	_, err := service.ListIntent(humanContext(t, workspaceID, userID), objectivesdomain.ListQuery{
		WorkspaceID: workspaceID, ActorID: userID, Search: " roadmap ", Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListIntent() error = %v", err)
	}
	if repository.listCalls != 1 || repository.listQuery.ActorID != userID || repository.listQuery.WorkspaceID != workspaceID {
		t.Fatalf("repository list query = %#v, calls=%d", repository.listQuery, repository.listCalls)
	}
	if repository.listQuery.Search != "roadmap" {
		t.Fatalf("repository search = %q, want trimmed query", repository.listQuery.Search)
	}
}

func TestObjectiveServiceRejectsCrossWorkspaceAndRestrictedCredentials(t *testing.T) {
	t.Parallel()

	workspaceID, otherWorkspaceID, userID := uuid.New(), uuid.New(), uuid.New()
	repository := &objectiveRepositoryStub{}
	service := New(nil, repository)

	_, err := service.ListIntent(humanContext(t, workspaceID, userID), objectivesdomain.ListQuery{
		WorkspaceID: otherWorkspaceID, ActorID: userID,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("cross-workspace ListIntent() error = %v, want ErrForbidden", err)
	}

	teamAccess, err := platformauth.RestrictedTeamAccess(uuid.New())
	if err != nil {
		t.Fatalf("create restricted team access: %v", err)
	}
	actor, err := platformauth.NewActor(
		userID, platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeObjectivesRead), teamAccess,
	)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor workspace: %v", err)
	}
	ctx, err := platformauth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("set actor: %v", err)
	}
	_, err = service.ListIntent(ctx, objectivesdomain.ListQuery{WorkspaceID: workspaceID, ActorID: userID})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("restricted ListIntent() error = %v, want ErrForbidden", err)
	}
	if repository.listCalls != 0 {
		t.Fatalf("repository called %d times for rejected actors", repository.listCalls)
	}

	_, err = service.Get(ctx, uuid.New(), workspaceID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("restricted Get() error = %v, want ErrForbidden", err)
	}
	if repository.getQuery.Internal {
		t.Fatal("restricted actor was widened into the trusted internal read path")
	}
}

func TestGetWithoutAnActorUsesNarrowInternalCompatibilityRead(t *testing.T) {
	t.Parallel()

	workspaceID, objectiveID := uuid.New(), uuid.New()
	repository := &objectiveRepositoryStub{}
	service := New(nil, repository)
	if _, err := service.Get(context.Background(), objectiveID, workspaceID); err != nil {
		t.Fatalf("actorless internal Get() error = %v", err)
	}
	if !repository.getQuery.Internal || repository.getQuery.WorkspaceID != workspaceID ||
		repository.getQuery.ObjectiveID != objectiveID || repository.getQuery.ActorID != uuid.Nil {
		t.Fatalf("internal get query = %#v, want tenant/id-scoped internal read", repository.getQuery)
	}
}

func TestUpdateIntentPropagatesCASAndPublishesOnlyAfterSuccess(t *testing.T) {
	t.Parallel()

	workspaceID, userID, objectiveID := uuid.New(), uuid.New(), uuid.New()
	expected := time.Date(2026, time.August, 28, 12, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	publisher := &objectivePublisherStub{}
	repository := &objectiveRepositoryStub{updateResult: objectivesdomain.Objective{
		ID: objectiveID, Workspace: workspaceID, UpdatedAt: expected.Add(time.Minute),
	}}
	service := New(nil, repository, withEventPublisher(publisher), withClock(func() time.Time {
		return time.Date(2026, time.August, 28, 10, 30, 0, 0, time.UTC)
	}))

	err := service.UpdateIntentIfUnchanged(
		humanContext(t, workspaceID, userID), objectiveID, workspaceID, userID, "reviewed",
		&expected, objectivesdomain.ObjectivePatch{Health: objectivesdomain.SetField(objectivesdomain.HealthOnTrack)},
	)
	if err != nil {
		t.Fatalf("UpdateIntentIfUnchanged() error = %v", err)
	}
	if repository.updateCommand.ExpectedUpdatedAt == nil || repository.updateCommand.ExpectedUpdatedAt.Location() != time.UTC ||
		!repository.updateCommand.ExpectedUpdatedAt.Equal(expected.UTC()) {
		t.Fatalf("repository expected timestamp = %v, want %v UTC", repository.updateCommand.ExpectedUpdatedAt, expected.UTC())
	}
	if repository.updateCommand.Comment != "reviewed" {
		t.Fatalf("repository comment = %q, want reviewed", repository.updateCommand.Comment)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != events.ObjectiveUpdated {
		t.Fatalf("published events = %#v, want one objective.updated", publisher.events)
	}

	repository.updateErr = ErrVersionConflict
	err = service.UpdateIntentIfUnchanged(
		humanContext(t, workspaceID, userID), objectiveID, workspaceID, userID, "stale",
		&expected, objectivesdomain.ObjectivePatch{Name: objectivesdomain.SetField("Stale")},
	)
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale update error = %v, want ErrVersionConflict", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("failed update published an event: %#v", publisher.events)
	}
}

func TestCompatibilityUpdateRejectsUnknownFieldsBeforeRepository(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	repository := &objectiveRepositoryStub{}
	service := New(nil, repository)
	err := service.Update(
		humanContext(t, workspaceID, userID), uuid.New(), workspaceID, userID, "",
		map[string]any{"name = 'compromised'": "ignored"},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Update() error = %v, want ErrInvalid", err)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("repository update calls = %d, want zero", repository.updateCalls)
	}
}

func humanContext(t *testing.T, workspaceID, userID uuid.UUID) context.Context {
	t.Helper()
	ctx, err := platformauth.BindWorkspace(platformauth.SetUserID(context.Background(), userID), workspaceID)
	if err != nil {
		t.Fatalf("bind human actor: %v", err)
	}
	return ctx
}

type objectivePublisherStub struct {
	events []events.Event
	err    error
}

func (publisher *objectivePublisherStub) Publish(_ context.Context, event events.Event) error {
	publisher.events = append(publisher.events, event)
	return publisher.err
}

type objectiveRepositoryStub struct {
	listQuery     objectivesdomain.ListQuery
	listCalls     int
	listResult    []objectivesdomain.Objective
	listErr       error
	getQuery      objectivesdomain.GetQuery
	getResult     objectivesdomain.Objective
	getErr        error
	createCommand objectivesdomain.CreateCommand
	createResult  objectivesdomain.CreateResult
	createErr     error
	updateCommand objectivesdomain.UpdateCommand
	updateResult  objectivesdomain.Objective
	updateErr     error
	updateCalls   int
	deleteCommand objectivesdomain.DeleteCommand
	deleteActorID uuid.UUID
	deleteErr     error
}

func (repository *objectiveRepositoryStub) List(_ context.Context, query objectivesdomain.ListQuery) ([]objectivesdomain.Objective, error) {
	repository.listCalls++
	repository.listQuery = query
	return repository.listResult, repository.listErr
}

func (repository *objectiveRepositoryStub) Get(_ context.Context, query objectivesdomain.GetQuery) (objectivesdomain.Objective, error) {
	repository.getQuery = query
	return repository.getResult, repository.getErr
}

func (repository *objectiveRepositoryStub) Create(_ context.Context, command objectivesdomain.CreateCommand) (objectivesdomain.CreateResult, error) {
	repository.createCommand = command
	return repository.createResult, repository.createErr
}

func (repository *objectiveRepositoryStub) Update(_ context.Context, command objectivesdomain.UpdateCommand) (objectivesdomain.Objective, error) {
	repository.updateCalls++
	repository.updateCommand = command
	return repository.updateResult, repository.updateErr
}

func (repository *objectiveRepositoryStub) Delete(_ context.Context, command objectivesdomain.DeleteCommand) error {
	repository.deleteCommand, repository.deleteActorID = command, command.ActorID
	return repository.deleteErr
}

func (repository *objectiveRepositoryStub) GetAnalytics(context.Context, objectivesdomain.AnalyticsQuery, time.Time) (objectivesdomain.ObjectiveAnalytics, error) {
	return objectivesdomain.ObjectiveAnalytics{}, nil
}

func (repository *objectiveRepositoryStub) GetStrategyMap(context.Context, objectivesdomain.StrategyQuery) (objectivesdomain.StrategyMap, error) {
	return objectivesdomain.StrategyMap{}, nil
}

func (repository *objectiveRepositoryStub) UpdateStrategy(context.Context, objectivesdomain.StrategyQuery, objectivesdomain.StrategyUpdate) error {
	return nil
}

func (repository *objectiveRepositoryStub) CreateStrategicPillar(context.Context, objectivesdomain.StrategyQuery, objectivesdomain.NewStrategicPillar) (objectivesdomain.StrategicPillar, error) {
	return objectivesdomain.StrategicPillar{}, nil
}

func (repository *objectiveRepositoryStub) UpdateStrategicPillar(context.Context, objectivesdomain.StrategyQuery, uuid.UUID, objectivesdomain.UpdateStrategicPillar) (objectivesdomain.StrategicPillar, error) {
	return objectivesdomain.StrategicPillar{}, nil
}

func (repository *objectiveRepositoryStub) DeleteStrategicPillar(context.Context, objectivesdomain.StrategyQuery, uuid.UUID) error {
	return nil
}

func (repository *objectiveRepositoryStub) AlignObjective(context.Context, objectivesdomain.StrategyQuery, uuid.UUID, *uuid.UUID) error {
	return nil
}
