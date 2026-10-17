package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
)

// CheckSchema verifies the candidate binary against PostgreSQL without creating
// migration tables, acquiring migration locks, or applying SQL migrations.
func CheckSchema(ctx context.Context, cfg platformdatabase.MigrationConfig) (uint64, error) {
	expected, err := latestSchemaVersion(FS)
	if err != nil {
		return 0, err
	}
	db, err := platformdatabase.OpenMigrationConnection(cfg)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, fmt.Errorf("begin read-only schema check: %w", err)
	}
	defer tx.Rollback()
	var actual uint64
	var dirty bool
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT version, dirty, count(*) OVER () FROM public.schema_migrations`).Scan(&actual, &dirty, &count); err != nil {
		return 0, fmt.Errorf("read migration state: %w", err)
	}
	if count != 1 {
		return 0, errors.New("expected exactly one migration state record")
	}
	if err := validateSchemaVersion(actual, expected, dirty); err != nil {
		return actual, err
	}
	return actual, nil
}

func latestSchemaVersion(source fs.FS) (uint64, error) {
	files, err := fs.Glob(source, "*.up.sql")
	if err != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", err)
	}
	var latest uint64
	for _, file := range files {
		prefix, _, found := strings.Cut(file, "_")
		version, err := strconv.ParseUint(prefix, 10, 64)
		if !found || err != nil || version == 0 {
			return 0, fmt.Errorf("invalid embedded migration filename %q", file)
		}
		latest = max(latest, version)
	}
	if latest == 0 {
		return 0, errors.New("no embedded migrations found")
	}
	return latest, nil
}

func validateSchemaVersion(actual, expected uint64, dirty bool) error {
	if dirty {
		return fmt.Errorf("database migration %d is dirty; operator repair required", actual)
	}
	if actual != expected {
		return fmt.Errorf("database schema is %d; candidate requires exactly %d; complete the reviewed migration procedure first", actual, expected)
	}
	return nil
}
