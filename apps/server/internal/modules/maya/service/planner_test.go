package maya

import (
	"context"
	"errors"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestPlannerChoosesEarliestAvailableLowLoadCandidate(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Improve onboarding", EstimateValue: int16Ptr(2)},
		DurationMinutes: 90,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:        firstUserID,
					FullName:      "Busy Person",
					OpenStories:   9,
					EstimateTotal: 22,
				},
				BusyWindows: []calendar.CoreBusyWindow{
					{StartAt: startAt, EndAt: startAt.Add(2 * time.Hour)},
				},
			},
			{
				Member: reports.CoreMemberWorkload{
					UserID:        secondUserID,
					FullName:      "Available Person",
					OpenStories:   2,
					EstimateTotal: 4,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != secondUserID {
		t.Fatalf("expected selected user %s, got %v", secondUserID, result.SelectedUserID)
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected assign and schedule actions, got %d", len(result.Actions))
	}
	if result.Actions[0].Type != ActionTypeAssignStory {
		t.Fatalf("expected first action %q, got %q", ActionTypeAssignStory, result.Actions[0].Type)
	}
	if result.Actions[1].Type != ActionTypeScheduleWorkBlock {
		t.Fatalf("expected second action %q, got %q", ActionTypeScheduleWorkBlock, result.Actions[1].Type)
	}
	if result.Actions[1].Payload.ScheduleBlock.StartAt != startAt {
		t.Fatalf("expected schedule start %s, got %s", startAt, result.Actions[1].Payload.ScheduleBlock.StartAt)
	}
}

func TestPlannerAvoidsInactiveCandidateWhenActiveAlternativeExists(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	inactiveUserID := uuid.New()
	activeUserID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	recentActivityAt := time.Now().UTC().Add(-24 * time.Hour)
	staleActivityAt := time.Now().UTC().Add(-60 * 24 * time.Hour)

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Route ownership"},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:              inactiveUserID,
					FullName:            "Dormant Person",
					OpenStories:         0,
					EstimateTotal:       0,
					LastStoryActivityAt: &staleActivityAt,
				},
			},
			{
				Member: reports.CoreMemberWorkload{
					UserID:              activeUserID,
					FullName:            "Active Person",
					OpenStories:         8,
					EstimateTotal:       16,
					LastStoryActivityAt: &recentActivityAt,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != activeUserID {
		t.Fatalf("expected active user %s, got %v", activeUserID, result.SelectedUserID)
	}
}

func TestPlannerUsesSprintWindowWhenStoryBelongsToSprint(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	windowStart := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 6, 30, 17, 0, 0, 0, time.UTC)
	sprintStart := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	sprintEnd := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	expectedStart := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Title:     "Build sprint scoped scheduling",
			SprintSummary: &stories.CoreSprintSummary{
				ID:        uuid.New(),
				Name:      "Sprint 12",
				StartDate: sprintStart,
				EndDate:   sprintEnd,
			},
		},
		DurationMinutes: 60,
		WindowStart:     windowStart,
		WindowEnd:       windowEnd,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:   userID,
					FullName: "Sprint Person",
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected assign and schedule actions, got %d", len(result.Actions))
	}
	if got := result.Actions[1].Payload.ScheduleBlock.StartAt; !got.Equal(expectedStart) {
		t.Fatalf("expected sprint-scoped schedule start %s, got %s", expectedStart, got)
	}
}

