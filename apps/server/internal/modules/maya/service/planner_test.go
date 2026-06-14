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
	if len(result.Actions) != 2 {
		t.Fatalf("expected assignment and risk actions, got %d", len(result.Actions))
	}
	if result.Actions[0].Type != ActionTypeAssignStory {
		t.Fatalf("expected first action %q, got %q", ActionTypeAssignStory, result.Actions[0].Type)
	}
	if result.Actions[1].Type != ActionTypeFlagScheduleRisk {
		t.Fatalf("expected risk action %q, got %q", ActionTypeFlagScheduleRisk, result.Actions[1].Type)
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

func int16Ptr(value int16) *int16 {
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
