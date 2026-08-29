package storyadapter

import (
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestToStoryPlanMapsOnlyFeedbackPlanningFields(t *testing.T) {
	t.Parallel()

	statusID := uuid.New()
	story := stories.CoreSingleStory{
		ID:     uuid.New(),
		Team:   uuid.New(),
		Status: &statusID,
		Title:  "does not leak through the feedback port",
	}

	require.Equal(t, feedback.StoryPlan{
		ID:          story.ID,
		WorkspaceID: story.Workspace,
		TeamID:      story.Team,
		StatusID:    &statusID,
	}, toStoryPlan(story))
}
