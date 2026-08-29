package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

var upMigrationPattern = regexp.MustCompile(`^([0-9]+)_.+\.up\.sql$`)

func main() {
	migrationsPath := flag.String("migrations", "internal/migrations", "migration directory")
	flag.Parse()

	databaseURL, err := validationDatabaseURL()
	if err != nil {
		fail(err)
	}
	expectedVersion, err := migrationHead(*migrationsPath)
	if err != nil {
		fail(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		fail(fmt.Errorf("connect to sqlc validation database: %w", err))
	}
	defer func() {
		_ = connection.Close(context.Background())
	}()

	var actualVersion int64
	var dirty bool
	if err := connection.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&actualVersion, &dirty); err != nil {
		fail(fmt.Errorf("read validation database migration state: %w", err))
	}
	if dirty {
		fail(fmt.Errorf("validation database migration %d is dirty", actualVersion))
	}
	if actualVersion != expectedVersion {
		fail(fmt.Errorf("validation database migration version = %d, want repository head %d", actualVersion, expectedVersion))
	}
}

func validationDatabaseURL() (string, error) {
	databaseURL := os.Getenv("SQLC_DATABASE_URL")
	if databaseURL == "" {
		return "", errors.New("SQLC_DATABASE_URL is required")
	}
	return databaseURL, nil
}

func migrationHead(directory string) (int64, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, fmt.Errorf("read migration directory: %w", err)
	}

	var head int64
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		matches := upMigrationPattern.FindStringSubmatch(filepath.Base(entry.Name()))
		if len(matches) != 2 {
			continue
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse migration version in %q: %w", entry.Name(), err)
		}
		if version > head {
			head = version
		}
	}
	if head == 0 {
		return 0, fmt.Errorf("no up migrations found in %q", directory)
	}
	return head, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
