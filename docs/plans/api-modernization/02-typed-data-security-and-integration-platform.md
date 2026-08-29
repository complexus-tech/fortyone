# FortyOne API Modernization: Typed Data, Security, and Integration Platform

**Status:** Active source of truth
**Scope:** persistence, authorization, public API, developer credentials, and provider integrations in `apps/server`  
**Companion plans:** [target Go architecture](./01-target-go-architecture.md) and [delivery, testing, and documentation](./03-delivery-testing-and-documentation-roadmap.md)
**Implementation status:** [evidence and crash-recovery ledger](./00-implementation-status.md)

### Relationship to the earlier sqlc plan

This document supersedes the architectural and sequencing decisions in the older [FortyOne SQLC Migration Implementation Plan](../../../apps/server/docs/superpowers/plans/2026-06-10-fortyone-sqlc-migration.md). That plan correctly identified pgx/sqlc, explicit tenant/user inputs, transaction helpers, module-owned SQL, and removal of middleware SQL as the direction. The current review updates it in material ways:

- `go.mod` now requires Go 1.25 rather than the older document's Go 1.23 baseline;
- generated code is module-local rather than one central generated repository package;
- sqlc is the end state for essentially every query, with a much stricter exception policy;
- security and tenant-isolation corrections precede mechanical query conversion;
- the migration includes jobs, consumers, webhooks, credentials, testing, CI, performance, and public API contracts;
- current module size, dependency, route, SQL, and test evidence determines the waves.

Keep the older file as historical context, but use this three-document plan as the implementation authority. Before execution, mark that relationship in the older plan so two apparently active roadmaps cannot drift independently.

## 1. Binding persistence decision

**sqlc is the default and the migration destination for all application SQL.**

FortyOne should migrate every normal query and command to handwritten SQL compiled by sqlc into typed Go methods. SQLx is not the target abstraction and should not be retained merely because a query is large. Large, complex SQL is one of the strongest reasons to use sqlc: the SQL remains visible while parameters and result shapes become compile-time checked.

The intended end state is:

- no application repository imports `github.com/jmoiron/sqlx`;
- no SQL strings in handlers, middleware, services, jobs, or task handlers;
- no untyped `map[string]any` persistence patches;
- no unchecked column, table, sort, or direction interpolation;
- no sqlc-generated row or parameter type exposed as a domain or HTTP contract;
- all tenant-owned queries include tenant scope and have isolation tests;
- all migrations, queries, generated code, and schema expectations are checked in CI.

### 1.1 Exceptional-query policy

An engineer may propose a non-sqlc query only after demonstrating that sqlc cannot express the required shape without unacceptable correctness or operational cost. “The query is dynamic,” “the query is long,” and “SQLx is easier” are not sufficient reasons.

An approved exception must have:

1. an ADR naming the exact query and owner;
2. evidence that static variants, typed filters, `CASE` ordering, arrays, views, or a report-specific projection are unsuitable;
3. a typed input object and typed result scanner;
4. strict allowlists for every identifier that cannot be a bind parameter;
5. workspace/team scoping and authorization tests;
6. query-plan and load evidence;
7. a removal condition and review date.

Even then, prefer native `pgx` rather than preserving SQLx as a second persistence framework. The reasonable target is **zero SQLx exceptions**.

## 2. Why sqlc, not SQLx

