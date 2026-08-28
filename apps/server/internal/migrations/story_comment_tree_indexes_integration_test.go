//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	projectmigrations "github.com/complexus-tech/projects-api/internal/migrations"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestStoryCommentTreeIndexesMigrateDownAndUpOnPostgres18(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()

	assertPostgres18(t, ctx, postgres.Pool)
	migrator := newStoryCommentIndexMigrator(t, postgres.DatabaseURL)
	headVersion := migrationVersion(t, migrator)
	if headVersion < 166 {
		t.Fatalf("migration head = %d, want at least 166", headVersion)
	}
	if headVersion > 166 {
		if err := migrator.Migrate(166); err != nil {
			t.Fatalf("move isolated database from head %d to migration 166: %v", headVersion, err)
		}
	}
	assertMigrationVersion(t, migrator, 166)
	assertStoryCommentIndexes(t, ctx, postgres.Pool, true)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("roll back story comment indexes: %v", err)
	}
	assertMigrationVersion(t, migrator, 165)
	assertStoryCommentIndexes(t, ctx, postgres.Pool, false)

	if err := migrator.Steps(1); err != nil {
		t.Fatalf("reapply story comment indexes: %v", err)
	}
	assertMigrationVersion(t, migrator, 166)
	assertStoryCommentIndexes(t, ctx, postgres.Pool, true)

	if headVersion > 166 {
		if err := migrator.Migrate(headVersion); err != nil {
			t.Fatalf("restore isolated database to migration head %d: %v", headVersion, err)
		}
		assertMigrationVersion(t, migrator, headVersion)
		assertStoryCommentIndexes(t, ctx, postgres.Pool, true)
	}
}

func newStoryCommentIndexMigrator(t *testing.T, databaseURL string) *migrate.Migrate {
	t.Helper()

	source, err := iofs.New(projectmigrations.FS, ".")
	if err != nil {
		t.Fatalf("load embedded migrations: %v", err)
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		source.Close()
		t.Fatalf("open isolated migration database: %v", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	driver, err := migratepostgres.WithInstance(database, &migratepostgres.Config{})
	if err != nil {
		database.Close()
		source.Close()
		t.Fatalf("create isolated migration driver: %v", err)
	}
	migrator, err := migrate.NewWithInstance("iofs", source, "story-comment-index-test", driver)
	if err != nil {
		driver.Close()
		source.Close()
		t.Fatalf("create story comment index migrator: %v", err)
	}
	t.Cleanup(func() {
		sourceErr, databaseErr := migrator.Close()
		if closeErr := errors.Join(sourceErr, databaseErr); closeErr != nil {
			t.Errorf("close story comment index migrator: %v", closeErr)
		}
	})
	return migrator
}

func assertMigrationVersion(t *testing.T, migrator *migrate.Migrate, want uint) {
	t.Helper()
	version := migrationVersion(t, migrator)
	if version != want {
		t.Fatalf("migration version = %d, want %d clean", version, want)
	}
}

func migrationVersion(t *testing.T, migrator *migrate.Migrate) uint {
	t.Helper()
	version, dirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if dirty {
		t.Fatalf("migration version = %d is dirty", version)
	}
	return version
}

func assertStoryCommentIndexes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, wantPresent bool) {
	t.Helper()
	for _, indexName := range []string{
		"idx_story_comments_roots_page",
		"idx_story_comments_replies_page",
	} {
		var found string
		if err := pool.QueryRow(
			ctx,
			"SELECT COALESCE(to_regclass(CAST($1 AS text))::text, '')",
			"public."+indexName,
		).Scan(&found); err != nil {
			t.Fatalf("inspect story comment index %s: %v", indexName, err)
		}
		if (found != "") != wantPresent {
			t.Fatalf("story comment index %s presence = %v, want %v", indexName, found != "", wantPresent)
		}
	}
}

func assertPostgres18(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var major int
	if err := pool.QueryRow(ctx, "SELECT current_setting('server_version_num')::integer / 10000").Scan(&major); err != nil {
		t.Fatalf("read PostgreSQL major version: %v", err)
	}
	if major != 18 {
		t.Fatalf("migration integration test requires PostgreSQL 18, got %d", major)
	}
}
