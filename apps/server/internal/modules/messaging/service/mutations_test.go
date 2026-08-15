package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testStoryMutationSecret = "a stable worker secret used only by messaging tests"

func TestFortyOneToolExecutorMutationCatalogIsStrictAndOptIn(t *testing.T) {
	t.Parallel()

	readOnly := newToolExecutorForTest(t, &teamsServiceStub{}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})
	require.Len(t, readOnly.Definitions(), 4)

	mutationStories := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{}, mutationStories, testStoryMutationSecret)
	definitions := executor.Definitions()
	require.Len(t, definitions, 9)
	require.NoError(t, validateToolDefinitions(definitions))
	require.Equal(t, toolCreateStory, definitions[4].Name)
	require.Equal(t, toolCreateStories, definitions[5].Name)
	require.Equal(t, toolUpdateStory, definitions[6].Name)
	updateProperties, ok := definitions[6].Parameters["properties"].(map[string]any)
	require.True(t, ok)
	for _, property := range []string{"estimated_duration_action", "estimated_duration_minutes", "minimum_focus_block_action", "minimum_focus_block_minutes"} {
		require.Contains(t, definitions[6].Parameters["required"], property)
		require.Contains(t, updateProperties, property)
	}
	for _, property := range []string{"estimated_duration_action", "minimum_focus_block_action"} {
		actionSchema, ok := updateProperties[property].(map[string]any)
		require.True(t, ok)
		require.Equal(t, []string{storyTimeActionUnchanged, storyTimeActionSet, storyTimeActionClear}, actionSchema["enum"])
	}
}

func TestStoryMutationsRequireDurableConfirmationStore(t *testing.T) {
	t.Parallel()

	_, err := NewFortyOneToolExecutor(
		&teamsServiceStub{},
		newMutationStoriesStub(),
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithStoryMutations(testStoryMutationSecret),
	)

	require.ErrorContains(t, err, "confirmation store is required")
}

func TestStoryCreateRequiresExplicitConfirmationAndIsRetrySafe(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	teamsService := &teamsServiceStub{joined: []teams.CoreTeam{team}}
	executor := newMutationToolExecutorForTest(t, teamsService, storyService, testStoryMutationSecret)
	estimatedDurationMinutes := 180
	minimumFocusBlockMinutes := 45

	arguments, err := json.Marshal(map[string]any{
		"team_id":                     team.ID.String(),
		"title":                       "  Fix checkout retries  ",
		"priority":                    "High",
		"assignee":                    assigneeActionMe,
		"estimated_duration_minutes":  estimatedDurationMinutes,
		"minimum_focus_block_minutes": minimumFocusBlockMinutes,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.createCalls, "a model tool call must never write")

	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, StoryMutationCreate, confirmation.Operation)
	require.Equal(t, "Fix checkout retries", confirmation.Story.Title)
	require.Equal(t, "WEB", confirmation.Story.TeamCode)
	require.Equal(t, &estimatedDurationMinutes, confirmation.Story.EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, confirmation.Story.MinimumFocusBlockMinutes)
	require.NotEmpty(t, confirmation.Token)

	deniedScope := scope
	deniedScope.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.ConfirmStoryMutation(context.Background(), deniedScope, confirmation.Token)
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	require.Empty(t, storyService.createCalls)
	disabledScope := scope
	disabledScope.AllowMutations = false
	_, err = executor.ConfirmStoryMutation(context.Background(), disabledScope, confirmation.Token)
	require.ErrorIs(t, err, ErrMutationNotAllowed)
	require.Empty(t, storyService.createCalls)

	// Confirm on another executor to prove the signed token requires no process-
	// local state and works across worker replicas.
	secondExecutor := newMutationToolExecutorForTest(t, teamsService, storyService, testStoryMutationSecret)
	result, err := secondExecutor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)
	require.Equal(t, StoryMutationCreate, result.Operation)
	require.Equal(t, "WEB-1", result.Reference)
	require.Equal(t, scope.UserID, *result.AssigneeID)
	require.Equal(t, &estimatedDurationMinutes, result.EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, result.MinimumFocusBlockMinutes)
	require.Len(t, storyService.persisted, 1)
	require.Len(t, storyService.createCalls, 1)
	created := storyService.createCalls[0]
	require.Equal(t, "Fix checkout retries", created.story.Title)
	require.Equal(t, storyService.defaultStatusID, created.story.Status)
	require.NotNil(t, created.story.CreationKey)
	require.True(t, strings.HasPrefix(*created.story.CreationKey, "messaging:create-story:"))
	require.Equal(t, scope.UserID, created.actorID)
	require.Equal(t, scope.UserID, created.contextActorID)
	require.Equal(t, &estimatedDurationMinutes, created.story.EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, created.story.MinimumFocusBlockMinutes)

	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Equal(t, result.StoryID, retry.StoryID)
	require.Len(t, storyService.persisted, 1, "a retried confirmation must not duplicate the story")

	tampered := confirmation.Token[:len(confirmation.Token)-1] + "x"
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, tampered)
	require.ErrorIs(t, err, ErrInvalidConfirmation)

	wrongUser := scope
	wrongUser.UserID = uuid.New()
	_, err = executor.ConfirmStoryMutation(context.Background(), wrongUser, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
}

func TestStoryCreateRejectsFocusBlockWithoutDuration(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	arguments, err := json.Marshal(map[string]any{
		"team_id":                     team.ID.String(),
		"title":                       "Invalid focus contract",
		"priority":                    nil,
		"assignee":                    assigneeActionUnassigned,
		"minimum_focus_block_minutes": 30,
	})
	require.NoError(t, err)

	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.ErrorIs(t, err, ErrInvalidToolArguments)
	require.ErrorContains(t, err, stories.ErrFocusBlockRequiresDuration.Error())
	require.Empty(t, storyService.createCalls)
}

