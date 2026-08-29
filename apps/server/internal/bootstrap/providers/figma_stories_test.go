package providers

import (
	"context"
	"testing"

	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type figmaStorySourceStub struct {
	created  stories.CoreNewStory
	activity stories.CoreActivity
	result   stories.CoreSingleStory
}

func (source *figmaStorySourceStub) Get(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (stories.CoreSingleStory, error) {
	return source.result, nil
}

func (source *figmaStorySourceStub) CreateExternal(
	_ context.Context,
	_ uuid.UUID,
	input stories.CoreNewStory,
	_ uuid.UUID,
) (stories.CoreSingleStory, error) {
	source.created = input
	return source.result, nil
}

func (source *figmaStorySourceStub) RecordActivity(
	_ context.Context,
	activity stories.CoreActivity,
) error {
	source.activity = activity
	return nil
}

func TestFigmaStoryAdapterTranslatesAtTheCompositionBoundary(t *testing.T) {
	t.Parallel()

	storyID, actorID, workspaceID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	source := &figmaStorySourceStub{result: stories.CoreSingleStory{
		ID: storyID, SequenceID: 41, TeamCode: "API", Title: "Typed boundary",
	}}
	adapter, err := NewFigmaStoryAdapter(source)
	require.NoError(t, err)

	created, err := adapter.CreateExternal(t.Context(), actorID, figma.NewStory{
		Title: "Typed boundary", TeamID: teamID, Priority: "high",
	}, workspaceID)
	require.NoError(t, err)
	require.Equal(t, teamID, source.created.Team)
	require.Equal(t, "high", source.created.Priority)
	require.Equal(t, figma.Story{
		ID: storyID, SequenceID: 41, TeamCode: "API", Title: "Typed boundary",
	}, created)

	err = adapter.RecordActivity(t.Context(), figma.StoryActivity{
		StoryID: storyID, ActorID: actorID, WorkspaceID: workspaceID,
		Type: "update", Field: "figma", Previous: "before", Current: "after",
	})
	require.NoError(t, err)
	require.Equal(t, actorID, source.activity.UserID)
	require.Equal(t, "before", source.activity.CurrentValue)
	require.Equal(t, "after", source.activity.NewValue)
}
