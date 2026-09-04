package workspaces

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceCreationHonorsExampleChoiceAfterCoreCommit(t *testing.T) {
	t.Parallel()
	include, exclude := true, false
	tests := []struct {
		name         string
		legacyMethod bool
		options      CreationOptions
		content      string
		workType     WorkType
		contentErr   error
	}{
		{name: "existing service callers", legacyMethod: true, content: "legacy"},
		{name: "omitted HTTP option", content: "legacy"},
		{name: "work type alone preserves legacy content", options: CreationOptions{WorkType: WorkTypeProduct}, content: "legacy"},
		{name: "explicitly empty", options: CreationOptions{IncludeExamples: &exclude, WorkType: WorkTypeProduct}},
		{name: "chosen examples", options: CreationOptions{IncludeExamples: &include, WorkType: WorkTypeProduct}, content: "examples", workType: WorkTypeProduct},
		{name: "default general examples", options: CreationOptions{IncludeExamples: &include}, content: "examples", workType: WorkTypeGeneral},
		{
			name:    "content failure does not misreport a committed workspace",
			options: CreationOptions{IncludeExamples: &include, WorkType: WorkTypePersonal},
			content: "examples", workType: WorkTypePersonal, contentErr: errors.New("story creation unavailable"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := newCreationRecorder()
			recorder.contentErr = test.contentErr
			service := recorder.service()
			input := CoreWorkspace{Name: "New workspace", Slug: " New-Workspace ", TeamSize: "1-10"}
			var workspace CoreWorkspace
			var err error
			if test.legacyMethod {
				workspace, err = service.Create(t.Context(), input, recorder.userID)
			} else {
				workspace, err = service.CreateWithOptions(t.Context(), input, recorder.userID, test.options)
			}

			require.NoError(t, err)
			require.Equal(t, recorder.workspaceID, workspace.ID)
			require.Equal(t, "new-workspace", workspace.Slug)
			require.Equal(t, input.Name, workspace.Name)
			require.Equal(t, input.TeamSize, workspace.TeamSize)
			wantEvents := []string{"begin", "workspace", "workspace member", "team", "team member", "last used", "settings", "commit"}
			if test.content != "" {
				wantEvents = append(wantEvents, test.content)
				require.Equal(t, [3]uuid.UUID{recorder.workspaceID, recorder.teamID, recorder.userID}, recorder.contentIDs)
			}
			wantEvents = append(wantEvents, "trial")
			require.Equal(t, wantEvents, recorder.events)
			require.Equal(t, test.workType, recorder.workType)
			require.Equal(t, "new-workspace", recorder.trial.WorkspaceSlug)
			require.Equal(t, recorder.userID, recorder.trial.UserID)
		})
	}
}

func TestWorkspaceCreationValidatesExampleOptionsBeforePersistence(t *testing.T) {
	t.Parallel()
	include, exclude := true, false
	for _, options := range []CreationOptions{
		{WorkType: "unknown"},
		{IncludeExamples: &exclude, WorkType: "unknown"},
		{IncludeExamples: &include, WorkType: "Product"},
		{IncludeExamples: &include, WorkType: " product "},
	} {
		recorder := newCreationRecorder()
		_, err := recorder.service().CreateWithOptions(t.Context(), CoreWorkspace{Slug: "new-workspace"}, recorder.userID, options)
		require.ErrorIs(t, err, ErrInvalidWorkType)
		require.Empty(t, recorder.events)
	}

	recorder := newCreationRecorder()
	service := recorder.service()
	service.examples = nil
	_, err := service.CreateWithOptions(t.Context(), CoreWorkspace{Slug: "new-workspace"}, recorder.userID, CreationOptions{IncludeExamples: &include})
	require.ErrorContains(t, err, "workspace examples are unavailable")
	require.Empty(t, recorder.events)
}

