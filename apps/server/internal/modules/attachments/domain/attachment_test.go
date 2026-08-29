package attachmentdomain

import "testing"

func TestAttachmentAvailabilityFailsClosedForKnownInfection(t *testing.T) {
	for name, testCase := range map[string]struct {
		status ScanStatus
		want   bool
	}{
		"unscanned compatibility state": {status: ScanStatusUnscanned, want: true},
		"pending scan":                  {status: ScanStatusPending, want: false},
		"clean":                         {status: ScanStatusClean, want: true},
		"failed scan":                   {status: ScanStatusFailed, want: false},
		"known infection":               {status: ScanStatusInfected, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := (Attachment{ScanStatus: testCase.status}).AvailableForDownload(); got != testCase.want {
				t.Fatalf("AvailableForDownload() = %t, want %t", got, testCase.want)
			}
		})
	}
}
