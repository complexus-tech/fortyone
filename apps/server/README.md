# Projects API

The backend API for FortyOne — a Go service that powers the web and mobile apps in this monorepo.

Use the [`API engineering guide`](docs/README.md) as the documentation index.

## Running the project locally

New to the API? Start with the internal
[`API first-hour guide`](docs/onboarding/first-hour.md). The companion
[`module change guide`](docs/onboarding/change-a-module.md) traces a safe change
from route and actor policy through SQLC, tests, and generated documentation.

### Prerequisites

- Go 1.25.0 (the version declared by `go.mod`)
- Approved PostgreSQL and Redis development endpoints
- No globally installed Go tools. `make dev` and the migration targets install
  repository-pinned binaries under the ignored `apps/server/.tools` directory.

### Setup

1. **Set up environment variables**

   ```bash
   cp .env.example .env
   ```

   Update `.env` with your local configuration.

   `APP_ENVIRONMENT` is the authoritative deployment mode for both the API and
   worker. Supported values are `development`, `test`, `staging`, and
   `production`. Production startup fails unless the application secrets are
   strong, PostgreSQL uses `verify-full`, Redis uses TLS, and certificate
   verification remains enabled.

   Email verification uses a dedicated versioned HMAC key. Set
   `APP_VERIFICATION_TOKEN_HMAC_KEY` to at least 32 random bytes and identify
   that generation with `APP_VERIFICATION_TOKEN_HMAC_KEY_ID`. Never reuse
   `APP_AUTH_SECRET_KEY`. The development default is rejected in production.
   The threat model, rotation guidance, and migration rollout are documented in
   [`docs/security/verification-tokens.md`](docs/security/verification-tokens.md).

   Workspace invitation links use a separate, shared API/worker keyring. Set
   `APP_INVITATION_TOKEN_HMAC_KEY` and its generation ID, and keep any bounded
   previous generations in `APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS`. The
   digest-only storage model, controlled migration `000155` cutover, rotation,
   contract cleanup, and forward-only recovery procedure are documented in
   [`docs/security/invitation-tokens.md`](docs/security/invitation-tokens.md).

   Feedback verification, unsubscribe links, and encrypted widget secrets use
   the shared API/worker `APP_FEEDBACK_SECURITY_KEY`. It must be independent of
   browser auth, ingress proof, token, OAuth, and vault keys. The key boundary,
   identity-only queue contract, recovery behavior, and rotation limitations
   are documented in
   [`docs/security/feedback-deliveries.md`](docs/security/feedback-deliveries.md).

   Brevo email-reply ingress and assistant mutation confirmations use the
   separate, shared API/worker `APP_EMAIL_REPLY_SECURITY_KEY` and
   `APP_MESSAGING_MUTATION_HMAC_KEY`. Production rejects missing, weak,
   development, or reused values. Email-reply rotation must drain pending
   encrypted work and update the provider header as documented in
   [`docs/maya-email-replies.md`](docs/maya-email-replies.md).

   The complete generated API/worker environment matrix is
   [`docs/configuration.md`](docs/configuration.md). `make config-check` fails
   when either Go config schema or `.env.example` drifts from that contract.

   Retained GitHub and Slack OAuth credentials use the shared envelope vault.
   Configure the same `APP_CREDENTIAL_VAULT_*` keyring for the API and worker;
   production rejects the public development key. The exact migration order,
   AAD bindings, rotation limits, verification queries, and SQLC compatibility
   debt are documented in
   [`docs/security/provider-credential-vault.md`](docs/security/provider-credential-vault.md).
   Operators must use the separate
   [`provider credential KEK rotation runbook`](docs/operations/provider-credential-rotation.md)
   before changing or retiring a vault key generation.

   PATs and service-account keys use a separate versioned HMAC keyring. Set the
   three `APP_API_CREDENTIAL_HMAC_*` values, replace the public development key
   in production, and never reuse another application or vault key. Read the
   [`developer credential security contract`](docs/security/developer-credentials.md),
   [`typed database model`](docs/database/developer-credentials.md), and
   [`rotation runbook`](docs/operations/developer-credential-rotation.md) before
   enabling a versioned machine-authenticated route.

   Provider webhook adapters share one typed durable inbox while retaining
   their own signature and payload semantics. The receive state machine,
   generation fencing, queue contract, recovery policy, payload retention, and
   provider adoption checklist are documented in the
   [`inbound webhook gateway`](docs/integrations/webhook-gateway.md) runbook.

