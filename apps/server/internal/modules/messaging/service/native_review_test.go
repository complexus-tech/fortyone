package messaging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNativeReviewIncludesEveryChangedFieldAndKeepsOriginalSnapshot(t *testing.T) {
	t.Parallel()
	executor, scope, backend, members, statuses := nativeFixture(t)
	story := backend.persisted[scope.SourceScope.StoryID]
	story.Status = &statuses.items[0].ID
	story.AutoSchedulingEnabled = true
	story.AutoSchedulingLocked = true
	story.AutoSchedulingStatus = storyAutoSchedulingStatusScheduled
	backend.persisted[story.ID] = story
	proposal := nativeProposal(t, executor, scope, map[string]any{
		"title": "Updated title", "description": "Complete rewritten description\nwith a second line.", "priority": "High",
		"assignee": "named", "assignee_id": members.items[0].ID.String(), "status_action": "close",
		"start_date": "2026-10-19", "end_date": "2026-10-20",
		"estimated_duration_action": "set", "estimated_duration_minutes": 120,
		"minimum_focus_block_action": "set", "minimum_focus_block_minutes": 30,
		"auto_scheduling_enabled": false, "command_intent": "proposal",
	})
	inspection, err := executor.InspectStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.False(t, inspection.DirectEligible)
	require.Len(t, inspection.ChangedFields, 11)
	for _, expected := range []string{
		"Title\nBefore:\nCurrent story\nAfter:\nUpdated title",
		"Description\nBefore:\nOld prose\nAfter:\nComplete rewritten description\nwith a second line.",
		"Priority\nBefore:\nLow\nAfter:\nHigh",
		"Assignee\nBefore:\nNot set\nAfter:\nAlex Example",
		"Status\nBefore:\nReady\nAfter:\nDone",
		"Start date\nBefore:\nNot set\nAfter:\n2026-10-19T",
		"Due date\nBefore:\nNot set\nAfter:\n2026-10-20T",
		"Time needed\nBefore:\nNot set\nAfter:\n120 minutes",
		"Minimum focus block\nBefore:\nNot set\nAfter:\n30 minutes",
		"Auto-scheduling\nBefore:\nYes\nAfter:\nNo",
		"Schedule locked\nBefore:\nYes\nAfter:\nNo",
	} {
		require.Contains(t, inspection.Review, expected)
	}
	for _, id := range []uuid.UUID{story.ID, members.items[0].ID, statuses.items[0].ID, statuses.items[1].ID} {
		require.NotContains(t, inspection.Review, id.String())
	}
	story.Title = "Later human title"
	story.UpdatedAt = story.UpdatedAt.Add(time.Second)
	backend.persisted[story.ID] = story
	stale, err := executor.InspectStoryMutation(context.Background(), scope, proposal.Token)
	require.NoError(t, err)
	require.Equal(t, inspection.Review, stale.Review)
	_, err = executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
	require.ErrorIs(t, err, ErrStaleMutation)
}

func TestNativeReviewRejectsUnnamedReferencesAndUnsupportedTargets(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"sprint_id", "objective_id", "key_result_id", "label_ids"} {
		t.Run(field, func(t *testing.T) {
			executor, scope, backend, _, _ := nativeFixture(t)
			var value any = uuid.New().String()
			if field == "label_ids" {
				value = []string{}
			}
			_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"title": "Changed", field: value})})
			require.ErrorIs(t, err, ErrInvalidToolArguments)
			require.Empty(t, backend.updateCalls)
			for _, definition := range sourceToolDefinitions(executor.Definitions(), scope.SourceScope) {
				if definition.Name == toolUpdateStory {
					require.NotContains(t, definition.Parameters["properties"], field)
				}
			}
			proposal := nativeProposal(t, executor, scope, map[string]any{"title": "Reviewable title"})
			id, _, err := batchConfirmationID(proposal.Token)
			require.NoError(t, err)
			entry := backend.confirmations.entries[id]
			var claims storyMutationClaims
			require.NoError(t, json.Unmarshal(entry.input.Proposal, &claims))
			target := uuid.New()
			switch field {
			case "sprint_id":
				claims.SprintID = &target
			case "objective_id":
				claims.ObjectiveID = &target
			case "key_result_id":
				claims.KeyResultID = &target
			case "label_ids":
				claims.LabelIDs = []uuid.UUID{target}
			}
			entry.input.Proposal, err = json.Marshal(claims)
			require.NoError(t, err)
			_, err = executor.ConfirmStoryMutation(context.Background(), scope, proposal.Token)
			require.ErrorIs(t, err, ErrInvalidConfirmation, "execution must reject unsupported claims independently of the tool schema")
			require.Empty(t, backend.updateCalls)
		})
	}
	executor, scope, backend, members, statuses := nativeFixture(t)
	members.items[0].FullName = ""
	_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"assignee": "named", "assignee_id": members.items[0].ID.String()})})
	require.ErrorIs(t, err, ErrInvalidToolArguments)
	statuses.items[1].Name = " "
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolUpdateStory, Arguments: nativeArguments(t, scope, map[string]any{"status_action": "close"})})
	require.ErrorIs(t, err, ErrInvalidToolArguments)
	require.Empty(t, backend.updateCalls)
}