func TestStoryBatchUsesOpaquePersistedProposalAndCreatesItemizedRetrySafeResult(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	scope.SourceURL = "https://acme.slack.com/archives/C123/p1750000000000000"
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	assigneeID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000010")
	usersService := &usersServiceStub{items: []users.CoreUser{{
		ID:       assigneeID,
		FullName: "Ada Lovelace",
		Username: "ada",
		IsActive: true,
	}}}
	storyService := newMutationStoriesStub()
	executor := newBatchMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		usersService,
		testStoryMutationSecret,
	)
	estimatedDurationMinutes := 240
	minimumFocusBlockMinutes := 60

	arguments, err := json.Marshal(map[string]any{
		"team_id": team.ID.String(),
		"stories": []map[string]any{
			{
				"title":                       "Add Microsoft authentication",
				"description":                 "Support Microsoft as an authentication provider.",
				"priority":                    "High",
				"assignee_id":                 assigneeID.String(),
				"estimated_duration_minutes":  estimatedDurationMinutes,
				"minimum_focus_block_minutes": minimumFocusBlockMinutes,
			},
			{
				"title":       "Add TikTok integration",
				"description": "Add TikTok to the supported integrations.",
				"priority":    nil,
				"assignee_id": nil,
			},
		},
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStories, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.createCalls, "a batch proposal must not write")

	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, StoryMutationCreateBatch, confirmation.Operation)
	require.Len(t, confirmation.Stories, 2)
	require.NotContains(t, confirmation.Token, ".", "batch token should be short and opaque")
	require.Len(t, confirmation.Token, 43)
	require.NotContains(t, confirmation.Token, "Microsoft")
	require.Contains(t, confirmation.Prompt, "Create 2 stories")
	require.Contains(t, confirmation.Prompt, "link to this Slack thread")
	require.Equal(t, "Ada Lovelace", confirmation.Stories[0].AssigneeName)
	require.Equal(t, scope.SourceURL, confirmation.Stories[0].SourceURL)
	require.Equal(t, &estimatedDurationMinutes, confirmation.Stories[0].EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, confirmation.Stories[0].MinimumFocusBlockMinutes)
	require.Contains(t, confirmation.Stories[0].Description, scope.SourceURL)

	confirmationID, opaque, err := batchConfirmationID(confirmation.Token)
	require.NoError(t, err)
	require.True(t, opaque)
	entry := storyService.confirmations.entries[confirmationID]
	require.NotNil(t, entry)
	require.NotEmpty(t, entry.input.Proposal, "batch contents must be persisted server-side")
	require.NotContains(t, confirmation.Token, string(entry.input.Proposal))

	wrongActor := scope
	wrongActor.UserID = uuid.New()
	_, err = executor.ConfirmStoryMutation(context.Background(), wrongActor, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.Empty(t, storyService.createCalls)

	denied := scope
	denied.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.ConfirmStoryMutation(context.Background(), denied, confirmation.Token)
	require.ErrorIs(t, err, ErrTeamNotAccessible)
	require.Empty(t, storyService.createCalls)

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)
	require.Equal(t, StoryMutationCreateBatch, result.Operation)
	require.Equal(t, team.ID, result.TeamID)
	require.Len(t, result.Items, 2)
	require.Equal(t, "WEB-1", result.Items[0].Reference)
	require.Equal(t, "WEB-2", result.Items[1].Reference)
	require.Equal(t, assigneeID, *result.Items[0].AssigneeID)
	require.Equal(t, &estimatedDurationMinutes, result.Items[0].EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, result.Items[0].MinimumFocusBlockMinutes)
	require.Nil(t, result.Items[1].AssigneeID)
	require.Len(t, storyService.createCalls, 2)
	require.Contains(t, *storyService.createCalls[0].story.Description, scope.SourceURL)
	require.Contains(t, *storyService.createCalls[1].story.Description, scope.SourceURL)
	require.NotEqual(t, *storyService.createCalls[0].story.CreationKey, *storyService.createCalls[1].story.CreationKey)
	require.Empty(t, entry.input.Proposal, "a completed batch must redact its Slack-derived proposal")
	require.Empty(t, entry.lastError)

	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Equal(t, result.Items, retry.Items)
	require.Len(t, storyService.createCalls, 2, "a provider retry must not execute the batch again")
}

func TestStoryBatchPrevalidatesEveryNamedAssigneeBeforeFirstWrite(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	firstID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000021")
	secondID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000022")
	activeMembers := []users.CoreUser{
		{ID: firstID, FullName: "Ada", IsActive: true},
		{ID: secondID, FullName: "Grace", IsActive: true},
	}
	usersService := &usersServiceStub{items: append([]users.CoreUser(nil), activeMembers...)}
	storyService := newMutationStoriesStub()
	executor := newBatchMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, usersService, testStoryMutationSecret)
	arguments, err := json.Marshal(map[string]any{
		"team_id": team.ID.String(),
		"stories": []map[string]any{
			{"title": "First", "description": nil, "priority": nil, "assignee_id": firstID.String()},
			{"title": "Second", "description": nil, "priority": nil, "assignee_id": secondID.String()},
		},
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStories, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)

	usersService.items = activeMembers[:1]
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.Empty(t, storyService.createCalls, "all assignees must be revalidated before the first write")

	usersService.items = activeMembers
	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	require.Len(t, storyService.createCalls, 2)
}

