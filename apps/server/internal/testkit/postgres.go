package testkit

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// TestDatabaseURLEnv is the single PostgreSQL connection contract used by
	// integration tests. The configured role must be allowed to create and drop
	// databases on a disposable, non-production PostgreSQL server.
	TestDatabaseURLEnv = "TEST_DATABASE_URL"

	testDatabasePrefix = "fortyone_test_"
	setupTimeout       = 2 * time.Minute
	cleanupTimeout     = 30 * time.Second
	testPoolMaxConns   = 8
)

// Postgres is a fully migrated, test-owned PostgreSQL database. Every call to
// NewPostgres creates a separate database so tests can run in parallel without
// sharing schema or tenant state.
type Postgres struct {
	DatabaseURL string
	Pool        *pgxpool.Pool
}

// NewPostgres creates an isolated database, applies the real migration chain,
// opens a bounded pgx pool, and registers cleanup with the test. Integration
// tests deliberately fail when TEST_DATABASE_URL or PostgreSQL is unavailable;
// they must never silently degrade into skipped coverage.
func NewPostgres(t testing.TB) *Postgres {
	return newPostgres(t, nil)
}

// NewPostgresAtMigration creates an isolated database migrated through the
// requested version. It is intended for testing migration boundaries without
// running historical down migrations against newer dependent schemas.
func NewPostgresAtMigration(t testing.TB, version uint) *Postgres {
	t.Helper()
	if version == 0 {
		t.Fatal("migration version must be positive")
	}
	return newPostgres(t, &version)
}

func newPostgres(t testing.TB, migrationVersion *uint) *Postgres {
	t.Helper()

	controlURL, err := parseControlDatabaseURL(os.Getenv(TestDatabaseURLEnv))
	if err != nil {
		t.Fatal(err)
	}
	databaseName, err := newTestDatabaseName()
	if err != nil {
		t.Fatalf("generate isolated PostgreSQL database name: %v", err)
	}
	databaseURL := databaseURLWithName(controlURL, databaseName)

	setupCtx, cancelSetup := context.WithTimeout(t.Context(), setupTimeout)
	defer cancelSetup()
	if err := createDatabase(setupCtx, controlURL.String(), databaseName); err != nil {
		t.Fatalf("create isolated PostgreSQL database: %v", err)
	}

	var pool *pgxpool.Pool
	t.Cleanup(func() {
		if pool != nil {
			pool.Close()
		}

		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancelCleanup()
		if err := dropDatabase(cleanupCtx, controlURL.String(), databaseName); err != nil {
			t.Errorf("drop isolated PostgreSQL database %q: %v", databaseName, err)
		}
	})

	if err := applyMigrations(setupCtx, databaseURL, databaseName, migrationVersion); err != nil {
		t.Fatalf("apply migrations to isolated PostgreSQL database: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("build isolated PostgreSQL pool configuration")
	}
	poolConfig.MaxConns = testPoolMaxConns
	poolConfig.MinConns = 0
	pool, err = pgxpool.NewWithConfig(setupCtx, poolConfig)
	if err != nil {
		t.Fatalf("open isolated PostgreSQL pool: %v", err)
	}
	if err := pool.Ping(setupCtx); err != nil {
		t.Fatalf("ping isolated PostgreSQL database: %v", err)
	}

	return &Postgres{
		DatabaseURL: databaseURL,
		Pool:        pool,
	}
}

func parseControlDatabaseURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf(
			"%s is required when integration tests are enabled; configure a disposable PostgreSQL control database and a role with CREATEDB",
			TestDatabaseURLEnv,
		)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		// Do not include the parser error: url.Error can contain the complete URL,
		// including its password.
		return nil, fmt.Errorf("%s must be a valid PostgreSQL URL", TestDatabaseURLEnv)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return nil, fmt.Errorf("%s must use the postgres or postgresql scheme", TestDatabaseURLEnv)
	}
	if parsed.User == nil || parsed.User.Username() == "" {
		return nil, fmt.Errorf("%s must include a PostgreSQL username", TestDatabaseURLEnv)
	}
	if parsed.Hostname() == "" || parsed.Port() == "" {
		return nil, fmt.Errorf("%s must include an explicit host and port", TestDatabaseURLEnv)
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" {
		return nil, fmt.Errorf("%s must include a control database name", TestDatabaseURLEnv)
	}
	if strings.Contains(databaseName, "/") {
		return nil, fmt.Errorf("%s must include exactly one control database name", TestDatabaseURLEnv)
	}
	if parsed.Fragment != "" {
		return nil, fmt.Errorf("%s must not include a URL fragment", TestDatabaseURLEnv)
	}

	return parsed, nil
}

func newTestDatabaseName() (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return testDatabasePrefix + hex.EncodeToString(random), nil
}

func databaseURLWithName(controlURL *url.URL, databaseName string) string {
	isolatedURL := *controlURL
	isolatedURL.Path = "/" + databaseName
	isolatedURL.RawPath = ""
	return isolatedURL.String()
}

func createDatabase(ctx context.Context, controlURL string, databaseName string) error {
	connection, err := pgx.Connect(ctx, controlURL)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL control database: %w", err)
	}
	defer connection.Close(context.Background())

	if _, err := connection.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	return nil
}

func dropDatabase(ctx context.Context, controlURL string, databaseName string) error {
	if !strings.HasPrefix(databaseName, testDatabasePrefix) {
		return fmt.Errorf("refusing to drop database without %q prefix", testDatabasePrefix)
	}

	connection, err := pgx.Connect(ctx, controlURL)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL control database: %w", err)
	}
	defer connection.Close(context.Background())

	statement := "DROP DATABASE IF EXISTS " + pgx.Identifier{databaseName}.Sanitize() + " WITH (FORCE)"
	if _, err := connection.Exec(ctx, statement); err != nil {
		return fmt.Errorf("drop database: %w", err)
	}
	return nil
}

func applyMigrations(ctx context.Context, databaseURL string, databaseName string, version *uint) error {
	migrationSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return errors.Join(
			fmt.Errorf("open migration database: %w", err),
			closeError("close migration source", migrationSource.Close()),
		)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	if err := database.PingContext(ctx); err != nil {
		return errors.Join(
			fmt.Errorf("ping migration database: %w", err),
			closeError("close migration database", database.Close()),
			closeError("close migration source", migrationSource.Close()),
		)
	}

	databaseDriver, err := migratepostgres.WithInstance(database, &migratepostgres.Config{})
	if err != nil {
		return errors.Join(
			fmt.Errorf("create PostgreSQL migration driver: %w", err),
			closeError("close migration database", database.Close()),
			closeError("close migration source", migrationSource.Close()),
		)
	}

	migrator, err := migrate.NewWithInstance("iofs", migrationSource, databaseName, databaseDriver)
	if err != nil {
		return errors.Join(
			fmt.Errorf("initialize migrator: %w", err),
			closeError("close migration database driver", databaseDriver.Close()),
			closeError("close migration source", migrationSource.Close()),
		)
	}

	var migrationErr error
	if version == nil {
		migrationErr = migrator.Up()
	} else {
		migrationErr = migrator.Migrate(*version)
	}
	if errors.Is(migrationErr, migrate.ErrNoChange) {
		migrationErr = nil
	}
	sourceCloseErr, databaseCloseErr := migrator.Close()

	return errors.Join(
		wrapError("run migrations", migrationErr),
		closeError("close migration source", sourceCloseErr),
		closeError("close migration database", databaseCloseErr),
	)
}

func wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func closeError(operation string, err error) error {
	return wrapError(operation, err)
}
