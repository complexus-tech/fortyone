package teamsettings

import (
	"context"
	"errors"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

func TestSprintAutomationIsScheduledOnlyAfterRepositoryCommit(t *testing.T) {
	t.Parallel()

	repo := &repositoryStub{sprint: CoreTeamSprintSettings{AutoCreateSprints: true}}
	scheduler := &schedulerStub{repo: repo}
	service := New(testLogger(), repo, scheduler)

	_, err := service.UpdateSprintSettings(
		context.Background(),
		adminAccess(t),
		CoreUpdateTeamSprintSettings{
			AutoCreateSprints: PatchField[bool]{Value: true, Present: true},
		},
	)
	if err != nil {
		t.Fatalf("UpdateSprintSettings() error = %v", err)
	}
	if scheduler.beforeCommit {
		t.Fatal("scheduler was called before repository commit")
	}
	if scheduler.sprintCalls != 1 {
		t.Fatalf("sprint schedule calls = %d, want 1", scheduler.sprintCalls)
	}
}

func TestRepositoryFailureDoesNotScheduleAutomation(t *testing.T) {
	t.Parallel()

	repositoryError := errors.New("repository unavailable")
	repo := &repositoryStub{err: repositoryError, sprint: CoreTeamSprintSettings{AutoCreateSprints: true}}
	scheduler := &schedulerStub{repo: repo}
	service := New(testLogger(), repo, scheduler)

	_, err := service.UpdateSprintSettings(
		context.Background(),
		adminAccess(t),
		CoreUpdateTeamSprintSettings{
			AutoCreateSprints: PatchField[bool]{Value: true, Present: true},
		},
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("UpdateSprintSettings() error = %v, want %v", err, repositoryError)
	}
	if scheduler.sprintCalls != 0 {
		t.Fatalf("sprint schedule calls = %d, want 0", scheduler.sprintCalls)
	}
}

func TestPostCommitDispatchFailureDoesNotMisreportCommittedUpdate(t *testing.T) {
	t.Parallel()

	dispatchError := errors.New("queue unavailable")
	repo := &repositoryStub{
		story: CoreTeamStoryAutomationSettings{
			AutoCloseInactiveEnabled: true,
			AutoArchiveEnabled:       true,
		},
	}
	scheduler := &schedulerStub{repo: repo, err: dispatchError}
	service := New(nil, repo, scheduler)

	_, err := service.UpdateStoryAutomationSettings(
		context.Background(),
		adminAccess(t),
		CoreUpdateTeamStoryAutomationSettings{
			AutoCloseInactiveEnabled: PatchField[bool]{Value: true, Present: true},
			AutoArchiveEnabled:       PatchField[bool]{Value: true, Present: true},
		},
	)
	if err != nil {
		t.Fatalf("committed update returned dispatch error: %v", err)
	}
	if !repo.committed {
		t.Fatal("repository update was not committed")
	}
	if scheduler.beforeCommit {
		t.Fatal("scheduler was called before repository commit")
	}
	if scheduler.closeCalls != 1 || scheduler.archiveCalls != 1 {
		t.Fatalf("story schedule calls = close:%d archive:%d, want 1 each", scheduler.closeCalls, scheduler.archiveCalls)
	}
}

type schedulerStub struct {
	repo         *repositoryStub
	err          error
	beforeCommit bool
	sprintCalls  int
	closeCalls   int
	archiveCalls int
}

func (s *schedulerStub) ScheduleSprintCreation() error {
	s.recordCommitState()
	s.sprintCalls++
	return s.err
}

func (s *schedulerStub) ScheduleStoryAutoClose() error {
	s.recordCommitState()
	s.closeCalls++
	return s.err
}

func (s *schedulerStub) ScheduleStoryAutoArchive() error {
	s.recordCommitState()
	s.archiveCalls++
	return s.err
}

func (s *schedulerStub) recordCommitState() {
	if s.repo == nil || !s.repo.committed {
		s.beforeCommit = true
	}
}

func adminAccess(t *testing.T) Access {
	t.Helper()
	workspaceID := uuid.New()
	return Access{
		Actor: testBoundActor(
			t,
			platformauth.PrincipalHumanUser,
			workspaceID,
			platformauth.MustScopeSet(platformauth.ScopeTeamsRead),
			platformauth.UnrestrictedTeamAccess(),
		),
		WorkspaceRole: authorization.WorkspaceRoleAdmin,
		WorkspaceID:   workspaceID,
		TeamID:        uuid.New(),
	}
}
