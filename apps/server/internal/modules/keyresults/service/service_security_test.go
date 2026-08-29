package keyresults

import (
	"context"
	"errors"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func TestUpdateIntentBindsActorAndPublishesOnlyAfterCommit(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, objectiveID, keyResultID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, time.August, 28, 10, 30, 0, 0, time.UTC)
	repository := &keyResultRepositoryStub{updateResult: keyresultsdomain.MutationResult{
		Before:        keyresultsdomain.KeyResult{ID: keyResultID, ObjectiveID: objectiveID, TargetValue: 80},
		After:         keyresultsdomain.KeyResult{ID: keyResultID, ObjectiveID: objectiveID, TargetValue: 100},
		ChangedFields: []string{"target_value"},
	}}
	publisher := &keyResultPublisherStub{}
	service := New(nil, repository, WithEventPublisher(publisher), withClock(func() time.Time { return now }))

	err := service.UpdateIntent(
		humanKeyResultContext(t, workspaceID, actorID),
		keyResultID,
		workspaceID,
		actorID,
		KeyResultPatch{TargetValue: SetField(100.0)},
		"ready",
	)
	if err != nil {
		t.Fatalf("UpdateIntent() error = %v", err)
	}
	if repository.updateCalls != 1 || repository.updateCommand.Access.ActorID != actorID ||
		repository.updateCommand.Access.WorkspaceID != workspaceID || !repository.updateCommand.Access.AllTeams {
		t.Fatalf("repository update command = %#v", repository.updateCommand)
	}
	if repository.updateCommand.Comment != "ready" {
		t.Fatalf("repository comment = %q", repository.updateCommand.Comment)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
	event := publisher.events[0]
	if event.Type != events.KeyResultUpdated || event.ActorID != actorID || !event.Timestamp.Equal(now) {
		t.Fatalf("event envelope = %#v", event)
	}
	payload, ok := event.Payload.(events.KeyResultUpdatedPayload)
	if !ok || payload.KeyResultID != keyResultID || payload.ObjectiveID != objectiveID || payload.WorkspaceID != workspaceID || payload.Updates["target_value"] != 100.0 {
		t.Fatalf("event payload = %#v", event.Payload)
	}

	repository.updateErr = errors.New("commit failed")
	err = service.UpdateIntent(
		humanKeyResultContext(t, workspaceID, actorID), keyResultID, workspaceID, actorID,
		KeyResultPatch{TargetValue: SetField(120.0)}, "failed",
	)
	if err == nil {
		t.Fatal("UpdateIntent() error = nil for repository failure")
	}
	if len(publisher.events) != 1 {
		t.Fatalf("failed mutation published an event: %#v", publisher.events)
	}
}

func TestWithPublisherAcceptsANilCompatibilityDependency(t *testing.T) {
	t.Parallel()

	service := New(nil, &keyResultRepositoryStub{}, WithPublisher(nil))
	if service.publisher != nil {
		t.Fatalf("nil compatibility publisher = %#v, want nil", service.publisher)
	}
}

func TestKeyResultServicePreservesRestrictedCredentialTeamScope(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, credentialID, teamID, keyResultID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	teamAccess, err := platformauth.RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatalf("RestrictedTeamAccess() error = %v", err)
	}
	actor, err := platformauth.NewActor(
		actorID,
		platformauth.PrincipalPersonalToken,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeObjectivesRead),
		teamAccess,
	)
	if err != nil {
		t.Fatalf("NewActor() error = %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("WithWorkspace() error = %v", err)
	}
	ctx, err := platformauth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("SetActor() error = %v", err)
	}
	repository := &keyResultRepositoryStub{getResult: keyresultsdomain.KeyResult{ID: keyResultID}}
	service := New(nil, repository)

	if _, err := service.GetForActor(ctx, keyResultID, workspaceID, actorID); err != nil {
		t.Fatalf("GetForActor() error = %v", err)
	}
	if repository.getQuery.Access.AllTeams || len(repository.getQuery.Access.TeamIDs) != 1 || repository.getQuery.Access.TeamIDs[0] != teamID {
		t.Fatalf("repository access = %#v", repository.getQuery.Access)
	}
}

func TestKeyResultServiceRejectsCrossWorkspaceActorAndMissingScope(t *testing.T) {
	t.Parallel()

	workspaceID, otherWorkspaceID, actorID := uuid.New(), uuid.New(), uuid.New()
	repository := &keyResultRepositoryStub{}
	service := New(nil, repository)

	_, err := service.GetForActor(humanKeyResultContext(t, workspaceID, actorID), uuid.New(), otherWorkspaceID, actorID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("cross-workspace GetForActor() error = %v, want ErrForbidden", err)
	}

	actor, err := platformauth.NewActor(
		actorID,
		platformauth.PrincipalPersonalToken,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeObjectivesRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("NewActor() error = %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("WithWorkspace() error = %v", err)
	}
	ctx, err := platformauth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("SetActor() error = %v", err)
	}
	err = service.UpdateIntent(ctx, uuid.New(), workspaceID, actorID, KeyResultPatch{Name: SetField("Nope")}, "")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("scope-less UpdateIntent() error = %v, want ErrForbidden", err)
	}
	if repository.getCalls != 0 || repository.updateCalls != 0 {
		t.Fatalf("rejected actor reached repository: get=%d update=%d", repository.getCalls, repository.updateCalls)
	}
}

