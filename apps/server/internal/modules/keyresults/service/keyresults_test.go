package keyresults

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasNotifiableKeyResultUpdate(t *testing.T) {
	require.True(t, hasNotifiableKeyResultUpdate(map[string]any{"current_value": 75}))
	require.True(t, hasNotifiableKeyResultUpdate(map[string]any{"contributors": []string{"user-id"}}))
	require.False(t, hasNotifiableKeyResultUpdate(map[string]any{"name": "autosaved title"}))
}