func TestStoryBatchRetriesSafelyAfterPartialProviderFailure(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyService.failCreateAt = 2
	executor := newBatchMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, &usersServiceStub{}, testStoryMutationSecret)
	arguments, err := json.Marshal(map[string]any{
		"team_id": team.ID.String(),
		"stories": []map[string]any{
			{"title": "One", "description": nil, "priority": nil, "assignee_id": nil},
			{"title": "Two", "description": nil, "priority": nil, "assignee_id": nil},
			{"title": "Three", "description": nil, "priority": nil, "assignee_id": nil},
		},
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStories, Arguments: arguments})
	require.NoError(t, err)
	confirmation, _, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)

	partial, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorContains(t, err, "transient create failure")
	require.Equal(t, storyMutationStatusPartial, partial.Status)
	require.Equal(t, StoryMutationCreateBatch, partial.Operation)
	require.Len(t, partial.Items, 1)
	require.Equal(t, "WEB-1", partial.Items[0].Reference)
	require.Len(t, storyService.persisted, 1)
	confirmationID, opaque, decodeErr := batchConfirmationID(confirmation.Token)
	require.NoError(t, decodeErr)
	require.True(t, opaque)
	entry := storyService.confirmations.entries[confirmationID]
	require.NotNil(t, entry)
	require.NotEmpty(t, entry.input.Proposal, "retryable partial batches must retain their proposal")
	require.NotNil(t, entry.result)
	require.Equal(t, partial.Items, entry.result.Items)
	require.NotEmpty(t, entry.lastError)

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Len(t, result.Items, 3)
	require.Len(t, storyService.persisted, 3, "the already-created first item must be reused by creation key")
	require.Equal(t, []string{"WEB-1", "WEB-2", "WEB-3"}, []string{
		result.Items[0].Reference,
		result.Items[1].Reference,
		result.Items[2].Reference,
	})
	require.Empty(t, entry.input.Proposal, "successful retry must redact its proposal")
	require.Empty(t, entry.lastError)
}

func TestStoryBatchRedactsProposalWhenCancelledOrExpired(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	scope := testToolScope()
	scope.AllowMutations = true
	scope.SourceURL = "https://acme.slack.com/archives/C123/p1750000000000000"
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newBatchMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, &usersServiceStub{}, testStoryMutationSecret)
	executor.mutations.now = func() time.Time { return fixedNow }
	arguments, err := json.Marshal(map[string]any{
		"team_id": team.ID.String(),
		"stories": []map[string]any{{
			"title": "Sensitive Slack context", "description": "Copied from the thread", "priority": nil, "assignee_id": nil,
		}},
	})
	require.NoError(t, err)

	propose := func() (*StoryMutationConfirmation, uuid.UUID) {
		t.Helper()
		output, executeErr := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStories, Arguments: arguments})
		require.NoError(t, executeErr)
		confirmation, proposed, parseErr := mutationConfirmationFromToolResult(output)
		require.NoError(t, parseErr)
		require.True(t, proposed)
		confirmationID, opaque, decodeErr := batchConfirmationID(confirmation.Token)
		require.NoError(t, decodeErr)
		require.True(t, opaque)
		return confirmation, confirmationID
	}

	cancelled, cancelledID := propose()
	_, err = executor.CancelStoryMutation(context.Background(), scope, cancelled.Token)
	require.NoError(t, err)
	require.Equal(t, StoryMutationConfirmationCancelled, storyService.confirmations.entries[cancelledID].status)
	require.Empty(t, storyService.confirmations.entries[cancelledID].input.Proposal)

	expired, expiredID := propose()
	executor.mutations.now = func() time.Time { return fixedNow.Add(storyMutationConfirmationTTL) }
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, expired.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	require.Equal(t, StoryMutationConfirmationExpired, storyService.confirmations.entries[expiredID].status)
	require.Empty(t, storyService.confirmations.entries[expiredID].input.Proposal)
}

func TestStoryMutationCancellationIsOwnerBoundTerminalAndIdempotent(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Cancel me")

	wrongActor := scope
	wrongActor.UserID = uuid.New()
	_, err := executor.CancelStoryMutation(context.Background(), wrongActor, confirmation.Token)
	require.ErrorIs(t, err, ErrInvalidConfirmation)

	cancelled, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)
	retry, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, "already_cancelled", retry.Status)

	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrCancelledConfirmation)
	require.Empty(t, storyService.createCalls, "Cancel followed by Confirm must never write")
}

func TestStoryMutationConfirmationCannotLaterClaimCancellation(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Confirm me")

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)

	_, err = executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrAppliedConfirmation)
	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Len(t, storyService.createCalls, 1)
}

func TestStoryMutationConfirmCancelRaceHasExactlyOneTerminalOutcome(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)
	confirmation := proposeCreateMutationForTest(t, executor, scope, team.ID, "Race me")

	start := make(chan struct{})
	confirmErr := make(chan error, 1)
	cancelErr := make(chan error, 1)
	go func() {
		<-start
		_, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
		confirmErr <- err
	}()
	go func() {
		<-start
		_, err := executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
		cancelErr <- err
	}()
	close(start)

	confirmed := <-confirmErr
	cancelled := <-cancelErr
	switch {
	case confirmed == nil:
		require.ErrorIs(t, cancelled, ErrAppliedConfirmation)
		require.Len(t, storyService.createCalls, 1)
	case cancelled == nil:
		require.ErrorIs(t, confirmed, ErrCancelledConfirmation)
		require.Empty(t, storyService.createCalls)
	default:
		t.Fatalf("neither terminal outcome won: confirm=%v cancel=%v", confirmed, cancelled)
	}
}