func TestCreateBatchRejectsForgedActorAndBindsAuthenticatedActor(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	repository := &keyResultRepositoryStub{createResult: []keyresultsdomain.KeyResult{{ID: uuid.New()}}}
	service := New(nil, repository)
	draft := validServiceDraft()
	draft.CreatedBy = uuid.New()

	_, err := service.CreateBatch(humanKeyResultContext(t, workspaceID, actorID), []CoreNewKeyResult{draft}, workspaceID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("forged CreateBatch() error = %v, want ErrForbidden", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("forged create reached repository %d times", repository.createCalls)
	}

	draft.CreatedBy = uuid.Nil
	if _, err := service.CreateBatch(humanKeyResultContext(t, workspaceID, actorID), []CoreNewKeyResult{draft}, workspaceID); err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if repository.createCommand.KeyResults[0].CreatedBy != actorID || repository.createCommand.Access.ActorID != actorID {
		t.Fatalf("create command = %#v", repository.createCommand)
	}
}

func TestExternalCASRecognizesAnAlreadyAppliedTypedPatch(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, keyResultID := uuid.New(), uuid.New(), uuid.New()
	expected := time.Date(2026, time.August, 28, 12, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	repository := &keyResultRepositoryStub{
		updateErr: ErrVersionConflict,
		getResult: keyresultsdomain.KeyResult{ID: keyResultID, CurrentValue: 75},
	}
	service := New(nil, repository)
	patch := KeyResultPatch{CurrentValue: SetField(75.0)}

	err := service.UpdateExternalUserActionIfUnchanged(
		humanKeyResultContext(t, workspaceID, actorID), keyResultID, workspaceID, actorID,
		expected, patch, "reviewed",
	)
	if err != nil {
		t.Fatalf("already-applied update error = %v", err)
	}
	if repository.updateCommand.ExpectedUpdatedAt == nil ||
		repository.updateCommand.ExpectedUpdatedAt.Location() != time.UTC ||
		!repository.updateCommand.ExpectedUpdatedAt.Equal(expected.UTC()) {
		t.Fatalf("expected timestamp = %v", repository.updateCommand.ExpectedUpdatedAt)
	}
	if repository.getCalls != 1 {
		t.Fatalf("CAS reconciliation reads = %d, want 1", repository.getCalls)
	}

	repository.getResult.CurrentValue = 70
	err = service.UpdateExternalUserActionIfUnchanged(
		humanKeyResultContext(t, workspaceID, actorID), keyResultID, workspaceID, actorID,
		expected, patch, "reviewed",
	)
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale CAS error = %v, want ErrVersionConflict", err)
	}
}

func humanKeyResultContext(t *testing.T, workspaceID, actorID uuid.UUID) context.Context {
	t.Helper()
	ctx, err := platformauth.BindWorkspace(platformauth.SetUserID(context.Background(), actorID), workspaceID)
	if err != nil {
		t.Fatalf("BindWorkspace() error = %v", err)
	}
	return ctx
}

func validServiceDraft() CoreNewKeyResult {
	start := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	return CoreNewKeyResult{
		ObjectiveID: uuid.New(), Name: "Ship API", MeasurementType: "percentage",
		TargetValue: 100, StartDate: &start, EndDate: &end,
	}
}

type keyResultPublisherStub struct {
	events []events.Event
	err    error
}

func (publisher *keyResultPublisherStub) Publish(_ context.Context, event events.Event) error {
	publisher.events = append(publisher.events, event)
	return publisher.err
}

type keyResultRepositoryStub struct {
	createCommand keyresultsdomain.CreateCommand
	createResult  []keyresultsdomain.KeyResult
	createErr     error
	createCalls   int
	updateCommand keyresultsdomain.UpdateCommand
	updateResult  keyresultsdomain.MutationResult
	updateErr     error
	updateCalls   int
	deleteCommand keyresultsdomain.DeleteCommand
	deleteErr     error
	getQuery      keyresultsdomain.GetQuery
	getResult     keyresultsdomain.KeyResult
	getErr        error
	getCalls      int
	listQuery     keyresultsdomain.ObjectiveListQuery
	listResult    []keyresultsdomain.KeyResult
	listErr       error
	pageQuery     keyresultsdomain.PaginatedListQuery
	pageResult    keyresultsdomain.ListResponse
	pageErr       error
}

func (repository *keyResultRepositoryStub) CreateBatch(_ context.Context, command keyresultsdomain.CreateCommand) ([]keyresultsdomain.KeyResult, error) {
	repository.createCalls++
	repository.createCommand = command
	return repository.createResult, repository.createErr
}

func (repository *keyResultRepositoryStub) Update(_ context.Context, command keyresultsdomain.UpdateCommand) (keyresultsdomain.MutationResult, error) {
	repository.updateCalls++
	repository.updateCommand = command
	return repository.updateResult, repository.updateErr
}

func (repository *keyResultRepositoryStub) Delete(_ context.Context, command keyresultsdomain.DeleteCommand) error {
	repository.deleteCommand = command
	return repository.deleteErr
}

func (repository *keyResultRepositoryStub) Get(_ context.Context, query keyresultsdomain.GetQuery) (keyresultsdomain.KeyResult, error) {
	repository.getCalls++
	repository.getQuery = query
	return repository.getResult, repository.getErr
}

func (repository *keyResultRepositoryStub) List(_ context.Context, query keyresultsdomain.ObjectiveListQuery) ([]keyresultsdomain.KeyResult, error) {
	repository.listQuery = query
	return repository.listResult, repository.listErr
}

func (repository *keyResultRepositoryStub) ListPaginated(_ context.Context, query keyresultsdomain.PaginatedListQuery) (keyresultsdomain.ListResponse, error) {
	repository.pageQuery = query
	return repository.pageResult, repository.pageErr
}
