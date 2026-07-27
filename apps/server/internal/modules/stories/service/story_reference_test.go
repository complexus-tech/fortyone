package stories

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStoryRef(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name       string
		input      string
		wantCode   string
		wantID     int
		shouldFail bool
	}{
		{
			name:     "canonical reference",
			input:    "PRD-571",
			wantCode: "PRD",
			wantID:   571,
		},
		{
			name:     "normalizes spaces and case",
			input:    " prd 571 ",
			wantCode: "PRD",
			wantID:   571,
		},
		{
			name:       "rejects missing sequence",
			input:      "PRD",
			shouldFail: true,
		},
		{
			name:       "rejects missing team code",
			input:      "571",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, sequenceID, err := service.parseStoryRef(tt.input)
			if tt.shouldFail {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantCode, code)
			require.Equal(t, tt.wantID, sequenceID)
		})
	}
}
