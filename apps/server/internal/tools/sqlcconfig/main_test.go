package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateRejectsDuplicateOutputs(t *testing.T) {
	root := t.TempDir()
	writeTestInput(t, root, "internal/migrations/schema.sql")
	writeTestInput(t, root, "internal/modules/links/repository/queries/links.sql")
	config := `version: "2"
sql:
  - name: "links"
    engine: "postgresql"
    schema: "internal/migrations"
    queries: "internal/modules/links/repository/queries"
    strict_function_checks: true
    strict_order_by: true
    analyzer:
      database: false
    database:
      uri: "${SQLC_DATABASE_URL}"
      managed: false
    rules:
      - "sqlc/db-prepare"
    gen:
      go:
        package: "linksql"
        out: "internal/modules/links/repository/sqlc"
        sql_package: "pgx/v5"
        sql_driver: "github.com/jackc/pgx/v5"
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        emit_pointers_for_null_enum_types: true
        emit_json_tags: false
        emit_db_tags: false
        emit_enum_valid_method: true
        omit_unused_structs: true
        query_parameter_limit: 0
        initialisms: ["id", "api", "url", "uri"]
  - name: "other"
    engine: "postgresql"
    schema: "internal/migrations"
    queries: "internal/modules/links/repository/queries"
    strict_function_checks: true
    strict_order_by: true
    analyzer:
      database: false
    database:
      uri: "${SQLC_DATABASE_URL}"
      managed: false
    rules:
      - "sqlc/db-prepare"
    gen:
      go:
        package: "othersql"
        out: "internal/modules/links/repository/sqlc"
        sql_package: "pgx/v5"
        sql_driver: "github.com/jackc/pgx/v5"
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        emit_pointers_for_null_enum_types: true
        emit_json_tags: false
        emit_db_tags: false
        emit_enum_valid_method: true
        omit_unused_structs: true
        query_parameter_limit: 0
        initialisms: ["id", "api", "url", "uri"]
`
	configPath := filepath.Join(root, "sqlc.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadAndValidate(configPath)
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate rejection", err)
	}
}

func TestValidateOutputDirectoryRejectsHiddenHandwrittenFile(t *testing.T) {
	root := t.TempDir()
	output := "internal/modules/links/repository/sqlc"
	writeTestInput(t, root, output+"/.keep")

	err := validateOutputDirectory(root, output)
	if err == nil || !strings.Contains(err.Error(), "handwritten") {
		t.Fatalf("error = %v, want handwritten-file rejection", err)
	}
}

func TestRejectSymlinkComponents(t *testing.T) {
	root := t.TempDir()
	realDirectory := filepath.Join(root, "real")
	if err := os.MkdirAll(realDirectory, 0o755); err != nil {
		t.Fatalf("create real directory: %v", err)
	}
	if err := os.Symlink(realDirectory, filepath.Join(root, "internal")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	err := rejectSymlinkComponents(root, "internal/modules/links/repository/sqlc")
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink rejection", err)
	}
}

func TestLoadAndValidateRejectsSQLCProfileDrift(t *testing.T) {
	serverRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	contents, err := os.ReadFile(filepath.Join(serverRoot, "sqlc.yaml"))
	if err != nil {
		t.Fatalf("read repository sqlc config: %v", err)
	}

	tests := []struct {
		name        string
		old         string
		new         string
		wantMessage string
	}{
		{
			name:        "typed option",
			old:         "emit_empty_slices: true",
			new:         "emit_empty_slices: false",
			wantMessage: "emit_empty_slices",
		},
		{
			name:        "unsupported generator",
			old:         "gen:\n      go:",
			new:         "gen:\n      json:",
			wantMessage: "Go generator",
		},
		{
			name:        "managed database",
			old:         "managed: false",
			new:         "managed: true",
			wantMessage: "database.managed",
		},
		{
			name:        "unknown SQL block option",
			old:         "strict_order_by: true",
			new:         "strict_order_by: true\n    database_uri_override: true",
			wantMessage: "database_uri_override",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for _, input := range []string{
				"internal/migrations/000001_test.up.sql",
				"internal/modules/links/repository/queries/links.sql",
				"internal/tools/sqlccontract/schema/contract.sql",
				"internal/tools/sqlccontract/queries/contract.sql",
			} {
				writeTestInput(t, root, input)
			}
			modified := strings.Replace(string(contents), tc.old, tc.new, 1)
			if modified == string(contents) {
				t.Fatalf("repository config no longer contains fixture %q", tc.old)
			}
			configPath := filepath.Join(root, "sqlc.yaml")
			if err := os.WriteFile(configPath, []byte(modified), 0o600); err != nil {
				t.Fatalf("write modified config: %v", err)
			}

			_, err := loadAndValidate(configPath)
			if err == nil || !strings.Contains(err.Error(), tc.wantMessage) {
				t.Fatalf("error = %v, want %q profile rejection", err, tc.wantMessage)
			}
		})
	}
}

func TestRejectOrphanGeneratedOutputs(t *testing.T) {
	root := t.TempDir()
	orphan := filepath.Join(root, "internal", "modules", "orphan", "repository", "sqlc")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatalf("create orphan output: %v", err)
	}

	err := rejectOrphanGeneratedOutputs(root, map[string]struct{}{})
	if err == nil || !strings.Contains(err.Error(), "no owning sqlc config block") {
		t.Fatalf("error = %v, want orphan output rejection", err)
	}
}

func TestLoadAndValidateAllowsPlatformRepositoryOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestInput(t, root, "internal/migrations/000001_test.up.sql")
	writeTestInput(t, root, "internal/platform/idempotency/repository/queries/receipts.sql")
	config := `version: "2"
sql:
  - name: "idempotency"
    engine: "postgresql"
    schema: "internal/migrations"
    queries: "internal/platform/idempotency/repository/queries"
    strict_function_checks: true
    strict_order_by: true
    analyzer:
      database: false
    database:
      uri: "${SQLC_DATABASE_URL}"
      managed: false
    rules:
      - "sqlc/db-prepare"
    gen:
      go:
        package: "idempotencysql"
        out: "internal/platform/idempotency/repository/sqlc"
        sql_package: "pgx/v5"
        sql_driver: "github.com/jackc/pgx/v5"
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        emit_pointers_for_null_enum_types: true
        emit_json_tags: false
        emit_db_tags: false
        emit_enum_valid_method: true
        omit_unused_structs: true
        query_parameter_limit: 0
        initialisms: ["id", "api", "url", "uri"]
`
	configPath := filepath.Join(root, "sqlc.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := loadAndValidate(configPath); err != nil {
		t.Fatalf("loadAndValidate() error = %v", err)
	}
}

func TestRejectOrphanGeneratedOutputsIncludesPlatformRepositories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	orphan := filepath.Join(root, "internal", "platform", "orphan", "repository", "sqlc")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatalf("create orphan output: %v", err)
	}

	err := rejectOrphanGeneratedOutputs(root, map[string]struct{}{})
	if err == nil || !strings.Contains(err.Error(), "no owning sqlc config block") {
		t.Fatalf("error = %v, want platform orphan output rejection", err)
	}
}

func writeTestInput(t *testing.T, root, relative string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
}
