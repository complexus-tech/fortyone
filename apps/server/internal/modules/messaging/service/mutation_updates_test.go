package messaging

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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

func TestStoryUpdateNullOptionalFieldsRemainUnchanged(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowMutations = true
	team := mutationTestTeam(scope.WorkspaceID)
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	storyService := newMutationStoriesStub()
	storyID := uuid.MustParse("baaaaaaa-0000-0000-0000-000000000099")
	currentStatusID := uuid.MustParse("daaaaaaa-0000-0000-0000-000000000010")
	completedStatusID := uuid.MustParse("daaaaaaa-0000-0000-0000-000000000011")
	sprintID := uuid.MustParse("eaaaaaaa-0000-0000-0000-000000000001")
	objectiveID := uuid.MustParse("faaaaaaa-0000-0000-0000-000000000001")
	startDate := time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	storyService.persisted[storyID] = stories.CoreSingleStory{
		ID:         storyID,
		SequenceID: 99,
		Title:      "Keep every unrelated field",
		Priority:   "Low",
		Status:     &currentStatusID,
		Sprint:     &sprintID,
		Objective:  &objectiveID,
		StartDate:  &startDate,
		EndDate:    &endDate,
		Team:       team.ID,
		Workspace:  scope.WorkspaceID,
		UpdatedAt:  time.Date(2026, time.August, 26, 8, 0, 0, 0, time.UTC),
	}
	executor := newMutationToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyService,
		testStoryMutationSecret,
	)

	arguments, err := json.Marshal(map[string]any{
		"story_id":                    storyID.String(),
		"story_reference":             nil,
		"title":                       nil,
		"priority":                    nil,
		"assignee":                    assigneeActionUnchanged,
		"status_id":                   completedStatusID.String(),
		"sprint_id":                   nil,
		"objective_id":                nil,
		"key_result_id":               nil,
		"start_date":                  nil,
		"end_date":                    nil,
		"label_ids":                   nil,
		"estimated_duration_action":   storyTimeActionUnchanged,
		"estimated_duration_minutes":  nil,
		"minimum_focus_block_action":  storyTimeActionUnchanged,
		"minimum_focus_block_minutes": nil,
		"auto_scheduling_enabled":     nil,
		"auto_scheduling_locked":      nil,
	})
	require.NoError(t, err)
	output, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolUpdateStory,
		Arguments: arguments,
	})
	require.NoError(t, err)
	confirmation, proposed, err := mutationConfirmationFromToolResult(output)
	require.NoError(t, err)
	require.True(t, proposed)
	require.Equal(t, []string{"status_id"}, confirmation.Story.ChangedFields)

	_, err = executor.ConfirmStoryMutation(context.Background(), scope, confirmation.Token)
	require.NoError(t, err)
	require.Len(t, storyService.updateCalls, 1)
	require.Equal(t, map[string]any{"status_id": completedStatusID}, storyService.updateCalls[0].updates)
	require.Equal(t, "Low", storyService.persisted[storyID].Priority)
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
