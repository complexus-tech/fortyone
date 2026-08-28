package testkit

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseControlDatabaseURLRequiresExplicitSafeContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		databaseURL string
		wantError   string
	}{
		{name: "missing", wantError: "is required"},
		{name: "wrong scheme", databaseURL: "mysql://user:secret@localhost:5432/control", wantError: "postgres or postgresql"},
		{name: "missing user", databaseURL: "postgresql://localhost:5432/control", wantError: "username"},
		{name: "missing port", databaseURL: "postgresql://user:secret@localhost/control", wantError: "host and port"},
		{name: "missing database", databaseURL: "postgresql://user:secret@localhost:5432", wantError: "database name"},
		{name: "multiple path segments", databaseURL: "postgresql://user:secret@localhost:5432/control/extra", wantError: "exactly one"},
		{name: "fragment", databaseURL: "postgresql://user:secret@localhost:5432/control#fragment", wantError: "fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseControlDatabaseURL(tt.databaseURL)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("parse error = %v, want error containing %q", err, tt.wantError)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("parse error exposed URL credentials")
			}
		})
	}
}

func TestParseControlDatabaseURLAcceptsExplicitPostgreSQLURL(t *testing.T) {
	t.Parallel()

	parsed, err := parseControlDatabaseURL("postgresql://user:fake_password@127.0.0.1:5432/control?sslmode=disable")
	if err != nil {
		t.Fatalf("parse valid control database URL: %v", err)
	}
	if parsed.Path != "/control" || parsed.Port() != "5432" {
		t.Fatalf("parsed control URL = path %q, port %q", parsed.Path, parsed.Port())
	}
}

func TestDatabaseURLWithNamePreservesConnectionSettings(t *testing.T) {
	t.Parallel()

	controlURL, err := url.Parse("postgresql://user:password@127.0.0.1:5432/control?sslmode=disable&application_name=tests")
	if err != nil {
		t.Fatalf("parse fixture URL: %v", err)
	}

	got := databaseURLWithName(controlURL, "fortyone_test_0123456789abcdef")
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse isolated URL: %v", err)
	}
	if parsed.Path != "/fortyone_test_0123456789abcdef" {
		t.Fatalf("database path = %q", parsed.Path)
	}
	if parsed.Query().Get("sslmode") != "disable" || parsed.Query().Get("application_name") != "tests" {
		t.Fatalf("connection settings changed: %v", parsed.Query())
	}
	if controlURL.Path != "/control" {
		t.Fatalf("control URL was mutated: %q", controlURL.Path)
	}
}

func TestNewTestDatabaseNameUsesCleanupPrefix(t *testing.T) {
	t.Parallel()

	databaseName, err := newTestDatabaseName()
	if err != nil {
		t.Fatalf("new test database name: %v", err)
	}
	if !strings.HasPrefix(databaseName, testDatabasePrefix) {
		t.Fatalf("database name = %q, want %q prefix", databaseName, testDatabasePrefix)
	}
}
