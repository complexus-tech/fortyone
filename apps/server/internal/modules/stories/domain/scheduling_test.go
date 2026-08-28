package domain

import (
	"errors"
	"testing"
)

func TestValidateAutoSchedulingStatus(t *testing.T) {
	t.Parallel()

	validStatuses := []string{
		AutoSchedulingStatusOff,
		AutoSchedulingStatusNeedsOwner,
		AutoSchedulingStatusNeedsTime,
		AutoSchedulingStatusPlanning,
		AutoSchedulingStatusScheduled,
		AutoSchedulingStatusAtRisk,
		AutoSchedulingStatusCannotFit,
		AutoSchedulingStatusLocked,
	}
	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			if err := ValidateAutoSchedulingStatus(status); err != nil {
				t.Fatalf("ValidateAutoSchedulingStatus(%q) error = %v", status, err)
			}
		})
	}

	err := ValidateAutoSchedulingStatus("scheduled_elsewhere")
	if !errors.Is(err, ErrInvalidAutoSchedulingStatus) {
		t.Fatalf("invalid status error = %v, want %v", err, ErrInvalidAutoSchedulingStatus)
	}
}

func TestValidateStoryAutoSchedulingContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
		locked  bool
		status  string
		wantErr error
	}{
		{
			name:    "enabled scheduling",
			enabled: true,
			status:  AutoSchedulingStatusPlanning,
		},
		{
			name:   "disabled scheduling",
			status: AutoSchedulingStatusOff,
		},
		{
			name:    "locked enabled schedule",
			enabled: true,
			locked:  true,
			status:  AutoSchedulingStatusLocked,
		},
		{
			name:    "locked disabled schedule",
			locked:  true,
			status:  AutoSchedulingStatusLocked,
			wantErr: ErrLockedAutoSchedulingOff,
		},
		{
			name:    "unknown status",
			enabled: true,
			status:  "unknown",
			wantErr: ErrInvalidAutoSchedulingStatus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateStoryAutoSchedulingContract(test.enabled, test.locked, test.status)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidateStoryAutoSchedulingContract() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
