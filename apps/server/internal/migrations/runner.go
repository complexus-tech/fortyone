package migrations

import (
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type migrateLogger struct{}

func (migrateLogger) Printf(format string, v ...any) {
	fmt.Printf(format, v...)
}

func (migrateLogger) Verbose() bool {
	return true
}

func Run(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	source, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("initializing migrator: %w", err)
	}
	m.Log = migrateLogger{}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