[sqlc](https://docs.sqlc.dev/en/latest/) parses SQL and generates type-safe Go code. Its generated methods expose typed parameter and result structures and a small database interface, as shown in its [query documentation](https://docs.sqlc.dev/en/latest/howto/select.html). SQLx remains a useful extension over `database/sql`, but its named arguments and struct scanning are runtime facilities; they do not compile the SQL into a checked Go contract. The [SQLx project description](https://github.com/jmoiron/sqlx) is explicit about those runtime conveniences.

sqlc improves FortyOne in the places that currently cause the most risk:

- renamed columns or changed nullability surface during generation/compilation;
- query parameter types become visible at call sites;
- row shapes are explicit rather than inferred from struct tags;
- complex SQL remains normal SQL that database engineers can inspect and tune;
- generated interfaces make database adapter testing possible without inventing broad repositories;
- migration and query validation can run before deployment.

sqlc does **not** prove authorization, logical correctness, useful indexing, or safe transaction boundaries. Those remain application responsibilities. A typed cross-workspace query is still unsafe.

## 3. Current persistence snapshot

The reviewed API uses pgx's `database/sql` driver through `*sqlx.DB`. A mechanical inventory found approximately:

| Pattern               | Occurrences | Migration consequence                                                |
| --------------------- | ----------: | -------------------------------------------------------------------- |
| `PrepareNamedContext` |         310 | Replace prepared named statements with generated methods.            |
| `NamedExecContext`    |          28 | Replace runtime named binding with typed command params.             |
| `GetContext`          |         448 | Replace with `:one` or an intentional optional result.               |
| `SelectContext`       |         212 | Replace with `:many` and explicit page bounds.                       |
| `BeginTxx`            |         104 | Move transaction ownership to use cases and `Queries.WithTx`.        |
| `ExecContext`         |         346 | Replace with typed `:exec`, `:execrows`, or `RETURNING`.             |
| `sqlx.In`             |           4 | Prefer PostgreSQL arrays/`ANY`, `unnest`, or generated variants.     |
| `Rebind`              |           6 | Remove; PostgreSQL placeholders are native in query files.           |
| Repository interfaces |    About 58 | Narrow them by consumer/use case rather than copying generated APIs. |

This count is directional because the working tree is active. It is sufficient to show that the migration must be systematic rather than opportunistic.

Confirmed complexity and correctness risks include:

- SQL outside repository packages, including HTTP, middleware, service, jobs, and task-handler code;
- map-based mutation methods in core work modules;
- dynamic update/sort construction in stories, objectives, key results, sprints, labels, and feedback;
- large query files that mix unrelated projections and mutations;
- transactions that span unclear ownership boundaries;
- tenant scope that is sometimes inferred from an ID instead of included in the query;
- a missing `rows.Err()` check in at least one manually iterated path;
- migration/schema validation absent from CI.

## 4. Target database stack

### 4.1 Driver and pool

Use native `pgx/v5` and `pgxpool` as the target runtime. The current system already uses the pgx stdlib driver underneath SQLx, so this is a controlled simplification rather than a database change.

The database platform package should own:

- pool construction and configuration;
- connection health/readiness;
- per-connection tracing and query telemetry;
- transaction runner;
- SQLSTATE classification;
- test database construction;
- shutdown.

It must not contain domain queries.

Before finalizing the pool settings, measure:

- PostgreSQL connection ceiling and reserved capacity;
- number of API and worker replicas;
- API request concurrency;
- Asynq worker concurrency and task query fan-out;
- long-running report/export behavior.

Record the selected formula and production override policy in an ADR. Do not copy a generic `MaxConns` value.

### 4.2 sqlc version and ownership

- Start the implementation on **sqlc v1.31.1**, the current stable release reviewed on 2026-08-27. Pin the exact version and checksum in developer tooling and CI; never install `@latest` in the supported workflow.
- Treat a sqlc upgrade as a generated-contract change: update the pin in one PR, regenerate every package, run the type-contract fixture and full repository suite, and review generated signature/nullability diffs before merging.
- Pin pgx independently in `go.mod`. The pgx version used to build the sqlc binary does not select FortyOne's runtime pgx version.
- Keep one root `apps/server/sqlc.yaml` so the schema source and generation policy are visible in one place.
- Use one explicit SQL block per module that owns queries, with each block generating into that module's `repository/sqlc` package. Do not create blocks for modules without persistence.
- Use the shared golang-migrate schema directory as input.
- Never hand-edit generated files.
- Include generated files in source control so application builds do not require sqlc, but make CI fail on generation drift.

sqlc understands migration directories and ignores down migrations while constructing the schema; its migration behavior is documented in [Using Go migrations](https://docs.sqlc.dev/en/latest/howto/ddl.html). FortyOne's zero-padded migration names are suitable for lexical ordering.

### 4.3 Reference-project comparison

The Art Circles API was reviewed at commit `36e9272a` as the locally available mature sqlc reference. Its checked-in generated headers identify sqlc v1.30.0. At that commit, 23 query files and 757 named operations generate into one central package containing roughly 48,000 lines of generated Go, a 757-method `Querier`, and a 3,248-line all-schema model file. That repository proves several pgx/sqlc mechanics at meaningful scale, but it also demonstrates why FortyOne must preserve module ownership.

| Reference choice or observed pattern                                                        | FortyOne decision                             | Reason                                                                                                                                                                                                  |
| ------------------------------------------------------------------------------------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `version: "2"`, PostgreSQL, `pgx/v5`                                                        | Adopt.                                        | This is the supported typed runtime direction and produces `DBTX`, `WithTx`, typed rows, and typed `:copyfrom` support.                                                                                 |
| One wildcard query input and one central generated package                                  | Reject.                                       | It creates a global query namespace, mega-interface, all-schema models, broad imports, and poor ownership. FortyOne generates per persisted module.                                                     |
| Shared UUID and time overrides in the reference's only block                                | Adopt globally after contract validation.     | FortyOne already uses `google/uuid` and pointer-based optional time/ID values; one top-level override policy prevents module drift.                                                                     |
| Separate nullable and non-nullable override entries                                         | Adopt.                                        | sqlc does not apply a non-null override to nullable columns.                                                                                                                                            |
| Both qualified and unqualified timestamp/date override names                                | Adopt for v1.31.1.                            | An isolated generation fixture confirmed that plain table columns match the unqualified names while expression inference can use catalog-qualified names. Both forms are locked by the config contract. |
| `emit_empty_slices: true`                                                                   | Adopt.                                        | Generated `:many` methods return an empty slice for zero rows and still check `rows.Err()`. Domain/HTTP mappers retain ownership of external collection semantics.                                      |
| `emit_interface: true`                                                                      | Adopt only module-locally.                    | A small generated query interface is useful inside an adapter; a cross-application generated interface is not a domain port.                                                                            |
| `emit_prepared_queries: false`                                                              | Adopt explicitly.                             | pgx has its own statement-cache behavior. Adding sqlc's explicit prepare lifecycle would increase pool/proxy lifecycle complexity without a demonstrated benefit.                                       |
| `ANY` with typed arrays                                                                     | Adopt.                                        | It avoids expanded SQL, but request/service code must cap array length and the query must retain tenant scope. FortyOne uses `CAST(value AS type)` in application SQL.                                  |
| Deliberate `sqlc.narg` filters                                                              | Adopt selectively.                            | Null must have documented query semantics; it must never become an ambiguous patch mechanism.                                                                                                           |
| Tenant-scoped `:execrows`                                                                   | Adopt.                                        | Affected-row counts can prove ownership, optimistic concurrency, or not-found behavior without leaking driver command tags.                                                                             |
| Bounded `:copyfrom`                                                                         | Adopt as the primary high-volume insert path. | The reference proves generated pgx copy code works; FortyOne additionally requires explicit batch limits, validation, atomicity, and repository mapping.                                                |
| Generated `Queries.WithTx`                                                                  | Adopt as a mechanism.                         | Transaction ownership stays in FortyOne's shared runner/use case, not scattered across repository methods.                                                                                              |
| `SELECT *`, `RETURNING *`, raw `pgx.CopyFrom`, and large parallel-array `unnest` operations | Reject.                                       | These weaken reviewability, type stability, bulk-row alignment, and the sqlc-everywhere end state.                                                                                                      |
| Global `geography -> string` and `vector -> pgvector.Vector` overrides                      | Do not copy.                                  | FortyOne has neither database type. Overrides are introduced only for an actual schema type with a proven pgx codec and domain mapping.                                                                 |
| Delete-selected-files generation target and whichever `sqlc` is on `PATH`                   | Reject.                                       | It can leave stale `querier.go`/`copyfrom.go` files and cannot reproduce generated output. FortyOne uses generated-only directories, an exact tool pin, and clean-output comparison.                    |

This comparison is evidence, not inheritance. In particular, no HTTP or service package may import a generated sqlc package or alias a generated table/row type.

### 4.4 Initial configuration shape

The following is the intended v1.31.1 policy for the first module. It uses top-level Go overrides so every module shares one type policy. The same config remains offline and migration-driven for generation because `analyzer.database` is false; the database URI is used only by database-aware vet rules. The command wrapper must reject an empty `SQLC_DATABASE_URL` before invoking `sqlc vet`, preventing an accidental fallback to a developer's default local socket/database.

```yaml
version: "2"

overrides:
  go:
    overrides:
      - db_type: "uuid"
        engine: "postgresql"
        go_type:
          import: "github.com/google/uuid"
          type: "UUID"
      - db_type: "uuid"
        engine: "postgresql"
        nullable: true
        go_type:
          import: "github.com/google/uuid"
          type: "UUID"
          pointer: true
      - db_type: "pg_catalog.timestamptz"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "pg_catalog.timestamptz"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true
      - db_type: "timestamptz"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "timestamptz"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true
      - db_type: "pg_catalog.timestamp"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "pg_catalog.timestamp"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true
      - db_type: "timestamp"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "timestamp"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true
      - db_type: "pg_catalog.date"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "pg_catalog.date"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true
      - db_type: "date"
        engine: "postgresql"
        go_type: "time.Time"
      - db_type: "date"
        engine: "postgresql"
        nullable: true
        go_type:
          type: "time.Time"
          pointer: true

sql:
  - name: "stories"
    engine: "postgresql"
    schema: "internal/migrations"
    queries: "internal/modules/stories/repository/queries"
    strict_function_checks: true
    strict_order_by: true
    analyzer:
      database: false
    database:
      uri: "${SQLC_DATABASE_URL}"
    rules:
      - "sqlc/db-prepare"
    gen:
      go:
        package: "storysql"
        out: "internal/modules/stories/repository/sqlc"
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
        initialisms:
          - "id"
          - "api"
          - "url"
          - "uri"
```

Repeat the module block explicitly for each persisted module. Keep blocks ordered by module name and require a config test that verifies unique `name`, `queries`, package, and output values. Introduce a manifest generator only if maintaining the explicit blocks becomes a measured source of drift; indirection is not the default.

`omit_unused_structs: true` is required because every block reads the shared schema; without it, each module may generate models for unrelated tables. Module-local output gives smaller APIs, prevents naming collisions, makes ownership obvious, and avoids coupling unrelated migrations to one generated model surface. `query_parameter_limit: 0` deliberately emits a parameter struct for every parameterized query, keeping call sites named and preventing a one-parameter query from changing invocation style when a second scope or concurrency field is added.

`emit_pointers_for_null_enum_types` is explicit because sqlc v1.31 changed nullable-enum behavior. A future upgrade must not silently swap `*Enum` and generated `NullEnum` shapes. Pointers still do not model the three states required by patch commands—leave unchanged, set null, and set a value—so patch intent remains separate typed commands/queries.

Keep the `initialisms` list narrow. sqlc uppercases the entire matched token; adding `ids`, for example, produces `IDS`, not idiomatic `IDs`. Generated spelling stays below the adapter and a module mapper exposes the correct domain name; use a targeted rename only when generated readability materially suffers. Keep `strict_function_checks` enabled. If the pinned parser cannot resolve a legitimate extension function, capture the smallest reproducer and choose a narrow, documented workaround rather than silently disabling function checks for every module.

Configuration options and type overrides should follow the pinned release's [configuration reference](https://docs.sqlc.dev/en/latest/reference/config.html) and [override guide](https://docs.sqlc.dev/en/latest/howto/overrides.html).

### 4.5 Type-mapping contract

The global override list is intentionally small. Every added override requires a schema use case, pgx encode/decode proof, domain mapping, round-trip integration test, and an ADR if it changes a public or cross-module semantic type.

| PostgreSQL shape              | Generated/repository policy                                                                             | Domain/API policy and required proof                                                                                                                                                                                                                      |
| ----------------------------- | ------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `uuid NOT NULL`               | `uuid.UUID`                                                                                             | Map to the module's ID type where one exists; reject zero IDs at the input/policy boundary.                                                                                                                                                               |
| nullable `uuid`               | `*uuid.UUID`                                                                                            | `nil` means SQL null only. Patch intent uses separate Set/Clear commands rather than overloading nil.                                                                                                                                                     |
| `timestamptz`                 | `time.Time` / `*time.Time`                                                                              | Treat as an instant, normalize/compare in UTC, preserve PostgreSQL microsecond precision in tests, and explicitly reject unsupported infinity values.                                                                                                     |
| `timestamp` without time zone | `time.Time` / `*time.Time` for migration compatibility                                                  | Each column documents its assumed location/meaning. Prefer migrating true instants to `timestamptz`; never silently call `UTC()` on an unspecified local wall time.                                                                                       |
| `date`                        | `time.Time` / `*time.Time` below the adapter initially                                                  | Map immediately to a date-only domain value, never interpret database midnight through a local timezone, and round-trip leap days/DST boundaries. Reconsider `pgtype.Date` or a codec-backed domain date if infinity/BC/valid-state fidelity is required. |
| PostgreSQL enum               | Generated enum and optional pointer, with generated `Valid()`                                           | Map to a module-owned domain enum. Generated values never become HTTP contracts. Unknown/invalid values are a hard mapping error and observable data-integrity event.                                                                                     |
| `numeric`                     | Keep `pgtype.Numeric` by default; no global float override                                              | Convert inside the repository according to column semantics. Currency uses an approved exact money/decimal or minor-unit model; key-result metrics get explicit precision/range rules. Never globally map `numeric` to `float64`.                         |
| `json`/`jsonb`                | Keep bytes/`json.RawMessage`-style storage at the generated boundary unless a column override is proven | Decode into a versioned typed payload with size/depth limits and unknown-field policy. Do not expose `map[string]any` as a reusable business contract.                                                                                                    |
| `uuid[]` and other arrays     | Typed slices inferred from element mapping                                                              | Bound count and total payload size before querying; define nil versus empty semantics; integration-test `ANY`, `unnest`, and empty inputs.                                                                                                                |
| `bytea`                       | `[]byte`                                                                                                | Copy when ownership crosses layers, cap size, and never log secret/blob content.                                                                                                                                                                          |
| future extension type         | No override by default                                                                                  | Add only with a real schema column, maintained pgx codec, null behavior, mapper, fixture, and operational ownership.                                                                                                                                      |

Before the first module freezes this policy, add an isolated sqlc configuration-contract fixture. Generation into a temporary directory must compile assertions for non-null/nullable UUIDs, timestamps, dates, nullable enums, numeric, JSONB, and UUID arrays; prove that one query parameter still produces a params struct; and prove that zero-row `:many` behavior is an empty slice. Do not golden-snapshot thousands of generated lines. Assert the small set of signatures and behaviors the application intentionally relies on.

The initial v1.31.1 fixture has already established three requirements that the first draft did not capture: global overrides work across query blocks, both qualified and unqualified timestamp/date override names are needed for the expected generated types, and the current FortyOne migration directory can compile into a tenant-scoped typed Story query with `uuid.UUID`, `*time.Time`, and `time.Time` fields. This was an isolated design check, not a repository integration test. Ticket 14 makes the experiment repository-owned and repeatable.

### 4.6 Generated interface policy

`emit_interface` can support repository adapter tests, but the generated `Querier` is not a domain port. A concrete repository wraps it:

```go
type Repository struct {
	q *storysql.Queries
}

func (r *Repository) Get(ctx context.Context, scope stories.Scope, id stories.ID) (stories.Story, error) {
	row, err := r.q.GetStory(ctx, storysql.GetStoryParams{
		WorkspaceID: scope.WorkspaceID,
		StoryID:     id,
	})
	if err != nil {
		return stories.Story{}, mapDatabaseError(err)
	}
	return mapStory(row), nil
}
```

The consuming service defines only the methods it needs. Generated parameter/row structs remain below the repository boundary.

### 4.7 Generation and verification workflow

Provide separate mutation and verification commands:

- `make sqlc-generate` intentionally removes only allowlisted generated-only `internal/modules/*/repository/sqlc` directories, regenerates with the pinned binary, formats generated Go if required by the pinned tool, and leaves reviewable working-tree changes;
- `make sqlc-check` asserts the tool version, runs offline `sqlc compile`, runs the configuration-contract fixture, generates into a temporary copy, compares every declared output directory, compiles the API, and fails on missing or stale files without mutating the developer's checkout;
- `make sqlc-vet` requires a non-empty ephemeral `SQLC_DATABASE_URL`, verifies that the database migration version is exactly the repository head, and runs the configured `sqlc/db-prepare` rule plus approved custom rules;
- `make generated-check` aggregates sqlc, OpenAPI, SDK, and generated configuration-reference drift checks.

Generated output directories contain generated files only—no mapper, repository adapter, hand-written test helper, or documentation. Cleanup resolves paths from the checked configuration, rejects symlinks and paths outside `internal/modules/*/repository/sqlc`, and fails closed if an unexpected hand-written file is present. Do not copy a `find | xargs rm` target that recognizes only today's generated filenames; sqlc can add `querier.go`, `copyfrom.go`, batch output, or renamed files as query annotations/options change.

`sqlc generate` and `sqlc compile` stay migration-source-driven and offline because `analyzer.database` is false. Database-backed `sqlc vet` runs only after CI/local integration setup has applied the exact migration chain; sqlc does not migrate that validation database itself. SQLC Cloud `verify` remains out of the required path unless an ADR deliberately accepts the external account, credentials, cost, availability, and vendor dependency.

## 5. SQL and query standards

### 5.1 Naming

Every query has a stable, intention-revealing name:

```sql
-- name: GetStoryForWorkspace :one
-- name: ListStoriesByTeam :many
-- name: ArchiveStoryIfVersion :one
-- name: InsertStoryActivity :exec
```

Avoid vague names such as `GetData`, `UpdateFields`, or `ListAll`. A query comment should explain non-obvious business, security, or performance intent—not restate SQL syntax.

### 5.2 Explicit columns

- Prohibit `SELECT *` in application query files.
- List insert/update return columns explicitly.
- Alias duplicate or computed fields intentionally.
- Keep row shapes small enough for the use case.
- Separate write models from report projections.

This reduces accidental contract expansion and makes schema changes reviewable.

### 5.3 Parameters

- Use positional parameters or `sqlc.arg(name)` for clarity.
- Use `sqlc.narg(name)` only when SQL null is a real part of the command semantics.
- Model optional, nullable, and zero values distinctly in the domain-to-database mapper.
- Bound array lengths before queries.
- Use PostgreSQL arrays with `ANY`, `unnest`, or typed temporary inputs instead of string-building `IN` clauses.
- Never bind identifiers; identifiers must come from a closed, typed allowlist or fixed query variant.

### 5.4 Results

- Use `:one` only when absence is exceptional or explicitly mapped to a domain not-found result.
- Use `:many` with a required maximum on user-controlled lists.
- Use `:execrows` or `RETURNING` when affected-row count proves optimistic concurrency or ownership.
- Map PostgreSQL nullable types at the repository boundary.
- Normalize empty list behavior at the API contract; `emit_empty_slices` avoids accidental JSON `null` where lists are expected.

### 5.5 Updates and patch semantics

Replace map-based updates with intent-specific commands. For example:

```text
UpdateStoryDetails(title, description, expectedVersion)
MoveStory(targetState, position, expectedVersion)
AssignStory(assignee, expectedVersion)
SetStoryDates(startDate, dueDate, expectedVersion)
ArchiveStory(expectedVersion)
```

This provides:

- a compile-time list of mutable fields;
- per-action authorization;
- correct omitted/null/value semantics;
- specific audit events;
- stable idempotency identity;
- smaller SQL statements;
- predictable optimistic concurrency.

If the product exposes a broad JSON `PATCH`, the transport layer should decompose it into a typed domain command. Do not pass a JSON map into persistence.

### 5.6 Filters and sorting

Use typed enums and one of these strategies, in order:

1. fixed query for the common filter/sort combination;
2. nullable typed filter parameters in a static query;
3. `CASE`-based ordering with a unique tie-breaker when the plan remains good;
4. a small set of generated query variants;
5. a view/materialized view for a stable reporting projection;
6. redesign the report/export boundary so a finite, typed sqlc contract exists.

Never concatenate raw request values. Every accepted sort field and direction must be parsed into an enum before reaching the repository. Do not execute user-authored SQL against the application database. If a genuinely unbounded analytics requirement still cannot be represented after the six strategies above, it must go through the exceptional-query ADR in Section 1.1; it is not a normal repository pattern.

### 5.7 Search

Keep search SQL stable and typed:

- normalize and bound query text;
- use appropriate PostgreSQL full-text/trigram/vector mechanisms only after measured need;
- make tenant/team scope part of the search query, not post-filtering;
- use stable score plus unique-ID ordering for cursors;
- prevent snippets or metadata from exposing inaccessible resources;
- maintain indexes and statistics with representative data tests.

### 5.8 Reports

Complex reports remain good sqlc candidates. Prefer:

- one named projection per report;
- CTEs for readability when plans remain sound;
- database views for stable shared projections;
- materialized views only with an explicit freshness and refresh contract;
- typed period, grouping, filter, and sort inputs;
- statement timeouts and maximum windows;
- async exports for expensive/high-volume results.

A 150-line readable SQL report is preferable to a Go query builder that hides joins and tenant rules.

### 5.9 Bulk operations

Use typed arrays, `unnest`, sqlc's supported `:copyfrom` command, or sqlc-managed staging-table queries according to measured volume. Direct `pgx.CopyFrom` is not a casual escape hatch; it requires the exceptional-query ADR if sqlc's generated copy method is genuinely insufficient. Every bulk command must define:

- maximum input size;
- all-or-nothing versus partial failure semantics;
- authorization of every referenced resource;
- deduplication behavior;
- transaction and lock strategy;
- emitted audit/outbox behavior;
- timeout and retry rules.

## 6. Tenant isolation and authorization in SQL

Compile-time query types do not solve broken object-level authorization. Every tenant-owned query must carry tenant scope:

```sql
-- name: DeleteComment :execrows
DELETE FROM comments
WHERE workspace_id = sqlc.arg(workspace_id)
  AND comment_id = sqlc.arg(comment_id);
```

Where the table does not currently hold `workspace_id`, join through the owning resource or add the denormalized tenant key with a consistency constraint. Prefer schemas that can express composite ownership constraints such as `(workspace_id, id)`.

Rules:

- never load by globally supplied resource ID and authorize only after mutation;
- include team scope when the actor has restricted team access;
- treat zero affected rows as not found/forbidden according to the API's non-enumeration policy;
- add cross-workspace negative repository tests for every resource family;
- preserve an explicit actor for audit, but do not trust an `actor_id` supplied by a client;
- consider PostgreSQL row-level security only as a later measured ADR/spike, not as a substitute for application policy.

The security phase must address confirmed ID-only mutation risks in comments, links, private user-memory paths, and related tenant resources before treating their sqlc migrations as complete. These are examples of OWASP's [Broken Object Level Authorization](https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/) risk.

## 7. Transactions and consistency

sqlc supports transaction-bound query sets through `WithTx`, documented in its [transaction guide](https://docs.sqlc.dev/en/latest/howto/transactions.html).

### 7.1 Ownership

The service use case owns a transaction when multiple database effects form one invariant. A repository may expose a transaction-aware adapter, but it should not surprise callers by starting unrelated transactions.

For a cross-module invariant, the transaction runner must supply a **tx-bound unit of work** containing only the consumer ports needed by that use case. Every port in the callback is constructed from the same `pgx.Tx`/`Queries.WithTx` binding. The callback must not retain the bundle, open a second transaction, or call a pool-backed sibling repository accidentally. This replaces today's raw `*sqlx.Tx` service signatures while still allowing workspace creation, invitations, membership, and related multi-module invariants to commit atomically.

### 7.2 Rules

- keep transactions short;
- load/lock only rows required for the invariant;
- use `FOR UPDATE` intentionally and document lock order;
- use optimistic version checks for collaborative work mutations where appropriate;
- never call Slack, GitHub, Stripe, email, object storage, or AI inside a transaction;
- write an outbox record in the same transaction, then deliver after commit;
- classify serialization/deadlock errors and retry only an explicitly safe, bounded callback;
- attach actor, workspace, operation, and trace metadata to transaction spans without SQL arguments containing secrets.

### 7.3 Migration of existing transactions

For each of the roughly 108 current transaction starts across the API's different transaction patterns (including about 104 `BeginTxx` calls):

1. identify the business invariant;
2. identify external side effects currently inside the boundary;
3. write a rollback/failure characterization test;
4. move the boundary to the service use case;
5. bind module sqlc queries with `WithTx`;
6. persist an outbox where external work follows;
7. test duplicate/retry and concurrent updates;
8. remove the old `BeginTxx` path.

## 8. PostgreSQL error mapping

Central database infrastructure may extract SQLSTATE, but modules decide meaning:

| PostgreSQL condition     | Adapter/domain mapping                                            |
| ------------------------ | ----------------------------------------------------------------- |
| no rows                  | not found or absent optional value                                |
| unique violation         | domain conflict, duplicate idempotency, or already exists         |
| foreign-key violation    | invalid reference or conflict; never expose table details         |
| check violation          | invalid domain state; should usually also be prevented before SQL |
| serialization/deadlock   | classified retryable only for safe transaction callbacks          |
| statement timeout/cancel | deadline/dependency timeout                                       |
| connection failure       | dependency unavailable                                            |

Do not compare provider/database error strings when a typed error or SQLSTATE is available.

## 9. Schema and migration policy

- Never edit a migration that has been applied; create a new migration.
- Continue checked-in paired up/down migrations, but classify each migration in the shared manifest as reversible or forward-only and make rollback semantics honest when data loss is unavoidable; an executable destructive down file is not proof of safe production rollback.
- Use expand/contract changes for zero-downtime deploys.
- Keep migrations focused and deterministically ordered.
- Add ownership/tenant keys, constraints, and indexes before relying on new query behavior.
- Avoid long table rewrites and unbounded backfills in a deploy transaction.
- Run backfills as resumable, observable jobs with checkpoints and verification.
- Add constraints as `NOT VALID` and validate later when appropriate for large tables.
- Verify empty and prior-version upgrades for all migrations; run down/up automatically only for migrations declared reversible, and require restore/forward-fix evidence for forward-only changes.
- Generate sqlc against the exact post-migration schema.
- Record production rollout, monitoring, and rollback steps for high-risk migrations.

CI should run `sqlc generate`, compile, and drift checks, plus sqlc's static/query checks. [`sqlc vet`](https://docs.sqlc.dev/en/latest/howto/vet.html) can use database-backed preparation for stronger validation. The available [CI guidance](https://docs.sqlc.dev/en/stable/howto/ci-cd.html) includes diff, vet, and verify workflows; cloud-dependent verification should remain an explicit product/vendor decision rather than an accidental requirement.

## 10. Per-query migration workflow

Each repository operation follows this checklist:

1. identify all callers and route/task exposure;
2. document current authorization, tenant, null, ordering, pagination, transaction, and error behavior;
3. add characterization and cross-tenant tests;
4. give the query an intention-revealing name;
5. move SQL into the module's `repository/queries` directory;
6. replace `SELECT *`, identifier interpolation, and map updates;
7. generate sqlc code;
8. add a repository mapper and domain error mapping;
9. replace the old implementation behind the existing service port;
10. run unit, PostgreSQL integration, race-relevant, and API contract tests;
11. compare query plan and latency for critical paths;
12. delete the old SQL and SQLx helper;
13. update module and data-access documentation;
14. mark the inventory item complete only when no caller uses the legacy path.

Do not maintain long-lived dual-write query implementations. If a staged rollout is required, make the compatibility window, comparison telemetry, and deletion condition explicit.

## 11. sqlc module migration waves

The waves are ordered to close security gaps, prove the tooling on small tenant-sensitive modules, establish identity/tenant semantics, and then tackle reference and high-complexity workflows. Module size alone does not decide sequence. Every module is listed so it is audited; a module with no owned SQL does not generate an empty sqlc package and completes its wave by proving that persistence is not applicable and its boundaries/tests are intentional.

| Wave | Module                                     | Primary migration concern                                       | Completion evidence beyond generation                                      |
| ---: | ------------------------------------------ | --------------------------------------------------------------- | -------------------------------------------------------------------------- |
|    0 | platform/database, bootstrap, taskhandlers | pgxpool, transaction runner, sqlc tooling, direct-SQL inventory | Pool/lifecycle tests; CI drift gate; architecture baseline.                |
|    1 | comments                                   | Cross-tenant mutation risk                                      | Workspace-scoped queries and negative authz tests.                         |
|    1 | links                                      | Cross-tenant ID-only access risk                                | Ownership joins/keys and resource policy tests.                            |
|    2 | users                                      | Authentication, sessions, private memories, system actors       | Credential separation, tenant-safe private data, deterministic tests.      |
|    2 | workspaces                                 | Membership lifecycle and bootstrap side effects                 | Transaction/outbox ownership; immediate authz invalidation.                |
|    2 | teams                                      | Membership and team scope                                       | Compound tenant keys and role/team policy.                                 |
|    2 | teamsettings                               | Admin/member policies and async effects                         | Typed settings commands and authorization matrix.                          |
|    2 | invitations                                | Privilege escalation and token handling                         | Admin policy, CSPRNG tokens, hashed lookup, rate limits.                   |
|    2 | subscriptions                              | Stripe idempotency/webhook correctness                          | Retry-safe webhook state machine and outbox.                               |
|    2 | admin                                      | Platform-admin separation                                       | Explicit admin principal/policy and immutable audit.                       |
|    2 | attachments                                | Storage metadata and ownership                                  | Tenant ownership plus async scan/optimization state.                       |
|    2 | documents                                  | Content ownership                                               | Scoped CRUD and object-storage consistency.                                |
|    2 | mentions                                   | Shared repository use                                           | Consumer-owned ports and typed references.                                 |
|    3 | stories                                    | Largest core repository, map updates, transactions              | Intent commands, optimistic concurrency, query plans, full auth matrix.    |
|    3 | activities                                 | Small read model; pagination                                    | Stable cursor and tenant filters.                                          |
|    3 | labels                                     | Dynamic update patterns                                         | Typed mutation commands and sort/filter enums.                             |
|    3 | states                                     | Small reference module                                          | Team/workspace scope and ordering contract.                                |
|    3 | objectivestatus                            | Small reference module                                          | Typed status lifecycle and uniqueness constraints.                         |
|    3 | epics                                      | No current SQL; core-work boundary audit                        | Intentional no-persistence archetype and consistent team ownership.        |
|    3 | okractivities                              | Event/audit-style writes                                        | Actor attribution and immutable append semantics.                          |
|    3 | workflowtemplates                          | Static service registry; no current SQL                         | Ownership/tests documented; no empty generated persistence package.        |
|    3 | sprints                                    | Dynamic filters/automation                                      | Typed filters, scheduling transactions, audit/outbox.                      |
|    3 | objectives                                 | Dynamic updates, cross-module activity                          | Intent commands and OKR invariant transactions.                            |
|    3 | keyresults                                 | Dynamic updates and OKR activity                                | Typed progress/update semantics and concurrency.                           |
|    3 | reports                                    | Large complex projections                                       | Named projections, bounded periods, plan budgets.                          |
|    3 | search                                     | Cross-resource visibility                                       | Tenant/team-safe results and stable ranked cursor.                         |
|    3 | notifications                              | Fan-out, reads, worker writes                                   | Outbox/fan-out ownership, cursor lists, retry safety.                      |
|    4 | messaging                                  | Existing strong inbox/outbox foundation                         | Preserve generation fencing and typed integration runtime ports.           |
|    4 | Slack                                      | Very large provider service/repository                          | Split capability stores; remove plaintext fallback; common gateway.        |
|    4 | GitHub                                     | Plaintext user token and synchronous webhook                    | Encrypt tokens, admin policy, async inbox, code-host adapter.              |
|    4 | calendar                                   | Google/Microsoft capability model                               | Consolidate secret storage; typed sync/watch/outbox stores.                |
|    4 | Figma                                      | Provider-specific crypto and design context                     | Shared credential vault; keep design-context boundary explicit.            |
|    4 | integrationrequests                        | Provider maps and prepared envelopes                            | Versioned envelope and provider-neutral acceptance contracts.              |
|    4 | emailreply                                 | Secure inbound delivery                                         | Retain derived auth, bounded body, encrypted durable inbox.                |
|    4 | emailagent                                 | Service-only email automation; no current SQL                   | Typed tasks, actor policy, delivery audit; no accidental persistence.      |
|    5 | feedback                                   | Large public ingress and contributor identity                   | Bounded/rate-limited ingress, explicit external identity, delivery outbox. |
|    5 | Maya                                       | AI tools, approvals, schedule recovery                          | Tool-level auth, transaction-free provider calls, deterministic clock.     |
|    5 | chatsessions                               | Approval/session state                                          | Optimistic leases, actor/audit model, expiry semantics.                    |
|    5 | agentreadiness                             | MCP OAuth and API exposure                                      | Scoped grant store, public/developer API separation.                       |
|    5 | health                                     | Dependency readiness queries                                    | Native pool checks and truthful readiness.                                 |

Wave completion requires removing SQLx usage for the listed modules, not merely adding sqlc beside it.

## 12. Security program before and during migration

sqlc migration must not delay confirmed security fixes. Phase 0 fixes should land behind regression tests before broad structural changes.

A mechanical route-token inventory found about 380 registrations: roughly 313 include authentication middleware, nine use optional authentication, 58 are public, 276 include workspace middleware, and only seven explicitly attach a rate-limit middleware. These counts do not prove that a route is correctly authorized—the service/resource policy still needs review—but they define the surface that the generated route/security inventory must classify. Public webhooks need signature/replay controls rather than user auth, and authenticated routes may still require role, team, resource, scope, and principal restrictions.

### 12.1 Immediate critical/high fixes

| Finding                                                                       | Required correction                                                                                                                                      | Proof                                                           |
| ----------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| Email OTP has weak generation and insufficient attempt throttling             | CSPRNG code/token, hashed challenge bound to purpose/user, short expiry, single use, attempt counter, per-account/IP/device limits                       | brute-force, replay, expiry, concurrency, and enumeration tests |
| Invitation create/list/revoke can be reached without a sufficient role check  | central workspace policy; require workspace admin for create, list, and revoke in route and service; validate allowed roles                              | exhaustive guest/member/admin matrix for all three operations   |
| Cross-tenant ID-only mutations/reads                                          | tenant key in policy and SQL; compound constraints where appropriate                                                                                     | negative tests using IDs from two workspaces                    |
| GitHub integration administration lacks admin policy                          | restrict install/settings/sync-link administration; distinguish read status from mutation                                                                | route and service policy matrix                                 |
| GitHub user OAuth tokens are plaintext                                        | use the minimal shared credential vault; backfill; remove plaintext read/write; plan migration to expiring GitHub App user tokens or eliminate the token | migration verification, decrypt/rotation tests, secret scanning |
| Slack installation still dual-writes plaintext token                          | stop dual-write, migrate/scrub legacy value atomically, monitor fallback use, then delete column/path                                                    | zero plaintext rows and no fallback metrics                     |
| Stripe webhook records failed processing as complete and acknowledges success | durable received/processing/succeeded/failed state; retryable response; idempotent handler                                                               | failure/retry/duplicate tests                                   |
| Unbounded/default request bodies and synchronous provider webhooks            | global bounded JSON decoder; provider-specific raw-body limits; durable quick-ack inbox                                                                  | oversized-body and provider deadline tests                      |
| Validation tags are often inert and trailing JSON is accepted                 | one enforced validator path; unknown fields rejected; exactly one JSON value; security-sensitive DTOs have explicit validation                           | malformed/trailing/unknown and invitation-role contract tests   |
| Query-string bearer tokens                                                    | remove from general auth; use purpose-built one-time signed URLs only where required                                                                     | logs/analytics test; query credential rejected                  |
| Shared JWT verification is under-specified                                    | credential-specific verifier with allowed algorithms, issuer, audience, type, expiry, key ID and rotation                                                | algorithm/issuer/audience/confusion tests                       |
| Authorization cache may retain revoked membership                             | authz epoch/version or authoritative membership validation that rejects stale cached authorization for every read and write after removal/demotion       | immediate read/write denial and cache-race tests                |
| Production-strength secret validation depends on email configuration          | one authoritative deployment mode; fail startup on default, missing, or too-short required production secrets                                            | production-config startup rejection tests                       |
| Asynqmon can be exposed without auth                                          | private bind/network policy plus authentication; disable by default                                                                                      | deployment/config test                                          |

These risks map directly to OWASP API Security categories including [Broken Authentication](https://owasp.org/API-Security/editions/2023/en/0xa2-broken-authentication/), [Broken Object Property Level Authorization](https://owasp.org/API-Security/editions/2023/en/0xa3-broken-object-property-level-authorization/), and [Unrestricted Resource Consumption](https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/).

### 12.2 Secret separation

Do not derive unrelated security mechanisms from one global application secret. Establish separate keys or managed key references for:

- browser sessions;
- first-party access tokens;
- OAuth authorization codes/access/refresh tokens;
- API key hashing/pepper;
- provider credential envelope encryption;
- outbound webhook signing;
- email reply ingress;
- temporary action links;
- messaging mutation confirmations.

Every key has an owner, purpose, algorithm, version/key ID, rotation process, validation at startup, and incident-revocation procedure. Logs and traces must contain key IDs at most—never key material, authorization codes, raw tokens, OTPs, or invite secrets.

### 12.3 Credential vault

Standardize the existing versioned AES-GCM approach into a credential vault boundary:

- generate a random data-encryption key (DEK) for the credential/record according to the approved reuse policy;
- encrypt the credential with an approved authenticated-encryption algorithm and random nonce;
- wrap the DEK with a managed key-encryption key (KEK/KMS key);
- AAD binding provider, workspace, installation, credential type, and generation;
- persist an explicit envelope such as `{algorithm, aad_version, nonce, ciphertext, wrapped_dek, kek_id, kek_version}`;
- support KEK rewrap without exposing or re-encrypting credential plaintext unnecessarily;
- atomic rotation with compare-and-swap generation;
- no plaintext fallback after migration;
- narrowly scoped decrypt operation at the provider call boundary;
- redacted types whose `String`/JSON/log behavior cannot expose secrets;
- audit for creation, access class, rotation, revocation, and migration without secret values.

Use envelope encryption backed by a managed KMS for production. If direct KMS encryption is deliberately chosen for a small payload instead, name it accurately and record its size, latency, availability, and rotation consequences in the ADR. Local development can use a documented local key provider with the same interface and envelope format.

## 13. Principal and authorization model

### 13.1 Principal types

| Principal             | Meaning                                                        | Attribution                                                    |
| --------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| Human user            | Interactive FortyOne member                                    | The user ID and credential/session used.                       |
| Personal token        | Automation acting with one user's rights                       | The user remains actor; token ID is credential.                |
| Service account       | Workspace-owned non-human automation                           | Service-account principal; never installer/user impersonation. |
| OAuth app, user actor | Third-party app acting for a consenting user                   | User actor plus app/grant/credential IDs.                      |
| OAuth app, app actor  | Installed application acting as itself                         | App installation principal with granted scopes.                |
| System actor          | First-party scheduled/internal operation                       | Named system principal and initiating cause.                   |
| External contributor  | Verified provider/feedback identity without a FortyOne account | Explicit external identity; never fabricated user membership.  |

Every audit event stores actor kind/ID, workspace, credential or installation ID, operation, resource, request/delivery ID, result, and safe metadata.

### 13.2 Authorization inputs

Authorization decisions may use:

- credential scopes;
- workspace role;
- team membership/restriction;
- resource ownership/visibility;
- actor type;
- app installation grant;
- product plan/capability;
- operation and requested property changes.

The service policy is authoritative. Middleware may reject missing coarse scope early, but it cannot prove resource ownership before the resource is loaded.

### 13.3 Scope catalog

Start with resource/action scopes rather than one coarse `mcp:access`:

```text
workspaces:read
teams:read
stories:read
stories:write
comments:read
comments:write
labels:read
labels:write
sprints:read
objectives:read
objectives:write
webhooks:manage
integrations:manage
service_accounts:manage
```

Separate administrative scopes from normal resource writes. Scopes narrow existing product permission; they never expand it. Publish the exact endpoint-to-scope mapping and version it compatibly.

## 14. Developer credentials

### 14.1 Personal access tokens

PATs are for scripts acting as a user:

- recognizable prefix and random 32-byte-or-stronger secret;
- select scopes and optional team restrictions;
- mandatory name and expiry, with bounded maximum lifetime unless policy explicitly allows otherwise;
- show secret once;
- store prefix plus keyed digest, never plaintext;
- constant-time verification;
- last-used timestamp updated asynchronously/coarsely to avoid a write per request;
- revoke and rotate controls;
- cannot outlive or exceed the user's access;
- automatically invalid or narrowed after user suspension/removal.

### 14.2 Service-account API keys

Service accounts are workspace-owned automation principals:

- admin-created and separately named;
- explicit scopes and optional team allowlist;
- one or more independently revocable keys;
- creator is audit metadata, not the runtime actor;
- expiry, rotation overlap, last-used, and revocation reason;
- plan/quantity controls where appropriate;
- no login, invitation, billing ownership, or privilege escalation by default.

A representative schema is:

```text
principals
  id, workspace_id, kind, name, status, created_by, created_at, disabled_at

api_credentials
  id, principal_id, workspace_id, prefix, secret_digest, scopes,
  team_restrictions, expires_at, last_used_at, revoked_at, revoked_by,
  rotated_from_id, created_at
```

Prefer normalized scope/team grant tables if array/JSON constraints and query plans are weaker. The exact schema requires an ADR and migration tests.

### 14.3 OAuth applications

For multi-workspace third-party integrations, implement authorization code flow with PKCE, exact redirect URI matching, state binding, short-lived audience/resource-bound access tokens, rotating refresh-token families, reuse detection, consent, revoke, and app-installation lifecycle. Follow the current OAuth security best practice in [RFC 9700](https://www.rfc-editor.org/info/rfc9700/).

Authorization codes and refresh tokens are high-entropy bearer secrets: store only a keyed digest/hash and lookup prefix, enforce single use/expiry, and never log them. Store confidential-client secrets as verifiable digests, show them once, and support overlapping rotation/revocation. When a rotated refresh token is reused, revoke the active token family and require reauthorization; test concurrent legitimate refresh separately from replay. Bind issued tokens to the intended FortyOne API resource/audience and granted installation/actor so a token for MCP or another service cannot be replayed as a general API credential. Evaluate sender-constrained tokens later if the ecosystem threat model requires them.

Separate:

- app registration/client metadata;
- workspace installation;
- user consent/grant;
- access credential;
- rotating refresh family;
- app/user actor choice;
- webhook subscriptions.

Dynamic client registration should not accept arbitrary long-lived public clients without governance. The existing MCP OAuth implementation has useful PKCE and refresh foundations but grants a coarse scope and should not simply become the public developer platform unchanged.

## 15. Public API contract

### 15.1 Version and source of truth

Create `/api/v1` as a deliberately supported external contract. Keep current web/mobile routes as internal/first-party while adapters are built. Do not expose all 380 routes by copying them into a spec.

Use one split OpenAPI source under `apps/server/api/openapi/v1`. Generate strict standard-library server interfaces/models with a pinned tool such as [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen), then implement adapters into existing services. The current [OpenAPI specification](https://spec.openapis.org/oas/latest.html) is the contract reference, while generator support must determine the exact spec version chosen.

There are currently two small agent-readiness OpenAPI documents with different path sets and no general product schema. Treat that contract as MCP-specific; consolidate copies and do not misrepresent it as the public FortyOne API.

The current product documentation also refers to API-token management, but the database and route review found no general customer PAT/service-account implementation. Treat this as contract debt: correct the product wording until the credential platform exists, then publish behavior only when its schema, lifecycle, and tests are real.

Start with REST rather than adding GraphQL during the modernization. Linear demonstrates a strong GraphQL developer platform, while Shortcut demonstrates a usable versioned REST/OpenAPI surface. FortyOne already has REST resources, `net/http`, and a stated OpenAPI/documentation need. REST therefore minimizes simultaneous change and gives generated contracts/SDKs quickly. Revisit GraphQL through an ADR only if measured client use cases require cross-resource selection that REST cannot serve cleanly; do not operate two public paradigms by default.

### 15.2 Initial resource surface

Start with stable collaboration primitives:

- workspaces visible to the actor;
- teams;
- stories;
- comments;
- labels;
- states/workflows where safe;
- sprints;
- objectives and key results;
- webhook endpoints/subscriptions/deliveries.

Do not expose billing administration, platform admin, unrestricted Maya tools, raw provider credentials, internal inbox/outbox records, or unstable internal settings in v1.

### 15.3 Contract rules

- resource IDs and workspace context are unambiguous;
- cursor pagination is consistent;
- filters/sorts are enumerated and documented;
- errors have stable codes and request IDs;
- create and retryable mutation endpoints support `Idempotency-Key`;
- optimistic updates use a version or ETag/`If-Match` where lost updates matter;
- timestamps are UTC RFC 3339 with defined precision;
- null, omitted, and empty collection behavior is specified;
- rate-limit headers and retry behavior are documented;
- deprecation includes announcement, sunset policy, changelog, and compatibility tests;
- webhook and API payloads are versioned independently when necessary;
- examples are tested against the schema.

### 15.4 SDKs

Generate at least TypeScript and Go SDKs from the committed contract after the API stabilizes. Generated SDKs should provide:

- typed resources and errors;
- credential configuration without logging tokens;
- pagination iterators;
- idempotency/request-ID support;
- retry only for documented safe/transient cases;
- webhook verification helpers;
- version compatibility metadata.

SDK generation is downstream of the OpenAPI contract, never the source of truth.

## 16. Integration platform architecture

### 16.1 Two different extension mechanisms

Do not confuse internal Go adapters with third-party integrations:

1. **First-party adapters** compile into `apps/server` and implement narrow Go capability interfaces.
2. **Third-party integrations** run out of process and use the public API, OAuth/service accounts, and webhooks.

Do not load arbitrary Go plugins into the API process. Go plugin ABI coupling, secret access, deployment coordination, and crash/security blast radius make that inappropriate for a public ecosystem. If an enterprise later requires deployable custom logic, use a versioned RPC/HTTP sidecar contract with isolation and an explicit threat model.

### 16.2 Shared control plane, capability-specific data planes

The common integration control plane owns:

- provider descriptor and configuration validation;
- installations and granted capabilities;
- encrypted credentials and rotation generation;
- external identity links;
- OAuth state/nonce lifecycle;
- webhook endpoint verification and durable inbox;
- outbound deliveries/outbox;
- retry/error classification;
- audit and operator redelivery;
- provider rate budgets and health.

Capability families own provider behavior:

| Family           | Providers                        | Example capabilities                                                                         |
| ---------------- | -------------------------------- | -------------------------------------------------------------------------------------------- |
| Code host        | GitHub, GitLab                   | repository catalog, issues/work items, comments, commits, pull/merge requests, sync mappings |
| Messaging        | Slack, Microsoft Teams, WhatsApp | messages, threads, commands, forms/cards, account linking, assistant requests                |
| Calendar         | Google, Microsoft                | calendars, events, watch subscriptions, scheduling and availability                          |
| Support/feedback | Zendesk, Intercom, email         | tickets/conversations, contacts, comments, attachments, delivery callbacks                   |
| Design context   | Figma                            | files/nodes/comments/webhooks and story design references                                    |

Figma design context should not be forced into the same semantics as customer feedback. Provider families share control-plane machinery without pretending their domain models are identical.

### 16.3 Capability-sized interfaces

Avoid `Provider` with dozens of methods. Prefer ports such as:

```go
type WebhookVerifier interface {
	Verify(ctx context.Context, request SignedRequest) (VerifiedDelivery, error)
}

type TokenRefresher interface {
	Refresh(ctx context.Context, installation InstallationRef) (RotatedCredential, error)
}

type RepositoryCatalog interface {
	ListRepositories(ctx context.Context, installation InstallationRef, cursor Cursor) (RepositoryPage, error)
}

type WorkItemWriter interface {
	CreateWorkItem(ctx context.Context, installation InstallationRef, cmd ExternalWorkItem) (ExternalWorkItemRef, error)
}

type CommentWriter interface {
	AddComment(ctx context.Context, installation InstallationRef, cmd ExternalComment) (ExternalCommentRef, error)
}
```

Provider SDK request/response types stay in adapter packages. Adapters declare capabilities so product behavior degrades deliberately when GitLab, Slack, Teams, or another provider lacks a feature.

### 16.4 GitHub/GitLab reuse boundary

GitHub and GitLab should share code-host domain ports for:

- installation/grant identity;
- repository/project catalog;
- issue/work-item mapping;
- comment synchronization;
- commit/branch/PR/MR references;
- webhook delivery normalization;
- rate-limit and retry classification;
- reconciliation and echo suppression.

They should not share raw payload structs, authentication implementation, API clients, or assumptions that GitHub pull requests equal GitLab merge requests in every detail. Provider-specific capabilities and mappings stay in adapters.

For GitHub specifically, encrypting the current standalone user OAuth token is the urgent containment step, not the desired final credential model. During the code-host migration, prefer expiring/refreshable GitHub App user access tokens when user attribution requires them, or remove stored user access entirely when installation credentials provide the required behavior. Document why each user token is needed and its expiry/revoke path.

### 16.5 Provider registry

A first-party provider descriptor should declare:

- stable provider key and display metadata;
- family/capabilities;
- configuration requirements;
- OAuth/install/account-link strategies;
- requested and supported scopes;
- webhook verifier and event normalizer factories;
- outbound adapter factories;
- task handlers and schedules;
- health checks;
- data retention and disconnect behavior.

Registration is explicit at compile time. Bootstrap reads descriptors to wire common control-plane behavior, reducing hard-coded maps without hiding dependencies through reflection.

## 17. Webhook platform

### 17.1 Inbound provider delivery

Standardize the strongest existing Slack/messaging behavior:

1. cap headers and raw body before allocation grows unbounded;
2. preserve untouched bytes for signature verification;
3. ask the provider verifier to validate the assurances that provider supplies: signature/authenticity, signed timestamp/replay window when available, and route/account identity;
4. derive and deduplicate a stable provider delivery ID even when the provider has no signed timestamp, as with GitHub's delivery ID;
5. store a versioned canonical envelope in a unique durable inbox;
6. encrypt retained raw provider payload when replay requires it;
7. acknowledge within the provider deadline;
8. process asynchronously and idempotently;
9. classify retryable, rate-limited, revoked, invalid, and terminal outcomes;
10. expose safe redelivery and dead-letter tooling;
11. expire payload content according to documented retention while retaining safe audit facts.

GitHub and Microsoft calendar ingress should migrate to this path rather than synchronously processing unbounded bodies.

A canonical envelope needs at least:

```text
version, provider, event_type, delivery_id,
external_account_id, installation_id, installation_generation,
workspace_id, received_at, trace_id, payload_ciphertext_ref
```

Queue the inbox ID, not the raw payload or credential.

### 17.2 Outbound FortyOne webhooks

Public app webhooks require:

- endpoint and event-subscription resources;
- generated per-endpoint signing secret, shown once and rotatable with overlap;
- HTTPS endpoint validation plus controlled egress on **every** attempt: resolve DNS, reject loopback/private/link-local/metadata/reserved addresses, connect to the validated address while preserving TLS hostname verification, disable redirects or re-resolve/revalidate every redirect, and bound connect/total time and response bytes;
- audit the resolved destination safely and defend against DNS rebinding rather than trusting creation-time URL validation;
- recheck app installation status/generation, current grant/scopes, event subscription, and endpoint status at enqueue and delivery so revoked apps cannot receive queued tenant events;
- signed delivery ID, timestamp, and exact body;
- replay-window guidance and verification helpers;
- immutable delivery attempts and response metadata with size limits;
- exponential backoff with jitter;
- operator/developer redelivery;
- disablement policy for persistent terminal failure;
- tenant fairness and per-endpoint concurrency limits;
- payload schema/version and event catalog;
- at-least-once semantics and explicit ordering limitations.

[Standard Webhooks](https://www.standardwebhooks.com/) is a useful interoperability model. Match its approach where compatible rather than inventing novel header/signature semantics.

### 17.3 Reconciliation and echo prevention

Bi-directional sync must persist mappings and origin metadata:

- FortyOne resource/version ↔ provider resource/version;
- last successfully applied source revision;
- origin/correlation/idempotency identity;
- conflict policy;
- tombstone/deletion semantics;
- periodic reconciliation cursor and status.

Events caused by FortyOne's outbound write must not loop back into an identical write. Suppression cannot rely only on message text or timing.

## 18. External product research and adopted lessons

The target combines relevant strengths rather than copying one product.

| Product/platform | Official behavior reviewed                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | FortyOne decision                                                                                                                |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| Linear           | OAuth scopes, PKCE, user/app actor choice, rotating credentials, app manifests, signed webhooks with delivery identity, cursor pagination, rate limits, typed SDK ([OAuth](https://linear.app/developers/oauth-2-0-authentication), [app actors](https://linear.app/developers/oauth-actor-authorization), [app manifests](https://linear.app/developers/oauth-app-manifests), [webhooks](https://linear.app/developers/webhooks), [pagination](https://linear.app/developers/pagination), [rate limits](https://linear.app/developers/rate-limiting), [SDK](https://linear.app/developers/sdk)) | Explicit app actors, least-privilege consent, cursor-first API, signed async webhooks, typed SDKs.                               |
| Shortcut         | Current REST v4 uses read/write token classes, workspace-scoped paths, cursor pagination and field filtering; webhooks expose signed event identity ([REST v4](https://developer.shortcut.com/api/rest/v4), [webhooks](https://developer.shortcut.com/api/webhook/v1))                                                                                                                                                                                                                                                                                                                           | Publish a usable REST/OpenAPI surface with finer resource/action and team restrictions than read/write alone.                    |
| GitHub Apps      | Fine-grained permissions, installation/repository scope, short-lived installation tokens ([permissions](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app), [installation tokens](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app))                                                                                                                                                                                             | Prefer install-scoped app credentials and least privilege over long-lived user tokens.                                           |
| GitLab           | Multiple token principal types and expirations; signed webhook evolution ([API authentication](https://docs.gitlab.com/api/rest/authentication/), [webhooks](https://docs.gitlab.com/user/project/integrations/webhooks/))                                                                                                                                                                                                                                                                                                                                                                       | Model credential kind explicitly and make signature/version handling an adapter capability.                                      |
| Slack            | Granular OAuth scopes, separate Sign in with Slack identity/OIDC, token rotation, and raw-body signing ([authentication](https://docs.slack.dev/authentication/), [Sign in with Slack](https://api.slack.com/authentication/sign-in-with-slack), [token rotation](https://docs.slack.dev/authentication/using-token-rotation), [request verification](https://docs.slack.dev/authentication/verifying-requests-from-slack))                                                                                                                                                                      | Use OIDC for account linking rather than bot-install scopes; minimize install scopes, rotate atomically, retain quick-ack inbox. |

The comparison is guidance, not compatibility. FortyOne should document its own invariants and threat model.

## 19. Security and integration work packages

### DS0 — Baseline and freeze unsafe growth

Deliverables:

- machine-readable inventory of every SQL call, dynamic fragment, transaction, table, module, caller, and test;
- route/auth/workspace/role/rate-limit inventory;
- credential and secret-flow inventory;
- webhook ingress/outbound inventory;
- fail CI when new SQLx/direct-SQL or unbounded webhook paths are added.

Acceptance:

- every current item has an owner and migration wave;
- baseline exceptions can decrease but not grow.

### DS1 — Close confirmed security defects

Deliverables:

- OTP/invitation/cross-tenant/GitHub admin/Stripe fixes;
- minimal shared envelope-encryption vault, GitHub token migration through it, and Slack plaintext removal;
- credential-specific JWT verification and query-token removal;
- enforced validation plus body/rate/resource limits and Asynqmon protection;
- revocation-safe authorization cache behavior for reads and writes;
- one production deployment mode with fail-fast required-secret validation.

Acceptance:

- regression and abuse tests pass;
- migration verification finds no prohibited plaintext token;
- security review signs off before structural refactors touch these paths.

### DS2 — Introduce pgx/sqlc foundation

Deliverables:

- pinned tools and configuration;
- module-local generated packages;
- pgxpool database platform and transaction runner;
- SQLSTATE mapper and test database harness;
- CI generation/compile/vet/drift/migration gates.

Acceptance:

- one Wave 1 module runs fully on sqlc without SQLx;
- generated drift fails locally and in CI;
- pool/startup/shutdown and transaction behavior are tested.

### DS3 — Complete module waves

Deliverables:

- execute Waves 1–5;
- replace map updates and identifier interpolation;
- move all direct SQL to module repositories;
- add tenant/transaction/query-plan evidence;
- remove SQLx dependency.

Acceptance:

- `rg 'jmoiron/sqlx' apps/server --glob '*.go'` returns no application use;
- direct SQL architecture check is clean except for an exact file/query allowlist generated from approved exceptional-query ADRs;
- every inventory item is closed or has the near-zero exceptional-query ADR, owner, tests, and removal condition required by Section 1.1;
- repository integration suite passes against a migrated schema.

### AP1 — Principal and credential platform

Deliverables:

- actor/principal model;
- centralized policy/scopes;
- PAT and service accounts;
- API key hash/rotation/revoke/expiry;
- OAuth app/grant/refresh model;
- unified immutable audit.

Acceptance:

- exhaustive scope × role × team × principal tests;
- secret lifecycle and incident runbooks exercised;
- app/service actions never appear as an installer action.

### AP2 — Public API v1

Deliverables:

- single OpenAPI source;
- generated strict adapters;
- initial resources, cursor/error/idempotency/version semantics;
- rate limiting and compatibility checks;
- TypeScript and Go SDK preview.

Acceptance:

- every v1 operation has contract, authorization, repository, and example tests;
- undocumented external routes are blocked from accidental exposure;
- breaking spec changes fail CI.

### IP1 — Common integration control plane

Deliverables:

- provider registry/capabilities;
- installation, credential, identity-link, inbox/outbox, and audit contracts;
- common webhook gateway and operator recovery;
- migrate GitHub then remaining provider ingress onto the reference runtime.

Acceptance:

- provider SDK types remain inside adapters;
- duplicate/replayed/out-of-order/revoked-generation tests pass;
- raw tokens/payloads never appear in logs or queue payloads.

### IP2 — Code-host and ecosystem extensibility

Deliverables:

- code-host capability ports extracted from GitHub behavior;
- GitLab adapter implementation against those ports;
- external developer app manifest/portal workflow;
- outbound webhook event catalog and SDK verification helpers.

Acceptance:

- GitLab addition does not change stories business rules;
- unsupported capability behavior is explicit;
- a sample external integration uses only documented API/OAuth/webhooks.

## 20. Typed data/security definition of done

A persistence or integration operation is complete only when:

- its caller, actor, tenant, scope, and policy are explicit;
- SQL is named, static, explicit-column, and sqlc-generated, unless the exact operation is governed by the near-zero exceptional-query ADR policy;
- domain and database types are mapped at the repository boundary;
- null/omitted/empty semantics are tested;
- cross-tenant and insufficient-scope tests fail closed;
- transaction and idempotency behavior are documented;
- query plans meet the agreed representative-data budget;
- credentials and provider payloads follow encryption/retention rules;
- duplicate, replay, retry, revoke, and disconnect paths are covered;
- public behavior matches OpenAPI and examples;
- old SQLx, plaintext, fallback, or synchronous webhook paths are deleted;
- operator and incident runbooks are updated.

The objective is not simply “typed SQL.” It is a typed, tenant-safe, auditable system whose integration surface can evolve without touching unrelated business logic.
