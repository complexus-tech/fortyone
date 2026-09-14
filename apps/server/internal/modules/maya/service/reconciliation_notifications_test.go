package maya

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReconciliationKeepsMeaningfulChangesWhileSkippingExpiredRollovers(t *testing.T) {
	asOf := time.Date(2026, time.September, 14, 13, 40, 0, 374_000_000, time.UTC)
	start := asOf.Truncate(time.Minute)
	userID := uuid.New()
	previous := []ScheduleBlock{{UserID: userID, StartAt: start.Add(-time.Hour), EndAt: start}}
	next := []ScheduleSegmentInput{{StartAt: start, EndAt: start.Add(time.Hour)}}
	story := Story{AutoSchedulingStatus: AutoSchedulingStatusScheduled}
	build := func(previous []ScheduleBlock, next []ScheduleSegmentInput, status string) *events.StoryScheduleTransition {
		return buildStoryScheduleTransitionAt(story, userID, previous, next, "Africa/Harare", status, "", asOf)
	}
	require.Nil(t, build(previous, next, AutoSchedulingStatusScheduled), "expired same-day rollover stays quiet")
	// Explicit schedule actions use the unfiltered builder.
	require.NotNil(t, buildStoryScheduleTransition(story, userID, previous, next, "Africa/Harare", AutoSchedulingStatusScheduled, ""))

	futurePrevious := []ScheduleBlock{{UserID: userID, StartAt: start.Add(time.Hour), EndAt: start.Add(2 * time.Hour)}}
	futureNext := []ScheduleSegmentInput{{StartAt: start.Add(2 * time.Hour), EndAt: start.Add(3 * time.Hour)}}
	transition := build(futurePrevious, futureNext, AutoSchedulingStatusScheduled)
	require.NotNil(t, transition, "displacing a future reservation stays visible")
	require.Equal(t, events.StoryScheduleTransitionMoved, transition.Kind)

	tomorrow := []ScheduleSegmentInput{{StartAt: start.Add(24 * time.Hour), EndAt: start.Add(25 * time.Hour)}}
	transition = build(previous, tomorrow, AutoSchedulingStatusScheduled)
	require.NotNil(t, transition)
	require.Equal(t, events.StoryScheduleTransitionDayChanged, transition.Kind)
	require.NotNil(t, build(previous, next, AutoSchedulingStatusAtRisk), "new risk stays visible")
	story.AutoSchedulingStatus = AutoSchedulingStatusAtRisk
	require.NotNil(t, build(previous, next, AutoSchedulingStatusScheduled), "risk recovery stays visible")
}

func TestReconciliationSelectsFutureMoveDespiteLargerExpiredRollover(t *testing.T) {
	asOf := time.Date(2026, time.September, 14, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	previous := []ScheduleBlock{
		{UserID: userID, SegmentIndex: 0, StartAt: asOf.Add(-3 * time.Hour), EndAt: asOf.Add(-2 * time.Hour)},
		{UserID: userID, SegmentIndex: 1, StartAt: asOf.Add(4 * time.Hour), EndAt: asOf.Add(5 * time.Hour)},
	}
	next := []ScheduleSegmentInput{
		{SegmentIndex: 0, StartAt: asOf.Add(2 * time.Hour), EndAt: asOf.Add(3 * time.Hour)},
		{SegmentIndex: 1, StartAt: asOf.Add(6 * time.Hour), EndAt: asOf.Add(7 * time.Hour)},
	}
	transition := buildStoryScheduleTransitionAt(
		Story{AutoSchedulingStatus: AutoSchedulingStatusScheduled}, userID, previous, next,
		"UTC", AutoSchedulingStatusScheduled, "", asOf,
	)
	require.NotNil(t, transition)
	require.Equal(t, 120, transition.ShiftMinutes)
	require.Equal(t, previous[1].StartAt, *transition.PreviousStartAt)
	require.Equal(t, next[1].StartAt, *transition.StartAt)
}