func TestPlannerSpreadsLargerWorkAcrossAvailableWindows(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 18, 17, 0, 0, 0, time.UTC)
	expectedPlannedEnd := time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC)
	minimumFocus := 60

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID:                       storyID,
			Workspace:                workspaceID,
			Title:                    "Implement WhatsApp campaigns end-to-end",
			EstimatedDurationMinutes: intPtr(8 * 60), MinimumFocusBlockMinutes: &minimumFocus,
		},
		WindowStart: startAt,
		WindowEnd:   endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:   userID,
					FullName: "Campaign Person",
				},
				Blocks: []calendar.CoreScheduleBlock{
					{StartAt: startAt, EndAt: startAt.Add(4 * time.Hour), Source: calendar.ScheduleBlockSourceUser},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 5 {
		t.Fatalf("expected assign and four schedule actions, got %d", len(result.Actions))
	}
	scheduleBlock := result.Actions[1].Payload.ScheduleBlock
	if scheduleBlock == nil {
		t.Fatal("expected schedule block payload")
	}
	if !scheduleBlock.StartAt.Equal(startAt.Add(4 * time.Hour)) {
		t.Fatalf("expected first focus block to start after existing work, got %s", scheduleBlock.StartAt)
	}
	if !scheduleBlock.PlannedEndAt.Equal(expectedPlannedEnd) {
		t.Fatalf("expected larger work to plan through %s, got %s", expectedPlannedEnd, scheduleBlock.PlannedEndAt)
	}
}

func TestPlannerKeepsNoChunkWorkContiguousWhenAvailable(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	duration := 4 * 60

	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Title: "Keep focused work together",
			Assignee: &userID, EstimatedDurationMinutes: &duration,
		},
		WindowStart: startAt, WindowEnd: endAt,
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID},
		}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected one contiguous schedule action, got %d", len(result.Actions))
	}

	segment := result.Actions[0].Payload.ScheduleBlock
	if segment == nil {
		t.Fatal("expected schedule block payload")
	}
	if !segment.StartAt.Equal(startAt) || !segment.EndAt.Equal(startAt.Add(4*time.Hour)) {
		t.Fatalf("expected one four-hour block, got %s-%s", segment.StartAt, segment.EndAt)
	}
}

func TestPlannerSplitsNoChunkWorkOnlyWhenConflictsRequireIt(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	duration := 6 * 60

	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Title: "Split around a commitment",
			Assignee: &userID, EstimatedDurationMinutes: &duration,
		},
		WindowStart: startAt, WindowEnd: endAt,
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID},
			BusyWindows: []calendar.CoreBusyWindow{{
				StartAt: startAt.Add(5 * time.Hour), EndAt: startAt.Add(7 * time.Hour),
			}},
		}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected two schedule segments, got %d", len(result.Actions))
	}

	first := result.Actions[0].Payload.ScheduleBlock
	second := result.Actions[1].Payload.ScheduleBlock
	if first == nil || second == nil {
		t.Fatal("expected schedule block payloads")
	}
	if !first.StartAt.Equal(startAt) || !first.EndAt.Equal(startAt.Add(5*time.Hour)) {
		t.Fatalf("expected first slice to use the five-hour window, got %s-%s", first.StartAt, first.EndAt)
	}
	if !second.StartAt.Equal(startAt.Add(7*time.Hour)) || !second.EndAt.Equal(endAt) {
		t.Fatalf("expected remaining hour after the conflict, got %s-%s", second.StartAt, second.EndAt)
	}
}

func TestPlannerCreatesExactMinimumSizedSegmentsAroundBusyWindows(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	duration := 300
	minimumFocus := 60

	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Title: "Segmented work", Assignee: &userID,
			EstimatedDurationMinutes: &duration, MinimumFocusBlockMinutes: &minimumFocus,
		},
		WindowStart: startAt, WindowEnd: endAt,
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID},
			BusyWindows: []calendar.CoreBusyWindow{
				{StartAt: startAt.Add(2 * time.Hour), EndAt: startAt.Add(3 * time.Hour)},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 3 {
		t.Fatalf("expected three schedule segments, got %d", len(result.Actions))
	}
	total := time.Duration(0)
	var previousEnd time.Time
	for index, action := range result.Actions {
		segment := action.Payload.ScheduleBlock
		if segment == nil || segment.SegmentIndex != index {
			t.Fatalf("unexpected segment %d: %#v", index, segment)
		}
		duration := segment.EndAt.Sub(segment.StartAt)
		if duration < time.Hour {
			t.Fatalf("segment %d is shorter than minimum focus: %s", index, duration)
		}
		if !previousEnd.IsZero() && segment.StartAt.Before(previousEnd) {
			t.Fatalf("segment %d overlaps the previous segment", index)
		}
		if segment.StartAt.Before(startAt.Add(3*time.Hour)) && segment.EndAt.After(startAt.Add(2*time.Hour)) {
			t.Fatalf("segment %d overlaps provider busy time: %#v", index, segment)
		}
		total += duration
		previousEnd = segment.EndAt
	}
	if total != 300*time.Minute {
		t.Fatalf("expected exactly 300 scheduled minutes, got %s", total)
	}
}

