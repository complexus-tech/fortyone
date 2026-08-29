package credentialvault

import "testing"

func TestMaintenanceBatchSizeIsBounded(t *testing.T) {
	t.Parallel()
	tests := []struct {
		requested int
		want      int
	}{
		{requested: -1, want: DefaultMaintenanceBatchSize},
		{requested: 0, want: DefaultMaintenanceBatchSize},
		{requested: 1, want: 1},
		{requested: MaxMaintenanceBatchSize, want: MaxMaintenanceBatchSize},
		{requested: MaxMaintenanceBatchSize + 1, want: MaxMaintenanceBatchSize},
	}
	for _, test := range tests {
		if got := MaintenanceBatchSize(test.requested); got != test.want {
			t.Errorf("MaintenanceBatchSize(%d) = %d, want %d", test.requested, got, test.want)
		}
	}
}

func TestRotationReportAdd(t *testing.T) {
	t.Parallel()
	key := KeyRef{ID: "provider-credentials", Version: 2}
	report := RotationReport{}
	report.Add(RotationReport{ActiveKey: key, Scanned: 3, Current: 1, Rewrapped: 1, Stale: 1})
	report.Add(RotationReport{ActiveKey: key, Scanned: 2, Current: 2})
	if report.ActiveKey != key || report.Scanned != 5 || report.Current != 3 || report.Rewrapped != 1 || report.Stale != 1 {
		t.Fatalf("combined report = %#v", report)
	}
}
