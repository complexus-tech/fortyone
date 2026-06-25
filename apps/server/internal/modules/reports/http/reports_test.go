package reportshttp

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseReportFiltersAcceptsDateOnlyValuesAndAssignees(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	assigneeID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()

	got, err := parseReportFilters(map[string]interface{}{
		"teamIds":      teamID.String(),
		"assigneeIds":  assigneeID.String(),
		"sprintIds":    sprintID.String(),
		"objectiveIds": objectiveID.String(),
		"startDate":    "2026-06-01",
		"endDate":      "2026-06-24",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.StartDate == nil || got.StartDate.Format(time.DateOnly) != "2026-06-01" {
		t.Fatalf("expected start date 2026-06-01, got %#v", got.StartDate)
	}
	if got.EndDate == nil || got.EndDate.Format(time.DateOnly) != "2026-06-24" {
		t.Fatalf("expected end date 2026-06-24, got %#v", got.EndDate)
	}
	if len(got.TeamIDs) != 1 || got.TeamIDs[0] != teamID {
		t.Fatalf("expected team id %s, got %#v", teamID, got.TeamIDs)
	}
	if len(got.AssigneeIDs) != 1 || got.AssigneeIDs[0] != assigneeID {
		t.Fatalf("expected assignee id %s, got %#v", assigneeID, got.AssigneeIDs)
	}
	if len(got.SprintIDs) != 1 || got.SprintIDs[0] != sprintID {
		t.Fatalf("expected sprint id %s, got %#v", sprintID, got.SprintIDs)
	}
	if len(got.ObjectiveIDs) != 1 || got.ObjectiveIDs[0] != objectiveID {
		t.Fatalf("expected objective id %s, got %#v", objectiveID, got.ObjectiveIDs)
	}
}

func TestParseReportFiltersRejectsInvalidDates(t *testing.T) {
	t.Parallel()

	_, err := parseReportFilters(map[string]interface{}{
		"startDate": "not-a-date",
		"endDate":   "2026-06-24",
	})

	if err == nil {
		t.Fatal("expected invalid date error")
	}
}
