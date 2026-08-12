package objectives

import (
	"testing"

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

func TestHasNotifiableObjectiveUpdate(t *testing.T) {
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"health": "At Risk"}))
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"lead_user_id": "user-id"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"description": "draft"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"name": "autosaved title"}))
}
