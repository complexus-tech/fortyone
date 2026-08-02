package objectives

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasNotifiableObjectiveUpdate(t *testing.T) {
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"health": "At Risk"}))
	require.True(t, hasNotifiableObjectiveUpdate(map[string]any{"lead_user_id": "user-id"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"description": "draft"}))
	require.False(t, hasNotifiableObjectiveUpdate(map[string]any{"name": "autosaved title"}))
}
