package domain

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/workweek"
)

// BurndownChange is the typed daily aggregate returned by persistence before
// the domain calculates actual and ideal remaining work.
type BurndownChange struct {
	Date             time.Time
	ScopeDelta       int
	CompletionDelta  int
	InitialStories   int
	InitialCompleted int
}

// CalculateOverview derives progress against the workspace working week.
func CalculateOverview(sprint Sprint, breakdown StoryBreakdown, workingDays []int, now time.Time) Overview {
	completionPercentage := 0
	if breakdown.Total > 0 {
		completionPercentage = (breakdown.Completed * 100) / breakdown.Total
	}

	totalDays := workweek.CountInclusive(sprint.StartDate, sprint.EndDate, workingDays)
	daysElapsed := 0
	daysRemaining := totalDays
	if !now.Before(sprint.StartDate) {
		daysElapsed = workweek.CountInclusive(sprint.StartDate, now.AddDate(0, 0, -1), workingDays)
		daysRemaining = workweek.CountInclusive(now, sprint.EndDate, workingDays)
	}
	if now.After(sprint.EndDate) {
		daysRemaining = 0
	}

	status := "on_track"
	switch {
	case now.After(sprint.EndDate):
		status = "completed"
	case now.Before(sprint.StartDate):
		status = "not_started"
	case totalDays > 0:
		timeProgress := float64(daysElapsed) / float64(totalDays)
		workProgress := float64(completionPercentage) / 100
		if workProgress < timeProgress-0.2 {
			status = "behind"
		} else if workProgress < timeProgress-0.1 {
			status = "at_risk"
		}
	}

	return Overview{
		CompletionPercentage: completionPercentage,
		DaysElapsed:          daysElapsed,
		DaysRemaining:        daysRemaining,
		Status:               status,
	}
}

// BuildBurndown applies scope and completion changes in date order. Future
// completion events are not counted in actual remaining work.
func BuildBurndown(
	changes []BurndownChange,
	startDate, endDate time.Time,
	workingDays []int,
	now time.Time,
) []BurndownDataPoint {
	if len(changes) == 0 {
		return []BurndownDataPoint{}
	}

	currentTotal := changes[0].InitialStories
	cumulativeCompleted := changes[0].InitialCompleted
	totalWorkingDays := workweek.CountInclusive(startDate, endDate, workingDays)
	result := make([]BurndownDataPoint, 0, len(changes))

	for index, change := range changes {
		currentTotal += change.ScopeDelta
		if !change.Date.After(now) {
			cumulativeCompleted += change.CompletionDelta
		}

		ideal := calculateIdealRemaining(
			currentTotal,
			changes[0].InitialStories,
			index,
			startDate,
			change.Date,
			totalWorkingDays,
			workingDays,
		)
		actual := max(currentTotal-cumulativeCompleted, 0)
		result = append(result, BurndownDataPoint{Date: change.Date, Remaining: actual, Ideal: max(ideal, 0)})
	}
	return result
}

func calculateIdealRemaining(
	currentTotal, initialStories, dayIndex int,
	startDate, currentDate time.Time,
	totalWorkingDays int,
	workingDays []int,
) int {
	if dayIndex == 0 {
		return max(initialStories, 0)
	}
	if totalWorkingDays <= 1 {
		return 0
	}
	workingDaysReached := workweek.CountInclusive(startDate, currentDate, workingDays)
	workingDayTransitions := max(workingDaysReached-1, 0)
	remainingTransitions := max(totalWorkingDays-1-workingDayTransitions, 0)
	return max(int(float64(currentTotal)*float64(remainingTransitions)/float64(totalWorkingDays-1)), 0)
}
