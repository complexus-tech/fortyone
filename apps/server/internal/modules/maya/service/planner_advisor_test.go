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
	asOf := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)
	startAt := asOf.Add(2 * time.Hour)
	deadline := asOf.Add(24 * time.Hour)
	lowerBlockID := uuid.New()

	result, err := NewPlanner().Plan(PlanInput{
		AsOf:        asOf,
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
	asOf := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)
	startAt := asOf.Add(2 * time.Hour)
	deadline := asOf.Add(72 * time.Hour)
	earlierDeadline := asOf.Add(24 * time.Hour)

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
		{name: "in progress", priority: "Medium", storyEnd: &deadline, blockEnd: &deadline, blockStart: asOf.Add(-30 * time.Minute), expectedStart: startAt},
		{name: "more urgent deadline", priority: "Medium", storyEnd: &deadline, blockEnd: &earlierDeadline, blockStart: startAt, expectedStart: startAt.Add(time.Hour)},
	} {
		t.Run(test.name, func(t *testing.T) {
			lowerStoryID := uuid.New()
			result, err := NewPlanner().Plan(PlanInput{
				AsOf:            asOf,
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
