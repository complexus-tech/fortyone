package maya

import (
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func TestBuildStoryScheduleTransitionUsesMostSignificantChangedSegment(t *testing.T) {
	userID := uuid.New()
	firstStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	secondStart := time.Date(2026, 8, 17, 13, 0, 0, 0, time.UTC)
	story := stories.CoreSingleStory{
		AutoSchedulingStatus: stories.AutoSchedulingStatusScheduled,
	}
	previousBlocks := []calendar.CoreScheduleBlock{
		{UserID: userID, SegmentIndex: 0, StartAt: firstStart, EndAt: firstStart.Add(time.Hour)},
		{UserID: userID, SegmentIndex: 1, StartAt: secondStart, EndAt: secondStart.Add(time.Hour)},
	}

	t.Run("later segment changes day while first segment stays fixed", func(t *testing.T) {
		movedStart := secondStart.Add(26 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			previousBlocks,
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: firstStart, EndAt: firstStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: movedStart, EndAt: movedStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionDayChanged {
			t.Fatalf("expected later segment day change, got %#v", transition)
		}
		if transition.PreviousStartAt == nil || !transition.PreviousStartAt.Equal(secondStart) ||
			transition.StartAt == nil || !transition.StartAt.Equal(movedStart) ||
			transition.ShiftMinutes != 26*60 {
			t.Fatalf("transition should describe the changed later segment: %#v", transition)
		}
	})

	t.Run("largest same-day shift wins", func(t *testing.T) {
		firstMovedStart := firstStart.Add(time.Hour)
		secondMovedStart := secondStart.Add(3 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			previousBlocks,
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: firstMovedStart, EndAt: firstMovedStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: secondMovedStart, EndAt: secondMovedStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionMoved {
			t.Fatalf("expected a same-day move, got %#v", transition)
		}
		if transition.PreviousStartAt == nil || !transition.PreviousStartAt.Equal(secondStart) ||
			transition.StartAt == nil || !transition.StartAt.Equal(secondMovedStart) ||
			transition.ShiftMinutes != 3*60 {
			t.Fatalf("transition should describe the largest changed segment: %#v", transition)
		}
	})

	t.Run("split onto another day is always meaningful", func(t *testing.T) {
		oldStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
		newDayStart := oldStart.Add(24 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			[]calendar.CoreScheduleBlock{{
				UserID: userID, SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(2 * time.Hour),
			}},
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: newDayStart, EndAt: newDayStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionDayChanged {
			t.Fatalf("a split onto another date must satisfy the always-meaningful day-change notification contract: %#v", transition)
		}
		if transition.StartAt == nil || !transition.StartAt.Equal(newDayStart) || transition.PreviousLocalDate == transition.LocalDate {
			t.Fatalf("transition should describe the newly scheduled day: %#v", transition)
		}
	})

	t.Run("material end-only change is moved", func(t *testing.T) {
		oldStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			[]calendar.CoreScheduleBlock{{
				UserID: userID, SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(2 * time.Hour),
			}},
			[]calendar.MayaScheduleSegmentInput{{
				SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(time.Hour),
			}},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya changed the work block.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionMoved || transition.ShiftMinutes != -60 {
			t.Fatalf("a material end-only change must meet the 60-minute notification threshold: %#v", transition)
		}
	})
}

func TestBuildStoryScheduleTransitionDoesNotTreatPlanningReplanAsFirstSchedule(t *testing.T) {
	userID := uuid.New()
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	transition := buildStoryScheduleTransition(
		stories.CoreSingleStory{AutoSchedulingStatus: stories.AutoSchedulingStatusPlanning},
		userID,
		[]calendar.CoreScheduleBlock{{UserID: userID, SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		[]calendar.MayaScheduleSegmentInput{{SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		"UTC",
		stories.AutoSchedulingStatusScheduled,
		"Maya kept the existing schedule.",
	)

	if transition == nil || transition.Kind != events.StoryScheduleTransitionStateChanged {
		t.Fatalf("an existing schedule leaving planning must be a state change, not first schedule: %#v", transition)
	}
}

func TestScheduleTransitionDoesNotAnnounceAnExistingSlotAgain(t *testing.T) {
	userID := uuid.New()
	start := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	block := func(index int, offset time.Duration) ScheduleBlock {
		return ScheduleBlock{UserID: userID, SegmentIndex: index, StartAt: start.Add(offset), EndAt: start.Add(offset + time.Hour)}
	}
	segment := func(index int, offset time.Duration) ScheduleSegmentInput {
		return ScheduleSegmentInput{SegmentIndex: index, StartAt: start.Add(offset), EndAt: start.Add(offset + time.Hour)}
	}
	story := Story{AutoSchedulingStatus: AutoSchedulingStatusScheduled}
	for _, test := range []struct {
		name     string
		segments []ScheduleSegmentInput
	}{
		{"earlier block removed", []ScheduleSegmentInput{segment(1, 24*time.Hour)}},
		{"remaining block renumbered", []ScheduleSegmentInput{segment(0, 24*time.Hour)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			transition := buildStoryScheduleTransition(story, userID,
				[]ScheduleBlock{block(0, 0), block(1, 24*time.Hour)}, test.segments,
				"UTC", AutoSchedulingStatusScheduled, "Availability changed.")
			if transition != nil {
				t.Fatalf("unchanged destination must not generate another alert: %#v", transition)
			}
		})
	}
	transition := buildStoryScheduleTransition(story, userID,
		[]ScheduleBlock{block(0, 0), block(1, 24*time.Hour)},
		[]ScheduleSegmentInput{segment(0, 24*time.Hour), segment(1, 48*time.Hour)},
		"UTC", AutoSchedulingStatusScheduled, "Availability changed.")
	if transition == nil || !transition.StartAt.Equal(start.Add(48*time.Hour)) {
		t.Fatalf("a genuinely new slot must still generate an alert: %#v", transition)
	}
}

func TestRefineScheduleOutcomeReasonKeepsFirstScheduleCopy(t *testing.T) {
	fallback := "Maya scheduled this story around the assignee's availability."
	startAt := time.Now().UTC().Add(24 * time.Hour)
	reason := refineScheduleOutcomeReason(
		nil,
		[]calendar.MayaScheduleSegmentInput{{SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		stories.AutoSchedulingStatusScheduled,
		fallback,
	)
	if reason != fallback || strings.Contains(strings.ToLower(reason), "moved") || strings.Contains(strings.ToLower(reason), "changed") {
		t.Fatalf("first schedule must keep first-placement copy, got %q", reason)
	}
}

func TestScheduleTransitionMatchesRetainedTimeAfterEarlierSegmentExpires(t *testing.T) {
	userID := uuid.New()
	activeStart := time.Date(2026, 9, 14, 14, 40, 0, 0, time.UTC)
	tomorrowStart := time.Date(2026, 9, 15, 11, 38, 0, 0, time.UTC)
	previousBlocks := []ScheduleBlock{
		{UserID: userID, SegmentIndex: 0, StartAt: activeStart, EndAt: activeStart.Add(20 * time.Minute)},
		{UserID: userID, SegmentIndex: 1, StartAt: tomorrowStart, EndAt: tomorrowStart.Add(40 * time.Minute)},
	}
	for _, test := range []struct {
		name       string
		startShift time.Duration
		endShift   time.Duration
		wantShift  int
	}{
		{name: "twenty minute extension", endShift: 20 * time.Minute},
		{name: "twenty minute shift", startShift: 20 * time.Minute, endShift: 20 * time.Minute},
		{name: "forty minute non-overlapping shift", startShift: 40 * time.Minute, endShift: 40 * time.Minute},
		{name: "fifty-nine minute non-overlapping shift", startShift: 59 * time.Minute, endShift: 59 * time.Minute},
		{name: "material extension", endShift: time.Hour, wantShift: 60},
		{name: "material earlier start", startShift: -time.Hour, wantShift: -60},
	} {
		t.Run(test.name, func(t *testing.T) {
			transition := buildStoryScheduleTransition(
				Story{AutoSchedulingStatus: AutoSchedulingStatusScheduled}, userID, previousBlocks,
				[]ScheduleSegmentInput{{
					SegmentIndex: 0,
					StartAt:      tomorrowStart.Add(test.startShift),
					EndAt:        tomorrowStart.Add(40*time.Minute + test.endShift),
				}},
				"Africa/Harare", AutoSchedulingStatusScheduled, "Availability changed.",
			)
			if test.wantShift == 0 {
				if transition != nil {
					t.Fatalf("a minor change to retained tomorrow work must not look like today's block moved a day: %#v", transition)
				}
				return
			}
			if transition == nil || transition.Kind != events.StoryScheduleTransitionMoved ||
				transition.ShiftMinutes != test.wantShift || transition.PreviousStartAt == nil ||
				!transition.PreviousStartAt.Equal(tomorrowStart) || transition.PreviousEndAt == nil ||
				!transition.PreviousEndAt.Equal(previousBlocks[1].EndAt) {
				t.Fatalf("a material retained-slot change must compare with its own prior time: %#v", transition)
			}
		})
	}
	t.Run("non-overlapping hour shift remains meaningful", func(t *testing.T) {
		transition := buildStoryScheduleTransition(
			Story{AutoSchedulingStatus: AutoSchedulingStatusScheduled}, userID, previousBlocks,
			[]ScheduleSegmentInput{{
				SegmentIndex: 0,
				StartAt:      tomorrowStart.Add(time.Hour),
				EndAt:        tomorrowStart.Add(100 * time.Minute),
			}},
			"Africa/Harare", AutoSchedulingStatusScheduled, "Availability changed.",
		)
		if transition == nil {
			t.Fatal("a non-overlapping hour-long move must still produce a schedule transition")
		}
	})
}

func TestLockedScheduleRisk(t *testing.T) {
	baseStart := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	asOf := baseStart.Add(-time.Hour)
	storyID := uuid.New()
	ownerID := uuid.New()
	block := func(index int, start time.Time, minutes int) calendar.CoreScheduleBlock {
		return calendar.CoreScheduleBlock{
			ID: uuid.New(), UserID: ownerID, StoryID: &storyID, SegmentIndex: index,
			StartAt: start, EndAt: start.Add(time.Duration(minutes) * time.Minute), IsLocked: true,
		}
	}
	intPointer := func(value int) *int { return &value }
	timePointer := func(value time.Time) *time.Time { return &value }

	tests := []struct {
		name       string
		story      stories.CoreSingleStory
		blocks     []calendar.CoreScheduleBlock
		wantRisk   bool
		reasonPart string
	}{
		{
			name: "valid", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks: []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
		},
		{
			name: "busy conflict", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks: func() []calendar.CoreScheduleBlock {
				value := block(0, baseStart, 60)
				value.HasConflict = true
				return []calendar.CoreScheduleBlock{value}
			}(),
			wantRisk: true, reasonPart: "conflicts with another calendar event",
		},
		{
			name: "elapsed", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks:   []calendar.CoreScheduleBlock{block(0, asOf.Add(-2*time.Hour), 60)},
			wantRisk: true, reasonPart: "locked work time has passed",
		},
		{
			name: "under allocated", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(90)},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "reserves 60 minutes",
		},
		{
			name: "over allocated", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(30)},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "now needs 30 minutes",
		},
		{
			name: "minimum focus violation", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), MinimumFocusBlockMinutes: intPointer(45),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 30), block(1, baseStart.Add(time.Hour), 30)},
			wantRisk: true, reasonPart: "45-minute minimum focus block",
		},
		{
			name: "before story start", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), StartDate: timePointer(baseStart.Add(time.Hour)),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "before this story's start date",
		},
		{
			name: "after story deadline", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), EndDate: timePointer(baseStart.Add(-48 * time.Hour)),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "after this story's deadline",
		},
		{
			name: "outside sprint", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), SprintSummary: &stories.CoreSprintSummary{
					StartDate: baseStart.Add(24 * time.Hour), EndDate: baseStart.Add(48 * time.Hour),
				},
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "outside this story's sprint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason, atRisk := lockedScheduleRisk(test.story, test.blocks, "UTC", asOf)
			if atRisk != test.wantRisk {
				t.Fatalf("lockedScheduleRisk() risk = %v, want %v; reason=%q", atRisk, test.wantRisk, reason)
			}
			if test.reasonPart != "" && !strings.Contains(reason, test.reasonPart) {
				t.Fatalf("lockedScheduleRisk() reason = %q, want substring %q", reason, test.reasonPart)
			}
		})
	}
}

func containsUUID(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func timePointer(value time.Time) *time.Time { return &value }
