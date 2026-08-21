package calendarrepository

import (
	"os"
	"strings"
	"testing"
)

func TestCalendarEventReadsAuthorizeProviderSpecificScopes(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("events.go")
	if err != nil {
		t.Fatalf("read events.go: %v", err)
	}
	source := string(data)

	for _, contract := range []string{
		"(cc.provider = $4 AND $5 = ANY(cc.scopes))",
		"OR (cc.provider = $6 AND $7 = ANY(cc.scopes))",
		"(cc.provider = $3 AND $4 = ANY(cc.scopes))",
		"OR (cc.provider = $5 AND $6 = ANY(cc.scopes))",
		"calendar.ProviderGoogle",
		"calendar.GoogleCalendarEventsReadonlyScope",
		"calendar.ProviderMicrosoft",
		"calendar.MicrosoftCalendarReadWriteScope",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("calendar event reads are missing provider authorization contract %q", contract)
		}
	}
}