func TestPlannerHonorsFifteenMinuteDurationWithoutDefaultClamp(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	duration := 15
	startAt := time.Date(2026, 6, 15, 9, 2, 0, 0, time.UTC)
	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story:       stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Quick follow-up", Assignee: &userID, EstimatedDurationMinutes: &duration},
		WindowStart: startAt, WindowEnd: startAt.Add(time.Hour),
		Candidates: []CandidateSchedule{{Member: reports.CoreMemberWorkload{UserID: userID}}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 1 || result.Actions[0].Payload.ScheduleBlock == nil {
		t.Fatalf("expected one schedule segment: %#v", result.Actions)
	}
	segment := result.Actions[0].Payload.ScheduleBlock
	if segment.EndAt.Sub(segment.StartAt) != 15*time.Minute || segment.StartAt.Minute() != 5 {
		t.Fatalf("expected exact 15-minute block at five-minute resolution: %#v", segment)
	}
}

func TestPlannerUsesCalendarTimezoneForWorkingHours(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	duration := 60
	windowStart := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story:       stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Timezone-aware work", Assignee: &userID, EstimatedDurationMinutes: &duration},
		WindowStart: windowStart, WindowEnd: windowStart.Add(24 * time.Hour),
		Candidates: []CandidateSchedule{{Member: reports.CoreMemberWorkload{UserID: userID}, Timezone: "Africa/Harare"}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	expectedStart := time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC)
	if len(result.Actions) != 1 || !result.Actions[0].Payload.ScheduleBlock.StartAt.Equal(expectedStart) {
		t.Fatalf("expected 09:00 Africa/Harare (%s), got %#v", expectedStart, result.Actions)
	}
}

func TestPlannerReturnsMissingDurationRiskInsteadOfUsingComplexity(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story:       stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Complex but untimed", Assignee: &userID, EstimateValue: int16Ptr(8), EstimateScheme: "points"},
		WindowStart: time.Now().UTC(), WindowEnd: time.Now().UTC().Add(24 * time.Hour),
		Candidates: []CandidateSchedule{{Member: reports.CoreMemberWorkload{UserID: userID}}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 2 || result.Actions[0].Payload.ScheduleBlock == nil || result.Actions[0].Payload.ScheduleBlock.Operation != ScheduleBlockOperationRetain || result.Actions[1].Payload.Risk == nil || result.Actions[1].Payload.Risk.Code != "missing_duration" {
		t.Fatalf("expected explicit missing-duration risk: %#v", result.Actions)
	}
}

func TestPlannerAssignsAndEnrollsMissingDurationForNonCandidateMayaAssignee(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	mayaActorID := uuid.New()
	humanCandidateID := uuid.New()
	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Title: "Batch story without time", Assignee: &mayaActorID,
		},
		WindowStart: time.Now().UTC(), WindowEnd: time.Now().UTC().Add(24 * time.Hour),
		Candidates: []CandidateSchedule{{Member: reports.CoreMemberWorkload{UserID: humanCandidateID}}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != humanCandidateID || len(result.Actions) != 3 ||
		result.Actions[0].Payload.AssignStory == nil || result.Actions[0].Payload.AssignStory.AssigneeID != humanCandidateID ||
		result.Actions[1].Payload.ScheduleBlock == nil || result.Actions[1].Payload.ScheduleBlock.Operation != ScheduleBlockOperationRetain ||
		result.Actions[2].Payload.Risk == nil || result.Actions[2].Payload.Risk.Code != "missing_duration" {
		t.Fatalf("expected eligible human assignment, ownership, and risk for the Maya-assigned story: %#v", result)
	}
}

func TestPlannerRejectsDurationAboveProductLimit(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	duration := stories.MaximumEstimatedDurationMinutes + 1
	_, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story:       stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Unbounded work", Assignee: &userID, EstimatedDurationMinutes: &duration},
		WindowStart: time.Now().UTC(), WindowEnd: time.Now().UTC().Add(90 * 24 * time.Hour),
		Candidates: []CandidateSchedule{{Member: reports.CoreMemberWorkload{UserID: userID}}},
	})
	if !errors.Is(err, ErrInvalidPlanInput) {
		t.Fatalf("expected defensive duration limit, got %v", err)
	}
}

