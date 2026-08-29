package calendarrepository

import (
	"strings"
	"testing"
)

func TestCalendarEventReadsAuthorizeProviderSpecificScopes(t *testing.T) {
	t.Parallel()

	for _, queryName := range []string{"ListCalendarEvents", "GetCalendarEvent"} {
		query := normalizedNamedQuery(t, "queries/events.sql", queryName)
		for _, contract := range []string{
			"connection.cleanup_pending_at is null",
			"connection.provider = sqlc.arg(google_provider)",
			"cast(sqlc.arg(google_read_scope) as text) = any(connection.scopes)",
			"connection.provider = sqlc.arg(microsoft_provider)",
			"cast(sqlc.arg(microsoft_read_scope) as text) = any(connection.scopes)",
			"event.user_id = sqlc.arg(user_id)",
		} {
			if !strings.Contains(query, contract) {
				t.Errorf("%s is missing provider authorization contract %q", queryName, contract)
			}
		}
	}

	source := readRepositorySource(t, "event_reads.go")
	for _, contract := range []string{
		"GoogleProvider: string(calendar.ProviderGoogle)",
		"GoogleReadScope: calendar.GoogleCalendarEventsReadonlyScope",
		"MicrosoftProvider: string(calendar.ProviderMicrosoft)",
		"MicrosoftReadScope: calendar.MicrosoftCalendarReadWriteScope",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("calendar event repository is missing typed SQLC argument %q", contract)
		}
	}
}
