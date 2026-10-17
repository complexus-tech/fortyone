package messaging

import (
	"context"
	"encoding/json"
	"testing"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatTaskProposalIsBoundToMessageAndDoesNotWriteBeforeApproval(t *testing.T) {
	t.Parallel()
	scope := ToolScope{WorkspaceID: uuid.New(), UserID: uuid.New(), AllowMutations: true}
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	scope.SourceScope = &SourceScope{Surface: SourceSurfaceChatMessage, TeamID: team.ID, ConversationID: uuid.New(), MessageID: uuid.New()}
	scope.SourceURL = "https://fortyone.app/acme/chat?message=" + scope.SourceScope.MessageID.String()
	stories := newMutationStoriesStub()
	executor := newBatchMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, stories, &usersServiceStub{}, testStoryMutationSecret)
	arguments := json.RawMessage(`{"stories":[{"title":"Improve onboarding","description":"Add the agreed welcome checklist.","priority":null,"assignee_id":null}]}`)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStories, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Empty(t, stories.createCalls)
	require.Contains(t, confirmation.Prompt, "this conversation")
	require.NotContains(t, confirmation.Prompt, "Slack")
	inspection, err := executor.InspectChatMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.False(t, inspection.DirectEligible)
	require.Contains(t, inspection.Review, "Improve onboarding")
	require.Contains(t, inspection.Review, "Unassigned")
	otherMessage := scope
	copy := *scope.SourceScope
	copy.MessageID = uuid.New()
	otherMessage.SourceScope = &copy
	_, err = executor.ConfirmStoryMutation(context.Background(), otherMessage, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	unscoped := scope
	unscoped.SourceScope = nil
	_, err = executor.ConfirmStoryMutation(context.Background(), unscoped, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.Empty(t, stories.createCalls)
	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.NotEqual(t, uuid.Nil, result.Items[0].StoryID)
	require.NotEmpty(t, result.Items[0].Reference)
	require.Len(t, stories.createCalls, 1)
	require.Equal(t, team.ID, result.Items[0].TeamID)
}

func TestChatSourceRejectsOtherTeamAndExistingStoryWrites(t *testing.T) {
	t.Parallel()
	scope := &SourceScope{Surface: SourceSurfaceChatMessage, TeamID: uuid.New(), ConversationID: uuid.New(), MessageID: uuid.New()}
	require.False(t, sourceAllowsTool(scope, toolUpdateStory))
	require.False(t, sourceAllowsTool(scope, toolAddComment))
	require.False(t, sourceAllowsTool(scope, toolCreateStory))
	require.True(t, sourceAllowsTool(scope, toolCreateStories))
	_, err := sourceTeamArguments(scope, json.RawMessage(`{"team_id":"`+uuid.NewString()+`","stories":[]}`))
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	_, err = normalizedSourceScope(&SourceScope{Surface: SourceSurfaceChatMessage, TeamID: scope.TeamID})
	require.ErrorIs(t, err, ErrInvalidRequest)
	instructions, err := instructionsForRequest("base", Request{SourceScope: scope, AllowMutations: true})
	require.NoError(t, err)
	require.Contains(t, instructions, "bounded excerpt")
	require.Contains(t, instructions, "explicitly approve")
}
