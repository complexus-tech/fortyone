package reportsrepository

import (
	"strings"
	"testing"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

func TestBuildWorkloadStoryFilterSupportsAssigneeIDs(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	namedParams := map[string]any{}

	filter := buildWorkloadStoryFilter(reports.ReportFilters{
		AssigneeIDs: []uuid.UUID{assigneeID},
	}, namedParams)

	if !strings.Contains(filter, "s.assignee_id = ANY(:assignee_ids)") {
		t.Fatalf("expected assignee filter in %q", filter)
	}
	if got := namedParams["assignee_ids"]; got == nil {
		t.Fatal("expected assignee_ids named parameter")
	}
}

func TestBuildPulseRequestFilterSupportsAllReportFilters(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	assigneeID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()
	namedParams := map[string]any{}

	filter := buildPulseRequestFilter(reports.ReportFilters{
		TeamIDs:      []uuid.UUID{teamID},
		AssigneeIDs:  []uuid.UUID{assigneeID},
		SprintIDs:    []uuid.UUID{sprintID},
		ObjectiveIDs: []uuid.UUID{objectiveID},
	}, namedParams)

	expectedFragments := []string{
		"ir.team_id = ANY(:team_ids)",
		"ir.assignee_id = ANY(:assignee_ids)",
		"ir.sprint_id = ANY(:sprint_ids)",
		"ir.objective_id = ANY(:objective_ids)",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(filter, fragment) {
			t.Fatalf("expected request filter to contain %q, got %q", fragment, filter)
		}
	}
	for _, key := range []string{"team_ids", "assignee_ids", "sprint_ids", "objective_ids"} {
		if namedParams[key] == nil {
			t.Fatalf("expected %s named parameter", key)
		}
	}
}
