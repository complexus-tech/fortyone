package objectivesrepository

import (
	"errors"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
)

func TestValidateObjectiveDatePatchUsesTheResultingDateRange(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	afterEnd := end.AddDate(0, 0, 1)
	beforeStart := start.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		patch   objectivesdomain.ObjectivePatch
		wantErr bool
	}{
		{name: "unrelated field", patch: objectivesdomain.ObjectivePatch{Name: objectivesdomain.SetField("Renamed")}},
		{name: "clear start", patch: objectivesdomain.ObjectivePatch{StartDate: objectivesdomain.ClearField[time.Time]()}},
		{name: "clear end", patch: objectivesdomain.ObjectivePatch{EndDate: objectivesdomain.ClearField[time.Time]()}},
		{name: "start after stored end", patch: objectivesdomain.ObjectivePatch{StartDate: objectivesdomain.SetField(afterEnd)}, wantErr: true},
		{name: "end before stored start", patch: objectivesdomain.ObjectivePatch{EndDate: objectivesdomain.SetField(beforeStart)}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateObjectiveDatePatch(&start, &end, test.patch)
			if test.wantErr && !errors.Is(err, objectivesdomain.ErrInvalid) {
				t.Fatalf("error = %v, want ErrInvalid", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
		})
	}
}
