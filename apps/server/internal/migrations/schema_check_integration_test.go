//go:build integration

package migrations_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/jackc/pgx/v5"
)

func TestSchemaPreflightReadsOnlyAndRejectsUnreadyDatabase(t *testing.T) {
	database := testkit.NewPostgres(t)
	parsed, err := pgx.ParseConfig(database.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	cfg := platformdatabase.MigrationConfig{Config: platformdatabase.Config{
		Host: parsed.Host, Port: strconv.Itoa(int(parsed.Port)), Name: parsed.Database,
		User: parsed.User, Password: parsed.Password, MaxOpenConns: 1,
		SSLMode: "disable", ConnectTimeout: 5 * time.Second,
	}}
	var expected uint64
	if err := database.Pool.QueryRow(t.Context(), "SELECT version FROM public.schema_migrations").Scan(&expected); err != nil {
		t.Fatal(err)
	}
	if actual, err := migrations.CheckSchema(t.Context(), cfg); err != nil || actual != expected {
		t.Fatalf("valid schema: version=%d error=%v", actual, err)
	}
	for _, tc := range []struct {
		name    string
		version uint64
		dirty   bool
	}{
		{"pending migration", expected - 1, false},
		{"dirty schema", expected, true},
		{"newer database", expected + 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := database.Pool.Exec(t.Context(), "UPDATE public.schema_migrations SET version = $1, dirty = $2", tc.version, tc.dirty); err != nil {
				t.Fatal(err)
			}
			if _, err := migrations.CheckSchema(t.Context(), cfg); err == nil {
				t.Fatal("unready schema passed")
			}
			var after uint64
			var dirty bool
			if err := database.Pool.QueryRow(t.Context(), "SELECT version, dirty FROM public.schema_migrations").Scan(&after, &dirty); err != nil {
				t.Fatal(err)
			}
			if after != tc.version || dirty != tc.dirty {
				t.Fatal("schema check changed migration state")
			}
		})
	}
	t.Run("missing table is not created", func(t *testing.T) {
		if _, err := database.Pool.Exec(t.Context(), "ALTER TABLE public.schema_migrations RENAME TO hidden_migration_state"); err != nil {
			t.Fatal(err)
		}
		if _, err := migrations.CheckSchema(t.Context(), cfg); err == nil {
			t.Fatal("missing migration state passed")
		}
		var absent bool
		if err := database.Pool.QueryRow(t.Context(), "SELECT to_regclass('public.schema_migrations') IS NULL").Scan(&absent); err != nil || !absent {
			t.Fatalf("check created migration state: absent=%v err=%v", absent, err)
		}
	})
}
