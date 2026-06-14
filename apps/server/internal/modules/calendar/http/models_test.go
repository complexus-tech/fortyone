package calendarhttp

import (
	"testing"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func TestToAppConnectionReportsEventDetailCapability(t *testing.T) {
	t.Parallel()

	connection := toAppConnection(calendar.CoreConnection{
		ID:       uuid.New(),
		Provider: calendar.ProviderGoogle,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events.readonly",
		},
	})

	if !connection.CanReadEventDetails {
		t.Fatal("expected Google event detail scope to enable event details")
	}
}
