package stories

import (
	"context"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type automationStateActorRepository struct {
	typedStoryMutationRepositoryStub
	readScopes  []storydomain.MutationScope
	writeScopes []storydomain.MutationScope
}

func (r *automationStateActorRepository) AuthorizedMayaScheduleBlocksExist(context.Context, storydomain.MutationScope, uuid.UUID) (bool, error) {
	return false, nil
}

func (r *automationStateActorRepository) GetStoryForMutation(_ context.Context, scope storydomain.MutationScope, _ uuid.UUID) (storydomain.Story, error) {
	r.readScopes = append(r.readScopes, scope)
	return r.story, nil
}

func (r *automationStateActorRepository) UpdateAuthorizedAutoSchedulingStateIfUnchanged(
	_ context.Context, scope storydomain.MutationScope, _ uuid.UUID, _ time.Time,
	_ string, _ *string, _ time.Time, _ *bool,
) (bool, error) {
	r.writeScopes = append(r.writeScopes, scope)
	return true, nil
}

func TestUpdateAutomationStateUsesSystemAuthorizedSnapshotWithoutHTTPActor(t *testing.T) {
	t.Parallel()
	workspaceID, storyID, actorID := uuid.New(), uuid.New(), uuid.New()
	version := time.Now().UTC()
	repo := &automationStateActorRepository{typedStoryMutationRepositoryStub: typedStoryMutationRepositoryStub{
		story: CoreSingleStory{ID: storyID, Workspace: workspaceID, UpdatedAt: version, AutoSchedulingEnabled: true, AutoSchedulingStatus: AutoSchedulingStatusPlanning},
	}}
	service := newTypedMutationService(repo)
	err := service.UpdateAutomationStateIfUnchanged(context.Background(), actorID, storyID, workspaceID, version, AutoSchedulingStatusScheduled, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, repo.readScopes, 1)
	require.Len(t, repo.writeScopes, 1)
	scope := repo.readScopes[0]
	require.Equal(t, auth.PrincipalSystem, scope.Actor.Kind)
	require.Equal(t, actorID, scope.Actor.PrincipalID)
	require.Equal(t, workspaceID, scope.Actor.WorkspaceID)
	require.Equal(t, scope, repo.writeScopes[0])
}

func TestUpdateAutomationStateRejectsStaleVersionBeforeWriting(t *testing.T) {
	t.Parallel()
	version := time.Now().UTC()
	repo := &automationStateActorRepository{typedStoryMutationRepositoryStub: typedStoryMutationRepositoryStub{
		story: CoreSingleStory{UpdatedAt: version.Add(time.Second)},
	}}
	service := newTypedMutationService(repo)
	err := service.UpdateAutomationStateIfUnchanged(context.Background(), uuid.New(), uuid.New(), uuid.New(), version, AutoSchedulingStatusScheduled, nil, nil, nil)
	require.ErrorIs(t, err, ErrStoryChanged)
	require.Empty(t, repo.writeScopes)
}

func TestUpdateAutomationStateRejectsMismatchedContextActor(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := auth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	repo := &automationStateActorRepository{}
	service := newTypedMutationService(repo)
	err = service.UpdateAutomationStateIfUnchanged(ctx, uuid.New(), uuid.New(), workspaceID, time.Now(), AutoSchedulingStatusScheduled, nil, nil, nil)
	require.ErrorIs(t, err, ErrStoryMutationForbidden)
	require.Empty(t, repo.readScopes)
	require.Empty(t, repo.writeScopes)
}