func TestPlannerHonorsCustomWorkingDays(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	windowStart := time.Date(2026, 6, 19, 9, 0, 0, 0, time.UTC)   // Friday
	windowEnd := time.Date(2026, 6, 22, 17, 0, 0, 0, time.UTC)    // Monday
	expectedStart := time.Date(2026, 6, 21, 9, 0, 0, 0, time.UTC) // Sunday

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Title:     "Plan around the team workweek",
		},
		DurationMinutes: 60,
		WindowStart:     windowStart,
		WindowEnd:       windowEnd,
		WorkingDays:     []int{7, 1, 2, 3, 4},
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID, FullName: "Custom Week Person"},
		}},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if got := result.Actions[1].Payload.ScheduleBlock.StartAt; !got.Equal(expectedStart) {
		t.Fatalf("expected work to start on configured Sunday %s, got %s", expectedStart, got)
	}
}

func TestPlannerUsesAdvisorRecommendationWhenCandidateIsValid(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)

	var advisorInput CandidateRecommendationInput
	planner := NewPlannerWithAdvisor(fakeCandidateAdvisor{
		input: &advisorInput,
		result: CandidateRecommendationResult{
			UserID: secondUserID,
			Reason: "Available Person owns the backend area and has enough calendar capacity.",
		},
	})
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Improve webhook retries", EstimateValue: int16Ptr(2)},
		DurationMinutes: 90,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:        firstUserID,
					FullName:      "Earlier Person",
					OpenStories:   1,
					EstimateTotal: 2,
				},
			},
			{
				Member: reports.CoreMemberWorkload{
					UserID:                secondUserID,
					FullName:              "Available Person",
					TeamAIRoleTitle:       "Backend engineer",
					TeamAIRoleDescription: "Owns webhook reliability and backend integrations.",
					OpenStories:           4,
					EstimateTotal:         8,
				},
				BusyWindows: []calendar.CoreBusyWindow{
					{StartAt: startAt, EndAt: startAt.Add(time.Hour)},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != secondUserID {
		t.Fatalf("expected advisor-selected user %s, got %v", secondUserID, result.SelectedUserID)
	}
	if len(result.Actions) != 2 {
		t.Fatalf("expected assign and schedule actions, got %d", len(result.Actions))
	}
	if got := result.Actions[0].Reason; got != "Available Person owns the backend area and has enough calendar capacity." {
		t.Fatalf("expected advisor reason, got %q", got)
	}
	if got := result.Actions[1].Payload.ScheduleBlock.StartAt; !got.Equal(startAt.Add(time.Hour)) {
		t.Fatalf("expected advisor candidate slot start %s, got %s", startAt.Add(time.Hour), got)
	}
	if len(advisorInput.Candidates) != 2 {
		t.Fatalf("expected two advisor candidates, got %d", len(advisorInput.Candidates))
	}
	if got := advisorInput.Candidates[1].TeamAIRoleTitle; got != "Backend engineer" {
		t.Fatalf("expected advisor role title, got %q", got)
	}
	if got := advisorInput.Candidates[1].TeamAIRoleDescription; got != "Owns webhook reliability and backend integrations." {
		t.Fatalf("expected advisor role description, got %q", got)
	}
}

func TestPlannerFallsBackWhenAdvisorRecommendationIsInvalid(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	firstUserID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)

	planner := NewPlannerWithAdvisor(fakeCandidateAdvisor{
		result: CandidateRecommendationResult{UserID: uuid.New(), Reason: "invalid"},
	})
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Improve onboarding"},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{
					UserID:        firstUserID,
					FullName:      "Fallback Person",
					OpenStories:   1,
					EstimateTotal: 2,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != firstUserID {
		t.Fatalf("expected deterministic fallback user %s, got %v", firstUserID, result.SelectedUserID)
	}
}

