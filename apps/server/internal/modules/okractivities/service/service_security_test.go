package okractivities

import (
	"context"
	"errors"
	"testing"

	okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestCreateBatchBindsTheAuthenticatedActorAndNormalizesInput(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, objectiveID := uuid.New(), uuid.New(), uuid.New()
	repository := &activityRepositoryStub{}
	service := New(repository)
	activity := CoreNewActivity{
		ObjectiveID: objectiveID, UserID: actorID, WorkspaceID: workspaceID,
		Type: ActivityTypeUpdate, UpdateType: UpdateTypeObjective,
		Field: " health ", Comment: " reviewed ",
	}

	if err := service.CreateBatch(activityHumanContext(t, workspaceID, actorID), []CoreNewActivity{activity}); err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	if repository.createCalls != 1 || len(repository.created) != 1 {
		t.Fatalf("repository create calls/rows = %d/%d", repository.createCalls, len(repository.created))
	}
	if repository.created[0].UserID != actorID || repository.created[0].WorkspaceID != workspaceID ||
		repository.created[0].Field != "health" || repository.created[0].Comment != "reviewed" {
		t.Fatalf("repository activity = %#v", repository.created[0])
	}
}

func TestCreateBatchRejectsForgedActorBeforeRepository(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	repository := &activityRepositoryStub{}
	service := New(repository)
	err := service.Create(activityHumanContext(t, workspaceID, actorID), CoreNewActivity{
		ObjectiveID: uuid.New(), UserID: uuid.New(), WorkspaceID: workspaceID,
		Type: ActivityTypeCreate, UpdateType: UpdateTypeObjective,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Create() error = %v, want ErrForbidden", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("forged actor reached repository %d times", repository.createCalls)
	}
}

func TestActivityReadsBindActorAndNormalizePagination(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, objectiveID := uuid.New(), uuid.New(), uuid.New()
	repository := &activityRepositoryStub{listResult: []okractivitiesdomain.Activity{{ID: uuid.New()}}, hasMore: true}
	service := New(repository)

	activities, hasMore, err := service.GetObjectiveActivities(
		activityHumanContext(t, workspaceID, actorID), objectiveID, workspaceID, -1, MaximumPageSize+10,
	)
	if err != nil {
		t.Fatalf("GetObjectiveActivities() error = %v", err)
	}
	if len(activities) != 1 || !hasMore {
		t.Fatalf("activities/hasMore = %d/%v", len(activities), hasMore)
	}
	if repository.listQuery.ActorID != actorID || repository.listQuery.WorkspaceID != workspaceID ||
		repository.listQuery.ObjectiveID != objectiveID || repository.listQuery.Page != 1 ||
		repository.listQuery.PageSize != MaximumPageSize {
		t.Fatalf("repository list query = %#v", repository.listQuery)
	}
}

func TestActivityServiceRejectsCrossWorkspaceAndRestrictedCredentials(t *testing.T) {
	t.Parallel()

	workspaceID, otherWorkspaceID, actorID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &activityRepositoryStub{}
	service := New(repository)
	_, _, err := service.GetObjectiveActivities(
		activityHumanContext(t, workspaceID, actorID), uuid.New(), otherWorkspaceID, 1, 20,
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("cross-workspace read error = %v, want ErrForbidden", err)
	}

	teamAccess, err := platformauth.RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatalf("RestrictedTeamAccess() error = %v", err)
	}
	actor, err := platformauth.NewActor(
		actorID, platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeObjectivesRead), teamAccess,
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
	_, _, err = service.GetObjectiveActivities(ctx, uuid.New(), workspaceID, 1, 20)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("restricted read error = %v, want ErrForbidden", err)
	}
	if repository.listCalls != 0 {
		t.Fatalf("rejected reads reached repository %d times", repository.listCalls)
	}
}

func activityHumanContext(t *testing.T, workspaceID, actorID uuid.UUID) context.Context {
	t.Helper()
	ctx, err := platformauth.BindWorkspace(platformauth.SetUserID(context.Background(), actorID), workspaceID)
	if err != nil {
		t.Fatalf("BindWorkspace() error = %v", err)
	}
	return ctx
}

type activityRepositoryStub struct {
	created     []okractivitiesdomain.NewActivity
	createErr   error
	createCalls int
	listQuery   okractivitiesdomain.ListQuery
	listResult  []okractivitiesdomain.Activity
	hasMore     bool
	listErr     error
	listCalls   int
}

func (repository *activityRepositoryStub) Create(_ context.Context, activity okractivitiesdomain.NewActivity) error {
	return repository.CreateBatch(context.Background(), []okractivitiesdomain.NewActivity{activity})
}

func (repository *activityRepositoryStub) CreateBatch(_ context.Context, activities []okractivitiesdomain.NewActivity) error {
	repository.createCalls++
	repository.created = append([]okractivitiesdomain.NewActivity(nil), activities...)
	return repository.createErr
}

func (repository *activityRepositoryStub) List(_ context.Context, query okractivitiesdomain.ListQuery) ([]okractivitiesdomain.Activity, bool, error) {
	repository.listCalls++
	repository.listQuery = query
	return repository.listResult, repository.hasMore, repository.listErr
}
