# ADR 0002: SQLC and native pgx

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering

## Context

Runtime SQL strings and reflection-based scans defer query, parameter, and result
errors until execution. Large queries make that risk more important, not less.
FortyOne also needs explicit tenant predicates and transaction binding that can
be reviewed beside each owning module.

## Decision

SQLC is the production boundary for every ordinary application query. Each API,
worker, or seed process owns one native pgx/v5 pool and shares it across its
repositories. Each persistence-owning module keeps named queries in
`repository/queries` and generated-only output in `repository/sqlc`.

- Select and return columns explicitly; do not use `SELECT *` or `RETURNING *`.
- Name every parameter and encode tenant/resource authority in the query where applicable.
- Keep SQLC parameter and row types inside the repository adapter.
- Map database nullability and enums to deliberate domain types.
- Bind generated queries to a transaction with `WithTx`; do not expose raw transactions upward.
- Keep one root SQLC configuration, an exact tool version, clean generation, and database-backed vet.

This decision was adopted with a temporary SQLx compatibility allowance so
modules could move in complete transaction-safe waves. That cutover is now
complete: SQLx and its compatibility connection view have been removed from
production and must not be reintroduced. The only production `database/sql`
connection is the short-lived handle required by the golang-migrate driver; it
is never injected into an application repository.

A permanent non-SQLC operation still requires a separate accepted ADR naming
the exact query, why the pinned SQLC version cannot express it, security and type
controls, benchmarks where performance is claimed, tests, owner, and revisit
condition. Dynamic or long SQL is not sufficient justification.

## Enforcement and adoption

- `make sqlc-check` performs offline compile, clean generation drift, and Go compilation.
- `make sqlc-vet` refuses an empty database URL and prepares queries against the exact migrated schema.
- Architecture tests reject SQLx dependencies, new direct production SQL, and
  SQLC imports outside the owning repository.
- Repository behavior is tested against real PostgreSQL, including negative tenant cases.

The staged adoption waves are complete. A new or extended persistence slice is
compliant only when its static SQL is module-owned and generated, its domain
adapter is the sole production path, and one business invariant uses one pgx
transaction end to end. Generated code beside a handwritten fallback is not
completion.

## Consequences

Queries remain readable SQL while Go receives compile-time signatures. Generated
code is intentionally checked in. Migrations may require narrow schema
compatibility work, but previously applied migrations are never edited.

## Revisit when

An exact operation satisfies the exceptional-query evidence above or SQLC is no
longer maintained. Convenience is not a revisit condition.
