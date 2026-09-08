package mayaadapter

import (
	"context"
	"testing"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type authorityStories struct {
	maya.StoriesService
	actor         auth.Actor
	reads, writes int
}

func (s *authorityStories) Get(ctx context.Context, storyID, workspaceID uuid.UUID) (storydomain.Story, error) {
	s.reads++
	s.actor, _ = auth.GetActor(ctx)
	return storydomain.Story{ID: storyID, Workspace: workspaceID}, nil
}

func (s *authorityStories) UpdateAutomationIfUnchanged(ctx context.Context, _, _, _ uuid.UUID, _ time.Time, _ map[string]any, _ string) error {
	s.writes++
	var err error
	s.actor, err = auth.GetActor(ctx)
	return err
}

func (s *authorityStories) UpdateAutomationStateIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, version time.Time, _ string, _ *string, _ *bool, _ *events.StoryScheduleTransition) error {
	return s.UpdateAutomationIfUnchanged(ctx, actorID, storyID, workspaceID, version, nil, "")
}

func TestInteractiveMayaKeepsUserReadsAndExplicitlyAttributesAutomation(t *testing.T) {
	t.Parallel()
	userID, systemID, workspaceID, storyID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	user, err := auth.NewHumanActor(userID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := auth.SetActor(context.Background(), user)
	require.NoError(t, err)
	backend := &authorityStories{}
	reader := &workerMayaStoryReaderStub{}
	adapter := New(backend, reader, systemID)
	_, err = adapter.Get(ctx, storyID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, userID, backend.actor.PrincipalID)
	require.Empty(t, reader.scopes)
	require.NoError(t, adapter.UpdateAutomationIfUnchanged(ctx, systemID, storyID, workspaceID, time.Now(), nil, "approved plan"))
	require.Equal(t, systemID, backend.actor.PrincipalID)
	require.Equal(t, auth.PrincipalSystem, backend.actor.Kind)
	require.NoError(t, adapter.UpdateAutomationStateIfUnchanged(ctx, systemID, storyID, workspaceID, time.Now(), "scheduled", nil, nil, nil))
	original, err := auth.GetActor(ctx)
	require.NoError(t, err)
	require.Equal(t, user, original)
	_, err = adapter.Get(ctx, storyID, uuid.New())
	require.ErrorIs(t, err, stories.ErrStoryReadForbidden)
	require.ErrorIs(t, adapter.UpdateAutomationIfUnchanged(ctx, systemID, storyID, uuid.New(), time.Now(), nil, ""), stories.ErrStoryMutationForbidden)
	require.ErrorIs(t, adapter.UpdateAutomationIfUnchanged(ctx, uuid.New(), storyID, workspaceID, time.Now(), nil, ""), stories.ErrStoryMutationForbidden)
	require.Equal(t, 2, backend.writes)
}

type authorityWorkload struct {
	actorID, workspaceID, teamID  uuid.UUID
	interactiveCalls, systemCalls int
}

func (s *authorityWorkload) GetSystemTeamWorkload(_ context.Context, actorID, workspaceID, teamID uuid.UUID) (reports.CoreWorkloadAnalysis, error) {
	s.systemCalls++
	s.actorID, s.workspaceID, s.teamID = actorID, workspaceID, teamID
	return reports.CoreWorkloadAnalysis{}, nil
}

func (s *authorityWorkload) GetWorkloadAnalysis(ctx context.Context, _ uuid.UUID, _ reports.ReportFilters) (reports.CoreWorkloadAnalysis, error) {
	s.interactiveCalls++
	actor, err := auth.GetActor(ctx)
	s.actorID = actor.PrincipalID
	return reports.CoreWorkloadAnalysis{}, err
}

func TestMayaBackgroundWorkloadRequiresOneExplicitTeam(t *testing.T) {
	t.Parallel()
	backend := &authorityWorkload{}
	systemID, workspaceID, teamID := uuid.New(), uuid.New(), uuid.New()
	adapter := NewReports(backend, backend, systemID)
	_, err := adapter.GetWorkloadAnalysis(context.Background(), workspaceID, reports.ReportFilters{TeamIDs: []uuid.UUID{teamID}})
	require.NoError(t, err)
	require.Equal(t, systemID, backend.actorID)
	require.Equal(t, workspaceID, backend.workspaceID)
	require.Equal(t, teamID, backend.teamID)
	_, err = adapter.GetWorkloadAnalysis(context.Background(), workspaceID, reports.ReportFilters{})
	require.ErrorIs(t, err, reports.ErrInvalidReportFilters)
	userID := uuid.New()
	_, err = adapter.GetWorkloadAnalysis(auth.SetUserID(context.Background(), userID), workspaceID, reports.ReportFilters{})
	require.NoError(t, err)
	require.Equal(t, userID, backend.actorID)
	require.Equal(t, 1, backend.systemCalls)
	require.Equal(t, 1, backend.interactiveCalls)
}

func TestMayaSystemReadDoesNotPromoteReducedScope(t *testing.T) {
	t.Parallel()
	systemID, workspaceID := uuid.New(), uuid.New()
	actor, err := auth.NewActor(systemID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeStoriesRead), auth.UnrestrictedTeamAccess())
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := auth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	reader := &workerMayaStoryReaderStub{}
	adapter := New(&authorityStories{}, reader, systemID)
	_, err = adapter.Get(ctx, uuid.New(), workspaceID)
	require.ErrorIs(t, err, stories.ErrStoryReadForbidden)
	require.Empty(t, reader.scopes)
}
