package messaging

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func TestFortyOneToolExecutorSprintSummaryUsesScopedAnalyticsAndStories(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{ID: uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"), Name: "Product", Code: "PRO", Workspace: scope.WorkspaceID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	sprintID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	completedAt := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	completedStory := planningTestStory(scope, team.ID, 1, "Audit the interface", &sprintID, nil)
	completedStory.CompletedAt = &completedAt
	remainingStory := planningTestStory(scope, team.ID, 2, "Implement the primitives", &sprintID, nil)
	storiesReader := &completedStoriesServiceStub{items: []stories.CoreStoryList{completedStory, remainingStory}}
	sprintsReader := &planningSprintsServiceStub{
		items: []sprints.CoreSprint{{
			ID:          sprintID,
			Name:        "Sprint 1",
			Goal:        stringPtr("Establish the UI foundation"),
			TeamID:      team.ID,
			WorkspaceID: scope.WorkspaceID,
			StartDate:   time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC),
		}},
		analytics: sprints.CoreSprintAnalytics{
			SprintID:       sprintID,
			Overview:       sprints.CoreSprintOverview{CompletionPercentage: 50, DaysElapsed: 5, DaysRemaining: 5, Status: "on_track"},
			StoryBreakdown: sprints.CoreStoryBreakdown{Total: 2, Completed: 1, InProgress: 1},
		},
	}
	executor := newToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storiesReader, &searchServiceStub{}, &objectivesServiceStub{}, WithPlanningTools(PlanningToolServices{Sprints: sprintsReader}))
	if err := validateToolDefinitions(executor.Definitions()); err != nil {
		t.Fatalf("planning tool catalog is not strict: %v", err)
	}

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolGetSprintSummary,
		Arguments: json.RawMessage(`{"name":"Sprint 1","team_name":null,"limit":null}`),
	})
	if err != nil {
		t.Fatalf("execute sprint summary: %v", err)
	}
	var result sprintSummaryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode sprint summary: %v", err)
	}
	if result.Sprint.Name != "Sprint 1" || result.Sprint.ProgressPercent != 50 || result.Sprint.StartDate != "2026-08-10" || result.Sprint.EndDate != "2026-08-21" {
		t.Fatalf("unexpected sprint summary: %#v", result.Sprint)
	}
	if len(result.Work.Completed) != 1 || result.Work.Completed[0].Reference != "PRO-1" || len(result.Work.Remaining) != 1 || result.Work.Remaining[0].Reference != "PRO-2" {
		t.Fatalf("expected scoped completed and remaining work, got %#v", result.Work)
	}
	call := storiesReader.lastCall()
	if len(call.filters.SprintIDs) != 1 || call.filters.SprintIDs[0] != sprintID {
		t.Fatalf("expected sprint-scoped story filter, got %#v", call.filters)
	}
}

func TestFortyOneToolExecutorObjectiveSummaryUsesObjectiveScopedStories(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{ID: uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"), Name: "Product", Code: "PRO", Workspace: scope.WorkspaceID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	objectiveID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	completedStory := planningTestStory(scope, team.ID, 7, "Accessibility audit", nil, &objectiveID)
	completedAt := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	completedStory.CompletedAt = &completedAt
	remainingStory := planningTestStory(scope, team.ID, 8, "Define feedback patterns", nil, &objectiveID)
	storiesReader := &completedStoriesServiceStub{items: []stories.CoreStoryList{completedStory, remainingStory}}
	objective := objectives.CoreObjective{
		ID:               objectiveID,
		Name:             "Build the UI foundation",
		ShortSummary:     stringPtr("Create a consistent product interface"),
		Team:             team.ID,
		Workspace:        scope.WorkspaceID,
		StartDate:        timePtr(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:          timePtr(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		TotalStories:     2,
		CompletedStories: 1,
		StartedStories:   1,
	}
	objectivesReader := &objectivesServiceStub{items: []objectives.CoreObjective{objective}}
	executor := newToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{team}}, storiesReader, &searchServiceStub{}, objectivesReader, WithPlanningTools(PlanningToolServices{Sprints: &planningSprintsServiceStub{}}))

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolGetObjectiveSummary,
		Arguments: json.RawMessage(`{"name":"Build the UI foundation","team_name":"Product","limit":10}`),
	})
	if err != nil {
		t.Fatalf("execute objective summary: %v", err)
	}
	var result objectiveSummaryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode objective summary: %v", err)
	}
	if result.Objective.Name != objective.Name || result.Objective.ProgressPercent != 50 || result.Objective.Health != "" {
		t.Fatalf("unexpected objective summary: %#v", result.Objective)
	}
	if len(result.Work.Completed) != 1 || len(result.Work.Remaining) != 1 {
		t.Fatalf("expected objective work split, got %#v", result.Work)
	}
	call := storiesReader.lastCall()
	if call.filters.Objective == nil || *call.filters.Objective != objectiveID {
		t.Fatalf("expected objective-scoped story filter, got %#v", call.filters)
	}
}

type planningSprintsServiceStub struct {
	items     []sprints.CoreSprint
	analytics sprints.CoreSprintAnalytics
}

func (s *planningSprintsServiceStub) List(_ context.Context, _ uuid.UUID, _ uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error) {
	search, _ := filters["search"].(string)
	teamID, _ := filters["team_id"].(uuid.UUID)
	result := make([]sprints.CoreSprint, 0, len(s.items))
	for _, item := range s.items {
		if teamID != uuid.Nil && item.TeamID != teamID {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(search)) {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *planningSprintsServiceStub) GetAnalytics(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) (sprints.CoreSprintAnalytics, error) {
	return s.analytics, nil
}

func planningTestStory(scope ToolScope, teamID uuid.UUID, sequenceID int, title string, sprintID, objectiveID *uuid.UUID) stories.CoreStoryList {
	return stories.CoreStoryList{
		ID:         uuid.New(),
		SequenceID: sequenceID,
		Title:      title,
		Team:       teamID,
		Workspace:  scope.WorkspaceID,
		Sprint:     sprintID,
		Objective:  objectiveID,
		UpdatedAt:  time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC),
	}
}

func stringPtr(value string) *string { return &value }

func timePtr(value time.Time) *time.Time { return &value }
