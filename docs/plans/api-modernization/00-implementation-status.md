# FortyOne API Modernization: Implementation Status

**Status:** Local implementation and hermetic acceptance complete; environment-backed production acceptance remains

**Snapshot:** 2026-08-28 19:31 CAT (`codex/api-sqlc-migration`, base `58e9f4220`)

**Scope:** the Go API, worker, database boundary, integration platform, public API, security controls, tests, documentation, and internal deployment posture described by the three governing plans

**Governing plans:** [target Go architecture](./01-target-go-architecture.md), [typed data, security, and integrations](./02-typed-data-security-and-integration-platform.md), and [delivery, testing, and documentation](./03-delivery-testing-and-documentation-roadmap.md)

This ledger is the current implementation source of truth. The governing plans
remain the source of truth for intent and acceptance criteria. A local file or
passing hermetic test does not prove migrated-schema, load, staging, or
production behavior.

The working tree is intentionally large and uncommitted. Nothing in this
document means that a commit, push, deployment, image publication, database
migration, secret migration, or production-state change has occurred.

## 1. Executive checkpoint

| Area                    | Local implementation                  | Current evidence                                                                                                                                                                                             | Acceptance still required                                                                 |
| ----------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------- |
| SQLC migration          | Complete                              | 41 generation units, 1,218 named SQLC operations, zero production SQLx imports, and no SQLx module dependency                                                                                                | Database-backed `sqlc vet` and migrated PostgreSQL suites                                 |
| Database runtime        | Complete                              | API, worker, seed, migrations, and tooling share `internal/platform/database`; runtime queries use native pgx and SQLC                                                                                       | Pool behavior under production-like load and dependency failure                           |
| Go structure            | Complete                              | Module-owned domain/repository/service/HTTP boundaries, dependency-cycle enforcement, and no production file above the architecture limit                                                                    | Newcomer trace exercise and continued ratchet enforcement                                 |
| Reusable primitives     | Complete                              | Shared pagination, request decoding, patching, safe conversion, actor, authorization, transaction, clock, idempotency, and integration contracts                                                             | Representative adoption and system evidence for every externally reachable flow           |
| Security implementation | Complete locally                      | Tenant/resource fences, scope/role checks, credential vaulting, token digests, browser-session epochs, login-reactivation policy, replay protection, and fail-closed security gates                          | Migrated-data verification, live abuse tests, secret-rotation exercise, and formal review |
| Integrations            | Complete for the planned platform     | Provider-neutral registries and ports, GitHub/Slack/Figma/calendar adapters, GitLab proof adapter, durable webhook inbox/outbox, OAuth applications, service accounts, PATs, and outbound developer webhooks | Live provider/database recovery and end-to-end external application exercise              |
| Scalability/reliability | Complete locally                      | Bounded keyset jobs, leases, locks, retry state, transactional outboxes, idempotent delivery, and planner indexes through migration 000175                                                                   | `EXPLAIN (ANALYZE, BUFFERS)`, representative data, load, queue soak, and alert thresholds |
| Public API/docs         | Complete locally                      | OpenAPI v1, generated Go/TypeScript SDKs, resource docs, executable external-integration example, module/database docs, generated inventory, and a successful 45-page docs build                             | Remote compatibility check and a live consumer flow                                       |
| Self-hosting posture    | Complete for current product decision | Public Compose/setup/licensing/support artifacts removed; private managed API/worker images and release workflow retained                                                                                    | Revisit only when self-hosting becomes an explicitly supported product                    |

The implementation percentage is therefore split deliberately:

- **Local code migration:** 100%.
- **Local hermetic acceptance:** 100%; the full Staticcheck, workflow, security,
  secret, fuzz, race, generated-contract, architecture, and documentation gates
  pass on this snapshot.
- **Production acceptance:** incomplete until approved PostgreSQL, Redis,
  Docker, staging, provider, load, and deployment environments are available.

## 2. What changed

### 2.1 Typed persistence

- Every production application query is a named SQLC query owned by its module
  or platform persistence package.
- SQLx imports, compatibility pools, named-query helpers, and the
  `github.com/jmoiron/sqlx` module dependency are removed.
- Native pgx is the single runtime database client. The standard
  `database/sql` bridge exists only for the `golang-migrate` driver; other
  standard-library uses are error/null/transaction types, not handwritten query
  execution.
- SQLC configuration, generation, drift checks, compile checks, and the
  database-backed vet entrypoint are pinned and documented.
- The generated API inventory currently records 426 routes across 46 modules;
  every module reports zero production SQLx files.

### 2.2 Simplicity and discoverability

- Each business capability follows the same navigable shape:
  `domain` for business vocabulary, `repository/queries` plus generated
  `repository/sqlc` for persistence, repository adapters for mapping and
  transactions, `service` for use cases and policies, and `http` for
  transport.
- Jobs depend on narrow capability interfaces. Worker bootstrap performs
  composition; job packages do not reach into global SQL or configuration.
