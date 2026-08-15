package stories

import "testing"

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
