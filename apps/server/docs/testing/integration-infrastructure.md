# Integration infrastructure contract

FortyOne's default Go suite is hermetic:

```bash
make test
```

Repository, Redis, and queue adapter tests use the single `integration` build
tag and require explicitly provisioned infrastructure:

```bash
TEST_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/test_control?sslmode=disable' \
TEST_REDIS_URL='redis://127.0.0.1:6379/0' \
  make test-integration
```

Use only internally approved, disposable, non-production endpoints. Tagged
tests fail with a safe configuration or connectivity error when either required
service is missing. They never call `t.Skip` for unavailable infrastructure.

## PostgreSQL contract

`TEST_DATABASE_URL` is a control connection. It must:

- use the `postgres` or `postgresql` URL scheme;
- include an explicit username, host, port, and control database;
- identify a role with `CREATEDB` on a disposable PostgreSQL server.

Each `internal/testkit.NewPostgres` call creates a cryptographically named
database, applies the complete embedded migration chain, opens a bounded pgx
pool, and registers cleanup with `t.Cleanup`. Parallel tests therefore do not
share schemas or tenant state. Cleanup refuses to drop a database unless its
name has the testkit ownership prefix.

## Redis contract

`TEST_REDIS_URL` must:

- use the `redis` or `rediss` URL scheme;
- include an explicit host and port;
- optionally select one non-negative numeric database in the URL path;
- identify a disposable Redis database that the test role may mutate.

Each `internal/testkit.NewRedis` call parses the URL without echoing it in
errors, applies bounded command and connection-pool settings, and verifies the
service with `PING`. It then generates a 128-bit cryptographically random key
namespace. Tests derive every mutable key with `Redis.Key`.

Cleanup scans and deletes only keys under the exact namespace owned by that
test, then closes the client. It never uses `FLUSHDB`, never deletes a caller
supplied broad prefix, and never touches another parallel test's namespace.
The integration smoke test concurrently writes identical logical keys in two
namespaces and proves that cleaning one leaves the other intact.

## CI contract

The `Server SQLC` workflow provides both dependencies and runs
`make test-integration`. PostgreSQL is pinned to the production major by its
multi-architecture digest. Redis is the official Redis 7.4.11 Alpine image,
pinned by its multi-architecture index digest. Both services have explicit
health checks, and CI supplies `TEST_DATABASE_URL` plus `TEST_REDIS_URL` before
running tagged tests.

When upgrading either service image, verify the official multi-architecture
index digest, update the version comment and digest together, and exercise the
complete tagged suite. Do not replace a digest with a floating tag.

## Hermetic test primitives

`internal/testkit` also owns small infrastructure primitives that remain usable
from the default hermetic suite:

| Primitive                                          | Contract                                                                                                                                                                                                            |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `NewFixedClock` / `NewManualClock`                 | Satisfy the application's structural `Now() time.Time` ports. The manual clock is concurrency-safe and advances decisions without sleeping.                                                                         |
| `NewUUIDSource`                                    | Emits a reproducible, concurrency-safe UUID sequence scoped by a descriptive seed. Use it only when output IDs are asserted.                                                                                        |
| `Eventually`                                       | Polls immediately and at a positive interval under a required context deadline. Deadline errors include a normalized, bounded last observation. It is not a general sleep helper.                                   |
| `NewHMACSHA256Signer` / `NewSignedProviderRequest` | Build bounded signed provider requests. Signer diagnostics redact the secret, and URL errors never echo a possibly credential-bearing target.                                                                       |
| `NewProviderServer`                                | Provides a timeout-bound client, one-MiB request/response limits, concurrency-safe capture, automatic cleanup, and defensive request snapshots whose formatted form omits headers, query values, paths, and bodies. |
| `RecordHandler`                                    | Invokes one `web.Handler` with deterministic request metadata and time while preserving caller-supplied context such as an authenticated actor.                                                                     |

Domain builders stay beside the module that owns their invariants. Testkit does
not create users, workspaces, stories, credentials, or provider-specific
payloads.

There is intentionally no full-API test builder yet. Constructing the current
root mux requires nearly every module, infrastructure adapter, and provider
credential, which would turn testkit into a second composition root and make
unrelated module changes invalidate all handler tests. HTTP modules should use
`RecordHandler` for one-handler contracts and the real bootstrap package for the
small number of lifecycle/system tests. A future API harness is appropriate only
after bootstrap exposes a narrow, stable dependency bundle and actor seam.
