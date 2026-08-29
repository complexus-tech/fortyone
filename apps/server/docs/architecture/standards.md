# API engineering standards

The three documents under `docs/plans/api-modernization` define the target and
delivery sequence. The ADRs in this directory define the cross-cutting choices.
This page is the short review standard applied to every API change.

| Concern               | Required implementation                                                                                                                                                                                                                                                                                                           | Automated evidence                                                                                               | Adoption path                                                                                                                                   |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Module boundaries     | Transport maps input; service owns use cases and policy; repository owns persistence; provider adapter owns SDK types                                                                                                                                                                                                             | `internal/bootstrap/architecture` import and debt gates                                                          | Migrate a complete behavior slice and then delete its legacy path                                                                               |
| SQL                   | Named SQLC queries in the owning module; explicit columns and tenant predicates; domain mapping outside generated code                                                                                                                                                                                                            | `make sqlc-check`, `make sqlc-vet`, repository integration tests                                                 | Comments and links are pilots; continue in the recorded migration waves                                                                         |
| Transactions          | Use the shared pgx transactor; one owner defines the invariant; no network I/O while a transaction is open                                                                                                                                                                                                                        | unit-of-work rollback tests and real PostgreSQL scenarios                                                        | Introduce transaction ownership while each mutation slice moves to SQLC                                                                         |
| Events                | Commit state and outbox entry together; deliver asynchronously and idempotently                                                                                                                                                                                                                                                   | duplicate, retry, crash-window, and poison-message tests                                                         | Move direct post-commit delivery when its owning use case is migrated                                                                           |
| Requests              | `web.Decode` for bounded strict JSON, `web.ParseURLForm` plus `PostForm` for bounded credential forms, `web.ParseMultipartForm` for bounded multipart, typed DTO tags for structural validation, and `Validate()` only for cross-field/domain-shaped input rules; validation responses contain stable value-free field violations | table, fuzz, body-limit, unknown-field, query-substitution, structural-tag, and sensitive-value reflection tests | The shared decoder now enforces tags for every decoded DTO; remove remaining local decoders and validators when a handler is touched            |
| Signed webhook bodies | Read exact bytes once with `web.ReadBoundedBody`; reject overflow with 413 before signature verification or decoding; never use an unbounded read or a truncating `io.LimitReader`                                                                                                                                                | exact-limit/overflow tests plus the architecture ingress gate                                                    | Provider-specific limits remain explicit; durable gateway adoption later moves verification and quick acknowledgement behind one inbox contract |
| Rate limits           | Every enforced quota emits bounded `RateLimit-Policy` and `RateLimit` metadata; a rejected request also emits authoritative `Retry-After`; quota partition identities never appear in headers or logs                                                                                                                             | allowed/rejected response contract tests, opaque-key assertions, and CORS exposure tests                         | Adopt `web.SetRateLimitHeaders` whenever an existing counter is touched; public API limits later reuse the same contract                        |
| Authorization         | Resolve a typed actor; service policy checks current tenant/resource authority; deny unknown states                                                                                                                                                                                                                               | actor/role/scope/team and two-tenant negative matrices                                                           | Privileged paths first, then each SQLC module wave                                                                                              |
| Errors                | Stable machine code, safe message, request identifier, and field details where useful; log the internal cause once                                                                                                                                                                                                                | transport contract tests                                                                                         | Preserve legacy responses only behind an explicit compatibility test                                                                            |
| Lists                 | Stable unique ordering and signed opaque cursor; bounded limit; offset only for a documented legacy route                                                                                                                                                                                                                         | tamper, expiry, mutation-stability, and fuzz tests                                                               | Use the shared cursor primitive for new endpoints, then migrate copied pagination                                                               |
| Tests                 | Pure policy tests, repository tests against real PostgreSQL, handler contract tests, and focused end-to-end journeys                                                                                                                                                                                                              | required CI jobs cannot silently skip infrastructure tests                                                       | Shared `internal/testkit` owns infrastructure, modules own scenarios                                                                            |
| Telemetry             | Correlated structured fields and bounded cardinality; no secrets or raw customer/provider payloads                                                                                                                                                                                                                                | security scans, log assertions, readiness/lifecycle tests                                                        | Add signals with the behavior they describe and assign an owner/runbook                                                                         |
| API contracts         | Only explicitly versioned integration routes are external contracts; OpenAPI is generated and diff-gated                                                                                                                                                                                                                          | OpenAPI lint/generation/breaking-change gates when introduced                                                    | Internal routes remain undocumented implementation details until deliberately promoted                                                          |

The rate-limit metadata grammar follows the current
[IETF HTTPAPI RateLimit fields draft](https://datatracker.ietf.org/doc/html/draft-ietf-httpapi-ratelimit-headers-11).
It remains a work in progress, so `Retry-After` is the authoritative rejection
signal and draft changes require an explicit contract review rather than an
unnoticed dependency upgrade.

## Architecture dependency gate

`internal/bootstrap/architecture` parses Go syntax and the production module
import graph. New code fails the gate when repositories depend on services,
authentication context, or concrete repositories; services depend on concrete
repository adapters or another module's concrete service; a module edge creates
or expands a dependency cycle; or generated SQLC/OpenAPI types escape their
owning adapters. The owning repository may import only its own generated
`repository/sqlc` package. Generated public API code belongs under
`api/openapi/<version>/generated` so the leakage rule can identify it
deterministically.

Cross-module database invariants use a bootstrap-owned or capability-owned unit
of work. Each participating repository exposes a narrow transaction-specific
port constructed with its own `Queries.WithTx(tx)`. The callback must not receive
the raw `pgx.Tx`, a pool, or an ordinary repository. Transaction scopes close
before the commit attempt and reject calls retained beyond the callback.

Production module HTTP code must read raw request bytes with
`web.ReadBoundedBody` or a rejecting `http.MaxBytesReader` flow. An
`io.LimitReader` is not an ingress limit because it silently truncates the
signed payload. Existing findings are recorded by file and exact count in
canonical `testdata/debt-baseline.json`; the comparison permits reductions but
fails every new file, occurrence, or larger handwritten file. Test-only
concrete collaborators are excluded from production dependency edges, while
the permanent direct-SQL and zero-SQLx guards continue to inspect test code for
prohibited persistence dependencies.

After an intentional cleanup, replace stale allowances with the exact reviewed
snapshot:

```bash
make architecture-baseline-generate
```

The command is deliberately separate from normal generation and requires an
explicit environment guard inside the test harness. Review the JSON diff before
accepting it. The generator itself refuses any new occurrence or larger file,
so it can only ratchet the snapshot downward. A separately approved rule/ADR
change requires a direct, visible baseline edit; the command is never a way to
make a growth failure green. Normal verification remains read-only through
`make architecture-check` and the required Go test suite.

## Review rule

An exception must name the exact operation, owner, evidence, expiry or revisit
condition, and ADR that permits it. File size, query complexity, or migration
convenience are not exceptions by themselves.
