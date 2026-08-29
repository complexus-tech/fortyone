package database

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"
)

func TestConnectionStringEscapesCredentialsAndSupportsIPv6(t *testing.T) {
	cfg := Config{
		User:       "api user",
		Password:   "p@ss/word",
		Host:       "2001:db8::1",
		Port:       "5432",
		Name:       "forty one",
		DisableTLS: true,
	}

	connectionString, err := ConnectionString(cfg)
	if err != nil {
		t.Fatalf("build connection string: %v", err)
	}
	parsed, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse connection string: %v", err)
	}
	if parsed.User.Username() != cfg.User {
		t.Fatalf("username = %q, want %q", parsed.User.Username(), cfg.User)
	}
	password, ok := parsed.User.Password()
	if !ok || password != cfg.Password {
		t.Fatalf("password = %q/%v, want %q/true", password, ok, cfg.Password)
	}
	if parsed.Hostname() != cfg.Host || parsed.Port() != cfg.Port {
		t.Fatalf("host = %q:%q, want %q:%q", parsed.Hostname(), parsed.Port(), cfg.Host, cfg.Port)
	}
	if parsed.Path != "/forty one" {
		t.Fatalf("database path = %q, want %q", parsed.Path, "/forty one")
	}
	if parsed.Query().Get("sslmode") != "disable" || parsed.Query().Get("timezone") != "utc" {
		t.Fatalf("connection options = %v", parsed.Query())
	}
}

func TestConnectionStringAcceptsBracketedIPv6Host(t *testing.T) {
	connectionString, err := ConnectionString(Config{
		Host:       "[2001:db8::1]",
		Port:       "5432",
		User:       "postgres",
		Password:   "password",
		Name:       "fortyone",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("build connection string: %v", err)
	}
	connectionURL, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse connection string: %v", err)
	}
	if got := connectionURL.Hostname(); got != "2001:db8::1" {
		t.Fatalf("hostname = %q, want %q", got, "2001:db8::1")
	}
	if got := connectionURL.Port(); got != "5432" {
		t.Fatalf("port = %q, want 5432", got)
	}
}

func TestConnectionStringRoundsConnectTimeoutUpToWholeSeconds(t *testing.T) {
	connectionString, err := ConnectionString(Config{
		Host:           "localhost",
		Port:           "5432",
		User:           "postgres",
		Password:       "password",
		Name:           "fortyone",
		ConnectTimeout: 1500 * time.Millisecond,
		DisableTLS:     true,
	})
	if err != nil {
		t.Fatalf("build connection string: %v", err)
	}
	connectionURL, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse connection string: %v", err)
	}
	if got := connectionURL.Query().Get("connect_timeout"); got != "2" {
		t.Fatalf("connect_timeout = %q, want 2", got)
	}
}

func TestConnectionStringRejectsNegativeConnectTimeout(t *testing.T) {
	_, err := ConnectionString(Config{ConnectTimeout: -time.Second})
	if err == nil {
		t.Fatal("expected a negative connect timeout to fail")
	}
}

func TestConnectionStringRejectsUnauthenticatedExplicitTLS(t *testing.T) {
	if _, err := ConnectionString(Config{SSLMode: "require"}); err == nil {
		t.Fatal("expected explicit require mode without a root certificate to fail")
	}
}

func TestConnectionStringRejectsRootCertificateWithDisabledTLS(t *testing.T) {
	if _, err := ConnectionString(Config{
		SSLMode:     "disable",
		SSLRootCert: "/run/secrets/database-ca.pem",
	}); err == nil {
		t.Fatal("expected a root certificate with disabled TLS to fail")
	}
}

func TestConnectionStringPreservesLegacyEncryptedTLSFallback(t *testing.T) {
	connectionString, err := ConnectionString(Config{
		Host:     "database.example.com",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		Name:     "fortyone",
	})
	if err != nil {
		t.Fatalf("build connection string: %v", err)
	}
	connectionURL, err := url.Parse(connectionString)
	if err != nil {
		t.Fatalf("parse connection string: %v", err)
	}
	if got := connectionURL.Query().Get("sslmode"); got != "require" {
		t.Fatalf("sslmode = %q, want require", got)
	}
	if got := connectionURL.Query().Get("sslrootcert"); got != "" {
		t.Fatalf("sslrootcert = %q, want empty", got)
	}
}

