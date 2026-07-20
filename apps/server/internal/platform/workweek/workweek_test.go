package workweek

import (
	"reflect"
	"testing"
	"time"
)

func TestNormalizeFallsBackToMondayThroughFriday(t *testing.T) {
	if got := Normalize(nil); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("expected default working days, got %v", got)
	}
}

func TestCountInclusiveSupportsCustomWorkweeks(t *testing.T) {
	start := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC) // Monday
	end := time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)   // Sunday

	if got := CountInclusive(start, end, []int{7, 1, 2, 3, 4}); got != 5 {
		t.Fatalf("expected five working days, got %d", got)
	}
	if IsWorkingDay(time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC), []int{7, 1, 2, 3, 4}) {
		t.Fatal("expected Friday to be excluded from the custom workweek")
	}
}
