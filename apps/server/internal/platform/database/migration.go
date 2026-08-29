package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenMigrationConnection opens the database/sql handle required by the
// golang-migrate PostgreSQL driver. Application repositories never receive it.
func OpenMigrationConnection(cfg MigrationConfig) (*sql.DB, error) {
	if cfg.Config.MaxOpenConns < 0 {
		return nil, errors.New("maximum migration connections cannot be negative")
	}
	if cfg.MaxIdleConns < 0 {
		return nil, errors.New("idle migration connections cannot be negative")
	}
	if cfg.Config.MaxOpenConns > 0 && cfg.MaxIdleConns > cfg.Config.MaxOpenConns {
		return nil, fmt.Errorf(
			"idle migration connections %d exceed maximum %d",
			cfg.MaxIdleConns,
			cfg.Config.MaxOpenConns,
		)
	}
	connectionString, err := ConnectionString(cfg.Config)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL migration connection: %w", err)
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Config.MaxOpenConns)
	return db, nil
}
