package notifications

import (
	"context"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type scheduleRulesStories struct {
	story      stories.CoreSingleStory
	activities []stories.CoreActivity
}

func (s *scheduleRulesStories) Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error) {
	return s.story, nil
}

func (s *scheduleRulesStories) RecordActivity(_ context.Context, activity stories.CoreActivity) error {
	s.activities = append(s.activities, activity)
	return nil
}

func TestMayaAssignmentNotificationsIncludeReason(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	oldAssigneeID := uuid.New()
	newAssigneeID := uuid.New()
	rules := NewRules(nil, &scheduleRulesStories{story: stories.CoreSingleStory{Title: "Ship calendar"}}, nil, nil)
	payload := events.StoryUpdatedPayload{
		StoryID:     uuid.New(),
		WorkspaceID: uuid.New(),
		Updates:     map[string]any{"assignee_id": newAssigneeID},
		AssigneeID:  &oldAssigneeID,
		Source:      events.StoryUpdateSourceMaya,
		Reason:      " Best frontend fit because <availability> is open. ",
	}

	notifications := rules.handleReassignment(context.Background(), payload, actorID)
	require.Len(t, notifications, 2)
	require.Equal(t, "Maya reassigned this task to {assignee}: {reason}", notifications[0].Message.Template)
	require.Equal(t, "Best frontend fit because ‹availability› is open.", notifications[0].Message.Variables["reason"].Value)
	require.Equal(t, "Maya assigned you this task: {reason}", notifications[1].Message.Template)
	require.Equal(t, notifications[0].Message.Variables["reason"], notifications[1].Message.Variables["reason"])
}

func TestNonMayaAssignmentDoesNotExposeReasonAsMayaReason(t *testing.T) {
	t.Parallel()

	newAssigneeID := uuid.New()
	rules := NewRules(nil, &scheduleRulesStories{}, nil, nil)
	notifications := rules.handleNewAssignment(context.Background(), events.StoryUpdatedPayload{
		StoryID: uuid.New(), WorkspaceID: uuid.New(), Updates: map[string]any{"assignee_id": newAssigneeID},
		Reason: "provider detail that is not a Maya recommendation",
	}, uuid.New())

	require.Len(t, notifications, 1)
	require.Equal(t, "{actor} assigned you a task", notifications[0].Message.Template)
	require.NotContains(t, notifications[0].Message.Variables, "reason")
}

func TestMeaningfulScheduleTransitionThresholds(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		transition events.StoryScheduleTransition
		notify     bool
		record     bool
	}{
		{name: "first schedule", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionFirstSchedule, State: events.StoryScheduleStateScheduled}, notify: true, record: true},
		{name: "day change", transition: events.StoryScheduleTransition{State: events.StoryScheduleStateScheduled, PreviousLocalDate: "2026-08-17", LocalDate: "2026-08-18"}, notify: true, record: true},
		{name: "sixty minute move", transition: events.StoryScheduleTransition{State: events.StoryScheduleStateScheduled, PreviousStartAt: &start, StartAt: timePointer(start.Add(time.Hour))}, notify: true, record: true},
		{name: "small same day move", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionMoved, State: events.StoryScheduleStateScheduled, PreviousStartAt: &start, StartAt: timePointer(start.Add(59 * time.Minute)), PreviousLocalDate: "2026-08-17", LocalDate: "2026-08-17"}},
		{name: "needs time", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionStateChanged, State: events.StoryScheduleStateNeedsTime}, notify: true, record: true},
		{name: "at risk", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionStateChanged, State: events.StoryScheduleStateAtRisk}, notify: true, record: true},
		{name: "cannot fit", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionStateChanged, State: events.StoryScheduleStateCannotFit}, notify: true, record: true},
		{name: "planning", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionStateChanged, State: events.StoryScheduleStatePlanning}},
		{name: "locked uses user activity", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionLocked, State: events.StoryScheduleStateLocked}},
		{name: "unlocked uses user activity", transition: events.StoryScheduleTransition{Kind: events.StoryScheduleTransitionUnlocked, State: events.StoryScheduleStateScheduled}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.notify, shouldNotifyScheduleTransition(&test.transition))
			require.Equal(t, test.record, shouldRecordScheduleTransition(&test.transition))
		})
	}
}

