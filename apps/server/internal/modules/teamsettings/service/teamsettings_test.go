package teamsettings

import "testing"

func TestValidateEstimateSchemeOnlyAllowsComplexitySchemes(t *testing.T) {
	service := &Service{}
	for _, scheme := range []string{"points", "tshirt"} {
		if err := service.validateEstimationSettingsUpdate(CoreUpdateTeamEstimationSettings{Scheme: &scheme}); err != nil {
			t.Fatalf("expected %q to be valid, got %v", scheme, err)
		}
	}

	for _, scheme := range []string{"hours", "ideal_days"} {
		if err := service.validateEstimationSettingsUpdate(CoreUpdateTeamEstimationSettings{Scheme: &scheme}); err == nil {
			t.Fatalf("expected legacy time-based scheme %q to be rejected", scheme)
		}
	}
}

func TestValidateSprintSettingsUpdateRejectsInvalidWorkingDays(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name string
		days []int
	}{
		{name: "empty", days: []int{}},
		{name: "outside ISO range", days: []int{1, 8}},
		{name: "duplicate", days: []int{1, 1, 2}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := service.validateSprintSettingsUpdate(CoreUpdateTeamSprintSettings{WorkingDays: &test.days}); err != ErrInvalidWorkingDays {
				t.Fatalf("expected ErrInvalidWorkingDays, got %v", err)
			}
		})
	}
}

func TestValidateSprintSettingsUpdateAcceptsCustomWorkingDays(t *testing.T) {
	service := &Service{}
	days := []int{7, 1, 2, 3, 4}
	if err := service.validateSprintSettingsUpdate(CoreUpdateTeamSprintSettings{WorkingDays: &days}); err != nil {
		t.Fatalf("expected custom workweek to be valid, got %v", err)
	}
}
