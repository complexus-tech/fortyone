package domain_test

import (
	"testing"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
)

func TestCalculateOverviewUsesConfiguredWorkingDays(t *testing.T) {
	t.Parallel()

	sprint := sprintdomain.Sprint{
		StartDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 6, 21, 23, 59, 0, 0, time.UTC),
	}
	overview := sprintdomain.CalculateOverview(
		sprint,
		sprintdomain.StoryBreakdown{},
		[]int{1, 2, 3, 4, 5},
		time.Date(2026, 6, 19, 9, 0, 0, 0, time.UTC),
	)
	if overview.DaysElapsed != 4 || overview.DaysRemaining != 1 {
		t.Fatalf("overview days = %d/%d, want 4/1", overview.DaysElapsed, overview.DaysRemaining)
	}
}

func TestBuildBurndownStaysFlatOnNonWorkingDays(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	changes := make([]sprintdomain.BurndownChange, 7)
	for index := range changes {
		changes[index] = sprintdomain.BurndownChange{
			Date: start.AddDate(0, 0, index), InitialStories: 10,
		}
	}
	points := sprintdomain.BuildBurndown(changes, start, start.AddDate(0, 0, 13), []int{1, 2, 3, 4, 5}, start.AddDate(0, 0, 7))
	if points[4].Ideal == 0 || points[5].Ideal != points[4].Ideal || points[6].Ideal != points[4].Ideal {
		t.Fatalf("ideal Friday/Saturday/Sunday = %d/%d/%d", points[4].Ideal, points[5].Ideal, points[6].Ideal)
	}
}
