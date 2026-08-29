package stories

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEstimateLabelDefaultsToTshirt(t *testing.T) {
	value := int16(2)

	label := EstimateLabelFromValue("", &value)
	if label == nil || *label != "S" {
		t.Fatalf("expected default estimate scheme to label 2 as S, got %v", label)
	}
}

func TestValidateEstimateSchemeOnlyAllowsComplexitySchemes(t *testing.T) {
	for _, scheme := range []string{"points", "tshirt", ""} {
		if err := ValidateEstimateScheme(scheme); err != nil {
			t.Fatalf("expected %q to be valid, got %v", scheme, err)
		}
	}

	for _, scheme := range []string{"hours", "ideal_days"} {
		if err := ValidateEstimateScheme(scheme); err == nil {
			t.Fatalf("expected legacy time-based scheme %q to be rejected", scheme)
		}
	}
}

func TestNormalizeEstimateUpdateValueRejectsOverflowAndFractions(t *testing.T) {
	t.Parallel()

	tests := []any{
		int(1 << 15),
		int(-1<<15 - 1),
		float64(1 << 15),
		float64(-1<<15 - 1),
		1.5,
	}
	for _, value := range tests {
		_, err := normalizeEstimateUpdateValue(value)
		require.Error(t, err)
	}
}

func TestNormalizeEstimateUpdateValueAcceptsExactSupportedRepresentations(t *testing.T) {
	t.Parallel()

	for _, value := range []any{int(5), int16(5), float64(5)} {
		normalized, err := normalizeEstimateUpdateValue(value)
		require.NoError(t, err)
		require.NotNil(t, normalized)
		require.Equal(t, int16(5), *normalized)
	}
}
