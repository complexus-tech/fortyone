# First-party system actor resolution

FortyOne has a closed catalog of first-party actors used for work that must not
be falsely attributed to a human session. The current keys are `system` (Maya)
and `github`. Application code refers to the stable key; it does not hard-code a
database UUID or accept an arbitrary email from a request.

## Dependency path

```text
API or worker bootstrap
  -> actors.Resolver port
    -> actors/repository SQLC adapter
      -> actors/repository/queries/actors.sql
        -> shared native pgx pool
```

Bootstrap resolves required actors before the process reports ready. The typed
query returns a user only when the email matches and both `is_system` and
`is_active` are true. A human account, inactive system account, missing account,
unknown actor key, or zero UUID fails startup rather than silently choosing a
different identity.

The resolver deliberately has no Redis cache. These are two startup reads, not
a request hot path, and a cache could keep an inactive identity authoritative
after database state changed. This follows the same live-state rule used for
workspace membership and credential revocation.

Generated SQLC values remain inside the repository adapter. The resolver
depends only on a one-method `Lookup` port, which makes the key policy easy to
unit test without PostgreSQL and allows bootstrap—not a handler or service—to
choose the persistence implementation.

## Adding a system actor

1. Decide why a separate persisted principal is required and document its audit
   attribution.
2. Seed or migrate the unique system user through the controlled deployment
   process; never create it lazily during a request.
3. Add a closed key-to-email entry in `internal/platform/actors`.
4. Resolve it at startup and pass only its typed UUID/capability to consumers.
5. Add unit and PostgreSQL 18 tests for active system, inactive system, human,
   missing, and zero-ID behavior.
6. Ensure logs contain only the non-secret actor key or configured system email,
   never a credential or provider token.

Run the focused checks from `apps/server`:

```bash
go test -race ./internal/platform/actors/...
TEST_DATABASE_URL='<PostgreSQL 18 control URL>' \
  go test -race -tags=integration -count=1 ./internal/platform/actors/repository
make sqlc-check
```

Related contracts: [Actor and authorization model](authorization.md),
[ADR 0004](../architecture/decisions/0004-actors-authorization-and-revocation.md),
and [Typed database access](../database/sqlc.md).
