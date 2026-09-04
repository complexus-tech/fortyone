package workspaces

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestExampleStoriesMatchWorkTypeWithoutInventingProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		workType WorkType
		titles   []string
	}{
		{WorkTypeProduct, []string{"Define the next product improvement", "Build a small prototype", "Review feedback and choose the next step"}},
		{WorkTypeMarketing, []string{"Plan the next campaign", "Draft the first campaign message", "Review campaign results"}},
		{WorkTypeOperations, []string{"Document a recurring process", "Improve one handoff", "Review the weekly workflow"}},
		{WorkTypePersonal, []string{"Choose one goal for this week", "Break the goal into a small next step", "Review what worked"}},
		{WorkTypeGeneral, []string{"Define the outcome you want", "Choose the first small task", "Review progress and decide what comes next"}},
	}
	for _, test := range tests {
		t.Run(string(test.workType), func(t *testing.T) {
			t.Parallel()
			teamID, userID, unstartedID := uuid.New(), uuid.New(), uuid.New()
			stories, err := BuildExampleStories(teamID, userID, []SeedStatus{
				{ID: uuid.New(), Category: "completed"},
				{ID: uuid.New(), Category: "backlog"},
				{ID: unstartedID, Category: "unstarted"},
			}, test.workType)

			require.NoError(t, err)
			require.Len(t, stories, 3)
			for index, story := range stories {
				require.Equal(t, "[Example] "+test.titles[index], story.Title)
				require.NotNil(t, story.Description)
				require.Contains(t, *story.Description, "This is an example. Edit or delete it")
				require.NotNil(t, story.DescriptionHTML)
				require.Contains(t, *story.DescriptionHTML, "This is an example.")
				require.Equal(t, userID, story.Reporter)
				require.Equal(t, userID, story.Assignee)
				require.Equal(t, teamID, story.Team)
				require.Equal(t, unstartedID, story.Status)
				require.Equal(t, "No Priority", story.Priority)
				require.Nil(t, story.StartDate)
				require.Nil(t, story.EndDate)
			}
		})
	}
}

func TestExampleStoriesUseGeneralWhenWorkTypeIsOmitted(t *testing.T) {
	t.Parallel()
	teamID, userID := uuid.New(), uuid.New()
	statuses := []SeedStatus{{ID: uuid.New(), Category: "unstarted"}}
	general, err := BuildExampleStories(teamID, userID, statuses, WorkTypeGeneral)
	require.NoError(t, err)
	omitted, err := BuildExampleStories(teamID, userID, statuses, "")
	require.NoError(t, err)
	require.Equal(t, general, omitted)
}

func TestExampleStoriesRequireAnUnstartedOrBacklogStatus(t *testing.T) {
	t.Parallel()

	backlogID := uuid.New()
	stories, err := BuildExampleStories(uuid.New(), uuid.New(), []SeedStatus{
		{ID: uuid.New(), Category: "started"},
		{ID: backlogID, Category: "backlog"},
	}, WorkTypeGeneral)
	require.NoError(t, err)
	for _, story := range stories {
		require.Equal(t, backlogID, story.Status)
	}

	for _, statuses := range [][]SeedStatus{
		nil,
		{{ID: uuid.New(), Category: "completed"}},
		{{ID: uuid.New(), Category: "started"}},
		{{ID: uuid.Nil, Category: "unstarted"}},
	} {
		stories, err := BuildExampleStories(uuid.New(), uuid.New(), statuses, WorkTypeGeneral)
		require.ErrorIs(t, err, ErrNoExampleStatuses)
		require.Nil(t, stories)
	}
}

func TestExampleStoriesRejectUnknownWorkType(t *testing.T) {
	t.Parallel()
	stories, err := BuildExampleStories(uuid.New(), uuid.New(), []SeedStatus{
		{ID: uuid.New(), Category: "unstarted"},
	}, "unknown")
	require.ErrorIs(t, err, ErrInvalidWorkType)
	require.Nil(t, stories)
}
