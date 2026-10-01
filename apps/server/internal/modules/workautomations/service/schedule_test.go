package workautomations

import (
	"encoding/json"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
)

func TestCalendarRecurrenceFollowsTimezoneDSTAndMonthEnds(t *testing.T) {
	tests := []struct {
		name, frequency, timezone, start, clock, after, want string
		weekday, monthDay                                    int
	}{
		{"daily local time", "daily", "Africa/Harare", "2026-10-01", "09:00", "2026-10-01T07:01:00Z", "2026-10-02T07:00:00Z", 0, 1},
		{"weekly local calendar", "weekly", "Africa/Harare", "2026-10-01", "09:00", "2026-10-01T10:00:00Z", "2026-10-02T07:00:00Z", 5, 1},
		{"monthly short month", "monthly", "UTC", "2026-01-01", "09:00", "2026-01-31T10:00:00Z", "2026-02-28T09:00:00Z", 0, 31},
		{"monthly leap year", "monthly", "UTC", "2028-01-01", "09:00", "2028-01-31T10:00:00Z", "2028-02-29T09:00:00Z", 0, 31},
		{"spring missing time", "daily", "America/New_York", "2026-01-01", "02:30", "2026-03-07T16:00:00Z", "2026-03-08T07:00:00Z", 0, 1},
		{"autumn first occurrence", "daily", "America/New_York", "2026-01-01", "01:30", "2026-10-31T16:00:00Z", "2026-11-01T05:30:00Z", 0, 1},
		{"autumn never twice", "daily", "America/New_York", "2026-01-01", "01:30", "2026-11-01T05:45:00Z", "2026-11-02T06:30:00Z", 0, 1},
		{"far future start", "daily", "UTC", "2029-01-01", "09:00", "2026-10-01T10:00:00Z", "2029-01-01T09:00:00Z", 0, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			after, _ := time.Parse(time.RFC3339, test.after)
			next, err := NextOccurrence(domain.Schedule{Frequency: test.frequency, Timezone: test.timezone, StartsOn: test.start, LocalTime: test.clock, Weekday: test.weekday, MonthDay: test.monthDay}, after)
			if err != nil || next.Format(time.RFC3339) != test.want {
				t.Fatalf("next %s %v, want %s", next, err, test.want)
			}
		})
	}
}

func TestAutomationRejectsUnknownFieldsConflictingActionsAndUnboundedDrafts(t *testing.T) {
	for _, raw := range []string{
		`{"version":2,"trigger":"story.created","conditions":{},"actions":{"priority":"High"}}`,
		`{"version":1,"trigger":"story.created","conditions":{"secret":"private-value"},"actions":{"priority":"High"}}`,
		`{"version":1,"trigger":"story.deleted","conditions":{},"actions":{"priority":"High"}}`,
		`{"version":1,"trigger":"story.created","conditions":{},"actions":{}}`,
		`{"version":1,"trigger":"story.created","conditions":{},"actions":{"assigneeId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","clearAssignee":true}}`,
	} {
		if _, err := validateRule(json.RawMessage(raw)); err == nil {
			t.Fatalf("invalid rule accepted: %s", raw)
		}
	}
	for _, schedule := range []domain.Schedule{
		{Frequency: "hourly", Timezone: "UTC", StartsOn: "2026-10-01", LocalTime: "09:00", MonthDay: 1},
		{Frequency: "daily", Timezone: "invalid/zone", StartsOn: "2026-10-01", LocalTime: "09:00", MonthDay: 1},
		{Frequency: "daily", Timezone: "UTC", StartsOn: "2026-02-31", LocalTime: "09:00", MonthDay: 1},
		{Frequency: "daily", Timezone: "UTC", StartsOn: "2026-10-01", LocalTime: "25:00", MonthDay: 1},
	} {
		if validateSchedule(schedule) == nil {
			t.Fatalf("invalid schedule accepted: %+v", schedule)
		}
	}
	if validateDraft(domain.Draft{Title: "Title", Priority: "High", MinimumFocusBlockMinutes: ptr(60), EstimatedDurationMinutes: ptr(30)}) == nil {
		t.Fatal("invalid time contract accepted")
	}
}
func ptr(value int) *int { return &value }
