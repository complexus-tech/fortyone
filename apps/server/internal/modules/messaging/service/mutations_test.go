package messaging

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
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
	createProperties, ok := definitions[4].Parameters["properties"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, createProperties, "auto_scheduling_enabled")
	require.Contains(t, definitions[4].Parameters["required"], "auto_scheduling_enabled")
	require.NotContains(t, createProperties, "auto_scheduling_locked")
	require.NotContains(t, definitions[4].Parameters["required"], "auto_scheduling_locked")

	batchProperties, ok := definitions[5].Parameters["properties"].(map[string]any)
	require.True(t, ok)
	storiesSchema, ok := batchProperties["stories"].(map[string]any)
	require.True(t, ok)
	batchItemSchema, ok := storiesSchema["items"].(map[string]any)
	require.True(t, ok)
	batchItemProperties, ok := batchItemSchema["properties"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, batchItemProperties, "auto_scheduling_enabled")
	require.Contains(t, batchItemSchema["required"], "auto_scheduling_enabled")
	require.NotContains(t, batchItemProperties, "auto_scheduling_locked")
	require.NotContains(t, batchItemSchema["required"], "auto_scheduling_locked")

	updateProperties, ok := definitions[6].Parameters["properties"].(map[string]any)
	require.True(t, ok)
	for _, property := range []string{"estimated_duration_action", "estimated_duration_minutes", "minimum_focus_block_action", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked"} {
		require.Contains(t, definitions[6].Parameters["required"], property)
		require.Contains(t, updateProperties, property)
	}
	for _, property := range []string{"estimated_duration_action", "minimum_focus_block_action"} {
		actionSchema, ok := updateProperties[property].(map[string]any)
		require.True(t, ok)
		require.Equal(t, []string{storyTimeActionUnchanged, storyTimeActionSet, storyTimeActionClear}, actionSchema["enum"])
	}
}

func TestDesiredStoryAutoSchedulingUpdatesEnforcesLockInvariant(t *testing.T) {
	t.Parallel()

	enabled := true
	locked := true
	updates, err := desiredStoryAutoSchedulingUpdates(stories.CoreSingleStory{
		AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
	}, storyAutoSchedulingMutation{
		enabled: &enabled,
		locked:  &locked,
	})
	require.Nil(t, updates)
	require.ErrorIs(t, err, ErrInvalidToolArguments)
	require.ErrorIs(t, err, stories.ErrAutoSchedulingLockEmpty)

	disabled := false
	updates, err = desiredStoryAutoSchedulingUpdates(stories.CoreSingleStory{
		AutoSchedulingEnabled: true,
		AutoSchedulingLocked:  true,
		AutoSchedulingStatus:  stories.AutoSchedulingStatusLocked,
	}, storyAutoSchedulingMutation{enabled: &disabled})
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"auto_scheduling_enabled": false,
		"auto_scheduling_locked":  false,
	}, updates)
}

func TestCreateStoryClaimsRejectAutoSchedulingLockInput(t *testing.T) {
	t.Parallel()

	title := "Create without a schedule lock"
	priority := "High"
	locked := false
	err := validateStoryMutationClaims(storyMutationClaims{
		Operation:            StoryMutationCreate,
		Title:                &title,
		Priority:             &priority,
		AssigneeAction:       assigneeActionUnassigned,
		AutoSchedulingLocked: &locked,
	})

	require.ErrorIs(t, err, ErrInvalidConfirmation)
	require.ErrorContains(t, err, "cannot set auto-scheduling lock state")
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
	require.Nil(t, confirmation.Story.AutoSchedulingLocked)
	require.NotEmpty(t, confirmation.Token)
	claims, err := executor.mutations.verifyClaims(confirmation.Token)
	require.NoError(t, err)
	require.Nil(t, claims.AutoSchedulingLocked)

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
	require.False(t, created.story.AutoSchedulingLocked)

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
	require.Nil(t, confirmation.Stories[0].AutoSchedulingLocked)
	require.Nil(t, confirmation.Stories[1].AutoSchedulingLocked)
	require.Contains(t, confirmation.Stories[0].Description, scope.SourceURL)

	confirmationID, opaque, err := batchConfirmationID(confirmation.Token)
	require.NoError(t, err)
	require.True(t, opaque)
	entry := storyService.confirmations.entries[confirmationID]
	require.NotNil(t, entry)
	require.NotEmpty(t, entry.input.Proposal, "batch contents must be persisted server-side")
	require.NotContains(t, string(entry.input.Proposal), "auto_scheduling_locked")
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
	require.False(t, storyService.createCalls[0].story.AutoSchedulingLocked)
	require.False(t, storyService.createCalls[1].story.AutoSchedulingLocked)
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
