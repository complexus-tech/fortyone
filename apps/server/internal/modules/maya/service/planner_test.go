package maya

import (
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
	if result.SelectedUserID != nil {
		t.Fatalf("expected no selected user, got %v", result.SelectedUserID)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("expected one risk action, got %d", len(result.Actions))
	}
	if result.Actions[0].Type != ActionTypeFlagScheduleRisk {
		t.Fatalf("expected risk action %q, got %q", ActionTypeFlagScheduleRisk, result.Actions[0].Type)
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
