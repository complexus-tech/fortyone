package reportshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
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

	query := url.Values{
		"teamIds":      []string{teamID.String() + "," + secondTeamID.String()},
		"assigneeIds":  []string{assigneeID.String()},
		"sprintIds":    []string{sprintID.String()},
		"objectiveIds": []string{objectiveID.String()},
		"startDate":    []string{"2026-06-01"},
		"endDate":      []string{"2026-06-24"},
	}

	got, err := parseReportFilters(query, time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC))
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

	_, err := parseReportFilters(url.Values{
		"startDate": {"not-a-date"},
		"endDate":   {"2026-06-24"},
	}, time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC))

	if err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestParseReportFiltersRejectsInvalidIdentifiers(t *testing.T) {
	t.Parallel()

	for name, query := range map[string]url.Values{
		"only invalid identifier": {"teamIds": {"not-a-uuid"}},
		"mixed identifiers":       {"teamIds": {uuid.NewString() + ",not-a-uuid"}},
		"nil UUID":                {"teamIds": {uuid.Nil.String()}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseReportFilters(query, time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)); err == nil {
				t.Fatal("expected invalid report filter identifier error")
			}
		})
	}
}

func TestParseStatsFiltersRejectsInvalidIdentifiers(t *testing.T) {
	t.Parallel()

	if _, err := parseStatsFilters(url.Values{"teamId": {"not-a-uuid"}}); err == nil {
		t.Fatal("expected invalid statistics filter identifier error")
	}
}

func TestParseReportFiltersUsesDeterministicDefaultWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 24, 14, 30, 0, 0, time.FixedZone("test", 2*60*60))
	filters, err := parseReportFilters(url.Values{}, now)
	if err != nil {
		t.Fatalf("parse default filters: %v", err)
	}
	wantEnd := now.UTC()
	wantStart := wantEnd.AddDate(0, 0, -defaultReportWindowDays)
	if filters.StartDate == nil || !filters.StartDate.Equal(wantStart) {
		t.Fatalf("start date = %v, want %v", filters.StartDate, wantStart)
	}
	if filters.EndDate == nil || !filters.EndDate.Equal(wantEnd) {
		t.Fatalf("end date = %v, want %v", filters.EndDate, wantEnd)
	}
}

func TestReportFilterParsingRejectsRepeatedParametersWithoutEchoingInput(t *testing.T) {
	t.Parallel()

	sensitiveValue := "sensitive-filter-value"
	_, err := parseReportFilters(url.Values{
		"teamIds": {uuid.NewString(), sensitiveValue},
	}, time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, web.ErrRepeatedQueryParameter) || strings.Contains(err.Error(), sensitiveValue) {
		t.Fatalf("repeated filter error = %v", err)
	}
}

func TestRespondReportErrorPreservesAuthorizationAndValidationSemantics(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "access denied", err: fmt.Errorf("wrapped: %w", reports.ErrReportsAccessDenied), wantStatus: http.StatusForbidden},
		{name: "invalid filters", err: reports.ErrInvalidReportFilters, wantStatus: http.StatusBadRequest},
		{name: "invalid event", err: reports.ErrInvalidWorkspaceAnalyticsEvent, wantStatus: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			if err := respondReportError(context.Background(), recorder, test.err); err != nil {
				t.Fatalf("respond report error: %v", err)
			}
			if recorder.Code != test.wantStatus {
				t.Fatalf("response status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