- Large legacy files were split by behavior. The generated inventory points to
  route ownership, middleware, query ownership, tests, and the largest
  handwritten production file for every module.
- The architecture baseline contains no cross-module HTTP/service leaks, direct
  SQL outside persistence, SQLx imports, generated-type leaks, dependency
  cycles, repository-layer inversions, raw database route configuration, or
  unsafe raw request-body reads. The only recorded size exception is a
  710-line Maya test file; no production exception remains.
- Shared request, cursor, pagination, validation, conversion, authorization,
  idempotency, clock, transaction, and safe HTTP packages replace copied local
  helpers.

### 2.3 Security

- Browser sessions store and enforce a server-side user session epoch. Admin,
  scheduled inactivity, and relevant account-state transitions advance the
  epoch so existing sessions fail closed.
- Migration 000174 introduces explicit `verified_sign_in`,
  `admin_only`, and fail-closed `legacy_admin_review` reactivation policies.
  Verified email/OAuth sign-in can reactivate only eligible accounts; an
  administrator-disabled or ambiguous account receives a generic credential
  failure without a cookie or account-state leak.
- Personal tokens, service-account keys, verification codes, invitation tokens,
  OAuth codes, and refresh tokens use one-time display plus digest/HMAC storage
  where applicable. Provider credentials use the versioned, context-bound
  credential vault.
- Invitation-token parsing requires canonical Base64URL segments. Equivalent
  alternate encodings are rejected before digest derivation, preventing
  inconsistent lookup identities for the same signed bytes.
- Actor identity, workspace membership, role, scope, team restriction, and
  resource ownership are explicit policies rather than repository reads from
  HTTP context.
- Inbound webhooks use provider authentication, bounded bodies, replay and
  generation fences, durable envelopes, safe audit metadata, and retention.
  Outbound webhooks use endpoint validation, signed deliveries, leases,
  attempts, retry policy, and secret rotation.
- The G104/G115 gate parses and classifies the complete gosec report. It allows
  only AST-verified standalone response writes and fails closed on every other
  security finding or unanalyzable package.

### 2.4 Integrations and external API

- Provider-neutral integration, code-host, webhook, credential, actor, and
  event contracts separate business behavior from GitHub, GitLab, Slack,
  Figma, Google, and Microsoft SDKs.
- Adding another provider should implement a registry capability/port and its
  adapter; it must not duplicate story or workspace business rules.
- External consumers can use personal access tokens, service accounts, OAuth
  authorization-code/PKCE, refresh tokens, or confidential application actors
  with explicit scopes and exact resource binding.
- The public `/api/v1` surface has OpenAPI as its contract, generated server
  bindings, generated Go and TypeScript clients, opaque cursor pagination,
  bounded requests, structured errors, rate limits, idempotency support, and
  outbound webhooks.
- The external-integration example exercises authentication, typed reads,
  pagination, safe retry behavior, idempotency helpers, and webhook signature
  verification without coupling to web-application internals.

### 2.5 Background work and scalability

- Story and sprint automation use bounded keyset pages, stable UTC snapshots,
  advisory/row locks, transactional state transitions, audit/event writes, and
  explicit cancellation/backlog outcomes.
- Story retention and interactive hard deletion atomically retire only
  unreferenced attachment metadata and enqueue credential-free object deletion.
  A separate leased worker performs idempotent provider deletion with bounded
  retry state.
- Workspace/user inactivity, purges, digests, strategy communications, calendar
  reconciliation, feedback, messaging, invitation, webhook, attachment, and
  credential jobs use typed narrow stores instead of SQLx or ad hoc SQL.
- Migration shutdown owns one reserved connection, stops scheduling, sends the
  PostgreSQL protocol cancellation request to that connection's exact live
  route, joins the cancellation forwarder before closing resources, and exposes
  only sanitized operation-level failures.
- Migrations 000173 and 000175 add the maintenance and automation keyset indexes.
  Their shape is locally verified, but planner choice and latency are not
  claimed without a representative PostgreSQL dataset.

## 3. Migration state

- No tracked migration at or below `000151` was edited.
- Twenty-four post-baseline migration pairs, `000152` through `000175`, are
  present in the manifest and pass the migration contract checker.
- Every migration already introduced by this branch is now immutable. Any
  correction must be a new migration numbered `000176` or later.
- None of these migrations has been applied by this work.
- The manifest documents schema compatibility, mixed-version rules, API/worker
  ordering, rollback or forward-fix behavior, evidence, and operational
  prerequisites for every migration.

## 4. Verification ledger

