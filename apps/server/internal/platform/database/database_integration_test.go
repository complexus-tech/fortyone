//go:build integration

package database_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
)

func TestNativePoolHonorsConnectionBudgetAndClosesCleanly(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	cfg := configFromTestURL(t, postgres.DatabaseURL)
	cfg.MaxOpenConns = 3
	cfg.ConnectTimeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connections, err := platformdatabase.Open(ctx, cfg)
	if err != nil {
		t.Fatalf("open database connections: %v", err)
	}

	var value int
	if err := connections.Pool.QueryRow(ctx, "SELECT 41").Scan(&value); err != nil {
		t.Fatalf("native query: %v", err)
	}
	tx, err := connections.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin pgx transaction: %v", err)
	}
	if err := tx.QueryRow(ctx, "SELECT 42").Scan(&value); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("query in pgx transaction: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback pgx transaction: %v", err)
	}

	if value != 42 {
		t.Fatalf("query value = %d, want 42", value)
	}
	if got := connections.Pool.Stat().MaxConns(); got != 3 {
		t.Fatalf("pool max connections = %d, want 3", got)
	}

	if err := connections.Close(); err != nil {
		t.Fatalf("close database connections: %v", err)
	}
	if got := connections.Pool.Stat().TotalConns(); got != 0 {
		t.Fatalf("pool has %d connections after close, want 0", got)
	}
}

func configFromTestURL(t *testing.T, databaseURL string) platformdatabase.Config {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if parsed.User == nil || parsed.Hostname() == "" || parsed.Port() == "" || len(parsed.Path) < 2 {
		t.Fatalf("TEST_DATABASE_URL must include credentials, host, port, and database name")
	}
	password, _ := parsed.User.Password()
	return platformdatabase.Config{
		Host:        parsed.Hostname(),
		Port:        parsed.Port(),
		User:        parsed.User.Username(),
		Password:    password,
		Name:        parsed.Path[1:],
		SSLMode:     parsed.Query().Get("sslmode"),
		SSLRootCert: parsed.Query().Get("sslrootcert"),
	}
}
