package main

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUsesSharedAuthenticatedTLSContract(t *testing.T) {
	t.Setenv("DB_URL", "")
	t.Setenv("APP_DB_HOST", "database.example.com")
	t.Setenv("APP_DB_PORT", "5432")
	t.Setenv("APP_DB_USER", "api user")
	t.Setenv("APP_DB_PASSWORD", "p@ss/word")
	t.Setenv("APP_DB_NAME", "forty one")
	t.Setenv("APP_DB_SSL_MODE", "verify-full")
	t.Setenv("APP_DB_SSL_ROOT_CERT", "/run/secrets/database-ca.pem")
	t.Setenv("APP_DB_DISABLE_TLS", "false")

	var output bytes.Buffer
	if err := runWithEnvironment(&output, filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("build migration database URL: %v", err)
	}
	parsed, err := url.Parse(output.String())
	if err != nil {
		t.Fatalf("parse migration database URL: %v", err)
	}
	password, _ := parsed.User.Password()
	if parsed.User.Username() != "api user" || password != "p@ss/word" {
		t.Fatal("database credentials were not escaped and restored correctly")
	}
	if parsed.Query().Get("sslmode") != "verify-full" ||
		parsed.Query().Get("sslrootcert") != "/run/secrets/database-ca.pem" {
		t.Fatalf("TLS query = %v, want authenticated private-CA settings", parsed.Query())
	}
}

func TestRunPreservesExplicitDatabaseURLOverride(t *testing.T) {
	const explicitURL = "postgresql://operator:secret@database.example.com/fortyone?sslmode=verify-full"
	t.Setenv("DB_URL", explicitURL)
	t.Setenv("APP_DB_DISABLE_TLS", "not-a-boolean")

	var output bytes.Buffer
	if err := runWithEnvironment(&output, filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("use explicit migration database URL: %v", err)
	}
	if got := output.String(); got != explicitURL {
		t.Fatalf("database URL = %q, want explicit override", got)
	}
}

func TestRunRejectsMalformedEnvironmentFile(t *testing.T) {
	const sentinelSecret = "do-not-leak-this-database-password"
	environmentFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(
		environmentFile,
		[]byte("APP_DB_PASSWORD='"+sentinelSecret+"\n"),
		0o600,
	); err != nil {
		t.Fatalf("write malformed environment file: %v", err)
	}

	var output bytes.Buffer
	err := runWithEnvironment(&output, environmentFile)
	if err == nil || !strings.Contains(err.Error(), "database environment file") {
		t.Fatalf("error = %v, want strict malformed dotenv rejection", err)
	}
	if strings.Contains(err.Error(), sentinelSecret) {
		t.Fatal("malformed dotenv error leaked secret-bearing source text")
	}
	if output.Len() != 0 {
		t.Fatal("malformed dotenv produced a database URL")
	}
}

func TestRunRejectsUnreadableEnvironmentPath(t *testing.T) {
	environmentDirectory := t.TempDir()

	var output bytes.Buffer
	err := runWithEnvironment(&output, environmentDirectory)
	if err == nil || !strings.Contains(err.Error(), "database environment file") {
		t.Fatalf("error = %v, want unreadable dotenv rejection", err)
	}
	if output.Len() != 0 {
		t.Fatal("unreadable dotenv produced a database URL")
	}
}
