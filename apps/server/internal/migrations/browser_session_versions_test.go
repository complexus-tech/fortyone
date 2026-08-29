package migrations

import (
	"strings"
	"testing"
)

func TestBrowserSessionVersionMigrationIsMonotonicAndGuarded(t *testing.T) {
	t.Parallel()

	up, err := FS.ReadFile("000171_browser_session_versions.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	normalizedUp := strings.Join(strings.Fields(strings.ToLower(string(up))), " ")
	for _, clause := range []string{
		"add column auth_session_version bigint not null default 1",
		"constraint users_auth_session_version_positive",
		"check (auth_session_version > 0)",
	} {
		if !strings.Contains(normalizedUp, clause) {
			t.Fatalf("browser session up migration is missing %q", clause)
		}
	}

	down, err := FS.ReadFile("000171_browser_session_versions.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	normalizedDown := strings.Join(strings.Fields(strings.ToLower(string(down))), " ")
	guard := strings.Index(normalizedDown, "where auth_session_version <> 1")
	drop := strings.Index(normalizedDown, "drop column auth_session_version")
	if guard < 0 || drop <= guard {
		t.Fatal("browser session down migration must guard adopted epochs before dropping the column")
	}
}
