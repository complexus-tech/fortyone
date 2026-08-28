# Stories module

Stories owns FortyOne work items, their lifecycle, relationships, scheduling
state, activity records, and integration outbox intents. It is the reference
module for actor-scoped reads and transactional mutations.

## Where to look

| Path                  | Responsibility                                                                              |
| --------------------- | ------------------------------------------------------------------------------------------- |
| `domain/`             | Persistence-independent read scopes, mutation commands, events, activities, and projections |
| `http/`               | Request limits, parameter decoding, wire models, response mapping, and status mapping       |
| `service/`            | Story use cases, authorization inputs, mutation intent, and post-commit orchestration       |
| `repository/queries/` | Named PostgreSQL statements reviewed with tenant and actor predicates visible               |
| `repository/sqlc/`    | Generated SQLC code; never edit by hand                                                     |
| `repository/`         | SQLC-to-domain mapping and native pgx transactions                                          |

HTTP files are grouped by use case: primary reads, grouped reads, bulk
lifecycle, comments and activities, attachments, collaboration, associations,
and mutations. Wire models are story-owned; this package must not import another
module's HTTP package.

## Security and consistency invariants

1. A request supplies an explicit workspace and actor scope. Repositories do
   not infer authority from a bare story UUID.
2. Tenant, team, membership, and resource-lifecycle predicates remain visible
   in the named query that reads or mutates the row.
3. Multi-table mutations, activity records, and integration/outbox intents
   commit in one native pgx transaction.
4. Provider or system actors retain their real identity. They are never
   silently attributed to the human who installed an integration.
5. Conditional updates use the expected story version or timestamp and return
   a typed conflict instead of overwriting concurrent work.
6. Bulk requests are bounded, normalize duplicate IDs, return deterministic
   per-item outcomes, and never authorize one item from another item's scope.
7. Repository errors are mapped at the repository boundary; HTTP adapters do
   not inspect PostgreSQL errors.
8. Queue payloads contain stable IDs and mutation identity, not mutable story
   snapshots or secrets.

## Adding or changing an operation

1. Add or reuse a domain command/projection and make actor authority explicit.
2. Add the smallest service-owned repository port required by the use case.
3. Add a named SQLC query with tenant and lifecycle constraints in the SQL.
4. Keep transaction ownership in the repository operation that owns the
   invariant; do not split atomic work across service calls.
5. Map domain results to a story-owned HTTP model and use the shared bounded
   decoder, pagination, patch, and error primitives.
6. Add pure validation tests, service authorization tests, HTTP contract tests,
   and PostgreSQL 18 transaction/concurrency tests as appropriate.

Compatibility interfaces in `service/service.go` support callers that have not
yet moved to the typed mutation/read ports. They are migration seams, not an
extension pattern: do not add another legacy map-based repository method.

## Verification

```bash
go test -race ./internal/modules/stories/...
go vet ./internal/modules/stories/...
make sqlc-check
make architecture-check

TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/stories/repository
```

The integration database must be disposable and the role must have
`CREATEDB`. The test kit creates and drops only prefixed, test-owned databases
and applies the real migration chain.
