package teamsettingsrepository

import (
	"testing"
	"time"
)

func TestNextSprintStartAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		anchor   string
		startDay string
		want     string
	}{
		{name: "next day", anchor: "2026-08-02", startDay: "Monday", want: "2026-08-03"},
		{name: "same weekday advances one week", anchor: "2026-08-03", startDay: "Monday", want: "2026-08-10"},
		{name: "realigns after active sprint", anchor: "2026-08-05", startDay: "Monday", want: "2026-08-10"},
		{name: "invalid day defaults to Monday", anchor: "2026-08-05", startDay: "invalid", want: "2026-08-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			anchor := mustParseScheduleDate(t, tt.anchor)
			got := nextSprintStartAfter(anchor, tt.startDay)
			if got.Format("2006-01-02") != tt.want {
				t.Fatalf("nextSprintStartAfter() = %s, want %s", got.Format("2006-01-02"), tt.want)
			}
		})
	}
}

func TestFindScheduleConflict(t *testing.T) {
	t.Parallel()

	custom := []scheduledSprint{{
		Name:      "Launch sprint",
		StartDate: mustParseScheduleDate(t, "2026-08-10"),
		EndDate:   mustParseScheduleDate(t, "2026-08-16"),
	}}

	if _, ok := findScheduleConflict(
		mustParseScheduleDate(t, "2026-08-03"),
		mustParseScheduleDate(t, "2026-08-09"),
		custom,
	); ok {
		t.Fatal("adjacent sprint must not be treated as an overlap")
	}

	conflict, ok := findScheduleConflict(
		mustParseScheduleDate(t, "2026-08-09"),
		mustParseScheduleDate(t, "2026-08-15"),
		custom,
	)
	if !ok || conflict.Name != "Launch sprint" {
		t.Fatal("overlapping custom sprint was not detected")
	}
}

func mustParseScheduleDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse test date: %v", err)
	}
	return parsed
}
