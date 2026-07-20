package jobs

import (
	"testing"
	"time"
)

func TestCalculateSprintStartDateAfter(t *testing.T) {
	t.Parallel()

	anchor, err := time.Parse("2006-01-02", "2026-08-02")
	if err != nil {
		t.Fatalf("parse anchor: %v", err)
	}

	got := calculateSprintStartDateAfter(anchor, "Monday")
	if got.Format("2006-01-02") != "2026-08-03" {
		t.Fatalf("calculateSprintStartDateAfter() = %s, want 2026-08-03", got.Format("2006-01-02"))
	}
}