func TestWorkspaceCreationDoesNotCreateContentBeforeSuccessfulCommit(t *testing.T) {
	t.Parallel()
	include, exclude := true, false
	for _, includeExamples := range []*bool{nil, &exclude, &include} {
		recorder := newCreationRecorder()
		recorder.commitErr = errors.New("commit failed")
		_, err := recorder.service().CreateWithOptions(t.Context(), CoreWorkspace{Slug: "new-workspace"}, recorder.userID, CreationOptions{IncludeExamples: includeExamples})
		require.ErrorIs(t, err, ErrTx)
		require.Equal(t, []string{"begin", "workspace", "workspace member", "team", "team member", "last used", "settings", "rollback"}, recorder.events)
		require.Empty(t, recorder.contentIDs)
		require.Empty(t, recorder.trial)
	}
}

// creationRecorder records the transaction boundary and external collaborators
// without connecting to PostgreSQL, a task queue, or the stories service.
type creationRecorder struct {
	workspaceID uuid.UUID
	teamID      uuid.UUID
	userID      uuid.UUID
	events      []string
	contentIDs  [3]uuid.UUID
	workType    WorkType
	trial       TrialStart
	contentErr  error
	commitErr   error
}

func newCreationRecorder() *creationRecorder {
	return &creationRecorder{workspaceID: uuid.New(), teamID: uuid.New(), userID: uuid.New()}
}

func (recorder *creationRecorder) service() *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "workspace-creation-test"), nil, recorder, Dependencies{
		SeedContent: recorder, Examples: recorder, Users: recorder, Trials: recorder,
	})
}

func (recorder *creationRecorder) WithinTransaction(ctx context.Context, operation func(Transaction) error) error {
	recorder.events = append(recorder.events, "begin")
	if err := operation(recorder); err != nil {
		recorder.events = append(recorder.events, "rollback")
		return err
	}
	if recorder.commitErr != nil {
		recorder.events = append(recorder.events, "rollback")
		return recorder.commitErr
	}
	recorder.events = append(recorder.events, "commit")
	return nil
}

func (recorder *creationRecorder) CreateWorkspace(_ context.Context, input CoreWorkspace, _ uuid.UUID) (CoreWorkspace, error) {
	recorder.events = append(recorder.events, "workspace")
	input.ID = recorder.workspaceID
	return input, nil
}

func (recorder *creationRecorder) AddWorkspaceMember(context.Context, uuid.UUID, uuid.UUID, string) error {
	recorder.events = append(recorder.events, "workspace member")
	return nil
}

func (recorder *creationRecorder) CreateTeam(context.Context, DefaultTeam) (CreatedTeam, error) {
	recorder.events = append(recorder.events, "team")
	return CreatedTeam{ID: recorder.teamID}, nil
}

func (recorder *creationRecorder) AddTeamMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	recorder.events = append(recorder.events, "team member")
	return nil
}

func (recorder *creationRecorder) UpdateLastUsedWorkspace(context.Context, uuid.UUID, uuid.UUID) error {
	recorder.events = append(recorder.events, "last used")
	return nil
}

func (recorder *creationRecorder) InitializeWorkspaceSettings(context.Context, uuid.UUID) error {
	recorder.events = append(recorder.events, "settings")
	return nil
}

func (recorder *creationRecorder) CreateWorkspaceSeedContent(_ context.Context, workspaceID, teamID, userID uuid.UUID) error {
	recorder.events = append(recorder.events, "legacy")
	recorder.contentIDs = [3]uuid.UUID{workspaceID, teamID, userID}
	return recorder.contentErr
}

func (recorder *creationRecorder) CreateWorkspaceExamples(_ context.Context, workspaceID, teamID, userID uuid.UUID, workType WorkType) error {
	recorder.events = append(recorder.events, "examples")
	recorder.contentIDs = [3]uuid.UUID{workspaceID, teamID, userID}
	recorder.workType = workType
	return recorder.contentErr
}

func (*creationRecorder) GetWorkspaceUser(context.Context, uuid.UUID) (WorkspaceUser, error) {
	return WorkspaceUser{Email: "creator@example.com", FullName: "Workspace creator"}, nil
}

func (recorder *creationRecorder) ScheduleWorkspaceTrialStart(trial TrialStart) error {
	recorder.events = append(recorder.events, "trial")
	recorder.trial = trial
	return nil
}
