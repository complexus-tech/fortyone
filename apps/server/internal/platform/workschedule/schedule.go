package workschedule

import (
	"errors"
	"slices"

	"github.com/complexus-tech/projects-api/internal/platform/workweek"
)

const (
	DefaultStartMinute = 9 * 60
	DefaultEndMinute   = 17 * 60
)

var (
	ErrInvalidWorkingDays  = errors.New("working days must contain unique weekday numbers between 1 and 7")
	ErrInvalidWorkingHours = errors.New("working hours must have a valid start before the end")
)

type Schedule struct {
	WorkingDays []int
	StartMinute int
	EndMinute   int
}

func Default() Schedule {
	return Schedule{
		WorkingDays: workweek.DefaultWorkingDays(),
		StartMinute: DefaultStartMinute,
		EndMinute:   DefaultEndMinute,
	}
}

func Normalize(workingDays []int, startMinute, endMinute int) Schedule {
	result := Schedule{
		WorkingDays: workweek.Normalize(workingDays),
		StartMinute: startMinute,
		EndMinute:   endMinute,
	}
	if !ValidHours(result.StartMinute, result.EndMinute) {
		result.StartMinute = DefaultStartMinute
		result.EndMinute = DefaultEndMinute
	}
	return result
}

func Resolve(defaults Schedule, workingDays []int, startMinute, endMinute *int) Schedule {
	resolved := Normalize(defaults.WorkingDays, defaults.StartMinute, defaults.EndMinute)
	if len(workingDays) > 0 {
		resolved.WorkingDays = workweek.Normalize(workingDays)
	}
	if startMinute != nil && endMinute != nil && ValidHours(*startMinute, *endMinute) {
		resolved.StartMinute = *startMinute
		resolved.EndMinute = *endMinute
	}
	return resolved
}

func ValidateWorkingDays(days []int) error {
	if len(days) == 0 || len(days) > 7 {
		return ErrInvalidWorkingDays
	}
	seen := make(map[int]struct{}, len(days))
	for _, day := range days {
		if day < 1 || day > 7 {
			return ErrInvalidWorkingDays
		}
		if _, exists := seen[day]; exists {
			return ErrInvalidWorkingDays
		}
		seen[day] = struct{}{}
	}
	return nil
}

func ValidHours(startMinute, endMinute int) bool {
	return startMinute >= 0 && startMinute < 24*60 && endMinute > 0 && endMinute <= 24*60 && endMinute > startMinute
}

func ValidateHours(startMinute, endMinute int) error {
	if !ValidHours(startMinute, endMinute) {
		return ErrInvalidWorkingHours
	}
	return nil
}

func Equal(left, right Schedule) bool {
	return left.StartMinute == right.StartMinute && left.EndMinute == right.EndMinute && slices.Equal(left.WorkingDays, right.WorkingDays)
}