func TestStoryUpdateRequiresConfirmationRejectsStaleWritesAndIsRetrySafe(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000001")
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID:         storyID,
		SequenceID: 42,
		Title:      "Old title",
		Priority:   "Low",
		Team:       team.ID,
		Workspace:  scope.WorkspaceID,
		UpdatedAt:  time.Date(2026, time.August, 9, 7, 0, 0, 123000000, time.UTC),
	}
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)
	estimatedDurationMinutes := 120
	minimumFocusBlockMinutes := 30

	arguments, err := json.Marshal(map[string]any{
		"story_id":                    nil,
		"story_reference":             "web-42",
		"title":                       "New title",
		"priority":                    "High",
		"assignee":                    assigneeActionMe,
		"estimated_duration_action":   storyTimeActionSet,
		"estimated_duration_minutes":  estimatedDurationMinutes,
		"minimum_focus_block_action":  storyTimeActionSet,
		"minimum_focus_block_minutes": minimumFocusBlockMinutes,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: arguments})
	require.NoError(t, err)
	require.Empty(t, storyService.updateCalls, "a model tool call must never write")
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, []string{"title", "priority", "assignee_id", "estimated_duration_minutes", "minimum_focus_block_minutes"}, confirmation.Story.ChangedFields)

	result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusApplied, result.Status)
	require.Equal(t, "WEB-42", result.Reference)
	require.Equal(t, "New title", result.Title)
	require.Equal(t, "High", result.Priority)
	require.Equal(t, scope.UserID, *result.AssigneeID)
	require.Equal(t, &estimatedDurationMinutes, result.EstimatedDurationMinutes)
	require.Equal(t, &minimumFocusBlockMinutes, result.MinimumFocusBlockMinutes)
	require.Len(t, storyService.updateCalls, 1)

	retry, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Equal(t, storyMutationStatusAlreadyApplied, retry.Status)
	require.Len(t, storyService.updateCalls, 1, "a retried update must be a no-op")

	staleArguments, err := json.Marshal(map[string]any{
		"story_id":                    storyID.String(),
		"story_reference":             nil,
		"title":                       nil,
		"priority":                    "Urgent",
		"assignee":                    assigneeActionUnchanged,
		"estimated_duration_action":   storyTimeActionUnchanged,
		"estimated_duration_minutes":  nil,
		"minimum_focus_block_action":  storyTimeActionUnchanged,
		"minimum_focus_block_minutes": nil,
	})
	require.NoError(t, err)
	staleOutput, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: staleArguments})
	require.NoError(t, err)
	staleConfirmation, proposed, err := mutationConfirmationFromToolResult(staleOutput)
	require.NoError(t, err)
	require.True(t, proposed)
	external := storyService.persisted[storyID]
	external.Title = "Changed elsewhere"
	external.UpdatedAt = external.UpdatedAt.Add(time.Second)
	storyService.persisted[storyID] = external

	_, err = executor.ConfirmStoryMutation(context.Background(), scope, staleConfirmation.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Len(t, storyService.updateCalls, 1, "stale confirmation must not write")

	raceArguments, err := json.Marshal(map[string]any{
		"story_id":                    storyID.String(),
		"story_reference":             nil,
		"title":                       nil,
		"priority":                    "Medium",
		"assignee":                    assigneeActionUnchanged,
		"estimated_duration_action":   storyTimeActionUnchanged,
		"estimated_duration_minutes":  nil,
		"minimum_focus_block_action":  storyTimeActionUnchanged,
		"minimum_focus_block_minutes": nil,
	})
	require.NoError(t, err)
	raceOutput, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: raceArguments})
	require.NoError(t, err)
	raceConfirmation, proposed, err := mutationConfirmationFromToolResult(raceOutput)
	require.NoError(t, err)
	require.True(t, proposed)
	storyService.forceConditionalConflict = true
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, raceConfirmation.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
	require.Len(t, storyService.updateCalls, 1, "a write-time version conflict must not update")
}

func TestStoryUpdateTimeActionsSetClearAndPreserveUnchanged(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000002")
	estimatedDurationMinutes := 120
	minimumFocusBlockMinutes := 30
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID:                       storyID,
		SequenceID:               43,
		Title:                    "Time-aware story",
		Priority:                 "Low",
		Team:                     team.ID,
		Workspace:                scope.WorkspaceID,
		EstimatedDurationMinutes: &estimatedDurationMinutes,
		MinimumFocusBlockMinutes: &minimumFocusBlockMinutes,
		UpdatedAt:                time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC),
	}
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	executeAndConfirm := func(arguments map[string]any) (*StoryMutationConfirmation, StoryMutationResult) {
		t.Helper()
		raw, err := json.Marshal(arguments)
		require.NoError(t, err)
		output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: raw})
		require.NoError(t, err)
		confirmation, proposed, err := mutationConfirmationFromToolResult(output)
		require.NoError(t, err)
		require.True(t, proposed)
		result, err := executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
		require.NoError(t, err)
		return confirmation, result
	}
	baseArguments := func() map[string]any {
		return map[string]any{
			"story_id":        storyID.String(),
			"story_reference": nil,
			"title":           nil,
			"priority":        nil,
			"assignee":        assigneeActionUnchanged,
		}
	}

	clearMinimum := baseArguments()
	clearMinimum["estimated_duration_action"] = storyTimeActionUnchanged
	clearMinimum["estimated_duration_minutes"] = nil
	clearMinimum["minimum_focus_block_action"] = storyTimeActionClear
	clearMinimum["minimum_focus_block_minutes"] = nil
	confirmation, result := executeAndConfirm(clearMinimum)
	require.Equal(t, []string{"minimum_focus_block_minutes"}, confirmation.Story.ChangedFields)
	require.Equal(t, &estimatedDurationMinutes, result.EstimatedDurationMinutes)
	require.Nil(t, result.MinimumFocusBlockMinutes)
	require.Equal(t, map[string]any{"minimum_focus_block_minutes": nil}, storyService.updateCalls[0].updates)

	replacementDurationMinutes := 90
	replacementMinimumMinutes := 20
	setTime := baseArguments()
	setTime["estimated_duration_action"] = storyTimeActionSet
	setTime["estimated_duration_minutes"] = replacementDurationMinutes
	setTime["minimum_focus_block_action"] = storyTimeActionSet
	setTime["minimum_focus_block_minutes"] = replacementMinimumMinutes
	_, result = executeAndConfirm(setTime)
	require.Equal(t, &replacementDurationMinutes, result.EstimatedDurationMinutes)
	require.Equal(t, &replacementMinimumMinutes, result.MinimumFocusBlockMinutes)
	require.Equal(t, map[string]any{
		"estimated_duration_minutes":  replacementDurationMinutes,
		"minimum_focus_block_minutes": replacementMinimumMinutes,
	}, storyService.updateCalls[1].updates)

	leaveTimeUnchanged := baseArguments()
	leaveTimeUnchanged["priority"] = "High"
	leaveTimeUnchanged["estimated_duration_action"] = storyTimeActionUnchanged
	leaveTimeUnchanged["estimated_duration_minutes"] = nil
	leaveTimeUnchanged["minimum_focus_block_action"] = storyTimeActionUnchanged
	leaveTimeUnchanged["minimum_focus_block_minutes"] = nil
	_, result = executeAndConfirm(leaveTimeUnchanged)
	require.Equal(t, &replacementDurationMinutes, result.EstimatedDurationMinutes)
	require.Equal(t, &replacementMinimumMinutes, result.MinimumFocusBlockMinutes)
	require.Equal(t, map[string]any{"priority": "High"}, storyService.updateCalls[2].updates)

	clearDuration := baseArguments()
	clearDuration["estimated_duration_action"] = storyTimeActionClear
	clearDuration["estimated_duration_minutes"] = nil
	clearDuration["minimum_focus_block_action"] = storyTimeActionUnchanged
	clearDuration["minimum_focus_block_minutes"] = nil
	confirmation, result = executeAndConfirm(clearDuration)
	require.Equal(t, []string{"estimated_duration_minutes", "minimum_focus_block_minutes"}, confirmation.Story.ChangedFields)
	require.Nil(t, result.EstimatedDurationMinutes)
	require.Nil(t, result.MinimumFocusBlockMinutes)
	require.Equal(t, map[string]any{
		"estimated_duration_minutes":  nil,
		"minimum_focus_block_minutes": nil,
	}, storyService.updateCalls[3].updates)
}

