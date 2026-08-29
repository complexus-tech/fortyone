package database

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config contains PostgreSQL transport and pool settings used by API, worker,
// seed, tooling, and migration composition roots.
type Config struct {
	User         string
	Password     string
	Host         string
	Port         string
	Name         string
	MaxOpenConns int
	MinConns     int

	ConnectTimeout    time.Duration
	MaxConnIdleTime   time.Duration
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration
	SSLMode           string
	SSLRootCert       string
	DisableTLS        bool // Deprecated fallback used only when SSLMode is empty.
}

// MigrationConfig contains the database/sql-only settings used by
// golang-migrate. Runtime pgx pools use time-based idle connection management
// instead of database/sql's idle-count setting.
type MigrationConfig struct {
	Config
	MaxIdleConns     int
	StatementTimeout time.Duration
}

// ConnectionString returns a validated, safely escaped PostgreSQL URL.
func ConnectionString(cfg Config) (string, error) {
	if cfg.ConnectTimeout < 0 {
		return "", errors.New("PostgreSQL connect timeout cannot be negative")
	}
	sslMode, rootCert, err := connectionSecurity(cfg)
	if err != nil {
		return "", err
	}

	query := make(url.Values)
	query.Set("sslmode", sslMode)
	query.Set("timezone", "utc")
	if cfg.ConnectTimeout > 0 {
		timeoutSeconds := int64(cfg.ConnectTimeout / time.Second)
		if cfg.ConnectTimeout%time.Second != 0 {
			timeoutSeconds++
		}
		query.Set("connect_timeout", strconv.FormatInt(timeoutSeconds, 10))
	}
	if rootCert != "" {
		query.Set("sslrootcert", rootCert)
	}

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(unbracketHost(cfg.Host), cfg.Port),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}).String(), nil
}

// EffectiveSSLMode returns the validated PostgreSQL transport mode without
// exposing credentials or a complete connection string to diagnostics.
func EffectiveSSLMode(cfg Config) (string, error) {
	mode, _, err := connectionSecurity(cfg)
	return mode, err
}

func connectionSecurity(cfg Config) (string, string, error) {
	mode := strings.TrimSpace(cfg.SSLMode)
	rootCertInput := strings.TrimSpace(cfg.SSLRootCert)
	rootCert := rootCertInput
	// pgx versions before v5.7 do not recognize the libpq-style `system`
	// sentinel. Omitting sslrootcert with verify-full uses Go's system roots.
	if rootCert == "system" {
		rootCert = ""
	}
	explicit := mode != ""
	if !explicit {
		mode = "verify-full"
		if cfg.DisableTLS {
			mode = "disable"
		}
	}

	switch mode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		return "", "", fmt.Errorf("unsupported PostgreSQL SSL mode %q", mode)
	}
	if mode == "disable" && rootCertInput != "" {
		return "", "", errors.New("PostgreSQL SSL root certificate cannot be used when SSL mode is disabled")
	}
	if explicit && mode == "require" && rootCert == "" {
		return "", "", errors.New("explicit PostgreSQL SSL mode require needs a root certificate; use verify-full with system roots")
	}
	return mode, rootCert, nil
}

func unbracketHost(host string) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	}
	return host
}
