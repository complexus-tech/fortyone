package stories

import "testing"

func TestEstimateDurationMinutesUsesHoursSchemeLabels(t *testing.T) {
	tests := []struct {
		name     string
		value    int16
		expected int
	}{
		{name: "half hour", value: 1, expected: 30},
		{name: "one hour", value: 2, expected: 60},
		{name: "two hours", value: 3, expected: 120},
		{name: "four hours", value: 5, expected: 240},
		{name: "eight hours", value: 8, expected: 480},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateDurationMinutes("hours", &tt.value); got != tt.expected {
				t.Fatalf("expected %d minutes, got %d", tt.expected, got)
			}
		})
	}
}

func TestEstimateDurationMinutesDefaultsToHours(t *testing.T) {
	value := int16(2)

	if got := EstimateDurationMinutes("", &value); got != 60 {
		t.Fatalf("expected default estimate scheme to schedule 2 as 60 minutes, got %d", got)
	}
}

func TestEstimateDurationMinutesKeepsPointHeuristic(t *testing.T) {
	value := int16(2)

	if got := EstimateDurationMinutes("points", &value); got != 240 {
		t.Fatalf("expected 2 points to schedule as 240 minutes, got %d", got)
	}
}
