package reportshttp

import (
	"net/url"
	"testing"
	"time"
)

func TestCalendarEndDateIncludesTheWholeDay(t *testing.T) {
	t.Parallel()

	for _, parser := range []string{"report", "workload"} {
		t.Run(parser, func(t *testing.T) {
			t.Parallel()
			query := url.Values{"startDate": {"2026-10-01"}, "endDate": {"2026-10-01"}}
			filters, err := parseReportFilters(query, time.Now())
			if parser == "workload" {
				filters, err = parseWorkloadAnalysisFilters(query)
			}
			if err != nil {
				t.Fatal(err)
			}
			start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 0, 1).Add(-time.Nanosecond)
			if !filters.StartDate.Equal(start) || !filters.EndDate.Equal(end) {
				t.Fatalf("calendar range = %v through %v, want %v through %v", filters.StartDate, filters.EndDate, start, end)
			}
		})
	}
}

func TestExplicitEndTimestampKeepsItsBoundary(t *testing.T) {
	t.Parallel()
	query := url.Values{"endDate": {"2026-10-01T12:30:00Z"}}
	filters, err := parseReportFilters(query, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 10, 1, 12, 30, 0, 0, time.UTC)
	if !filters.EndDate.Equal(want) {
		t.Fatalf("end = %v, want %v", filters.EndDate, want)
	}
}
