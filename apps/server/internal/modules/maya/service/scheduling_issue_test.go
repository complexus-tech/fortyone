package maya

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPartialSchedulingIssueIdentitySurvivesChangedPlacement(t *testing.T) {
	first := PlanResult{ScheduledMinutes: 60, RemainingMinutes: 30}
	next := PlanResult{ScheduledMinutes: 30, RemainingMinutes: 60}
	firstStatus, firstReason := autoSchedulingOutcome(first, nil)
	nextStatus, nextReason := autoSchedulingOutcome(next, nil)
	require.NotEqual(t, firstReason, nextReason)
	require.Equal(t, "no_available_slot", schedulingIssueCode(firstStatus, first))
	require.Equal(t, schedulingIssueCode(firstStatus, first), schedulingIssueCode(nextStatus, next))

	escalation := PlanResult{Actions: []CoreAction{{Payload: ActionPayload{Risk: &RiskPayload{Code: "deadline_passed"}}}}}
	status, _ := autoSchedulingOutcome(escalation, nil)
	require.Equal(t, "deadline_passed", schedulingIssueCode(status, escalation))
}

func TestPersistentFailureRetainsStructuredIssueFacts(t *testing.T) {
	ownerID := uuid.New()
	reason := "30 minutes left to schedule."
	story := Story{AutoSchedulingStatus: AutoSchedulingStatusCannotFit, AutoSchedulingReason: &reason}
	transition := buildStoryScheduleTransitionAt(story, ownerID, nil, nil, "UTC", AutoSchedulingStatusCannotFit, reason, time.Now())
	require.NotNil(t, transition, "the durable issue lifecycle must still see an unchanged failure")
	require.Equal(t, "no_available_slot", transition.IssueCode)
	require.Equal(t, events.StoryScheduleTransitionStateChanged, transition.Kind)

	previousOwner := uuid.New()
	transition = buildStoryScheduleTransitionAt(story, ownerID,
		[]ScheduleBlock{{UserID: previousOwner}}, nil, "UTC", AutoSchedulingStatusCannotFit, reason, time.Now())
	require.NotNil(t, transition, "a new calendar owner must receive their own unresolved issue")
}

func TestLockedSchedulingIssueCodeUsesCauseInsteadOfChangingAmounts(t *testing.T) {
	now := time.Now().UTC()
	minutes := 90
	story := Story{EstimatedDurationMinutes: &minutes}
	blocks := []ScheduleBlock{{StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour)}}
	firstCode, firstReason, atRisk := lockedScheduleIssue(story, blocks, "UTC", now)
	require.True(t, atRisk)
	require.Equal(t, "locked_duration_mismatch", firstCode)
	minutes = 120
	nextCode, nextReason, atRisk := lockedScheduleIssue(story, blocks, "UTC", now)
	require.True(t, atRisk)
	require.NotEqual(t, firstReason, nextReason)
	require.Equal(t, firstCode, nextCode)
	blocks[0].HasConflict = true
	changedCode, _, atRisk := lockedScheduleIssue(story, blocks, "UTC", now)
	require.True(t, atRisk)
	require.Equal(t, "locked_calendar_conflict", changedCode)
}
