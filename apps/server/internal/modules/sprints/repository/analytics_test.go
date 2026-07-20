package sprintsrepository

import (
	"testing"
	"time"

	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
)

func TestCalculateOverviewAtUsesConfiguredWorkingDays(t *testing.T) {
	sprint := sprints.CoreSprint{
		StartDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),   // Monday
		EndDate:   time.Date(2026, 6, 21, 23, 59, 0, 0, time.UTC), // Sunday
	}
	now := time.Date(2026, 6, 19, 9, 0, 0, 0, time.UTC) // Friday

	overview := calculateOverviewAt(sprint, sprints.CoreStoryBreakdown{}, []int{1, 2, 3, 4, 5}, now)
	if overview.DaysElapsed != 4 {
		t.Fatalf("expected four elapsed working days, got %d", overview.DaysElapsed)
	}
	if overview.DaysRemaining != 1 {
		t.Fatalf("expected one remaining working day, got %d", overview.DaysRemaining)
	}
}

func TestCalculateOverviewAtSupportsSundayThroughThursday(t *testing.T) {
	sprint := sprints.CoreSprint{
		StartDate: time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC),   // Sunday
		EndDate:   time.Date(2026, 6, 20, 23, 59, 0, 0, time.UTC), // Saturday
	}
	now := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC) // Thursday

	overview := calculateOverviewAt(sprint, sprints.CoreStoryBreakdown{}, []int{7, 1, 2, 3, 4}, now)
	if overview.DaysElapsed != 4 || overview.DaysRemaining != 1 {
		t.Fatalf("expected 4 elapsed and 1 remaining working day, got %d and %d", overview.DaysElapsed, overview.DaysRemaining)
	}
}

func TestCalculateIdealRemainingStaysFlatOnNonWorkingDays(t *testing.T) {
	start := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC) // Monday
	workingDays := []int{1, 2, 3, 4, 5}
	totalWorkingDays := 10

	friday := calculateIdealRemaining(10, 10, 4, start, start.AddDate(0, 0, 4), totalWorkingDays, workingDays)
	saturday := calculateIdealRemaining(10, 10, 5, start, start.AddDate(0, 0, 5), totalWorkingDays, workingDays)
	sunday := calculateIdealRemaining(10, 10, 6, start, start.AddDate(0, 0, 6), totalWorkingDays, workingDays)

	if friday == 0 || saturday != friday || sunday != friday {
		t.Fatalf("expected ideal line to remain flat through the weekend, got Friday=%d Saturday=%d Sunday=%d", friday, saturday, sunday)
	}
}
