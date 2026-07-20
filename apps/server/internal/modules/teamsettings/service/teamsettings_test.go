package teamsettings

import "testing"

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