func TestStoryUpdateTimeActionsRejectAmbiguousOrInvalidValues(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000003")
	durationMinutes := 120
	minimumFocusBlockMinutes := 30
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID:                       storyID,
		SequenceID:               44,
		Title:                    "Validate time actions",
		Priority:                 "Low",
		Team:                     team.ID,
		Workspace:                scope.WorkspaceID,
		EstimatedDurationMinutes: &durationMinutes,
		MinimumFocusBlockMinutes: &minimumFocusBlockMinutes,
		UpdatedAt:                time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC),
	}
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)

	tests := []struct {
		name               string
		durationAction     string
		durationMinutes    any
		minimumAction      string
		minimumMinutes     any
		expectedErrorMatch string
	}{
		{name: "set duration requires value", durationAction: storyTimeActionSet, durationMinutes: nil, minimumAction: storyTimeActionUnchanged, minimumMinutes: nil, expectedErrorMatch: "estimated_duration_minutes is required"},
		{name: "unchanged duration rejects value", durationAction: storyTimeActionUnchanged, durationMinutes: 60, minimumAction: storyTimeActionUnchanged, minimumMinutes: nil, expectedErrorMatch: "must be null"},
		{name: "clear duration rejects value", durationAction: storyTimeActionClear, durationMinutes: 60, minimumAction: storyTimeActionUnchanged, minimumMinutes: nil, expectedErrorMatch: "must be null"},
		{name: "duration has upper bound", durationAction: storyTimeActionSet, durationMinutes: stories.MaximumEstimatedDurationMinutes + 1, minimumAction: storyTimeActionClear, minimumMinutes: nil, expectedErrorMatch: "must be between 1 and"},
		{name: "minimum has lower bound", durationAction: storyTimeActionUnchanged, durationMinutes: nil, minimumAction: storyTimeActionSet, minimumMinutes: 0, expectedErrorMatch: "must be between 1 and"},
		{name: "minimum cannot exceed resulting duration", durationAction: storyTimeActionSet, durationMinutes: 30, minimumAction: storyTimeActionSet, minimumMinutes: 60, expectedErrorMatch: "must not exceed estimated duration"},
		{name: "minimum cannot be set while duration is cleared", durationAction: storyTimeActionClear, durationMinutes: nil, minimumAction: storyTimeActionSet, minimumMinutes: 15, expectedErrorMatch: "require estimated duration"},
		{name: "action is required", durationAction: "", durationMinutes: nil, minimumAction: storyTimeActionUnchanged, minimumMinutes: nil, expectedErrorMatch: "action must be unchanged, set, or clear"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arguments, err := json.Marshal(map[string]any{
				"story_id":                    storyID.String(),
				"story_reference":             nil,
				"title":                       nil,
				"priority":                    nil,
				"assignee":                    assigneeActionUnchanged,
				"estimated_duration_action":   test.durationAction,
				"estimated_duration_minutes":  test.durationMinutes,
				"minimum_focus_block_action":  test.minimumAction,
				"minimum_focus_block_minutes": test.minimumMinutes,
			})
			require.NoError(t, err)
			_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: arguments})
			require.ErrorIs(t, err, ErrInvalidToolArguments)
			require.ErrorContains(t, err, test.expectedErrorMatch)
		})
	}
	require.Empty(t, storyService.updateCalls)
}

