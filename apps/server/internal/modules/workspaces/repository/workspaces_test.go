package workspacesrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateRandomColorReturnsPaletteValue(t *testing.T) {
	t.Parallel()

	valid := make(map[string]struct{}, len(colors))
	for _, color := range colors {
		valid[color] = struct{}{}
	}

	for range 100 {
		color, err := generateRandomColor()
		require.NoError(t, err)
		_, ok := valid[color]
		require.True(t, ok, "generated color %q is not in the palette", color)
	}
}
