package migrations

import (
	"strings"
	"testing"
)

func TestPrimaryCalendarConnectionMigrationEnforcesAccountInvariant(t *testing.T) {
	t.Parallel()

	migration, err := FS.ReadFile("000143_primary_calendar_connection.up.sql")
	if err != nil {
		t.Fatalf("read primary calendar migration: %v", err)
	}
	source := strings.ToLower(string(migration))
	for _, contract := range []string{
		"add column is_primary boolean not null default false",
		"partition by user_id",
		"calendar_connections_one_primary_per_account",
		"on public.calendar_connections (user_id)",
		"where is_primary = true",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("primary calendar migration is missing %q", contract)
		}
	}
}
