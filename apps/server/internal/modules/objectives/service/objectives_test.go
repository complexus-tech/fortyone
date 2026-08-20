package objectives

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestObjectiveExternalUpdatesAlreadyAppliedIsExplicitAndIdempotent(t *testing.T) {
	t.Parallel()

	health := HealthOnTrack
	objective := CoreObjective{Health: &health}

	require.True(t, objectiveExternalUpdatesAlreadyApplied(objective, map[string]any{"health": "On Track"}))
	require.False(t, objectiveExternalUpdatesAlreadyApplied(objective, map[string]any{"health": "At Risk"}))
	require.False(t, objectiveExternalUpdatesAlreadyApplied(objective, map[string]any{"status": "Done"}))
}

func TestApplyScheduleForecastPreservesTargetAndDerivesRisk(t *testing.T) {
	t.Parallel()

	target := time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC)
	forecast := time.Date(2026, time.June, 29, 0, 0, 0, 0, time.UTC)
	objective := CoreObjective{
		EndDate:         &target,
		ForecastEndDate: &forecast,
	}

	objective.ApplyScheduleForecast()

	require.Equal(t, ScheduleStatusAtRisk, objective.ScheduleStatus)
	require.Equal(t, 4, objective.ForecastDaysDelta)
	require.Equal(t, target, *objective.EndDate)
}

func TestApplyScheduleForecastHandlesIncompletePlanning(t *testing.T) {
	t.Parallel()

	target := time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC)
	forecast := time.Date(2026, time.June, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		objective CoreObjective
		expected  ObjectiveScheduleStatus
	}{
		{name: "missing target", objective: CoreObjective{ForecastEndDate: &forecast}, expected: ScheduleStatusNoTarget},
		{name: "missing forecast", objective: CoreObjective{EndDate: &target}, expected: ScheduleStatusNoSchedule},
		{name: "on track", objective: CoreObjective{EndDate: &target, ForecastEndDate: &forecast}, expected: ScheduleStatusOnTrack},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.objective.ApplyScheduleForecast()
			require.Equal(t, test.expected, test.objective.ScheduleStatus)
		})
	}
}

func TestHasNotifiableObjectiveUpdate(t *testing.T) {
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"health": "At Risk"}))
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"lead_user_id": "user-id"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"description": "draft"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"name": "autosaved title"}))
}
