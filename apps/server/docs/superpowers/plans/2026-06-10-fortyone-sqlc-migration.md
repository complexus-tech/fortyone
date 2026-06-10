# FortyOne SQLC Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move FortyOne's server database layer from handwritten `sqlx` SQL toward generated, type-safe `sqlc` queries while preserving the existing domain-first architecture and improving the boundaries around auth, workspace access, repositories, jobs, and AI assistant data tools.

**Architecture:** FortyOne already has the right high-level shape: `internal/bootstrap`, `internal/platform`, and `internal/modules/<domain>/{http,service,repository}`. This plan keeps that structure, adds generated `internal/repository` code from module-local `.sql` files, migrates modules one at a time, and removes raw SQL from middleware and repositories after each module is verified.

**Tech Stack:** Go 1.23, PostgreSQL, `pgx/v5`, `sqlc`, existing `sqlx` bridge during migration, Redis cache, `golang-migrate`, OpenTelemetry tracing, existing `pkg/web` HTTP stack, existing `internal/bootstrap/architecture` guard tests.

---

## Current State

FortyOne already has a domain-first backend layout:

```txt
apps/server/
  cmd/
    api/
    worker/
    seed/
  internal/
    bootstrap/
      api/
      architecture/
      worker/
    migrations/
    modules/
      github/
      reports/
      stories/
      objectives/
      keyresults/
      sprints/
      ...
    platform/
      auth/
      http/
    taskhandlers/
  pkg/
    database/
    jobs/
    web/
```

The server currently uses:

- `github.com/jmoiron/sqlx` for DB access.
- `github.com/jackc/pgx/v5/stdlib` as the SQL driver behind `sqlx`.
- Raw SQL strings inside repositories, jobs, task handlers, and workspace middleware.
- `BeginTxx` for multi-step writes.
- Architecture tests in `internal/bootstrap/architecture/boundaries_test.go`.

The migration target is:

- `pgxpool.Pool` as the primary DB pool.
- `internal/repository` as the generated `sqlc` package.
- Module-local SQL files such as `internal/modules/stories/repository/queries.sql`.
- Thin module repository wrappers that call generated query methods and map generated rows into existing domain models.
- No SQL inside HTTP middleware.
- No repository package reading auth/user/workspace values from context.
- Services and handlers pass explicit `userID`, `workspaceID`, `teamID`, and object IDs.

---

## Migration Principles

1. **No big-bang rewrite.** Add `sqlc` beside `sqlx`, then migrate one module at a time.
2. **Preserve public behavior.** API response contracts should not change unless a task explicitly says so.
3. **Keep module boundaries.** Generated query files live near the module that owns the behavior; generated Go lives centrally.
4. **Do not move business logic into SQL.** SQL returns data; services keep workflow decisions.
5. **Do not let repositories read auth context.** Repositories receive IDs as arguments.
6. **Prefer explicit typed filters over generic dynamic SQL.** Where dynamic filters are unavoidable, isolate them behind small reviewed query builders.
7. **Migrate high-risk writes with tests first.** Stories, objectives, key results, and sprints must have failing tests before repository replacement.
8. **Keep frequent commits.** Each task should leave the server compiling and tests runnable.

---

## Target File Responsibilities

### New Files

- `apps/server/sqlc.yaml`
  - Configures `sqlc` generation.
  - Reads schema from `internal/migrations`.
  - Reads query files from `internal/modules/*/repository/*.sql` and selected `internal/platform/*/repository/*.sql`.
  - Generates Go into `internal/repository`.

- `apps/server/internal/repository/`
  - Generated code only.
  - No manual edits.
  - Package name: `repository`.

- `apps/server/internal/platform/database/pool.go`
  - Opens `*pgxpool.Pool`.
  - Uses the existing DB config values.
  - Does not remove `pkg/database.Open` until migration completion.

- `apps/server/internal/platform/database/tx.go`
  - Contains small transaction helpers for generated query usage.
  - Centralizes rollback/commit behavior.

- `apps/server/internal/platform/workspace/repository/queries.sql`
  - Generated SQL for resolving workspace access, last-login updates, and last-accessed updates.

- `apps/server/internal/platform/workspace/repository/repository.go`
  - Wraps generated workspace queries.
  - Gives middleware a service/repository interface instead of raw SQL access.

- `apps/server/internal/platform/workspace/service.go`
  - Resolves workspace context and cache interaction.
  - Keeps HTTP middleware thin.

- `apps/server/internal/modules/<module>/repository/queries.sql`
  - Module-owned query definitions.

- `apps/server/docs/database/sqlc.md`
  - Documents repository patterns, transaction patterns, naming conventions, and migration rules.

### Existing Files To Modify

- `apps/server/Makefile`
  - Add `sqlc-generate`.
  - Add `sqlc-verify`.
  - Optionally add `test`.

- `apps/server/go.mod`
  - Add `github.com/sqlc-dev/pqtype` only if needed for JSON/nullable types.
  - Add no runtime `sqlc` dependency unless generated code requires one.
  - Keep `github.com/jackc/pgx/v5`.

- `apps/server/pkg/database/database.go`
  - Keep existing `sqlx` opener during migration.
  - Move new pgx pool opener into `internal/platform/database` to avoid mixing old and new concerns.