func TestStoryMutationProposalHonorsChannelScopeAndExpires(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	storyService := newMutationStoriesStub()
	executor := newMutationToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storyService, testStoryMutationSecret)
	fixedNow := time.Date(2026, time.August, 9, 8, 0, 0, 0, time.UTC)
	executor.mutations.now = func() time.Time { return fixedNow }

	arguments, err := json.Marshal(map[string]any{
		"team_id":  team.ID.String(),
		"title":    "Scoped story",
		"priority": nil,
		"assignee": assigneeActionUnassigned,
	})
	require.NoError(t, err)
	deniedScope := scope
	deniedScope.AllowedTeamIDs = []uuid.UUID{}
	_, err = executor.Execute(context.Background(), deniedScope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.ErrorIs(t, err, ErrTeamNotAccessible)

	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, fixedNow.Add(storyMutationConfirmationTTL), confirmation.ExpiresAt)

	executor.mutations.now = func() time.Time { return fixedNow.Add(storyMutationConfirmationTTL) }
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	require.Empty(t, storyService.createCalls)

	// Expiration is durable. Even if a skewed replica reports an earlier clock,
	// the proposal can no longer transition to either terminal user outcome.
	executor.mutations.now = func() time.Time { return fixedNow }
	_, err = executor.CancelStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.ErrorIs(t, err, ErrExpiredConfirmation)
	require.Empty(t, storyService.createCalls)
}

func TestStoryMutationTokenFitsSlackActionValueAtMaximumTitleLength(t *testing.T) {
	t.Parallel()

	executor := newMutationStoriesStub()
	mutations := newStoryMutationExecutor(executor, testStoryMutationSecret, executor.confirmations)
	title := strings.Repeat("😀", maximumStoryTitleRunes)
	priority := "Urgent"
	storyID := uuid.New()
	expectedUpdatedAt := time.Now().UTC()
	token, err := mutations.signClaims(storyMutationClaims{
		Version:                 storyMutationTokenVersion,
		ConfirmationID:          uuid.New(),
		Operation:               StoryMutationUpdate,
		WorkspaceID:             uuid.New(),
		UserID:                  uuid.New(),
		TeamID:                  uuid.New(),
		StoryID:                 &storyID,
		ExpectedUpdatedAt:       &expectedUpdatedAt,
		Title:                   &title,
		Priority:                &priority,
		AssigneeAction:          assigneeActionUnassigned,
		EstimatedDurationAction: storyTimeActionUnchanged,
		MinimumFocusBlockAction: storyTimeActionUnchanged,
		ExpiresAt:               time.Now().UTC().Add(storyMutationConfirmationTTL),
	})

	require.NoError(t, err)
	require.LessOrEqual(t, len(token), maximumStoryMutationTokenSize)
}

func TestStoryMutationTokenNormalizesLegacyUpdateTimeClaims(t *testing.T) {
	t.Parallel()

	executor := newMutationStoriesStub()
	mutations := newStoryMutationExecutor(executor, testStoryMutationSecret, executor.confirmations)
	priority := "High"
	storyID := uuid.New()
	expectedUpdatedAt := time.Now().UTC()
	durationMinutes := 75
	token, err := mutations.signClaims(storyMutationClaims{
		Version:                  storyMutationTokenVersion,
		ConfirmationID:           uuid.New(),
		Operation:                StoryMutationUpdate,
		WorkspaceID:              uuid.New(),
		UserID:                   uuid.New(),
		TeamID:                   uuid.New(),
		StoryID:                  &storyID,
		ExpectedUpdatedAt:        &expectedUpdatedAt,
		Priority:                 &priority,
		AssigneeAction:           assigneeActionUnchanged,
		EstimatedDurationMinutes: &durationMinutes,
		ExpiresAt:                time.Now().UTC().Add(storyMutationConfirmationTTL),
	})
	require.NoError(t, err)

	claims, err := mutations.verifyClaims(token)
	require.NoError(t, err)
	require.Equal(t, storyTimeActionSet, claims.EstimatedDurationAction)
	require.Equal(t, storyTimeActionUnchanged, claims.MinimumFocusBlockAction)
	require.Equal(t, &durationMinutes, claims.EstimatedDurationMinutes)
}

func newMutationToolExecutorForTest(
	t *testing.T,
	teamsService TeamsService,
	storiesService StoryMutationService,
	secret string,
) *FortyOneToolExecutor {
	t.Helper()
	mutationStories, ok := storiesService.(*mutationStoriesStub)
	require.True(t, ok, "mutation tests require the durable confirmation store stub")
	executor, err := NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithStoryMutations(secret),
		WithStoryMutationConfirmationStore(mutationStories.confirmations),
	)
	require.NoError(t, err)
	return executor
}

func newBatchMutationToolExecutorForTest(
	t *testing.T,
	teamsService TeamsService,
	storiesService *mutationStoriesStub,
	usersService UsersService,
	secret string,
) *FortyOneToolExecutor {
	t.Helper()
	executor, err := NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{},
			Users:  usersService,
		}),
		WithStoryMutations(secret),
		WithStoryMutationConfirmationStore(storiesService.confirmations),
	)
	require.NoError(t, err)
	return executor
}

func mutationTestTeam(workspaceID uuid.UUID) teams.CoreTeam {
	return teams.CoreTeam{
		ID:        uuid.MustParse("caaaaaaa-0000-0000-0000-000000000001"),
		Name:      "Web",
		Code:      "web",
		Workspace: workspaceID,
	}
}

func proposeCreateMutationForTest(
	t *testing.T,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	teamID uuid.UUID,
	title string,
) *StoryMutationConfirmation {
	t.Helper()
	arguments, err := json.Marshal(map[string]any{
		"team_id":  teamID.String(),
		"title":    title,
		"priority": nil,
		"assignee": assigneeActionUnassigned,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolCreateStory, Arguments: arguments})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	return confirmation
}

type mutationCreateCall struct {
	actorID        uuid.UUID
	contextActorID uuid.UUID
	story          stories.CoreNewStory
	workspaceID    uuid.UUID
}

