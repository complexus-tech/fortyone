package stories

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDesktopMediaReconciliationRetainsClientVersionAndUsesDurableDeletion(t *testing.T) {
	workspace, story, actor, attachment := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	expected := time.Date(2026, 10, 15, 9, 0, 0, 123456000, time.UTC)
	repository := &typedStoryMutationRepositoryStub{story: CoreSingleStory{ID: story, Workspace: workspace, Team: uuid.New(), Title: "Story", Priority: "Low", UpdatedAt: expected}}
	service := newTypedMutationService(repository)
	ctx := storyMutationActorContext(t, workspace, actor)
	err := service.UpdatePatchWithMediaReconciliationIfUnchanged(ctx, story, workspace, expected, StoryPatch{DescriptionHTML: SetField("<p>Changed</p>")}, []uuid.UUID{attachment})
	require.NoError(t, err)
	require.NotNil(t, repository.updateCommand)
	require.Equal(t, expected, repository.updateCommand.ExpectedUpdatedAt)
	require.True(t, repository.updateCommand.ReconcileMedia)
	require.True(t, repository.updateCommand.DeferMediaDeletion)
	require.Equal(t, []uuid.UUID{attachment}, repository.updateCommand.ReferencedMedia)
	repository.updateCommand = nil
	err = service.UpdatePatchWithMediaReconciliationIfUnchanged(ctx, story, workspace, expected.Add(-time.Second), StoryPatch{DescriptionHTML: SetField("<p>Stale</p>")}, nil)
	require.ErrorIs(t, err, ErrStoryChanged)
	require.Nil(t, repository.updateCommand)
}
