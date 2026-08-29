package stories

import (
	"errors"
	"testing"
)

func TestValidateStoryTimeContract(t *testing.T) {
	positiveDuration := 120
	shorterFocusBlock := 45
	zero := 0
	negative := -15
	longerFocusBlock := 180
	maximum := MaximumEstimatedDurationMinutes
	tooLarge := MaximumEstimatedDurationMinutes + 1

	tests := []struct {
		name      string
		duration  *int
		focus     *int
		wantError error
	}{
		{name: "both omitted"},
		{name: "duration only", duration: &positiveDuration},
		{name: "valid duration and focus block", duration: &positiveDuration, focus: &shorterFocusBlock},
		{name: "focus block without duration", focus: &shorterFocusBlock, wantError: ErrFocusBlockRequiresDuration},
		{name: "zero duration", duration: &zero, wantError: ErrInvalidEstimatedDuration},
		{name: "negative duration", duration: &negative, wantError: ErrInvalidEstimatedDuration},
		{name: "duration over product limit", duration: &tooLarge, wantError: ErrEstimatedDurationTooLarge},
		{name: "zero focus block", duration: &positiveDuration, focus: &zero, wantError: ErrInvalidMinimumFocusBlock},
		{name: "negative focus block", duration: &positiveDuration, focus: &negative, wantError: ErrInvalidMinimumFocusBlock},
		{name: "focus block over product limit", duration: &maximum, focus: &tooLarge, wantError: ErrMinimumFocusBlockTooLarge},
		{name: "focus block exceeds duration", duration: &positiveDuration, focus: &longerFocusBlock, wantError: ErrFocusBlockExceedsDuration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStoryTimeContract(tt.duration, tt.focus)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("ValidateStoryTimeContract() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

func TestNormalizeComparableValueUsesPointedIntegerValue(t *testing.T) {
	left := 90
	right := 90

	if normalizeComparableValue(&left) != normalizeComparableValue(&right) {
		t.Fatal("expected equal duration values to compare independently of pointer identity")
	}
	if normalizeComparableValue((*int)(nil)) != "nil" {
		t.Fatal("expected a nil duration pointer to normalize as nil")
	}
}

func TestFormatValueFormatsDurationPointersAsMinutes(t *testing.T) {
	duration := 90
	service := Service{}

	if got := service.formatValue(&duration); got != "90" {
		t.Fatalf("formatValue() = %q, want %q", got, "90")
	}
	if got := service.formatValue((*int)(nil)); got != "nil" {
		t.Fatalf("formatValue(nil duration) = %q, want %q", got, "nil")
	}
}

func TestApplyStoryTimeContractUpdateValidatesMergedState(t *testing.T) {
	duration := 120
	focusBlock := 60
	story := CoreSingleStory{
		EstimatedDurationMinutes: &duration,
		MinimumFocusBlockMinutes: &focusBlock,
	}

	t.Run("rejects reducing duration below existing focus block", func(t *testing.T) {
		updates := map[string]any{"estimated_duration_minutes": 30}

		err := applyStoryTimeContractUpdate(story, updates)

		if !errors.Is(err, ErrFocusBlockExceedsDuration) {
			t.Fatalf("applyStoryTimeContractUpdate() error = %v, want %v", err, ErrFocusBlockExceedsDuration)
		}
	})

	t.Run("rejects clearing duration while focus block remains", func(t *testing.T) {
		updates := map[string]any{"estimated_duration_minutes": nil}

		err := applyStoryTimeContractUpdate(story, updates)

		if !errors.Is(err, ErrFocusBlockRequiresDuration) {
			t.Fatalf("applyStoryTimeContractUpdate() error = %v, want %v", err, ErrFocusBlockRequiresDuration)
		}
	})

	t.Run("allows clearing duration and focus block together", func(t *testing.T) {
		updates := map[string]any{
			"estimated_duration_minutes":  nil,
			"minimum_focus_block_minutes": nil,
		}

		if err := applyStoryTimeContractUpdate(story, updates); err != nil {
			t.Fatalf("applyStoryTimeContractUpdate() unexpected error: %v", err)
		}
		clearedDuration, ok := updates["estimated_duration_minutes"].(*int)
		if !ok || clearedDuration != nil {
			t.Fatalf("expected duration to be a typed nil pointer, got %#v", updates["estimated_duration_minutes"])
		}
		clearedFocusBlock, ok := updates["minimum_focus_block_minutes"].(*int)
		if !ok || clearedFocusBlock != nil {
			t.Fatalf("expected focus block to be a typed nil pointer, got %#v", updates["minimum_focus_block_minutes"])
		}
	})
}
