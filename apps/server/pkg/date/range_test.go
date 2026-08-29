package date

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRangeFromQueryAtUsesDeterministicDefaults(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 24, 14, 30, 0, 0, time.FixedZone("test", 2*60*60))
	start, end, err := RangeFromQueryAt(url.Values{}, 30, now)
	if err != nil {
		t.Fatalf("parse default range: %v", err)
	}
	if want := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %v, want %v", start, want)
	}
	if want := time.Date(2026, 6, 24, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC); !end.Equal(want) {
		t.Fatalf("end = %v, want %v", end, want)
	}
}

func TestRangeFromQueryAtParsesExplicitDates(t *testing.T) {
	t.Parallel()

	start, end, err := RangeFromQueryAt(url.Values{
		"startDate": {"2026-06-01"},
		"endDate":   {"2026-06-03"},
	}, 30, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parse range: %v", err)
	}
	if start.Format(time.DateOnly) != "2026-06-01" || end.Format(time.DateOnly) != "2026-06-03" {
		t.Fatalf("range = %v through %v", start, end)
	}
}

func TestRangeFromQueryAtRejectsAmbiguousOrUnsafeInput(t *testing.T) {
	t.Parallel()

	sensitiveValue := "sensitive-date-value"
	for name, query := range map[string]url.Values{
		"repeated": {"startDate": {"2026-06-01", sensitiveValue}},
		"oversized": {"startDate": {
			strings.Repeat("x", maximumDateQueryParameterBytes+1),
		}},
		"invalid":   {"startDate": {sensitiveValue}},
		"backwards": {"startDate": {"2026-06-03"}, "endDate": {"2026-06-01"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := RangeFromQueryAt(query, 30, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC))
			if !errors.Is(err, ErrInvalidRangeQuery) || strings.Contains(err.Error(), sensitiveValue) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
