package storieshttp

import (
	"net/http/httptest"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestToCoreNewStoryMapsLabelIDs(t *testing.T) {
	userID := uuid.New()
	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}

	coreStory := toCoreNewStory(AppNewStory{
		Title:    "Add reporting filters",
		LabelIDs: labelIDs,
		Team:     uuid.New(),
		Priority: "High",
	}, userID)

	if len(coreStory.LabelIDs) != len(labelIDs) {
		t.Fatalf("expected %d labels, got %d", len(labelIDs), len(coreStory.LabelIDs))
	}

	for i, labelID := range labelIDs {
		if coreStory.LabelIDs[i] != labelID {
			t.Fatalf("expected label %d to be %s, got %s", i, labelID, coreStory.LabelIDs[i])
		}
	}
}

func TestParseStoryQueryMapsEstimateValues(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/stories?estimateValues=1,5&estimateValues=8",
		nil,
	)

	query, err := parseStoryQuery(request, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("expected query to parse, got error: %v", err)
	}

	expected := []int16{1, 5, 8}
	if len(query.Filters.EstimateValues) != len(expected) {
		t.Fatalf("expected %d estimates, got %d", len(expected), len(query.Filters.EstimateValues))
	}

	for i, estimateValue := range expected {
		if query.Filters.EstimateValues[i] != estimateValue {
			t.Fatalf("expected estimate %d to be %d, got %d", i, estimateValue, query.Filters.EstimateValues[i])
		}
	}
}

func TestToAppStoryListItemIncludesEmbeddedSummaries(t *testing.T) {
	teamID := uuid.New()
	objectiveID := uuid.New()
	sprintID := uuid.New()
	workspaceID := uuid.New()
	startDate := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.June, 14, 0, 0, 0, 0, time.UTC)
	description := "Improve conversion from trial teams"
	goal := "Ship the onboarding pass"

	story := stories.CoreStoryList{
		ID:        uuid.New(),
		Title:     "Add activation checklist",
		Objective: &objectiveID,
		Sprint:    &sprintID,
		Team:      teamID,
		Workspace: workspaceID,
		TeamSummary: &stories.CoreTeamSummary{
			ID:   teamID,
			Name: "Growth",
			Code: "GRO",
		},
		ObjectiveSummary: &stories.CoreObjectiveSummary{
			ID:          objectiveID,
			Name:        "Increase activation",
			Description: &description,
		},
		SprintSummary: &stories.CoreSprintSummary{
			ID:        sprintID,
			Name:      "June hardening",
			Goal:      &goal,
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	appStory := toAppStoryListItem(story, nil)

	if appStory.TeamSummary == nil {
		t.Fatal("expected team summary to be embedded")
	}
	if appStory.TeamSummary.Code != "GRO" {
		t.Fatalf("expected team code GRO, got %q", appStory.TeamSummary.Code)
	}
	if appStory.ObjectiveSummary == nil {
		t.Fatal("expected objective summary to be embedded")
	}
	if appStory.ObjectiveSummary.Name != "Increase activation" {
		t.Fatalf("expected objective name to be embedded, got %q", appStory.ObjectiveSummary.Name)
	}
	if appStory.ObjectiveSummary.Description == nil || *appStory.ObjectiveSummary.Description != description {
		t.Fatalf("expected objective description %q, got %#v", description, appStory.ObjectiveSummary.Description)
	}
	if appStory.SprintSummary == nil {
		t.Fatal("expected sprint summary to be embedded")
	}
	if appStory.SprintSummary.Name != "June hardening" {
		t.Fatalf("expected sprint name to be embedded, got %q", appStory.SprintSummary.Name)
	}
	if appStory.SprintSummary.Goal == nil || *appStory.SprintSummary.Goal != goal {
		t.Fatalf("expected sprint goal %q, got %#v", goal, appStory.SprintSummary.Goal)
	}
}
