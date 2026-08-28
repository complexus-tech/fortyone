# Comment mentions data access

The mentions repository is a small SQLC/native-pgx adapter used through the
story service's consumer-owned `MentionsRepository` port. It owns no HTTP or
business policy. The story use case supplies the authoritative workspace and
comment identity after resolving the parent story.

## Invariants

- Every read, delete, and replacement resolves the comment through its story's
  `workspace_id`; an ID from another tenant is reported as not found.
- A replacement locks the scoped comment and deletes/inserts mentions in one
  transaction. Invalid targets roll the deletion back, preserving the previous
  complete set.
- Mention targets must be distinct non-zero IDs, at most 100 per comment, active
  users, and current members of the same workspace.
- Target IDs are sorted before persistence and reads are ordered, keeping tests,
  events, and logs deterministic.
- Empty replacement is an intentional clear operation.
- Generated SQLC types remain inside the repository.

The repository does not infer a workspace from the comment ID and does not
accept a generic parameter map. Application SQL uses `CAST(...)` for arrays and
contains no SQLx fallback.

## Concurrency

`SaveMentions` locks the comment row with `FOR UPDATE`, then performs delete and
bulk insert through the transaction-bound SQLC query set. Concurrent writers
serialize and the final state is one complete requested set, never a partial
mix. `GetMentions` uses a repeatable-read, read-only transaction so the scoped
comment existence check and mention list share one snapshot.

## Verification

```bash
go test -race ./internal/modules/mentions/...
.tools/bin/sqlc compile -f sqlc.yaml
./scripts/check-sqlc.sh
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/mentions/repository
```

The PostgreSQL suite covers cross-tenant comment IDs, inactive/non-member and
cross-workspace targets, rollback of invalid replacements, scoped deletion, and
concurrent replacement atomicity.