type mutationUpdateCall struct {
	actorID     uuid.UUID
	storyID     uuid.UUID
	workspaceID uuid.UUID
	updates     map[string]any
}

type mutationStoriesStub struct {
	persisted                map[uuid.UUID]stories.CoreSingleStory
	byCreationKey            map[string]uuid.UUID
	defaultStatusID          *uuid.UUID
	forceConditionalConflict bool
	createCalls              []mutationCreateCall
	updateCalls              []mutationUpdateCall
	confirmations            *mutationConfirmationStoreStub
	failCreateAt             int
	failedCreate             bool
}

func newMutationStoriesStub() *mutationStoriesStub {
	statusID := uuid.MustParse("daaaaaaa-0000-0000-0000-000000000001")
	return &mutationStoriesStub{
		persisted:       make(map[uuid.UUID]stories.CoreSingleStory),
		byCreationKey:   make(map[string]uuid.UUID),
		defaultStatusID: &statusID,
		confirmations:   newMutationConfirmationStoreStub(),
	}
}

type mutationConfirmationStoreEntry struct {
	input     StoryMutationConfirmationStateInput
	status    StoryMutationConfirmationStatus
	result    *StoryMutationResult
	lastError string
}

type mutationConfirmationStoreStub struct {
	mu      sync.Mutex
	entries map[uuid.UUID]*mutationConfirmationStoreEntry
}

func newMutationConfirmationStoreStub() *mutationConfirmationStoreStub {
	return &mutationConfirmationStoreStub{entries: make(map[uuid.UUID]*mutationConfirmationStoreEntry)}
}

func (s *mutationConfirmationStoreStub) RegisterStoryMutationConfirmation(
	_ context.Context,
	input StoryMutationConfirmationStateInput,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[input.ConfirmationID]; exists {
		return errors.New("duplicate confirmation")
	}
	input.TokenHash = append([]byte(nil), input.TokenHash...)
	input.Proposal = append(json.RawMessage(nil), input.Proposal...)
	s.entries[input.ConfirmationID] = &mutationConfirmationStoreEntry{
		input:  input,
		status: StoryMutationConfirmationPending,
	}
	return nil
}

func (s *mutationConfirmationStoreStub) LoadStoryMutationConfirmation(
	_ context.Context,
	binding StoryMutationConfirmationBinding,
) (StoryMutationConfirmationRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationConfirmationRecord{}, err
	}
	return StoryMutationConfirmationRecord{
		TeamID:    entry.input.TeamID,
		Operation: entry.input.Operation,
		Status:    entry.status,
		Proposal:  append(json.RawMessage(nil), entry.input.Proposal...),
		Result:    entry.result,
		LastError: entry.lastError,
	}, nil
}

func (s *mutationConfirmationStoreStub) ApplyStoryMutationConfirmation(
	ctx context.Context,
	binding StoryMutationConfirmationBinding,
	now time.Time,
	apply func(context.Context) (StoryMutationResult, error),
) (StoryMutationResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationResult{}, false, err
	}
	if (entry.status == StoryMutationConfirmationPending ||
		(entry.status == StoryMutationConfirmationApplied && len(entry.input.Proposal) != 0)) &&
		!now.Before(entry.input.ExpiresAt) {
		entry.status = StoryMutationConfirmationExpired
		entry.input.Proposal = nil
	}
	switch entry.status {
	case StoryMutationConfirmationCancelled:
		return StoryMutationResult{}, false, ErrCancelledConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationResult{}, false, ErrExpiredConfirmation
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationApplied
	case StoryMutationConfirmationApplied:
		if entry.result != nil && entry.lastError == "" {
			return *entry.result, true, nil
		}
	}
	result, err := apply(ctx)
	if err != nil {
		if result.Operation == StoryMutationCreateBatch && (entry.result == nil || len(result.Items) >= len(entry.result.Items)) {
			stored := result
			entry.result = &stored
		}
		entry.lastError = err.Error()
		if entry.result != nil {
			return *entry.result, false, err
		}
		return result, false, err
	}
	stored := result
	entry.result = &stored
	entry.input.Proposal = nil
	entry.lastError = ""
	return result, false, nil
}

func (s *mutationConfirmationStoreStub) CancelStoryMutationConfirmation(
	_ context.Context,
	binding StoryMutationConfirmationBinding,
	now time.Time,
) (StoryMutationCancellationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, err := s.boundEntry(binding)
	if err != nil {
		return StoryMutationCancellationResult{}, err
	}
	if entry.status == StoryMutationConfirmationPending && !now.Before(entry.input.ExpiresAt) {
		entry.status = StoryMutationConfirmationExpired
		entry.input.Proposal = nil
	}
	switch entry.status {
	case StoryMutationConfirmationPending:
		entry.status = StoryMutationConfirmationCancelled
		entry.input.Proposal = nil
		return StoryMutationCancellationResult{Status: "cancelled"}, nil
	case StoryMutationConfirmationCancelled:
		return StoryMutationCancellationResult{Status: "already_cancelled"}, nil
	case StoryMutationConfirmationApplied:
		return StoryMutationCancellationResult{}, ErrAppliedConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationCancellationResult{}, ErrExpiredConfirmation
	default:
		return StoryMutationCancellationResult{}, ErrInvalidConfirmation
	}
}

func (s *mutationConfirmationStoreStub) boundEntry(
	binding StoryMutationConfirmationBinding,
) (*mutationConfirmationStoreEntry, error) {
	entry := s.entries[binding.ConfirmationID]
	if entry == nil || entry.input.WorkspaceID != binding.WorkspaceID || entry.input.UserID != binding.UserID || !bytes.Equal(entry.input.TokenHash, binding.TokenHash) {
		return nil, ErrInvalidConfirmation
	}
	return entry, nil
}

