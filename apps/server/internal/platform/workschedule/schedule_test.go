package workschedule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveUsesIndependentUserOverrides(t *testing.T) {
	defaults := Schedule{WorkingDays: []int{1, 2, 3, 4, 5}, StartMinute: 9 * 60, EndMinute: 17 * 60}
	start := 8 * 60
	end := 16 * 60

	resolved := Resolve(defaults, []int{1, 2, 3, 4}, &start, &end)

	require.Equal(t, []int{1, 2, 3, 4}, resolved.WorkingDays)
	require.Equal(t, start, resolved.StartMinute)
	require.Equal(t, end, resolved.EndMinute)
}

func TestResolveFallsBackToWorkspaceDefaults(t *testing.T) {
	defaults := Schedule{WorkingDays: []int{2, 3, 4, 5, 6}, StartMinute: 7 * 60, EndMinute: 15 * 60}
	require.True(t, Equal(defaults, Resolve(defaults, nil, nil, nil)))
}

func TestValidateHoursRejectsOvernightSchedules(t *testing.T) {
	require.ErrorIs(t, ValidateHours(17*60, 9*60), ErrInvalidWorkingHours)
}