func TestScheduleTransitionNotificationAndActivityPreserveReason(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	assigneeID := uuid.New()
	watcherID := uuid.New()
	start := time.Date(2026, time.August, 18, 7, 30, 0, 0, time.UTC)
	storyService := &scheduleRulesStories{story: stories.CoreSingleStory{Title: "Review campaign"}}
	rules := NewRules(nil, storyService, nil, nil)
	payload := events.StoryUpdatedPayload{
		StoryID: uuid.New(), WorkspaceID: uuid.New(), AssigneeID: &assigneeID,
		AudienceIDs: []uuid.UUID{assigneeID, watcherID}, AudienceResolved: true,
		Source: events.StoryUpdateSourceMaya, Reason: "A new busy event displaced the original block.",
		Schedule: &events.StoryScheduleTransition{
			Kind: events.StoryScheduleTransitionDayChanged, UserID: assigneeID,
			PreviousState: events.StoryScheduleStateScheduled, State: events.StoryScheduleStateScheduled,
			StartAt: &start, Timezone: "Africa/Harare", PreviousLocalDate: "2026-08-17", LocalDate: "2026-08-18",
		},
	}

	notifications := rules.handleScheduleTransition(context.Background(), payload, actorID, nil)
	require.Len(t, notifications, 1)
	require.Equal(t, assigneeID, notifications[0].RecipientID, "calendar movement must notify only the affected calendar owner")
	require.Equal(t, "moved this task to {scheduled_for}: {reason}", notifications[0].Message.Template)
	require.Equal(t, "18 Aug 2026 at 09:30", notifications[0].Message.Variables["scheduled_for"].Value)
	require.Equal(t, payload.Reason, notifications[0].Message.Variables["reason"].Value)

	eventTimestamp := time.Date(2026, time.August, 18, 7, 0, 0, 0, time.UTC)
	require.NoError(t, rules.RecordScheduleTransitionActivity(context.Background(), payload, actorID, eventTimestamp))
	require.Len(t, storyService.activities, 1)
	activity := storyService.activities[0]
	require.NotEqual(t, uuid.Nil, activity.ID)
	require.Equal(t, "auto_scheduling_time", activity.Field)
	require.Equal(t, "18 Aug 2026 at 09:30", activity.CurrentValue)
	require.NotNil(t, activity.Reason)
	require.Equal(t, payload.Reason, *activity.Reason)
}

func TestSmallSameDayScheduleMoveDoesNotCreateActivity(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 18, 9, 0, 0, 0, time.UTC)
	storyService := &scheduleRulesStories{}
	rules := NewRules(nil, storyService, nil, nil)
	payload := events.StoryUpdatedPayload{
		StoryID: uuid.New(), WorkspaceID: uuid.New(), Source: events.StoryUpdateSourceMaya,
		Schedule: &events.StoryScheduleTransition{
			Kind: events.StoryScheduleTransitionMoved, State: events.StoryScheduleStateScheduled,
			PreviousStartAt: &start, StartAt: timePointer(start.Add(30 * time.Minute)),
			PreviousLocalDate: "2026-08-18", LocalDate: "2026-08-18",
		},
	}

	require.NoError(t, rules.RecordScheduleTransitionActivity(context.Background(), payload, uuid.New(), time.Now()))
	require.Empty(t, storyService.activities)
}

func TestScheduleTransitionActivityIDIsStableAcrossEventRedelivery(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	userID := uuid.New()
	payload := events.StoryUpdatedPayload{
		StoryID: uuid.New(), WorkspaceID: uuid.New(),
		Schedule: &events.StoryScheduleTransition{
			Kind:   events.StoryScheduleTransitionStateChanged,
			UserID: userID,
			State:  events.StoryScheduleStateAtRisk,
		},
	}
	eventTimestamp := time.Date(2026, time.August, 18, 7, 0, 0, 123, time.UTC)

	first := scheduleTransitionActivityID(payload, actorID, eventTimestamp)
	retry := scheduleTransitionActivityID(payload, actorID, eventTimestamp)
	nextEvent := scheduleTransitionActivityID(payload, actorID, eventTimestamp.Add(time.Nanosecond))

	require.NotEqual(t, uuid.Nil, first)
	require.Equal(t, first, retry)
	require.NotEqual(t, first, nextEvent)
}

func TestFirstScheduleActivityShowsReservedTimeInsteadOfInternalState(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 18, 9, 0, 0, 0, time.UTC)
	field, currentValue, oldValue, newValue := scheduleTransitionActivityValues(&events.StoryScheduleTransition{
		Kind:          events.StoryScheduleTransitionFirstSchedule,
		PreviousState: events.StoryScheduleStatePlanning,
		State:         events.StoryScheduleStateScheduled,
		StartAt:       &start,
		Timezone:      "Africa/Harare",
	}, "Africa/Harare")

	require.Equal(t, "auto_scheduling_time", field)
	require.Equal(t, "18 Aug 2026 at 11:00", currentValue)
	require.Nil(t, oldValue)
	require.Equal(t, &start, newValue)
}

func TestRepeatedRiskWithNewReasonRecordsStatusActivity(t *testing.T) {
	t.Parallel()

	field, currentValue, oldValue, newValue := scheduleTransitionActivityValues(&events.StoryScheduleTransition{
		Kind:          events.StoryScheduleTransitionStateChanged,
		PreviousState: events.StoryScheduleStateAtRisk,
		State:         events.StoryScheduleStateAtRisk,
	}, "UTC")

	require.Equal(t, "auto_scheduling_status", field)
	require.Equal(t, "At risk", currentValue)
	require.Equal(t, events.StoryScheduleStateAtRisk, oldValue)
	require.Equal(t, events.StoryScheduleStateAtRisk, newValue)
}

func TestScheduleTransitionDisplayUsesUserTimezone(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 20, 7, 45, 0, 0, time.UTC)
	transition := &events.StoryScheduleTransition{
		Kind:     events.StoryScheduleTransitionMoved,
		StartAt:  &start,
		Timezone: "UTC",
	}

	field, currentValue, _, _ := scheduleTransitionActivityValues(transition, "Africa/Harare")

	require.Equal(t, "auto_scheduling_time", field)
	require.Equal(t, "20 Aug 2026 at 09:45", currentValue)
}

func timePointer(value time.Time) *time.Time {
	return &value
}
