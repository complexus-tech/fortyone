package migrations

import (
	"strings"
	"testing"
)

func TestBrowserSessionMetadataMigrationPreservesUnknownHistory(t *testing.T) {
	t.Parallel()
	up, err := FS.ReadFile("000216_browser_session_metadata.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	normalized := strings.Join(strings.Fields(strings.ToLower(string(up))), " ")
	if !strings.Contains(normalized, "add column browser_name varchar(64)") || strings.Contains(normalized, "default") || strings.Contains(normalized, "update public.workspace_browser_sessions") {
		t.Fatal("browser metadata must be optional without fabricating historical names")
	}
	down, err := FS.ReadFile("000216_browser_session_metadata.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	normalized = strings.Join(strings.Fields(strings.ToLower(string(down))), " ")
	guard := strings.Index(normalized, "where browser_name is not null")
	drop := strings.Index(normalized, "drop column browser_name")
	if guard < 0 || strings.Index(normalized, "raise exception") < guard || drop < guard {
		t.Fatal("rollback must preserve recorded browser names")
	}
}
