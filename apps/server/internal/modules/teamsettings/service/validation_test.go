package teamsettings

import "testing"

func TestValidateEstimateSchemeOnlyAllowsComplexitySchemes(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"points", "tshirt"} {
		if err := validateEstimationSettingsUpdate(CoreUpdateTeamEstimationSettings{
			Scheme: PatchField[string]{Value: scheme, Present: true},
		}); err != nil {
			t.Fatalf("expected %q to be valid, got %v", scheme, err)
		}
	}
	for _, scheme := range []string{"hours", "ideal_days"} {
		if err := validateEstimationSettingsUpdate(CoreUpdateTeamEstimationSettings{
			Scheme: PatchField[string]{Value: scheme, Present: true},
		}); err != ErrInvalidEstimateScheme {
			t.Fatalf("expected legacy scheme %q to be rejected, got %v", scheme, err)
		}
	}
}

func TestValidateSprintSettingsUpdateRejectsInvalidWorkingDays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		days []int
	}{
		{name: "empty", days: []int{}},
		{name: "outside ISO range", days: []int{1, 8}},
		{name: "duplicate", days: []int{1, 1, 2}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateSprintSettingsUpdate(CoreUpdateTeamSprintSettings{
				WorkingDays: PatchField[[]int]{Value: test.days, Present: true},
			})
			if err != ErrInvalidWorkingDays {
				t.Fatalf("validation error = %v, want ErrInvalidWorkingDays", err)
			}
		})
	}
}

func TestValidationRejectsEmptySettingsPatches(t *testing.T) {
	t.Parallel()

	if err := validateSprintSettingsUpdate(CoreUpdateTeamSprintSettings{}); err != ErrNoSettingsChanges {
		t.Fatalf("empty sprint patch error = %v", err)
	}
	if err := validateStoryAutomationSettingsUpdate(CoreUpdateTeamStoryAutomationSettings{}); err != ErrNoSettingsChanges {
		t.Fatalf("empty story patch error = %v", err)
	}
	if err := validateEstimationSettingsUpdate(CoreUpdateTeamEstimationSettings{}); err != ErrNoSettingsChanges {
		t.Fatalf("empty estimation patch error = %v", err)
	}
}

func TestPatchFromPointerPreservesFalseAndZeroPresence(t *testing.T) {
	t.Parallel()

	boolValue := false
	intValue := 0
	if patch := PatchFromPointer(&boolValue); !patch.Present || patch.Value {
		t.Fatalf("false patch = %#v", patch)
	}
	if patch := PatchFromPointer(&intValue); !patch.Present || patch.Value != 0 {
		t.Fatalf("zero patch = %#v", patch)
	}
	if patch := PatchFromPointer[int](nil); patch.Present {
		t.Fatalf("nil patch = %#v", patch)
	}
}
