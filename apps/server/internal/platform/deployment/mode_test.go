package deployment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  Mode
	}{
		{name: "development", value: "development", want: Development},
		{name: "test", value: " TEST ", want: Test},
		{name: "staging", value: "staging", want: Staging},
		{name: "production", value: "PRODUCTION", want: Production},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseRejectsUnknownMode(t *testing.T) {
	t.Parallel()

	_, err := Parse("prod")
	require.ErrorContains(t, err, "APP_ENVIRONMENT")
}