2. **Start dependencies**
   Ensure the PostgreSQL, Redis, and optional development services configured in
   `.env` are reachable. Obtain credentials through the internal team process.

3. **Run the server**

   ```bash
   make dev
   ```

   This command starts the API using `air` for hot-reloading.

4. **Run the worker**

   ```bash
   make worker
   ```

## API health and lifecycle

The API exposes load-balancer probes. Liveness preserves its existing runtime
diagnostic fields; readiness exposes only stable check names and safe states,
never dependency error strings or secrets:

| Route        | Purpose                                                                                    |
| ------------ | ------------------------------------------------------------------------------------------ |
| `/liveness`  | Confirms that the process is alive. It remains available while the process drains.         |
| `/readiness` | Requires the supervisor to be accepting traffic and PostgreSQL plus Redis to be reachable. |

`SIGTERM` and `SIGINT` immediately change readiness from `ready` to `draining`,
cancel HTTP/SSE/consumer work from one root context, drain the HTTP server, flush
tracing, and close the tasks client, Redis client, and shared PostgreSQL pool
once. A stream consumer-group startup error, listener bind failure, or terminal
supervised component error is fatal rather than leaving a partially working API.

See [`docs/architecture/api-lifecycle.md`](docs/architecture/api-lifecycle.md)
for the startup order, failure policy, timeout configuration, and testing seams.
Request IDs, trace/log correlation, safe field rules, and incident lookup are
documented in
[`docs/operations/observability.md`](docs/operations/observability.md).
Release OIDC, ECS task-role storage access, least-privilege boundaries, and
recovery checks are documented in
[`docs/operations/aws-identity.md`](docs/operations/aws-identity.md).

## Worker health and queue monitoring

The worker serves a small operational endpoint on
`APP_WORKER_HTTP_HOST` (default `0.0.0.0:8080`):

| Route           | Purpose                                                                |
| --------------- | ---------------------------------------------------------------------- |
| `/health/live`  | Process liveness. It remains available while the worker is draining.   |
| `/health/ready` | Readiness. It requires an active worker/scheduler and a healthy Redis. |
| `/admin/queues` | Read-only Asynq queue console. Disabled unless explicitly configured.  |

Queue monitoring is an internal operator surface, not a public application
route. To enable it, set `APP_WORKER_MONITOR_ENABLED=true` and supply
`APP_WORKER_MONITOR_USERNAME` plus `APP_WORKER_MONITOR_PASSWORD`. Production
requires a password of at least 32 bytes. The endpoint always requires HTTP
Basic authentication, but credentials are only safe behind HTTPS; keep the
entire worker port restricted by the deployment network policy or an internal
authenticated proxy. Never expose it through the public API load balancer.

`SIGTERM` and `SIGINT` mark readiness false, stop scheduling and reserving new
tasks, let active tasks finish within `APP_WORKER_HTTP_SHUTDOWN_TIMEOUT`, shut
down HTTP, and close Redis/PostgreSQL once. ECS/container stop timeouts must be
longer than that configured application timeout.

## Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

### Usage

| Command                                    | Description                          |
| ------------------------------------------ | ------------------------------------ |
| `make migrate-create name=add_users_table` | Create a new migration               |
| `make migrate-up`                          | Apply all pending migrations         |
| `make migrate-down n=1`                    | Rollback last N migrations           |
| `make migrate-version`                     | Show current version                 |
| `make migrate-force v=2`                   | Force set version (use with caution) |
| `make migration-docs`                      | Regenerate the migration runbook     |
| `make migration-check`                     | Validate migration files and policy  |

