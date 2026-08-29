package reports

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeReportFiltersBindsActorAndDeduplicatesIDs(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24 * time.Hour)

	filters, err := normalizeReportFilters(reportTestContext(actorID), workspaceID, ReportFilters{
		TeamIDs:   []uuid.UUID{teamID, teamID},
		StartDate: &startDate,
		EndDate:   &endDate,
	}, true)
	if err != nil {
		t.Fatalf("normalize report filters: %v", err)
	}
	if filters.ActorID != actorID {
		t.Fatalf("expected actor %s, got %s", actorID, filters.ActorID)
	}
	if len(filters.TeamIDs) != 1 || filters.TeamIDs[0] != teamID {
		t.Fatalf("expected one deduplicated team, got %#v", filters.TeamIDs)
	}
}

func TestNormalizeReportFiltersRejectsUnboundedInputs(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	workspaceID := uuid.New()
	tooManyTeams := make([]uuid.UUID, maxReportFilterIDs+1)
	for i := range tooManyTeams {
		tooManyTeams[i] = uuid.New()
	}

	_, err := normalizeReportFilters(reportTestContext(actorID), workspaceID, ReportFilters{
		TeamIDs: tooManyTeams,
	}, false)
	if !errors.Is(err, ErrInvalidReportFilters) {
		t.Fatalf("expected invalid filters, got %v", err)
	}
}

func TestNormalizeReportFiltersRejectsExcessiveDateRange(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(maxReportDateRange + time.Hour)
	_, err := normalizeReportFilters(reportTestContext(uuid.New()), uuid.New(), ReportFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}, true)
	if !errors.Is(err, ErrInvalidReportFilters) {
		t.Fatalf("expected invalid filters, got %v", err)
	}
}

func TestNormalizeReportFiltersRequiresAuthenticatedActor(t *testing.T) {
	t.Parallel()

	_, err := normalizeReportFilters(t.Context(), uuid.New(), ReportFilters{}, false)
	if !errors.Is(err, ErrReportsAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}
