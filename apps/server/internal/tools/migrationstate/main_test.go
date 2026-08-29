package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationHeadUsesHighestUpMigration(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{
		"000001_initial.up.sql",
		"000001_initial.down.sql",
		"000147_latest.up.sql",
		"README.md",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0o600); err != nil {
			t.Fatalf("write migration fixture: %v", err)
		}
	}

	got, err := migrationHead(directory)
	if err != nil {
		t.Fatalf("migration head: %v", err)
	}
	if got != 147 {
		t.Fatalf("migration head = %d, want 147", got)
	}
}

func TestMigrationHeadRejectsDirectoryWithoutUpMigrations(t *testing.T) {
	if _, err := migrationHead(t.TempDir()); err == nil {
		t.Fatal("expected missing migration head to fail")
	}
}

func TestValidationDatabaseURLRequiresExplicitEnvironment(t *testing.T) {
	t.Setenv("SQLC_DATABASE_URL", "")
	if _, err := validationDatabaseURL(); err == nil {
		t.Fatal("expected missing SQLC_DATABASE_URL to fail")
	}

	const want = "postgresql://sqlc-validation.example/fortyone"
	t.Setenv("SQLC_DATABASE_URL", want)
	got, err := validationDatabaseURL()
	if err != nil {
		t.Fatalf("validation database URL: %v", err)
	}
	if got != want {
		t.Fatalf("validation database URL = %q, want %q", got, want)
	}
}