The first migration command builds the repository-pinned CLI into
`apps/server/.tools`; it never assumes a developer's global Go binary path.

**Note:** Set `DB_URL` or the individual env vars (`APP_DB_USER`, `APP_DB_PASSWORD`, `APP_DB_HOST`, `APP_DB_PORT`, `APP_DB_NAME`) before running.

For an external database, set `APP_DB_SSL_MODE=verify-full`. Optionally set
`APP_DB_SSL_ROOT_CERT` to `system` or to a CA path available inside the running
API or worker runtime. Production CA files must be supplied through the managed
deployment secret mount. `disable` is for trusted local PostgreSQL only; see
[`docs/database/sqlc.md`](docs/database/sqlc.md#database-tls-contract).

### Best Practices

- Keep migrations small and focused (one logical change per migration)
- Always write both `.up.sql` and `.down.sql` files
- Use conditional DDL only when its behavior is intentional and tested; do not hide schema drift
- Test migrations locally before applying to production
- Never edit migrations that have been applied to production

Migration `000151` is the established baseline. Every later migration must be
represented in the machine-readable manifest and generated operator contract.
Read [`docs/database/migration-operations.md`](docs/database/migration-operations.md)
before rollout or recovery; `make migration-check` rejects missing pairs,
unlisted migrations, invalid classifications, and stale documentation.

Migration `000156` adds the shared API idempotency receipt primitive. It ships
unwired: each route must first prove how its domain transaction or outbox closes
the crash window between a committed mutation and receipt completion. Limits,
safe replay fields, lease fencing, expiry, rolling-deploy compatibility, and
the guarded rollback procedure are documented in
[`docs/security/idempotency-receipts.md`](docs/security/idempotency-receipts.md).

## Typed SQL with sqlc

Production repository queries use sqlc with the native pgx/v5 runtime. Each API,
worker, or seed process owns one `pgxpool.Pool` and shares it across its
repositories. SQLx is a prohibited production dependency. The only runtime
`*sql.DB` connection is the short-lived golang-migrate driver connection; it is
never injected into an application repository.

Start with [`docs/database/sqlc.md`](docs/database/sqlc.md) before adding or
changing a generated query. Module examples include the teams reference slice
in [`docs/database/teams.md`](docs/database/teams.md) and the typed,
tenant-scoped search contract in
[`docs/database/search.md`](docs/database/search.md). Maya's atomic realtime
quota and tool-call replay boundary is documented in
[`docs/database/maya-realtime.md`](docs/database/maya-realtime.md).
Sprint reads, mutations, audit transactions, and analytics are documented in
[`docs/database/sprints.md`](docs/database/sprints.md), with their live
membership rules in
[`docs/security/sprint-authorization.md`](docs/security/sprint-authorization.md).
Notification inbox, preferences, portal feedback, and email-delivery
persistence are documented in
[`docs/database/notifications.md`](docs/database/notifications.md), with their
live actor, tenant, resource, and contributor rules in
[`docs/security/notification-authorization.md`](docs/security/notification-authorization.md).
Story lifecycle mutations, bounded story/sprint automation, and durable
attachment-object deletion are documented in
[`docs/database/stories-mutations.md`](docs/database/stories-mutations.md).
The common commands are:

| Command               | Description                                                        |
| --------------------- | ------------------------------------------------------------------ |
| `make sqlc-bootstrap` | Download the checksum-verified, repository-pinned sqlc binary      |
| `make sqlc-generate`  | Safely replace every config-declared generated directory           |
| `make sqlc-check`     | Compile SQL and verify generated output without changing the tree  |
| `make sqlc-vet`       | Validate queries against an explicitly supplied, fully migrated DB |

`make test` runs the complete hermetic server test suite without external
infrastructure. PostgreSQL and Redis integration tests run separately in CI
with the `integration` build tag:

```bash
TEST_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/test_control?sslmode=disable' \
TEST_REDIS_URL='redis://127.0.0.1:6379/0' \
  make test-integration
```

`TEST_DATABASE_URL` is a control connection, not a schema that tests mutate.
The configured role must have `CREATEDB` on a disposable, non-production
PostgreSQL server. The shared testkit creates a randomly named database per
test, applies every embedded migration, and drops that database with
`t.Cleanup`. `TEST_REDIS_URL` identifies a disposable test database; every test
receives a cryptographically random key namespace and cleanup deletes only that
owned prefix. Tagged tests fail with a clear, credential-safe error when a
required variable or service is unavailable; they never silently skip coverage.
The default `make test` command does not read either variable or require
PostgreSQL or Redis. See the complete
[`integration infrastructure contract`](docs/testing/integration-infrastructure.md)
for isolation, cleanup, CI image, upgrade rules, and the shared hermetic clock,
ID, polling, signed-provider, and handler-recording primitives.

## Quality and security checks

Before handing off an API change, run:

```bash
make bootstrap-tools # first run only
make tool-versions
make check-fast
make test-race
make test-fuzz
make ci
```

`make check-fast` is the read-only local baseline: formatting, module drift,
vet, hermetic tests, generated artifacts, and architecture debt. `make ci` adds
the slower static, workflow, vulnerability, source-security, and secret scans.
Use `make generate` only when intentionally updating generated artifacts; the
architecture debt baseline always has its own review-only command.

Tool versions, CI behavior, the narrowly governed G104 transport exception, and safe secret-scan
scope are documented in
[`docs/security/quality-gates.md`](docs/security/quality-gates.md).

For code navigation and migration progress, use the generated
[`docs/inventory/api.md`](docs/inventory/api.md). It maps every registered route
to its module, handler, middleware, tests, SQLC queries, and largest handwritten
file. `make inventory-check` prevents the map from drifting.

## Developer API and SDK previews

Only operations committed under [`api/openapi/v1`](api/openapi/v1/README.md)
are public integration contracts. The repository contains private preview Go
and TypeScript clients generated from that contract plus a
[`runnable external integration`](examples/external-integration/README.md) that
uses PAT authentication, paginated story reads, and verified outbound
webhooks without importing server internals.

```bash
make sdk-generate # intentionally replace generated preview artifacts
make sdk-check    # read-only drift, type, tidy, test, and race checks
```

The packages are not published to npm or the Go module proxy. Generator pins,
allowed handwritten behavior, upgrade steps, and honest preview limitations
are documented in
[`docs/integrations/public-sdks.md`](docs/integrations/public-sdks.md) and
[ADR 0011](docs/architecture/decisions/0011-public-sdk-generation.md).

## Database Seeding

The project includes a Go-based seeder to quickly set up a development environment with a user, workspace, and default data (teams, statuses, stories).

### Usage

```bash
# Seed with default values (admin@example.com, "Development" workspace)
make seed

# Seed with custom values
make seed name="My Project" slug="my-project" email="joseph@example.com" fullname="Joseph Mukorivo"
```

The seeder leverages application side-effects, so creating a workspace will automatically:

- Create a default team ("Team 1")
- Create default story statuses
- Create initial "Welcome" stories
- Initialize workspace settings

## Brevo Integration

This service integrates with Brevo (formerly Sendinblue) to manage subscriber onboarding and emails.

### Required Environment Variables

Add this environment variable to your `.env` file:

```bash
# Brevo Configuration
APP_BREVO_API_KEY=your_brevo_api_key_here
```

### How It Works

1. **Onboarding**: When a user registers, they are added to the Brevo contact list (Default List ID: 6).
2. **Trials**: Workspace trial starts add users to the Trial list (Default List ID: 12).
3. **Emails**: System notifications are sent via Brevo's transactional email service.

## Architecture

The Projects API follows a **domain-first architecture** with clear boundaries between HTTP, domain services, and data access.

Accepted cross-cutting decisions and the enforceable review standard live in
[`docs/architecture/decisions`](docs/architecture/decisions/README.md) and
[`docs/architecture/standards.md`](docs/architecture/standards.md). New module
work follows those records; existing departures are migration debt, not a second
supported architecture.

Authentication, actor attribution, scopes, service policy, and immediate
workspace-role revocation are documented in
[`docs/security/authorization.md`](docs/security/authorization.md).

### System Overview

The system is organized around domains:

1.  **Entrypoints (`cmd`)**: Starts API, worker, seed, and metrics processes.
2.  **App Wiring (`internal/bootstrap`, `internal/platform/http`)**: Router setup and route registration.
3.  **Domain Modules (`internal/modules/<domain>`)**: Each domain owns `service`, `repository`, and `http` packages.
4.  **Shared Libraries (`pkg`)**: Cross-cutting utilities and third-party integrations.

### Directory Structure

```mermaid
graph TD
    cmd[cmd/] --> app[internal/bootstrap/api]
    app --> mux[internal/platform/http/mux]
    mux --> http[internal/modules/<domain>/http]
    http --> service[internal/modules/<domain>/service]
    service --> repo[internal/modules/<domain>/repository]
    service --> pkg[pkg/]
    repo --> pkg
```

#### Key Directories

- **`cmd/`**: Application entry points.
  - `api/`: The main REST API server.
  - `worker/`: Background job processor (using Asynq).
- **`internal/bootstrap/api`**: API route composition and top-level route registration.
- **`internal/platform/http/mux`**: Router setup and middleware registration.
- **`internal/modules/<domain>/http`**: Domain HTTP handlers (request parsing, validation, response writing).
- **`internal/modules/<domain>/service`**: Domain business logic.
- **`internal/modules/<domain>/repository`**: Domain-owned persistence adapters.
  - Static application SQL lives in `queries/*.sql`; generated code lives only in `sqlc/`.
  - Handwritten adapters map generated rows and errors into service-owned types.
- **`pkg/`**: Library code that is not specific to the application's business domain (e.g., `logger`, `database`, `brevo`, `azure`).

### Data Flow

1.  **Request**: An HTTP request hits `cmd/api`, then passes through `internal/platform/http/mux`.
2.  **Handler**: The domain handler in `internal/modules/<domain>/http` validates input and calls the domain service.
3.  **Service**: The service in `internal/modules/<domain>/service` executes business rules.
4.  **Repository**: The repository adapter executes typed sqlc queries through the process-owned pgx pool or a transaction derived from it.
5.  **Response**: Data flows back up: Repository → Service → Handler → HTTP response.

### Technology Stack

- **Language**: Go 1.25.0
- **Database**: PostgreSQL via pgx/v5 and sqlc
- **Caching**: Redis (via `go-redis`)
- **Background Jobs**: Asynq (Redis-backed queue)
- **Routing**: Standard `net/http` with custom Mux
- **Observability**: OpenTelemetry & Slog
- **Integrations**:
  - **Brevo**: Email & Transactional messaging
  - **Google**: OAuth2 Authentication
  - **Azure**: Blob Storage
  - **Stripe**: Billing

### Development Patterns

- **Dependency Injection**: Dependencies are created in `main.go` and passed down explicitly (no global state).
- **Tracing**: Key operations are wrapped in `web.AddSpan` for observability.
- **Transactions**: A business invariant uses one pgx transaction end to end; generated query sets bind to that transaction.
- **Boundaries**: `internal/modules/<domain>/service` and `internal/modules/<domain>/repository` must not import HTTP-layer packages.
- **Package Naming**: Domain adapters use explicit package names like `<domain>http` and `<domain>repository`.