func (s *mutationStoriesStub) MyStories(context.Context, uuid.UUID) ([]stories.CoreStoryList, error) {
	return nil, nil
}

func (s *mutationStoriesStub) Get(_ context.Context, id, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	story, ok := s.persisted[id]
	if !ok || story.Workspace != workspaceID {
		return stories.CoreSingleStory{}, errors.New("story not found")
	}
	return story, nil
}

func (s *mutationStoriesStub) QueryByRef(_ context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error) {
	for _, story := range s.persisted {
		if story.Workspace == workspaceID && storyRef == "WEB-"+strconv.Itoa(story.SequenceID) {
			return story, nil
		}
	}
	return stories.CoreSingleStory{}, errors.New("story not found")
}

func (s *mutationStoriesStub) CreateCommentExternal(_ context.Context, _ uuid.UUID, _ uuid.UUID, comment stories.CoreNewComment) (comments.CoreComment, error) {
	return comments.CoreComment{ID: uuid.New(), StoryID: comment.StoryID, UserID: comment.UserID, Comment: comment.Comment}, nil
}

func (s *mutationStoriesStub) UpdateLabels(_ context.Context, storyID, workspaceID uuid.UUID, labels []uuid.UUID) error {
	story, ok := s.persisted[storyID]
	if !ok || story.Workspace != workspaceID {
		return errors.New("story not found")
	}
	story.Labels = append([]uuid.UUID(nil), labels...)
	s.persisted[storyID] = story
	return nil
}

func (s *mutationStoriesStub) AddAssociation(_ context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error) {
	from, fromOK := s.persisted[fromID]
	to, toOK := s.persisted[toID]
	if !fromOK || !toOK || from.Workspace != workspaceID || to.Workspace != workspaceID {
		return stories.CoreStoryAssociation{}, errors.New("story not found")
	}
	return stories.CoreStoryAssociation{ID: uuid.New(), FromStoryID: fromID, ToStoryID: toID, Type: associationType}, nil
}

func (s *mutationStoriesStub) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, story stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	contextActorID, _ := platformauth.GetUserID(ctx)
	s.createCalls = append(s.createCalls, mutationCreateCall{
		actorID:        actorID,
		contextActorID: contextActorID,
		story:          story,
		workspaceID:    workspaceID,
	})
	if s.failCreateAt > 0 && len(s.createCalls) == s.failCreateAt && !s.failedCreate {
		s.failedCreate = true
		return stories.CoreSingleStory{}, errors.New("transient create failure")
	}
	if story.CreationKey != nil {
		if existingID, ok := s.byCreationKey[*story.CreationKey]; ok {
			existing := s.persisted[existingID]
			existing.CreatedNow = false
			return existing, nil
		}
	}
	id := uuid.New()
	sequenceID := len(s.persisted) + 1
	created := stories.CoreSingleStory{
		ID:                       id,
		SequenceID:               sequenceID,
		Title:                    story.Title,
		Status:                   story.Status,
		Assignee:                 story.Assignee,
		Reporter:                 story.Reporter,
		Priority:                 story.Priority,
		Team:                     story.Team,
		Workspace:                workspaceID,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		UpdatedAt:                time.Date(2026, time.August, 9, 9, 0, 0, 0, time.UTC),
		CreatedNow:               true,
		CreationKey:              story.CreationKey,
	}
	s.persisted[id] = created
	if story.CreationKey != nil {
		s.byCreationKey[*story.CreationKey] = id
	}
	return created, nil
}

func (s *mutationStoriesStub) UpdateExternalUserActionIfUnchanged(
	_ context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	story, ok := s.persisted[storyID]
	if !ok || story.Workspace != workspaceID {
		return errors.New("story not found")
	}
	if !story.UpdatedAt.Equal(expectedUpdatedAt) {
		return stories.ErrStoryChanged
	}
	if s.forceConditionalConflict {
		return stories.ErrStoryChanged
	}
	copyUpdates := make(map[string]any, len(updates))
	for key, value := range updates {
		copyUpdates[key] = value
	}
	s.updateCalls = append(s.updateCalls, mutationUpdateCall{
		actorID:     actorID,
		storyID:     storyID,
		workspaceID: workspaceID,
		updates:     copyUpdates,
	})
	if value, ok := updates["title"].(string); ok {
		story.Title = value
	}
	if value, ok := updates["priority"].(string); ok {
		story.Priority = value
	}
	if value, exists := updates["assignee_id"]; exists {
		switch typed := value.(type) {
		case nil:
			story.Assignee = nil
		case uuid.UUID:
			valueCopy := typed
			story.Assignee = &valueCopy
		case *uuid.UUID:
			story.Assignee = typed
		}
	}
	if value, exists := updates["estimated_duration_minutes"]; exists {
		switch typed := value.(type) {
		case nil:
			story.EstimatedDurationMinutes = nil
		case int:
			story.EstimatedDurationMinutes = cloneIntPointer(&typed)
		case *int:
			story.EstimatedDurationMinutes = cloneIntPointer(typed)
		}
	}
	if value, exists := updates["minimum_focus_block_minutes"]; exists {
		switch typed := value.(type) {
		case nil:
			story.MinimumFocusBlockMinutes = nil
		case int:
			story.MinimumFocusBlockMinutes = cloneIntPointer(&typed)
		case *int:
			story.MinimumFocusBlockMinutes = cloneIntPointer(typed)
		}
	}
	story.UpdatedAt = story.UpdatedAt.Add(time.Second)
	s.persisted[storyID] = story
	return nil
}

func (s *mutationStoriesStub) FindFirstStatusByCategory(_ context.Context, _, _ uuid.UUID, category string) (*uuid.UUID, error) {
	if category != "unstarted" {
		return nil, errors.New("unexpected category")
	}
	return s.defaultStatusID, nil
}
