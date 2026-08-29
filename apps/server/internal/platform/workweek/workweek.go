package workweek

import "time"

var defaultWorkingDays = []int{1, 2, 3, 4, 5}

// DefaultWorkingDays returns the default ISO weekdays (Monday through Friday).
func DefaultWorkingDays() []int {
	return append([]int(nil), defaultWorkingDays...)
}

// Normalize returns a valid, de-duplicated set of ISO weekdays. Invalid or
// empty input falls back to Monday through Friday so callers remain safe while
// older settings are being migrated.
func Normalize(days []int) []int {
	seen := make(map[int]struct{}, len(days))
	normalized := make([]int, 0, len(days))
	for _, day := range days {
		if day < 1 || day > 7 {
			continue
		}
		if _, exists := seen[day]; exists {
			continue
		}
		seen[day] = struct{}{}
		normalized = append(normalized, day)
	}
	if len(normalized) == 0 {
		return DefaultWorkingDays()
	}
	return normalized
}

// IsWorkingDay reports whether a date belongs to the configured ISO workweek.
func IsWorkingDay(value time.Time, days []int) bool {
	isoWeekday := int(value.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}
	for _, day := range Normalize(days) {
		if day == isoWeekday {
			return true
		}
	}
	return false
}

// CountInclusive counts configured working days between two calendar dates.
func CountInclusive(start, end time.Time, days []int) int {
	if end.Before(start) {
		return 0
	}

	location := start.Location()
	cursor := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, location)
	last := time.Date(end.In(location).Year(), end.In(location).Month(), end.In(location).Day(), 0, 0, 0, 0, location)
	count := 0
	for !cursor.After(last) {
		if IsWorkingDay(cursor, days) {
			count++
		}
		cursor = cursor.AddDate(0, 0, 1)
	}
	return count
}
