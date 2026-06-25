package reportsrepository

import (
	"strings"
	"testing"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func TestBuildWorkloadStoryFilterSupportsAssigneeIDs(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	namedParams := map[string]any{}

	filter := buildWorkloadStoryFilter(reports.ReportFilters{
		AssigneeIDs: []uuid.UUID{assigneeID},
	}, namedParams)

	if !strings.Contains(filter, "s.assignee_id = ANY(CAST(:assignee_ids AS uuid[]))") {
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
		"ir.team_id = ANY(CAST(:team_ids AS uuid[]))",
		"ir.assignee_id = ANY(CAST(:assignee_ids AS uuid[]))",
		"ir.sprint_id = ANY(CAST(:sprint_ids AS uuid[]))",
		"ir.objective_id = ANY(CAST(:objective_ids AS uuid[]))",
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

func TestBuildRequestSourceFilterSupportsReportFilters(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	assigneeID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()
	namedParams := map[string]any{}

	filter := buildRequestSourceFilter(reports.ReportFilters{
		TeamIDs:      []uuid.UUID{teamID},
		AssigneeIDs:  []uuid.UUID{assigneeID},
		SprintIDs:    []uuid.UUID{sprintID},
		ObjectiveIDs: []uuid.UUID{objectiveID},
	}, namedParams)

	expectedFragments := []string{
		"ir.team_id = ANY(CAST(:team_ids AS uuid[]))",
		"ir.assignee_id = ANY(CAST(:assignee_ids AS uuid[]))",
		"ir.sprint_id = ANY(CAST(:sprint_ids AS uuid[]))",
		"ir.objective_id = ANY(CAST(:objective_ids AS uuid[]))",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(filter, fragment) {
			t.Fatalf("expected request source filter to contain %q, got %q", fragment, filter)
		}
	}
}

func TestBuildWorkspaceEngagementFilterSupportsReportFilters(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	userID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()
	namedParams := map[string]any{}

	filter := buildWorkspaceEngagementFilter(reports.ReportFilters{
		TeamIDs:      []uuid.UUID{teamID},
		AssigneeIDs:  []uuid.UUID{userID},
		SprintIDs:    []uuid.UUID{sprintID},
		ObjectiveIDs: []uuid.UUID{objectiveID},
	}, namedParams)

	expectedFragments := []string{
		"wae.team_id = ANY(CAST(:team_ids AS uuid[]))",
		"wae.user_id = ANY(CAST(:assignee_ids AS uuid[]))",
		"wae.sprint_id = ANY(CAST(:sprint_ids AS uuid[]))",
		"wae.objective_id = ANY(CAST(:objective_ids AS uuid[]))",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(filter, fragment) {
			t.Fatalf("expected engagement filter to contain %q, got %q", fragment, filter)
		}
	}
}

func TestReportUUIDArrayFiltersBindWithoutColonSyntax(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	namedParams := map[string]any{}
	filter := buildUUIDArrayFilter("st.team_id", "team_ids", []uuid.UUID{teamID}, namedParams)
	query := "SELECT 1 WHERE TRUE" + filter

	bound, _, err := sqlx.BindNamed(sqlx.DOLLAR, query, namedParams)
	if err != nil {
		t.Fatalf("expected report filter to bind: %v", err)
	}
	if strings.Contains(bound, ":") {
		t.Fatalf("expected bound query not to contain colon syntax, got %q", bound)
	}
	if !strings.Contains(bound, "st.team_id = ANY(CAST($1 AS uuid[]))") {
		t.Fatalf("expected typed pgx array binding, got %q", bound)
	}
}