func TestPlannerReturnsRiskActionWhenNoCandidateHasCapacity(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Ship billing"},
		DurationMinutes: 120,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{UserID: userID, FullName: "Packed Person"},
				BusyWindows: []calendar.CoreBusyWindow{
					{StartAt: startAt, EndAt: endAt},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.SelectedUserID == nil || *result.SelectedUserID != userID {
		t.Fatalf("expected selected user %s, got %v", userID, result.SelectedUserID)
	}
	if len(result.Actions) != 3 {
		t.Fatalf("expected assignment, ownership retention, and risk actions, got %d", len(result.Actions))
	}
	if result.Actions[0].Type != ActionTypeAssignStory {
		t.Fatalf("expected first action %q, got %q", ActionTypeAssignStory, result.Actions[0].Type)
	}
	if result.Actions[1].Type != ActionTypeScheduleWorkBlock || result.Actions[1].Payload.ScheduleBlock == nil || result.Actions[1].Payload.ScheduleBlock.Operation != ScheduleBlockOperationRetain {
		t.Fatalf("expected ownership retention action, got %#v", result.Actions[1])
	}
	if result.Actions[2].Type != ActionTypeFlagScheduleRisk {
		t.Fatalf("expected risk action %q, got %q", ActionTypeFlagScheduleRisk, result.Actions[2].Type)
	}
}

func TestPlannerSkipsScheduleActionWhenStoryAlreadyHasBlock(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)

	planner := NewPlanner()
	result, err := planner.Plan(PlanInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Already scheduled", Assignee: &userID},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		Candidates: []CandidateSchedule{
			{
				Member: reports.CoreMemberWorkload{UserID: userID, FullName: "Scheduled Person"},
				Blocks: []calendar.CoreScheduleBlock{
					{StoryID: &storyID, StartAt: startAt, EndAt: startAt.Add(time.Hour), Source: calendar.ScheduleBlockSourceMaya},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 0 {
		t.Fatalf("expected no actions for already assigned and scheduled story, got %d", len(result.Actions))
	}
}

func TestPlannerPreemptsMovableLowerPriorityBlock(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	lowerStoryID := uuid.New()
	userID := uuid.New()
	startAt := time.Now().UTC().Add(2 * time.Hour).Truncate(5 * time.Minute)
	deadline := time.Now().UTC().Add(24 * time.Hour)
	lowerBlockID := uuid.New()

	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Assignee: &userID,
			Priority: "High", EndDate: &deadline,
		},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       startAt.Add(8 * time.Hour),
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID},
			Blocks: []calendar.CoreScheduleBlock{{
				ID: lowerBlockID, WorkspaceID: workspaceID, UserID: userID, StoryID: &lowerStoryID,
				Source: calendar.ScheduleBlockSourceMaya, StartAt: startAt, EndAt: startAt.Add(time.Hour),
				StoryPriority: "Medium", StoryEndDate: &deadline,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 1 || result.Actions[0].Payload.ScheduleBlock == nil {
		t.Fatalf("expected one schedule action, got %#v", result.Actions)
	}
	if got := result.Actions[0].Payload.ScheduleBlock.StartAt; !got.Equal(startAt) {
		t.Fatalf("expected high-priority work to claim the lower-priority slot, got %s", got)
	}
	if len(result.PreemptedBlockIDs) != 1 || result.PreemptedBlockIDs[0] != lowerBlockID {
		t.Fatalf("expected lower-priority block to be marked preemptible, got %v", result.PreemptedBlockIDs)
	}
	if len(result.Actions[0].Payload.ScheduleBlock.PreemptBlockIDs) != 1 || result.Actions[0].Payload.ScheduleBlock.PreemptBlockIDs[0] != lowerBlockID {
		t.Fatalf("expected preemption metadata on schedule action, got %v", result.Actions[0].Payload.ScheduleBlock.PreemptBlockIDs)
	}
}

func TestPlannerDoesNotPreemptProtectedOrMoreUrgentWork(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	startAt := now.Add(2 * time.Hour).Truncate(5 * time.Minute)
	deadline := now.Add(72 * time.Hour)
	earlierDeadline := now.Add(24 * time.Hour)

	for _, test := range []struct {
		name          string
		priority      string
		storyEnd      *time.Time
		blockEnd      *time.Time
		locked        bool
		blockStart    time.Time
		expectedStart time.Time
	}{
		{name: "locked", priority: "Medium", storyEnd: &deadline, blockEnd: &deadline, locked: true, blockStart: startAt, expectedStart: startAt.Add(time.Hour)},
		{name: "in progress", priority: "Medium", storyEnd: &deadline, blockEnd: &deadline, blockStart: now.Add(-30 * time.Minute), expectedStart: startAt},
		{name: "more urgent deadline", priority: "Medium", storyEnd: &deadline, blockEnd: &earlierDeadline, blockStart: startAt, expectedStart: startAt.Add(time.Hour)},
	} {
		t.Run(test.name, func(t *testing.T) {
			lowerStoryID := uuid.New()
			result, err := NewPlanner().Plan(PlanInput{
				WorkspaceID:     workspaceID,
				Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Assignee: &userID, Priority: "High", EndDate: test.storyEnd},
				DurationMinutes: 60,
				WindowStart:     startAt,
				WindowEnd:       startAt.Add(8 * time.Hour),
				Candidates: []CandidateSchedule{{
					Member: reports.CoreMemberWorkload{UserID: userID},
					Blocks: []calendar.CoreScheduleBlock{{
						ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, StoryID: &lowerStoryID,
						Source: calendar.ScheduleBlockSourceMaya, StartAt: test.blockStart, EndAt: test.blockStart.Add(time.Hour),
						StoryPriority: "Medium", StoryEndDate: test.blockEnd, IsLocked: test.locked,
					}},
				}},
			})
			if err != nil {
				t.Fatalf("Plan returned error: %v", err)
			}
			if len(result.Actions) != 1 || !result.Actions[0].Payload.ScheduleBlock.StartAt.Equal(test.expectedStart) {
				t.Fatalf("expected protected work to remain occupied, got %#v", result.Actions)
			}
			if len(result.PreemptedBlockIDs) != 0 {
				t.Fatalf("expected no preemption, got %v", result.PreemptedBlockIDs)
			}
		})
	}
}

func TestPlannerUsesManualSchedulePreferenceAsSoftStartTime(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	preferredStartMinute := 16 * 60
	startAt := time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)

	result, err := NewPlanner().Plan(PlanInput{
		WorkspaceID: workspaceID,
		Story: stories.CoreSingleStory{
			ID:        uuid.New(),
			Workspace: workspaceID,
			Assignee:  &userID,
		},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       startAt.Add(8 * time.Hour),
		Candidates: []CandidateSchedule{{
			Member:               reports.CoreMemberWorkload{UserID: userID},
			Timezone:             "UTC",
			PreferredStartMinute: &preferredStartMinute,
		}},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(result.Actions) != 1 || result.Actions[0].Payload.ScheduleBlock == nil {
		t.Fatalf("expected one schedule action, got %#v", result.Actions)
	}
	if got := result.Actions[0].Payload.ScheduleBlock.StartAt; !got.Equal(startAt.Add(7 * time.Hour)) {
		t.Fatalf("expected preferred start at 16:00 UTC, got %s", got)
	}
}

func int16Ptr(value int16) *int16 {
	return &value
}

func intPtr(value int) *int {
	return &value
}

type fakeCandidateAdvisor struct {
	result CandidateRecommendationResult
	err    error
	input  *CandidateRecommendationInput
}

func (f fakeCandidateAdvisor) RecommendCandidate(_ context.Context, input CandidateRecommendationInput) (CandidateRecommendationResult, error) {
	if f.input != nil {
		*f.input = input
	}
	if f.err != nil {
		return CandidateRecommendationResult{}, f.err
	}
	return f.result, nil
}

var _ CandidateAdvisor = fakeCandidateAdvisor{err: errors.New("unused")}
