package keyresults

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasNotifiableKeyResultUpdate(t *testing.T) {
	require.False(t, hasNotifiableKeyResultUpdate([]string{"current_value"}))
	require.True(t, hasNotifiableKeyResultUpdate([]string{"contributors"}))
	require.False(t, hasNotifiableKeyResultUpdate([]string{"name"}))
}
