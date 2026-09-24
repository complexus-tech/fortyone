package messaging

import (
	"context"
	"encoding/json"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func slackFileForMutation() StoryAttachmentSource {
	return StoryAttachmentSource{
		Provider: "slack", ExternalWorkspaceID: "T1", ChannelID: "C1",
		ThreadTS: "171.100", MessageTS: "171.101", FileID: "F1",
		Name: "design.png",
	}
}

func TestStoryCreateWithSlackFileBindsExactSourceUntilConfirmation(t *testing.T) {
	t.Parallel()
	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.AvailableFiles = []StoryAttachmentSource{slackFileForMutation()}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	arguments, err := json.Marshal(map[string]any{
		"team_id": team.ID.String(), "title": "Review design", "priority": nil,
		"assignee": assigneeActionUnassigned, "file_id": "F1",
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.createCalls)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Contains(t, confirmation.Prompt, "design.png")
	claims, err := executor.mutations.verifyClaims(confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, &scope.AvailableFiles[0], claims.Attachment)

	// The source is in signed claims, so confirmation does not depend on the
	// provider request's ephemeral file list or model-authored arguments.
	confirmScope := scope
	confirmScope.AvailableFiles = nil
	result, err := executor.ConfirmStoryMutation(context.Background(), confirmScope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, &scope.AvailableFiles[0], result.Attachment)
	require.Len(t, storyService.createCalls, 1)
	retry, err := executor.ConfirmStoryMutation(context.Background(), confirmScope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, &scope.AvailableFiles[0], retry.Attachment)
	require.Len(t, storyService.createCalls, 1)
}

func TestAttachSlackFileRejectsUnlistedIDAndConfirmsAccessibleStory(t *testing.T) {
	t.Parallel()
	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.AvailableFiles = []StoryAttachmentSource{slackFileForMutation()}
	storyService := newMutationStoriesStub()
	storyID := uuid.New()
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID: storyID, Workspace: scope.WorkspaceID, Team: team.ID,
		SequenceID: 7, Title: "Existing story",
	}
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	arguments, err := json.Marshal(map[string]any{
		"story_id": storyID.String(), "story_reference": nil, "file_id": "F-other",
	})
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolAttachStoryFile, Arguments: arguments})
	require.ErrorIs(t, err, ErrInvalidToolArguments)

	arguments, err = json.Marshal(map[string]any{
		"story_id": storyID.String(), "story_reference": nil, "file_id": "F1",
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolAttachStoryFile, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, StoryMutationAttachFile, confirmation.Operation)
	require.Contains(t, confirmation.Prompt, "WEB-7")
	require.Empty(t, storyService.updateCalls)

	denied := scope
	denied.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.ConfirmStoryMutation(context.Background(), denied, confirmation.Token)
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyID, result.StoryID)
	require.Equal(t, &scope.AvailableFiles[0], result.Attachment)
	require.Empty(t, storyService.updateCalls, "attachment import occurs outside messaging confirmation")
}