- `apps/server/internal/bootstrap/api/*.go`
  - Wire `pgxpool.Pool` and `repository.Queries`.
  - Keep `sqlx.DB` until migrated modules no longer need it.

- `apps/server/internal/bootstrap/worker/*.go`
  - Same bridge as API bootstrap.
  - Jobs should migrate after module repositories are stable.

- `apps/server/internal/platform/http/middleware/workspace.go`
  - Remove raw SQL.
  - Depend on workspace service interface.

- `apps/server/internal/bootstrap/architecture/boundaries_test.go`
  - Add checks for DB and auth boundary violations.

---

## Phase 0: Baseline Inventory And Guardrails

### Task 0.1: Capture Current Raw SQL Surface

**Files:**

- Create: `apps/server/docs/database/raw-sql-inventory.md`

- [ ] **Step 1: Generate the current SQL usage inventory**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
rg -n "sqlx|BeginTxx|NamedExec|Queryx|Select\\(|Get\\(|QueryRow|Exec\\(|database/sql|sqlx\\.In" . \
  --glob '*.go' \
  > docs/database/raw-sql-inventory.md
```

Expected:

- The command exits with status `0`.
- `docs/database/raw-sql-inventory.md` contains hits for `pkg/jobs`, `internal/modules/*/repository`, route wiring structs with `*sqlx.DB`, and platform middleware.

- [ ] **Step 2: Review the inventory and classify modules**

Append this classification header manually at the top of `docs/database/raw-sql-inventory.md`:

```markdown
# Raw SQL Inventory

## Migration Categories

- Platform first: auth/session/workspace lookup and workspace access updates.
- Read/report first: reports, notifications, search, GitHub integration reads.
- Core write flows: stories, objectives, key results, sprints, teams, users, workspaces.
- Worker/job flows: purge jobs, automation jobs, inactivity emails, GitHub/Slack background processing.

## Raw Hits
```

- [ ] **Step 3: Commit inventory**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/docs/database/raw-sql-inventory.md
git commit -m "docs(server): inventory raw sql usage"
```

Expected:

- Commit succeeds.

### Task 0.2: Add Architecture Guard For New Raw SQL

**Files:**

- Modify: `apps/server/internal/bootstrap/architecture/boundaries_test.go`

- [ ] **Step 1: Add a failing guard test**

Add this test to `boundaries_test.go`:

```go
func TestRepositoriesDoNotImportPlatformAuth(t *testing.T) {
	internalRoot := internalDir(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		cleanPath := filepath.ToSlash(path)
		if !strings.Contains(cleanPath, "/repository/") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("parse imports for %s: %v", path, parseErr)
		}

		for _, imp := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(imp.Path.Value)
			if unquoteErr != nil {
				t.Fatalf("unquote import in %s: %v", path, unquoteErr)
			}

			if importPath == modulePrefix+"platform/auth" {
				t.Errorf("repository boundary violation: %s imports %s", path, importPath)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk internal dir: %v", err)
	}
}
```

- [ ] **Step 2: Run the architecture tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/bootstrap/architecture
```

Expected:

- If any repository imports `internal/platform/auth`, the test fails and lists the file.
- If there are no current violations, the test passes and future violations are blocked.

- [ ] **Step 3: Commit guard test**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/bootstrap/architecture/boundaries_test.go
git commit -m "test(server): guard repository auth boundaries"
```

Expected:

- Commit succeeds.

---

## Phase 1: Add SQLC Without Changing Runtime Behavior

### Task 1.1: Add SQLC Config

**Files:**

- Create: `apps/server/sqlc.yaml`
- Modify: `apps/server/Makefile`

- [ ] **Step 1: Add `sqlc.yaml`**

Create `apps/server/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema:
      - "internal/migrations"
    queries:
      - "internal/modules/*/repository/*.sql"
      - "internal/platform/*/repository/*.sql"
    gen:
      go:
        package: "repository"
        out: "internal/repository"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: false
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true
        overrides:
          - db_type: "uuid"
            go_type:
              import: "github.com/google/uuid"
              type: "UUID"
          - db_type: "pg_catalog.uuid"
            go_type:
              import: "github.com/google/uuid"
              type: "UUID"
          - db_type: "timestamptz"
            go_type:
              import: "time"
              type: "Time"
          - db_type: "timestamp"
            go_type:
              import: "time"
              type: "Time"
```

- [ ] **Step 2: Add Makefile targets**

Append this section to `apps/server/Makefile`:

```makefile
# =============================================================================
# SQLC
# =============================================================================

SQLC ?= sqlc

sqlc-generate:
	$(SQLC) generate

sqlc-verify:
	$(SQLC) generate
	git diff --exit-code -- internal/repository
```

- [ ] **Step 3: Run generation**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
```

Expected:

- This may fail because there are no `.sql` files yet. If it fails with "no queries contained in paths", continue to Task 1.2 before committing.
- It must not fail due to invalid schema parsing. If it fails on a migration, fix the migration syntax for sqlc compatibility in a separate commit before continuing.

### Task 1.2: Add A Minimal Platform Query To Prove Generation

**Files:**

- Create: `apps/server/internal/platform/workspace/repository/queries.sql`
- Generated: `apps/server/internal/repository/*.go`

- [ ] **Step 1: Create the first query file**

Create `apps/server/internal/platform/workspace/repository/queries.sql`:

```sql
-- name: PingDatabase :one
SELECT 1::int AS value;
```

- [ ] **Step 2: Run sqlc**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
```

Expected:

- `internal/repository` is created.
- Generated code includes `PingDatabase`.

- [ ] **Step 3: Run server tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./...
```

Expected:

- Tests compile.
- Existing failures unrelated to generated code should be recorded before continuing.

- [ ] **Step 4: Commit SQLC foundation**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/sqlc.yaml apps/server/Makefile apps/server/internal/platform/workspace/repository/queries.sql apps/server/internal/repository
git commit -m "build(server): add sqlc generation foundation"
```

Expected:

- Commit succeeds.

---

## Phase 2: Add PGX Pool Bridge

### Task 2.1: Add New Pool Opener

**Files:**

- Create: `apps/server/internal/platform/database/pool.go`
- Test: `apps/server/internal/platform/database/pool_test.go`

- [ ] **Step 1: Write tests for DSN behavior**

Create `apps/server/internal/platform/database/pool_test.go`:

```go
package database

import "testing"

func TestBuildPostgresURLDisablesTLS(t *testing.T) {
	cfg := Config{
		User:       "postgres",
		Password:   "password",
		Host:       "localhost",
		Port:       "5432",
		Name:       "complexus",
		DisableTLS: true,
	}

	got := buildPostgresURL(cfg)

	if got.Scheme != "postgres" {
		t.Fatalf("scheme = %q, want postgres", got.Scheme)
	}
	if got.Query().Get("sslmode") != "disable" {
		t.Fatalf("sslmode = %q, want disable", got.Query().Get("sslmode"))
	}
	if got.Query().Get("timezone") != "utc" {
		t.Fatalf("timezone = %q, want utc", got.Query().Get("timezone"))
	}
}

func TestBuildPostgresURLRequiresTLS(t *testing.T) {
	cfg := Config{
		User:       "postgres",
		Password:   "password",
		Host:       "db.example.com",
		Port:       "5432",
		Name:       "complexus",
		DisableTLS: false,
	}

	got := buildPostgresURL(cfg)

	if got.Query().Get("sslmode") != "require" {
		t.Fatalf("sslmode = %q, want require", got.Query().Get("sslmode"))
	}
}
```

- [ ] **Step 2: Run the failing test**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/platform/database
```

Expected:

- FAIL because `Config` and `buildPostgresURL` do not exist yet.

- [ ] **Step 3: Add pool opener**

Create `apps/server/internal/platform/database/pool.go`:

```go
package database

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User         string
	Password     string
	Host         string
	Port         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

func OpenPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(buildPostgresURL(cfg).String())
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		poolConfig.MinConns = int32(cfg.MaxIdleConns)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func buildPostgresURL(cfg Config) url.URL {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	return url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host + ":" + cfg.Port,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}
}
```

- [ ] **Step 4: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/platform/database
```

Expected:

- PASS.

- [ ] **Step 5: Commit pool bridge**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/platform/database
git commit -m "feat(server): add pgx pool database opener"
```

Expected:

- Commit succeeds.

### Task 2.2: Add Transaction Helper

**Files:**

- Create: `apps/server/internal/platform/database/tx.go`
- Test: `apps/server/internal/platform/database/tx_test.go`

- [ ] **Step 1: Add transaction helper**

Create `apps/server/internal/platform/database/tx.go`:

```go
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
```

- [ ] **Step 2: Add a no-op compile test**

Create `apps/server/internal/platform/database/tx_test.go`:

```go
package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestWithTxSignature(t *testing.T) {
	var fn func(context.Context, interface{ Begin(context.Context) (pgx.Tx, error) }, func(pgx.Tx) error) error
	if fn != nil {
		t.Fatal("signature sentinel should stay nil")
	}
}
```

Then delete this test before committing if it does not add value. It only exists to force review of the transaction helper signature.

- [ ] **Step 3: Run package tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/platform/database
```

Expected:

- PASS.

- [ ] **Step 4: Commit helper**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/platform/database/tx.go
git commit -m "feat(server): add pgx transaction helper"
```

Expected:

- Commit succeeds.

---

## Phase 3: Move Workspace Access Out Of HTTP Middleware

### Task 3.1: Add Generated Workspace Queries

**Files:**

- Modify: `apps/server/internal/platform/workspace/repository/queries.sql`
- Generated: `apps/server/internal/repository/*.go`

- [ ] **Step 1: Replace ping query with real workspace queries**

Replace `apps/server/internal/platform/workspace/repository/queries.sql` with:

```sql
-- name: GetWorkspaceForUserBySlug :one
SELECT
  w.id,
  w.name,
  w.slug,
  w.avatar_url,
  w.plan,
  w.status,
  wm.role,
  wm.user_id
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE w.slug = $1
  AND wm.user_id = $2
  AND w.deleted_at IS NULL
  AND wm.deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserLastLoginAt :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: UpdateWorkspaceLastAccessedAt :exec
UPDATE workspaces
SET last_accessed_at = NOW(), updated_at = NOW()
WHERE id = $1;
```

If actual column names differ, inspect migrations and adjust before running `sqlc`.

- [ ] **Step 2: Generate code**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
```

Expected:

- Generated methods:
  - `GetWorkspaceForUserBySlug`
  - `UpdateUserLastLoginAt`
  - `UpdateWorkspaceLastAccessedAt`

- [ ] **Step 3: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./...
```

Expected:

- Compile passes.

### Task 3.2: Add Workspace Repository And Service

**Files:**

- Create: `apps/server/internal/platform/workspace/repository/repository.go`
- Create: `apps/server/internal/platform/workspace/service.go`

- [ ] **Step 1: Add repository wrapper**

Create `apps/server/internal/platform/workspace/repository/repository.go`:

```go
package repository

import (
	"context"

	generated "github.com/complexus-tech/projects-api/internal/repository"
	"github.com/google/uuid"
)

type WorkspaceInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	UserID    uuid.UUID `json:"userId"`
}

type Repository struct {
	q *generated.Queries
}

func New(q *generated.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetWorkspaceForUserBySlug(ctx context.Context, slug string, userID uuid.UUID) (WorkspaceInfo, error) {
	row, err := r.q.GetWorkspaceForUserBySlug(ctx, generated.GetWorkspaceForUserBySlugParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		return WorkspaceInfo{}, err
	}

	return WorkspaceInfo{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		AvatarURL: row.AvatarUrl,
		Plan:      row.Plan,
		Status:    row.Status,
		Role:      row.Role,
		UserID:    row.UserID,
	}, nil
}

func (r *Repository) MarkAccessed(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if err := r.q.UpdateUserLastLoginAt(ctx, userID); err != nil {
		return err
	}
	return r.q.UpdateWorkspaceLastAccessedAt(ctx, workspaceID)
}
```

- [ ] **Step 2: Add service**

Create `apps/server/internal/platform/workspace/service.go`:

```go
package workspace

import (
	"context"

	workspaceRepo "github.com/complexus-tech/projects-api/internal/platform/workspace/repository"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
)

type Repository interface {
	GetWorkspaceForUserBySlug(ctx context.Context, slug string, userID uuid.UUID) (workspaceRepo.WorkspaceInfo, error)
	MarkAccessed(ctx context.Context, workspaceID, userID uuid.UUID) error
}

type Service struct {
	repo  Repository
	cache *cache.Service
}

func NewService(repo Repository, cache *cache.Service) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Resolve(ctx context.Context, slug string, userID uuid.UUID) (workspaceRepo.WorkspaceInfo, error) {
	key := cache.WorkspaceCacheKey(slug, userID.String())

	var cached workspaceRepo.WorkspaceInfo
	if s.cache != nil {
		if err := s.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	info, err := s.repo.GetWorkspaceForUserBySlug(ctx, slug, userID)
	if err != nil {
		return workspaceRepo.WorkspaceInfo{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, key, info, cache.DefaultExpiration)
	}

	_ = s.repo.MarkAccessed(ctx, info.ID, userID)

	return info, nil
}
```

If `cache.WorkspaceCacheKey` or `cache.DefaultExpiration` names differ, use the existing cache helper names from `pkg/cache`.

- [ ] **Step 3: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/platform/workspace/...
```

Expected:

- PASS after aligning cache helper names and generated row nullable types.

### Task 3.3: Update Workspace Middleware

**Files:**

- Modify: `apps/server/internal/platform/http/middleware/workspace.go`
- Modify: API bootstrap wiring files under `apps/server/internal/bootstrap/api`

- [ ] **Step 1: Change middleware dependency**

Refactor `workspace.go` so it depends on an interface:

```go
type WorkspaceResolver interface {
	Resolve(ctx context.Context, slug string, userID uuid.UUID) (workspaceRepo.WorkspaceInfo, error)
}
```

The middleware should:

1. Read the workspace slug from route vars.
2. Read authenticated user ID from auth context.
3. Call `resolver.Resolve`.
4. Store workspace info in request context.
5. Return `401` or `403` consistently when user/workspace access is invalid.

- [ ] **Step 2: Remove direct SQL from middleware**

Delete direct usage of:

```go
*sqlx.DB
NamedQueryContext
NamedExecContext
```

from `workspace.go`.

- [ ] **Step 3: Wire service in API bootstrap**

In bootstrap, construct:

```go
queries := repository.New(pool)
workspaceRepository := workspaceRepository.New(queries)
workspaceService := workspace.NewService(workspaceRepository, cacheService)
```

Then pass `workspaceService` to the middleware builder.

- [ ] **Step 4: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/platform/http/... ./internal/bootstrap/api/... ./...
```

Expected:

- The server compiles.
- Workspace-scoped routes still resolve workspace context.

- [ ] **Step 5: Commit workspace migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/platform/workspace apps/server/internal/platform/http/middleware/workspace.go apps/server/internal/bootstrap/api apps/server/internal/repository
git commit -m "refactor(server): move workspace access to sqlc repository"
```

Expected:

- Commit succeeds.

---

## Phase 4: Migrate Read-Heavy Modules First

## Candidate Order

1. `reports`
2. `github`
3. `notifications`
4. `search`
5. `chatsessions`

Reports and GitHub should be prioritized because they feed the assistant's generative UI and integration tools.

### Task 4.1: Migrate Reports Repository

**Files:**

- Create: `apps/server/internal/modules/reports/repository/queries.sql`
- Modify: `apps/server/internal/modules/reports/repository/reports.go`
- Modify: `apps/server/internal/modules/reports/repository/queries.go`
- Test: `apps/server/internal/modules/reports/repository/reports_test.go`

- [ ] **Step 1: Extract current SQL into named queries**

Move stable queries from `queries.go` into `queries.sql` using names like:

```sql
-- name: GetSprintPerformanceReport :one
-- name: ListSprintStoryThroughput :many
-- name: ListMemberPerformanceRows :many
-- name: ListObjectiveProgressRows :many
```

Each query must include `workspace_id` and any relevant `team_id`, `sprint_id`, `objective_id`, or `user_id` filter directly in SQL.

- [ ] **Step 2: Add repository tests before replacing implementation**

Create tests that compare expected report shape against seeded or fixture data:

```go
func TestGetSprintPerformanceReportScopesByWorkspace(t *testing.T) {
	t.Skip("enable when test database harness is available")
}
```

If the server does not yet have a database test harness, add this test as skipped with a clear skip reason, then add the harness in a separate task before unskipping.

- [ ] **Step 3: Generate queries**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
```

Expected:

- Generated report params include explicit IDs.
- No generated param uses `interface{}`.

- [ ] **Step 4: Replace repository internals**

Keep public repository method signatures stable. Inside each method:

1. Convert service filters into generated params.
2. Call generated query.
3. Map generated rows into existing report models.
4. Return the same errors as before where possible.

- [ ] **Step 5: Run focused tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./internal/modules/reports/...
```

Expected:

- PASS or skipped DB tests with explicit skip reason.

- [ ] **Step 6: Run full compile**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./...
```

Expected:

- PASS or only known baseline failures unrelated to reports.

- [ ] **Step 7: Commit reports migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/modules/reports apps/server/internal/repository
git commit -m "refactor(server): migrate reports repository to sqlc"
```

Expected:

- Commit succeeds.

### Task 4.2: Migrate GitHub Repository

**Files:**

- Create: `apps/server/internal/modules/github/repository/queries.sql`
- Modify: `apps/server/internal/modules/github/repository/repository.go`
- Test: `apps/server/internal/modules/github/repository/repository_test.go`

- [ ] **Step 1: Split GitHub repository by responsibility**

If `repository.go` is too large to safely migrate in one pass, split it first:

```txt
internal/modules/github/repository/
  repository.go
  installations.go
  repositories.go
  links.go
  comments.go
  sync.go
  queries.sql
```

The split should move existing methods without changing behavior.

- [ ] **Step 2: Add generated query groups**

Use query names like:

```sql
-- name: GetGitHubInstallationByWorkspace :one
-- name: ListGitHubRepositoriesByWorkspace :many
-- name: ListActiveGitHubRepositoriesByWorkspace :many
-- name: GetGitHubWorkspaceSettings :one
-- name: UpsertGitHubRepository :one
-- name: CreateGitHubStoryLink :one
-- name: ListGitHubStoryLinks :many
-- name: RecordGitHubSyncResult :exec
```

- [ ] **Step 3: Replace `sqlx.In` with `ANY($1::uuid[])` or `ANY($1::text[])`**

Where the current implementation builds `IN (...)`, write generated SQL like:

```sql
WHERE id = ANY($1::uuid[])
```

or:

```sql
WHERE repository_full_name = ANY($1::text[])
```

- [ ] **Step 4: Generate and map types**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
go test ./internal/modules/github/...
```

Expected:

- Generated GitHub rows map cleanly into existing domain structs.
- GitHub service tests still compile.

- [ ] **Step 5: Commit GitHub migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/modules/github apps/server/internal/repository
git commit -m "refactor(server): migrate github repository to sqlc"
```

Expected:

- Commit succeeds.

---

## Phase 5: Migrate Core Write Modules

## Candidate Order

1. `stories`
2. `objectives`
3. `keyresults`
4. `sprints`
5. `teams`
6. `workspaces`
7. `users`

These modules should not be migrated without tests around creation, updates, ordering, permissions, and transaction rollback.

### Task 5.1: Migrate Stories Repository

**Files:**

- Create: `apps/server/internal/modules/stories/repository/queries.sql`
- Modify: `apps/server/internal/modules/stories/repository/*.go`
- Test: `apps/server/internal/modules/stories/service/stories_test.go`
- Test: `apps/server/internal/modules/stories/repository/stories_test.go`

- [ ] **Step 1: Identify write invariants**

Document these invariants in `apps/server/internal/modules/stories/repository/queries.sql` comments above the relevant queries:

```sql
-- Story write invariants:
-- 1. Every story must belong to exactly one workspace.
-- 2. Team-scoped story creation must validate team ownership by workspace.
-- 3. Sprint assignment must validate sprint ownership by workspace.
-- 4. State assignment must validate state ownership by workspace.
-- 5. Ordering updates must happen in one transaction.
-- 6. Deleted stories must not be returned by normal reads.
```

- [ ] **Step 2: Add tests for UI-created story path**

Create a service test that covers the same path the UI uses:

```go
func TestCreateStoryFromUIPathPersistsStoryAndActivity(t *testing.T) {
	t.Skip("enable when DB test harness exists")
}
```

The test should verify:

- story row exists
- workspace ID matches
- creator ID matches
- state ID belongs to the workspace
- activity row exists
- no duplicate order position is created

- [ ] **Step 3: Add generated story create query**

Use a query shape like:

```sql
-- name: CreateStory :one
INSERT INTO stories (
  id,
  workspace_id,
  team_id,
  sprint_id,
  state_id,
  title,
  description,
  created_by,
  assignee_id,
  position,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
)
RETURNING *;
```

Adjust columns to the actual migrations before generation.

- [ ] **Step 4: Move multi-step story writes to `pgx` transactions**

Replace:

```go
tx, err := r.db.BeginTxx(ctx, nil)
```

with:

```go
err := database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
	qtx := r.q.WithTx(tx)
	// generated calls here
	return nil
})
```

- [ ] **Step 5: Run focused tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
go test ./internal/modules/stories/...
```

Expected:

- Story package compiles.
- Any skipped DB tests have explicit harness skip reasons.

- [ ] **Step 6: Commit stories migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/modules/stories apps/server/internal/repository
git commit -m "refactor(server): migrate stories repository to sqlc"
```

Expected:

- Commit succeeds.

### Task 5.2: Migrate Objectives And Key Results Together

**Files:**

- Create: `apps/server/internal/modules/objectives/repository/queries.sql`
- Create: `apps/server/internal/modules/keyresults/repository/queries.sql`
- Modify: `apps/server/internal/modules/objectives/repository/*.go`
- Modify: `apps/server/internal/modules/keyresults/repository/*.go`
- Test: `apps/server/internal/modules/objectives/service/objectives_test.go`
- Test: `apps/server/internal/modules/keyresults/service/keyresults_test.go`

- [ ] **Step 1: Write failing tests for UI-created objectives and key results**

Add service-level tests for:

- create objective from UI payload
- create key result from UI payload
- update objective progress
- update key result progress
- reject cross-workspace status/owner references

Use skipped DB tests only if no harness exists yet.

- [ ] **Step 2: Replace dynamic pagination builders**

Where key results use dynamic query builders, prefer explicit generated queries:

```sql
-- name: ListKeyResultsByObjective :many
SELECT *
FROM key_results
WHERE workspace_id = $1
  AND objective_id = $2
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountKeyResultsByObjective :one
SELECT COUNT(*)::bigint
FROM key_results
WHERE workspace_id = $1
  AND objective_id = $2
  AND deleted_at IS NULL;
```

For optional filters, create separate named queries instead of a generic builder when there are only a few valid combinations.

- [ ] **Step 3: Generate and replace repository internals**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
go test ./internal/modules/objectives/... ./internal/modules/keyresults/...
```

Expected:

- Objective and key result packages compile.
- UI creation bugs caused by wrong column names are caught at generation time.

- [ ] **Step 4: Commit objectives and key results migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/modules/objectives apps/server/internal/modules/keyresults apps/server/internal/repository
git commit -m "refactor(server): migrate okr repositories to sqlc"
```

Expected:

- Commit succeeds.

### Task 5.3: Migrate Sprints And Sprint Analytics

**Files:**

- Create: `apps/server/internal/modules/sprints/repository/queries.sql`
- Modify: `apps/server/internal/modules/sprints/repository/*.go`
- Modify: `apps/server/internal/modules/reports/repository/queries.sql`
- Test: `apps/server/internal/modules/sprints/service/sprints_test.go`

- [ ] **Step 1: Add sprint detail and analytics queries**

Add named queries for:

```sql
-- name: GetSprintByID :one
-- name: ListSprintStories :many
-- name: GetSprintStoryCountsByState :many
-- name: GetSprintAssigneePerformance :many
-- name: GetSprintBurndownRows :many
```

- [ ] **Step 2: Use these queries for assistant-facing sprint tools**

Keep AI assistant tools calling services, but make those services use the generated report/sprint repository paths.

- [ ] **Step 3: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
go test ./internal/modules/sprints/... ./internal/modules/reports/...
```

Expected:

- PASS.

- [ ] **Step 4: Commit sprint migration**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/internal/modules/sprints apps/server/internal/modules/reports apps/server/internal/repository
git commit -m "refactor(server): migrate sprint data access to sqlc"
```

Expected:

- Commit succeeds.

---

## Phase 6: Migrate Worker Jobs And Task Handlers

### Task 6.1: Move Job SQL Behind Module Services

**Files:**

- Modify: `apps/server/pkg/jobs/*.go`
- Modify: `apps/server/internal/taskhandlers/*.go`
- Modify: `apps/server/internal/bootstrap/worker/*.go`

- [ ] **Step 1: Classify jobs**

Classify each job:

```txt
purge jobs: can use generated direct queries
automation jobs: should call module services where business rules exist
notification jobs: should call notification/user/story services
integration sync jobs: should call GitHub/Slack services
```

- [ ] **Step 2: Replace direct DB job signatures**

Move from:

```go
func ProcessSprintAutoCreation(ctx context.Context, db *sqlx.DB, log *logger.Logger) error
```

to either:

```go
func ProcessSprintAutoCreation(ctx context.Context, svc *sprints.Service, log *logger.Logger) error
```

or:

```go
func ProcessSprintAutoCreation(ctx context.Context, repo *repository.Queries, log *logger.Logger) error
```

Use services when the job changes product state. Use generated queries directly only for simple purge/cleanup jobs.

- [ ] **Step 3: Migrate purge jobs first**

For each purge job, add generated queries:

```sql
-- name: PurgeDeletedStories :execrows
DELETE FROM stories
WHERE deleted_at IS NOT NULL
  AND deleted_at < NOW() - INTERVAL '30 days';
```

Adjust retention windows to the existing job logic.

- [ ] **Step 4: Run worker compile**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./pkg/jobs/... ./internal/taskhandlers/... ./internal/bootstrap/worker/...
```

Expected:

- PASS.

- [ ] **Step 5: Commit worker migration batch**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/pkg/jobs apps/server/internal/taskhandlers apps/server/internal/bootstrap/worker apps/server/internal/repository
git commit -m "refactor(server): migrate worker data access to sqlc"
```

Expected:

- Commit succeeds.

---

## Phase 7: Expand Assistant Capabilities On Top Of Safe Queries

This phase should happen after `reports`, `github`, `stories`, `objectives`, `keyresults`, and `sprints` have sqlc-backed service methods.

### Task 7.1: Add Performance Service Methods

**Files:**

- Modify: `apps/server/internal/modules/reports/service/*.go`
- Modify: `apps/server/internal/modules/reports/repository/queries.sql`
- Test: `apps/server/internal/modules/reports/service/reports_test.go`

- [ ] **Step 1: Add service methods**

Add service methods for:

```go
GetWorkspacePerformance(ctx context.Context, workspaceID uuid.UUID, filters PerformanceFilters) (WorkspacePerformanceReport, error)
GetTeamPerformance(ctx context.Context, workspaceID, teamID uuid.UUID, filters PerformanceFilters) (TeamPerformanceReport, error)
GetSprintPerformance(ctx context.Context, workspaceID, sprintID uuid.UUID) (SprintPerformanceReport, error)
GetMemberPerformance(ctx context.Context, workspaceID, userID uuid.UUID, filters PerformanceFilters) (MemberPerformanceReport, error)
GetObjectivePerformance(ctx context.Context, workspaceID, objectiveID uuid.UUID) (ObjectivePerformanceReport, error)
```

- [ ] **Step 2: Add generated queries**

Add query groups:

```sql
-- name: GetWorkspacePerformanceSummary :one
-- name: GetTeamPerformanceSummary :one
-- name: GetSprintPerformanceSummary :one
-- name: GetMemberPerformanceSummary :one
-- name: GetObjectivePerformanceSummary :one
-- name: ListPerformanceTrendRows :many
-- name: ListPerformanceBreakdownRows :many
```

- [ ] **Step 3: Run tests**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
make sqlc-generate
go test ./internal/modules/reports/...
```

Expected:

- PASS.

### Task 7.2: Expose Assistant Tools Through Services

**Files:**

- Modify server AI assistant tool wiring files after locating their current path with:

```bash
cd /Users/joseph/development/complexus/fortyone
rg -n "tool|assistant|Maya|chat" apps/server apps/projects -S
```

- [ ] **Step 1: Add assistant tools that call services**

Assistant tools should call service methods, not repositories:

```txt
github.getWorkspaceIntegrationStatus
github.listLinkedRepositories
reports.getWorkspacePerformance
reports.getSprintPerformance
reports.getMemberPerformance
reports.getObjectivePerformance
stories.listBlockedStories
keyresults.listAtRiskKeyResults
```

- [ ] **Step 2: Add generated UI payloads**

Return typed payloads that the frontend can render:

```txt
githubIntegrationReport
workspacePerformanceReport
sprintPerformanceReport
memberPerformanceReport
objectivePerformanceReport
blockedStoriesReport
atRiskKeyResultsReport
```

- [ ] **Step 3: Verify no assistant tool bypasses auth**

Every assistant tool must receive:

- authenticated user ID
- workspace ID
- workspace role/access info

Every service call must scope by workspace ID.

- [ ] **Step 4: Run server and projects checks**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go test ./...

cd /Users/joseph/development/complexus/fortyone
pnpm --filter projects type-check
```

Expected:

- Server tests pass.
- Projects typecheck passes.

---

## Phase 8: Remove SQLX

This phase only starts after raw SQL inventory shows no runtime dependency on `sqlx` outside migration tooling or explicitly approved legacy code.

### Task 8.1: Remove SQLX From Runtime Wiring

**Files:**

- Modify: `apps/server/pkg/database/database.go`
- Modify: `apps/server/internal/bootstrap/api/*.go`
- Modify: `apps/server/internal/bootstrap/worker/*.go`
- Modify: `apps/server/go.mod`

- [ ] **Step 1: Confirm no runtime sqlx usage**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
rg -n "github.com/jmoiron/sqlx|\\*sqlx\\.DB|\\*sqlx\\.Tx|BeginTxx|NamedExec|Queryx|Select\\(|Get\\(" . --glob '*.go'
```

Expected:

- No hits in runtime files.
- If hits remain, migrate those files before continuing.

- [ ] **Step 2: Remove old database opener**

Delete or reduce `pkg/database/database.go` to only shared DB errors if still needed:

```go
package database

import (
	"database/sql"
	"errors"
)

var (
	ErrNoRows         = sql.ErrNoRows
	ErrDuplicateEntry = errors.New("duplicate entry")
)
```

- [ ] **Step 3: Remove sqlx dependency**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone/apps/server
go mod tidy
go test ./...
```

Expected:

- `github.com/jmoiron/sqlx` is removed from `go.mod`.
- Tests pass.

- [ ] **Step 4: Commit removal**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server
git commit -m "refactor(server): remove sqlx runtime dependency"
```

Expected:

- Commit succeeds.

---

## Phase 9: Documentation And CI

### Task 9.1: Add SQLC Repository Guide

**Files:**

- Create: `apps/server/docs/database/sqlc.md`

- [ ] **Step 1: Add repository guide**

Create `apps/server/docs/database/sqlc.md`:

```markdown
# SQLC Repository Guide

## Query Ownership

Module queries live beside the module repository:

`internal/modules/<module>/repository/queries.sql`

Platform queries live under:

`internal/platform/<area>/repository/queries.sql`

Generated code lives under:

`internal/repository`

Do not edit generated files manually.

## Repository Rules

- Repositories receive explicit IDs. Do not read user ID, workspace ID, or route values from context.
- Services own workflow and permission decisions.
- Repositories own persistence details and row-to-domain mapping.
- SQL must scope workspace-owned records by `workspace_id`.
- Soft-deleted records must be excluded by default.
- Use `ANY($1::uuid[])` instead of dynamically building `IN (...)`.
- Prefer multiple explicit queries over generic dynamic query builders.

## Transactions

Use `internal/platform/database.WithTx` for multi-step writes.

Inside a transaction:

```go
err := database.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
	qtx := r.q.WithTx(tx)
	return qtx.SomeWrite(ctx, params)
})
```

## Verification

Run:

```bash
make sqlc-generate
go test ./...
```

Before committing generated changes, verify:

```bash
git diff -- internal/repository
```
```

- [ ] **Step 2: Commit docs**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add apps/server/docs/database/sqlc.md
git commit -m "docs(server): add sqlc repository guide"
```

Expected:

- Commit succeeds.

### Task 9.2: Add CI Verification

**Files:**

- Modify the existing CI workflow after locating it:

```bash
cd /Users/joseph/development/complexus/fortyone
rg -n "go test|apps/server|server" .github . --glob '*.yml' --glob '*.yaml'
```

- [ ] **Step 1: Add sqlc generate check**

Add this to the server CI job:

```bash
cd apps/server
make sqlc-generate
git diff --exit-code -- internal/repository
go test ./...
```

- [ ] **Step 2: Commit CI update**

Run:

```bash
cd /Users/joseph/development/complexus/fortyone
git add .github apps/server
git commit -m "ci(server): verify generated sqlc code"
```

Expected:

- Commit succeeds.

---

## Risk Register

| Risk | Why It Matters | Mitigation |
| --- | --- | --- |
| Migration parsing fails in `sqlc` | Some migrations may contain SQL that runtime Postgres accepts but `sqlc` cannot parse | Fix migrations only when necessary; prefer adding minimal `sqlc`-friendly schema files if needed |
| Dual DB pools during migration | `sqlx.DB` and `pgxpool.Pool` could have different pool behavior | Use same config values; keep bridge short; remove `sqlx` after module migration |
| Nullable generated types differ from existing models | Generated rows may use pointers or pgtype wrappers | Map rows explicitly in repositories; do not leak generated rows into services |
| Dynamic filters become awkward | `sqlc` prefers static queries | Use explicit query variants; reserve tiny query builders for truly combinatorial filters |
| Transaction semantics change | Existing `BeginTxx` behavior may differ from pgx transaction behavior | Centralize `WithTx`; test rollback on create/update paths |
| Workspace scoping regressions | Unsafe queries could leak cross-workspace data | Every query for workspace-owned data includes `workspace_id`; add tests for cross-workspace rejection |
| Jobs bypass services | Jobs may duplicate business rules | Route state-changing jobs through services; allow direct generated queries only for purge jobs |
| Assistant tools bypass permissions | AI tools can expose sensitive workspace data | Assistant tools call services with authenticated user and workspace context only |

---

## Definition Of Done

The migration is complete when:

- `apps/server/sqlc.yaml` exists and is used in local and CI checks.
- Generated query code lives in `apps/server/internal/repository`.
- Runtime data access uses `pgxpool.Pool`.
- `github.com/jmoiron/sqlx` is removed from `apps/server/go.mod`.
- No repository imports `internal/platform/auth`.
- No HTTP middleware owns raw SQL.
- No module repository uses `NamedExec`, `Queryx`, `Select`, `Get`, `BeginTxx`, or `sqlx.In`.
- Story, objective, key result, sprint, GitHub, report, workspace, and assistant-facing service paths compile and have focused tests.
- AI assistant tools use service methods backed by typed queries.
- `go test ./...` passes from `apps/server`.
- `pnpm --filter projects type-check` passes when frontend generated UI contracts change.

---

## Recommended Execution Order

1. Phase 0: Inventory and guardrails.
2. Phase 1: SQLC foundation.
3. Phase 2: PGX pool bridge.
4. Phase 3: Workspace access migration.
5. Phase 4: Reports and GitHub migration.
6. Phase 5: Stories, objectives, key results, and sprints.
7. Phase 6: Worker jobs.
8. Phase 7: Assistant capabilities backed by safe queries.
9. Phase 8: Remove `sqlx`.
10. Phase 9: Documentation and CI.

This order gives the assistant better safe data access early without blocking the whole backend migration, and it addresses the UI-created story/objective/key-result failures in the same pass as the typed repository migration.