| Gate                      | Snapshot result        | Notes                                                                                                                                                                             |
| ------------------------- | ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `make check-fast`         | Pass                   | Formatting, module verification, vet, all Go tests, generated SQLC/OpenAPI/SDK/config/inventory/migration drift, SDK checks, and architecture checks passed on the final snapshot |
| `make sqlc-check`         | Pass                   | Generated SQLC is current and every package compiles                                                                                                                              |
| `make migration-check`    | Pass                   | 24 post-baseline migrations through 000175                                                                                                                                        |
| `make architecture-check` | Pass                   | Zero prohibited dependency/SQLx/direct-SQL debt and no production size exception                                                                                                  |
| `make staticcheck`        | Pass                   | Staticcheck 0.7.0 reports zero diagnostics; no check was excluded or weakened                                                                                                     |
| `make workflow-check`     | Pass                   | Local workflow policy and pinning checks pass; no remote CI run is claimed                                                                                                        |
| `make security-check`     | Pass                   | `govulncheck` reports zero reachable vulnerabilities; the gosec guard permits 298 AST-verified response writes and rejects zero findings                                          |
| `make gitleaks-check`     | Pass                   | 8.10 MB scanned with no leaks found                                                                                                                                               |
| `make test-fuzz`          | Pass                   | Every governed bounded fuzz target passes, including invitation-token canonicalization                                                                                            |
| `make test-race`          | Pass                   | The complete Go package tree passes with the race detector; the invitation-token regression also passes repeated focused race runs                                                |
| Docs-site build           | Pass                   | 45 static pages built; 919 local paths, 4 anchors, and 392 source references were checked with zero failures                                                                      |
| `git diff --check`        | Pass                   | No whitespace errors at the recorded checkpoint                                                                                                                                   |
| `make sqlc-vet`           | Blocked by environment | `SQLC_DATABASE_URL` is unset                                                                                                                                                      |
| `make test-integration`   | Blocked by environment | `TEST_DATABASE_URL` and `TEST_REDIS_URL` are unset                                                                                                                                |
| Managed image build/scan  | Blocked by environment | The Docker CLI is present, but its daemon is unavailable                                                                                                                          |

The hermetic suite is complete. It is deliberately not being used as a proxy
for PostgreSQL planner behavior, Redis recovery, live provider behavior,
container construction, staging rollout, or production operation.

## 5. External acceptance backlog

The remaining work requires environments or authority not present in this
session:

1. Apply migrations 000152-000175, in documented order, to a disposable
   PostgreSQL database and run SQLC vet plus the complete tagged integration
   suite without skips.
2. Run queue, lease-expiry, duplicate-delivery, cancellation, and recovery tests
   against isolated Redis and PostgreSQL.
3. Capture `EXPLAIN (ANALYZE, BUFFERS)` for representative reads and job pages,
   then run endpoint, worker, queue, and soak workloads against realistic data.
4. Exercise provider install, revoke, rotate, replay, retry, and recovery flows
   with approved GitHub, GitLab, Slack, Figma, Google, and Microsoft test
   integrations.
5. Build and scan the managed API and worker images with a healthy Docker
   daemon.
6. Run a staging cutover, readiness/drain test, dependency-failure exercise,
   alert/runbook exercise, rollback or forward-fix drill, and public API
   consumer flow.
7. Obtain the formal security and operational review, then make the separately
   authorized commit, pull request, migration application, and deployment.

## 6. Resume order

The local implementation task is complete. Resume only when the required
environment or authority exists, in this order:

1. Provision disposable PostgreSQL and Redis instances with approved test-only
   credentials.
2. Apply migrations 000152-000175 and run database-backed SQLC vet plus the
   tagged integration suite.
3. Capture query plans and run representative endpoint, worker, queue, and soak
   workloads.
4. Exercise live provider recovery, build and scan managed images, and complete
   the staging cutover and operational drills.
5. Commit, push, open a pull request, apply production migrations, or deploy
   only under separate explicit authorization.

## 7. Evidence anchors

- Architecture and navigation: `apps/server/docs/architecture`,
  `apps/server/docs/onboarding`, and `apps/server/docs/inventory/api.md`.
- SQLC and database: `apps/server/sqlc.yaml`,
  `apps/server/docs/database`, and module-local `repository/queries` plus
  generated `repository/sqlc`.
- Security: `apps/server/docs/security`,
  `internal/platform/auth`, `internal/platform/authorization`,
  `internal/platform/credentialvault`, and the user/admin modules.
- Integrations: `internal/platform/integrations`,
  `internal/platform/codehost`, `internal/platform/webhooks`, provider
  modules, and `docs/integration-runtime-contract.md`.
- Public API: `apps/server/api/openapi/v1`,
  `internal/modules/apiv1`, developer credential/OAuth/webhook modules,
  `apps/server/sdk/go`, `packages/sdk-typescript`, and
  `apps/server/examples/external-integration`.
- Migrations and rollout: `apps/server/internal/migrations/manifest.json` and
  `apps/server/docs/database/migration-operations.md`.
- Gates: `apps/server/Makefile`, `.github/workflows/server-quality.yml`, and
  `.github/workflows/weekly-assurance.yml`.

## 8. Status update rule

Record a gate as complete only with reproducible evidence from the current
snapshot. Record unavailable dependencies as blocked verification, not as a
pass. Preserve the distinction between local implementation, local acceptance,
staging acceptance, and production acceptance.
