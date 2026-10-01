package workautomations

import (
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
)

func validateSchedule(schedule domain.Schedule) error {
	if schedule.Frequency != "daily" && schedule.Frequency != "weekly" && schedule.Frequency != "monthly" {
		return domain.ErrInvalidInput
	}
	if schedule.Weekday < 0 || schedule.Weekday > 6 || schedule.MonthDay < 1 || schedule.MonthDay > 31 || len(schedule.Timezone) > 100 {
		return domain.ErrInvalidInput
	}
	if _, err := time.LoadLocation(schedule.Timezone); err != nil || schedule.Timezone == "" {
		return domain.ErrInvalidInput
	}
	if _, err := time.Parse("2006-01-02", schedule.StartsOn); err != nil {
		return domain.ErrInvalidInput
	}
	if _, err := time.Parse("15:04", schedule.LocalTime); err != nil {
		return domain.ErrInvalidInput
	}
	return nil
}

// NextOccurrence follows calendar dates in the selected timezone. Monthly
// schedules clamp to the final day of short months. An ambiguous DST time runs
// once at its first occurrence; a missing wall time runs at the first valid
// minute after the gap. Paused or delayed jobs resume with the next future date,
// so recovery does not flood a team with historical tasks.
func NextOccurrence(schedule domain.Schedule, after time.Time) (time.Time, error) {
	if err := validateSchedule(schedule); err != nil {
		return time.Time{}, err
	}
	location, _ := time.LoadLocation(schedule.Timezone)
	start, _ := time.Parse("2006-01-02", schedule.StartsOn)
	clock, _ := time.Parse("15:04", schedule.LocalTime)
	local := after.In(location)
	date := time.Date(local.Year(), local.Month(), local.Day(), 12, 0, 0, 0, location)
	if date.Format("2006-01-02") < schedule.StartsOn {
		date = time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, location)
	}
	for day := 0; day < 400; day++ {
		if date.Format("2006-01-02") >= start.Format("2006-01-02") && matchesDate(schedule, date) {
			occurrence, ok := wallTime(date, clock.Hour()*60+clock.Minute(), location)
			if ok && occurrence.After(after) {
				return occurrence.UTC(), nil
			}
		}
		date = date.AddDate(0, 0, 1)
	}
	return time.Time{}, fmt.Errorf("%w: no next occurrence", domain.ErrInvalidInput)
}

func matchesDate(schedule domain.Schedule, date time.Time) bool {
	switch schedule.Frequency {
	case "weekly":
		return int(date.Weekday()) == schedule.Weekday
	case "monthly":
		last := time.Date(date.Year(), date.Month()+1, 0, 12, 0, 0, 0, date.Location()).Day()
		day := schedule.MonthDay
		if day > last {
			day = last
		}
		return date.Day() == day
	default:
		return true
	}
}

func wallTime(date time.Time, desiredMinutes int, location *time.Location) (time.Time, bool) {
	// Scanning UTC minutes makes the repeated and missing local hours explicit.
	first := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location).UTC().Add(-18 * time.Hour)
	for minute := 0; minute < 36*60; minute++ {
		candidate := first.Add(time.Duration(minute) * time.Minute)
		local := candidate.In(location)
		if local.Year() == date.Year() && local.Month() == date.Month() && local.Day() == date.Day() && local.Hour()*60+local.Minute() >= desiredMinutes {
			return candidate, true
		}
	}
	return time.Time{}, false
}
