// Command databaseurl prints the validated PostgreSQL connection URL described
// by APP_DB_* variables. It exists so operational migration commands use the
// same escaping and TLS policy as application startup.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/joho/godotenv"
)

type runtimeConfig struct {
	DB struct {
		Host        string
		Port        string
		User        string
		Password    string
		Name        string
		SSLMode     string
		SSLRootCert string
		DisableTLS  bool
	}
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(output io.Writer) error {
	return runWithEnvironment(output, ".env")
}

func runWithEnvironment(output io.Writer, environmentFile string) error {
	if err := loadEnvironmentFile(environmentFile); err != nil {
		return err
	}
	if explicitURL := os.Getenv("DB_URL"); explicitURL != "" {
		if _, err := io.WriteString(output, explicitURL); err != nil {
			return fmt.Errorf("write explicit database connection URL: %w", err)
		}
		return nil
	}
	cfg, err := loadRuntimeConfig()
	if err != nil {
		return err
	}

	connectionURL, err := platformdatabase.ConnectionString(platformdatabase.Config{
		Host:        cfg.DB.Host,
		Port:        cfg.DB.Port,
		User:        cfg.DB.User,
		Password:    cfg.DB.Password,
		Name:        cfg.DB.Name,
		SSLMode:     cfg.DB.SSLMode,
		SSLRootCert: cfg.DB.SSLRootCert,
		DisableTLS:  cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("build database connection URL: %w", err)
	}
	if _, err := io.WriteString(output, connectionURL); err != nil {
		return fmt.Errorf("write database connection URL: %w", err)
	}
	return nil
}

func loadEnvironmentFile(path string) error {
	err := godotenv.Load(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	// godotenv parse errors can include the offending line, which may contain a
	// secret. Keep the operator-facing error deliberately path-only.
	return fmt.Errorf("database environment file %q is malformed or unreadable", path)
}

func loadRuntimeConfig() (runtimeConfig, error) {
	var cfg runtimeConfig
	cfg.DB.Host = environmentValue("APP_DB_HOST", "localhost")
	cfg.DB.Port = environmentValue("APP_DB_PORT", "5432")
	cfg.DB.User = environmentValue("APP_DB_USER", "postgres")
	cfg.DB.Password = environmentValue("APP_DB_PASSWORD", "password")
	cfg.DB.Name = environmentValue("APP_DB_NAME", "complexus")
	cfg.DB.SSLMode = os.Getenv("APP_DB_SSL_MODE")
	cfg.DB.SSLRootCert = os.Getenv("APP_DB_SSL_ROOT_CERT")

	disableTLS, err := strconv.ParseBool(environmentValue("APP_DB_DISABLE_TLS", "true"))
	if err != nil {
		return runtimeConfig{}, fmt.Errorf("parse APP_DB_DISABLE_TLS: %w", err)
	}
	cfg.DB.DisableTLS = disableTLS
	return cfg, nil
}

func environmentValue(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}
