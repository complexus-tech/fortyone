# API first hour

This guide gets an engineer from an unfamiliar checkout to a small, reviewable
API change. FortyOne is currently a privately operated application. The
repository does not promise a public self-hosting workflow, public container
stack, or support for arbitrary deployments.

## 1. Build the mental map

Read these in order:

1. [`README.md`](../../README.md) for commands, runtime processes, and the
   directory map.
2. [`docs/architecture/standards.md`](../architecture/standards.md) for the
   rules enforced on every change.
3. [`docs/inventory/api.md`](../inventory/api.md) to locate a route, its
   registered guards, handler, tests, SQLC queries, and largest handwritten
   files.
4. The ADR that owns the concern you are changing under
   [`docs/architecture/decisions`](../architecture/decisions/README.md).

The normal request path is:

```text
cmd/api
  -> internal/bootstrap/api
  -> internal/modules/<domain>/http
  -> internal/modules/<domain>/service
  -> internal/modules/<domain>/repository
  -> repository/queries/*.sql -> generated repository/sqlc
```

Background work follows the same ownership boundary:

```text
cmd/worker
  -> internal/bootstrap/worker
  -> internal/taskhandlers
  -> pkg/jobs
  -> narrow domain-owned store port
  -> internal/modules/<domain>/repository -> SQLC
```

HTTP translates protocol data. Services own use cases, current-state policy,
transaction intent, and narrow caller-owned ports. Stable values shared with a
persistence adapter live in the module's transport-neutral `domain` package.
Repositories own persistence and map generated rows to those domain values.
Provider adapters own external SDK types. If a change does not fit that path,
stop and identify the actual owning capability before adding a shortcut.

Each API, worker, or seed process owns one native pgx pool and shares it across
its repositories. Application transactions are pgx transactions derived from
that pool. The only production `*sql.DB` connection is the migration-driver
handle, and application repositories never receive it.

## 2. Establish the local contract

From `apps/server`:

```bash
make tool-versions
go env GOMOD GOWORK
make check-fast
```

Use the Go version declared by `go.mod`. Run `make bootstrap-tools` once when
the checksum-pinned local SQLC and secret-scanning binaries are absent.
`make check-fast` is hermetic and must not need PostgreSQL or Redis. Its
generation checks are read-only: they verify SQLC, configuration, inventory,
migration, and architecture artifacts without accepting drift.

Obtain development configuration through the internal team process and copy
`.env.example` to `.env` only on your own machine. Never paste `.env`, a database
URL, an authorization header, a cookie, an OAuth code, a provider payload, or a
token into a ticket, terminal transcript, log assertion, or test failure.

Run the API and worker separately when live behavior is required:

```bash
make dev
make worker
```

There is intentionally no repository-supported public Docker Compose or
one-command self-host flow. The checked-in runtime images and deployment
workflow serve the managed FortyOne environment.

## 3. Trace one behavior before editing

Search the generated inventory by route or module. Then read, in this order:

1. `http/routes.go` to see registered authentication, workspace, role/scope,
   rate-limit, and transport middleware;
2. the handler method and request/response DTO;
3. the service port and use case;
4. the repository adapter and named SQLC query;
5. the closest unit, HTTP, repository, and integration tests.

Registered middleware is not proof of authorization. Confirm where the service
checks the current actor, workspace role, team restriction, resource ownership,
and resource visibility. Confirm the SQL also scopes tenant-owned data by the
workspace and, where relevant, actor or team.

## 4. Make the smallest complete slice

A complete change usually contains:

- a typed request or command with bounded input;
- `web.Decode`, typed path/query parsing, and a safe response mapping;
- a service use case with explicit actor and policy inputs;
- named SQLC query changes and handwritten domain mapping;
- one transaction for one invariant, with its outbox record when an external
  side effect follows;
- negative tenant/role tests as well as the success path;
- updated generated inventory, configuration, migration, or API artifacts when
  the source contract changed.

Do not add a generic map update, a repository-to-repository call, provider SDK
types in a service contract, SQL in a handler, a second pagination helper, or a
handwritten Go SQL path for a static query. SQLx is a prohibited production
dependency; use the owning module's SQLC/pgx boundary.

## 5. Validate before handoff

Start narrow and finish with the required gates:

```bash
go test -race ./internal/modules/<domain>/...
go vet ./internal/modules/<domain>/...

make quality-check
make test
make security-check
make gitleaks-check
make generated-check
```

Repository, Redis, and queue integration tests use the single documented
contract in [`docs/testing/integration-infrastructure.md`](../testing/integration-infrastructure.md).
They run with the `integration` build tag against disposable infrastructure and
must fail clearly—not skip—when that infrastructure is missing.

Handoff notes state the behavior preserved or changed, tenant/actor/security
effect, migrations and rollout mode, checks actually run, and any baseline
failure that remains outside the change. Do not describe an unrun integration,
browser, migration, or deployment check as passing.