func TestNewPoolConfigUsesAuthenticatedTLS(t *testing.T) {
	poolConfig, err := newPoolConfig(Config{
		Host:     "database.example.com",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		Name:     "fortyone",
		SSLMode:  "verify-full",
	})
	if err != nil {
		t.Fatalf("new secure pool config: %v", err)
	}
	if poolConfig.ConnConfig.TLSConfig == nil {
		t.Fatal("verify-full did not configure TLS")
	}
	if poolConfig.ConnConfig.TLSConfig.InsecureSkipVerify {
		t.Fatal("verify-full unexpectedly disables certificate verification")
	}
	if got := poolConfig.ConnConfig.TLSConfig.ServerName; got != "database.example.com" {
		t.Fatalf("TLS server name = %q, want database.example.com", got)
	}
}

func TestNewPoolConfigSupportsSystemRootAliasWithPinnedPGX(t *testing.T) {
	poolConfig, err := newPoolConfig(Config{
		Host:        "database.example.com",
		Port:        "5432",
		User:        "postgres",
		Password:    "password",
		Name:        "fortyone",
		SSLMode:     "verify-full",
		SSLRootCert: "system",
	})
	if err != nil {
		t.Fatalf("new secure pool config with system roots: %v", err)
	}
	if poolConfig.ConnConfig.TLSConfig == nil || poolConfig.ConnConfig.TLSConfig.InsecureSkipVerify {
		t.Fatal("system-root alias did not configure authenticated TLS")
	}
}

func TestNewPoolConfigUsesOneConnectionBudget(t *testing.T) {
	poolConfig, err := newPoolConfig(Config{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Password:     "password",
		Name:         "fortyone",
		MaxOpenConns: 17,
		MinConns:     2,

		ConnectTimeout:    4 * time.Second,
		MaxConnIdleTime:   8 * time.Minute,
		MaxConnLifetime:   45 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		DisableTLS:        true,
	})
	if err != nil {
		t.Fatalf("new pool config: %v", err)
	}
	if poolConfig.MaxConns != 17 {
		t.Fatalf("max connections = %d, want 17", poolConfig.MaxConns)
	}
	if poolConfig.MinConns != 2 {
		t.Fatalf("minimum connections = %d, want 2", poolConfig.MinConns)
	}
	if poolConfig.ConnConfig.ConnectTimeout != 4*time.Second ||
		poolConfig.MaxConnIdleTime != 8*time.Minute ||
		poolConfig.MaxConnLifetime != 45*time.Minute ||
		poolConfig.HealthCheckPeriod != 30*time.Second {
		t.Fatalf("pool durations were not applied: %#v", poolConfig)
	}
}

func TestNewPoolConfigRejectsNegativeConnectionLimit(t *testing.T) {
	if _, err := newPoolConfig(Config{MaxOpenConns: -1}); err == nil {
		t.Fatal("expected negative max connections to fail")
	}
}

func TestNewPoolConfigDefaultsConnectionTimeout(t *testing.T) {
	poolConfig, err := newPoolConfig(Config{
		Host:       "localhost",
		Port:       "5432",
		User:       "postgres",
		Password:   "password",
		Name:       "fortyone",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("new pool config: %v", err)
	}
	if poolConfig.ConnConfig.ConnectTimeout != 10*time.Second {
		t.Fatalf("connect timeout = %s, want 10s", poolConfig.ConnConfig.ConnectTimeout)
	}
}

func TestOpenPoolStopsWhenStartupContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := OpenPool(ctx, Config{
		Host:       "127.0.0.1",
		Port:       "1",
		User:       "postgres",
		Password:   "password",
		Name:       "fortyone",
		DisableTLS: true,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("open pool error = %v, want context cancellation", err)
	}
}
