package reportshttp

import (
	"net/url"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestReportFiltersDecodeCommaSeparatedQueryValues(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	secondTeamID := uuid.New()
	assigneeID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()

	var queryFilters AppReportFilterQuery
	query, err := web.GetFilters(url.Values{
		"teamIds":      []string{teamID.String() + "," + secondTeamID.String()},
		"assigneeIds":  []string{assigneeID.String()},
		"sprintIds":    []string{sprintID.String()},
		"objectiveIds": []string{objectiveID.String()},
		"startDate":    []string{"2026-06-01"},
		"endDate":      []string{"2026-06-24"},
	}, &queryFilters)
	if err != nil {
		t.Fatalf("expected query decoding to succeed, got %v", err)
	}

	got, err := parseReportFilters(query)
	if err != nil {
		t.Fatalf("expected report filter parsing to succeed, got %v", err)
	}
	if got.StartDate == nil || got.StartDate.Format(time.DateOnly) != "2026-06-01" {
		t.Fatalf("expected start date 2026-06-01, got %#v", got.StartDate)
	}
	if got.EndDate == nil || got.EndDate.Format(time.DateOnly) != "2026-06-24" {
		t.Fatalf("expected end date 2026-06-24, got %#v", got.EndDate)
	}
	if len(got.TeamIDs) != 2 || got.TeamIDs[0] != teamID || got.TeamIDs[1] != secondTeamID {
		t.Fatalf("expected team ids %s and %s, got %#v", teamID, secondTeamID, got.TeamIDs)
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
