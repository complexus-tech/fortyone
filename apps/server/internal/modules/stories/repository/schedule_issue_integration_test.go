//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestScheduleIssueSurvivesReplanningAndReopensAfterResolution(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	now := time.Now().UTC()
	storyID := createSecondaryMutationStory(t, ctx, repository, fixture, now)
	version := secondaryStoryUpdatedAt(t, ctx, postgres, storyID)
	write := func(status, reason string, ownerID uuid.UUID, code string, stateOnly bool) uuid.UUID {
		t.Helper()
		now = now.Add(time.Minute)
		input := scheduleIssueInput(t, fixture, storyID, ownerID, status, code, now)
		input.StateOnly = stateOnly
		applied, claim, err := repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
			ctx, storyID, fixture.workspaceID, version, status, &reason, now, nil, input,
		)
		require.NoError(t, err)
		require.True(t, applied)
		require.Nil(t, claim)
		return input.EventID
	}
	readIssue := func() (uuid.UUID, *time.Time) {
		t.Helper()
		var issueID uuid.UUID
		var resolvedAt *time.Time
		require.NoError(t, postgres.Pool.QueryRow(ctx,
			"SELECT issue_id, resolved_at FROM story_schedule_issues WHERE workspace_id = $1 AND story_id = $2",
			fixture.workspaceID, storyID,
		).Scan(&issueID, &resolvedAt))
		return issueID, resolvedAt
	}

	write("cannot_fit", "30 minutes left to schedule.", fixture.assigneeID, "no_available_slot", false)
	firstID, resolvedAt := readIssue()
	require.Nil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 1)

	// A user edit resets the story's state and version, but the outstanding
	// issue remains the same even when the remaining duration and prose change.
	_, err := postgres.Pool.Exec(ctx,
		"UPDATE stories SET auto_scheduling_status = 'planning', updated_at = updated_at + INTERVAL '1 second' WHERE id = $1",
		storyID,
	)
	require.NoError(t, err)
	version = secondaryStoryUpdatedAt(t, ctx, postgres, storyID)
	write("cannot_fit", "45 minutes left to schedule.", fixture.assigneeID, "no_available_slot", false)
	issueID, resolvedAt := readIssue()
	require.Equal(t, firstID, issueID)
	require.Nil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 1)
	reason := "45 minutes left to schedule."
	assertScheduleTransitionState(t, ctx, postgres, storyID, "cannot_fit", &reason)

	// Deadline escalation and an affected-owner change are actionable new
	// conditions, even while the story has never reached a healthy schedule.
	write("at_risk", "The deadline has passed.", fixture.assigneeID, "deadline_passed", false)
	escalatedID, _ := readIssue()
	require.NotEqual(t, firstID, escalatedID)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 2)
	write("at_risk", "The deadline has passed.", fixture.actorID, "deadline_passed", false)
	ownerIssueID, _ := readIssue()
	require.NotEqual(t, escalatedID, ownerIssueID)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 3)

	write("scheduled", "Capacity found.", fixture.actorID, "", false)
	_, resolvedAt = readIssue()
	require.NotNil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 4)
	write("at_risk", "The deadline has passed again.", fixture.actorID, "deadline_passed", false)
	reopenedID, resolvedAt := readIssue()
	require.NotEqual(t, ownerIssueID, reopenedID)
	require.Nil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 5)

	// Pausing can resolve a problem without any calendar movement to announce.
	write("off", "", fixture.actorID, "", true)
	_, resolvedAt = readIssue()
	require.NotNil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 5)
	write("at_risk", "The deadline has passed.", fixture.actorID, "deadline_passed", false)
	activeID, _ := readIssue()
	require.NotEqual(t, reopenedID, activeID)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 6)

	// An outbox failure must roll back the new episode and scheduling state.
	failed := scheduleIssueInput(t, fixture, storyID, fixture.actorID, "needs_time", "missing_duration", now.Add(time.Minute))
	var existingEventID uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx,
		"SELECT schedule_transition_event_id FROM story_schedule_transition_outbox WHERE story_id = $1 ORDER BY transition_sequence LIMIT 1",
		storyID,
	).Scan(&existingEventID))
	failed.EventID = existingEventID
	_, _, err = repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx, storyID, fixture.workspaceID, version, "needs_time", nil, now.Add(time.Minute), nil, failed,
	)
	require.Error(t, err)
	unchangedID, resolvedAt := readIssue()
	require.Equal(t, activeID, unchangedID)
	require.Nil(t, resolvedAt)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 6)

	// Independent reconciliation workers must open only one episode even
	// though each proposes a different event, fingerprint, and episode ID.
	inputs := []storydomain.ScheduleTransitionOutboxInput{
		scheduleIssueInput(t, fixture, storyID, fixture.actorID, "at_risk", "locked_calendar_conflict", now.Add(2*time.Minute)),
		scheduleIssueInput(t, fixture, storyID, fixture.actorID, "at_risk", "locked_calendar_conflict", now.Add(2*time.Minute)),
	}
	var workers sync.WaitGroup
	failures := make(chan error, len(inputs))
	for _, input := range inputs {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, _, err := repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
				ctx, storyID, fixture.workspaceID, version, "at_risk", nil, now.Add(2*time.Minute), nil, input,
			)
			failures <- err
		}()
	}
	workers.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 7)
}

func scheduleIssueInput(t *testing.T, fixture storyMutationFixture, storyID, ownerID uuid.UUID, status, code string, observedAt time.Time) storydomain.ScheduleTransitionOutboxInput {
	t.Helper()
	transition := &events.StoryScheduleTransition{
		Kind: events.StoryScheduleTransitionStateChanged, State: events.StoryScheduleState(status), UserID: ownerID,
	}
	var issue *storydomain.ScheduleIssue
	if code != "" {
		transition.IssueID, transition.IssueCode = uuid.New(), code
		issue = &storydomain.ScheduleIssue{ID: transition.IssueID, OwnerID: ownerID, Code: code}
	}
	event := events.Event{
		Type: events.StoryUpdated, ActorID: fixture.actorID, Timestamp: observedAt,
		Payload: events.StoryUpdatedPayload{StoryID: storyID, WorkspaceID: fixture.workspaceID, Schedule: transition},
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)
	return storydomain.ScheduleTransitionOutboxInput{
		EventID: uuid.New(), StoryID: storyID, WorkspaceID: fixture.workspaceID, ActorID: fixture.actorID,
		SemanticFingerprint: uuid.NewString(), EventPayload: payload, Issue: issue,
	}
}
