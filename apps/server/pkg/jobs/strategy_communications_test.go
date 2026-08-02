package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextQuarterStart(t *testing.T) {
	tests := []struct {
		name     string
		current  time.Time
		expected time.Time
	}{
		{
			name:     "advances within the same year",
			current:  time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC),
			expected: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "rolls Q4 into the next year",
			current:  time.Date(2026, time.December, 20, 9, 0, 0, 0, time.UTC),
			expected: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, nextQuarterStart(test.current))
		})
	}
}

func TestCalendarDaysBetweenIgnoresDaylightSavingTransitions(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	beforeTransition := time.Date(2026, time.March, 7, 9, 0, 0, 0, location)
	afterTransition := time.Date(2026, time.March, 9, 9, 0, 0, 0, location)

	require.Equal(t, 2, calendarDaysBetween(beforeTransition, afterTransition))
}

func TestMissingStrategyElements(t *testing.T) {
	require.Equal(
		t,
		"your ultimate goal, strategic pillars, the next period's objectives",
		missingStrategyElements(strategyFoundation{}),
	)
	require.Equal(
		t,
		"the next period's objectives",
		missingStrategyElements(strategyFoundation{HasUltimateGoal: true, PillarCount: 3}),
	)
}

func TestStrategyCheckInSummaryOmitsZeroSignals(t *testing.T) {
	summary := strategyCheckInSummary(strategyCheckIn{
		AtRiskObjectives: 2,
		StaleKeyResults:  3,
	})

	require.Equal(t, "2 at-risk objectives, 3 stalled key results", summary)
}
