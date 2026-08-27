# FortyOne API Modernization: Target Go Architecture

**Status:** Proposed implementation plan  
**Scope:** `apps/server` only  
**Audience:** API maintainers, reviewers, new Go engineers, integration authors  
**Companion plans:** [typed data, security, and integrations](./02-typed-data-security-and-integration-platform.md) and [delivery, testing, and documentation](./03-delivery-testing-and-documentation-roadmap.md)

## 1. Executive decision

Keep FortyOne as a **modular Go monolith**. The current problem is not that the API needs microservices; it is that module boundaries, authorization, persistence, and composition are inconsistently enforced. Splitting the system into independently deployed services now would move those ambiguities onto a network and make them harder to repair.

The target is a codebase where an engineer can predict where code lives:

- HTTP routes and transport models live in a module's `http` package.
- Business use cases and policies live in its `service` package or domain root.
- Handwritten SQL lives in `repository/queries/*.sql`.
- sqlc-generated database code lives in `repository/sqlc` and never leaks into HTTP or domain contracts.
- Cross-module dependencies are narrow, consumer-owned interfaces.
- Shared code is promoted into `internal/platform` only after it has stable, genuinely cross-domain semantics.
- API and worker composition are explicit, small, and split by module capability.

The current API has useful foundations, but it is not yet predictable or uniformly safe. On the reviewed snapshot, it is approximately **5.5/10 overall**. The architecture target in this plan is intentionally measurable: the API reaches the proposed 10/10 bar only when the security, sqlc, test, documentation, operability, and module-boundary gates in all three plans are satisfied.

## 2. What was reviewed

This plan is based on a repository-wide read-only review of:

- all 39 directories under `apps/server/internal/modules`;
- API and worker bootstrapping;
- shared packages under `internal/platform` and `pkg`;
- all current migrations and repository query patterns;
- route registration and middleware usage;
- test files, CI, deployment, and local developer commands;
- existing Slack, GitHub, Figma, calendar, messaging, feedback, and integration-request plans;
- the existing [messaging integration runtime contract](../../integration-runtime-contract.md).

The working tree already contained unrelated, in-progress changes in chatsessions, Maya, stories, and migrations 146–147. This planning work does not reinterpret those changes as completed architecture and does not require modifying them.

### 2.1 Current scale snapshot

The exact counts will change, but the scale explains why discoverability must become a first-class constraint:

| Measure                                            |                  Reviewed snapshot | Architectural implication                                               |
| -------------------------------------------------- | ---------------------------------: | ----------------------------------------------------------------------- |
| Domain module directories under `internal/modules` |                                 39 | A standard module shape is valuable.                                    |
| Approximate Go lines, including tests              |                           206,000+ | Navigation cannot depend on tribal knowledge.                           |
| Registered HTTP routes                             |                          About 380 | Route, authorization, and API-contract inventories must be generated.   |
| Production files at or above 800 lines             |                           About 40 | Use-case and capability splits are overdue.                             |
| Largest service file                               |   Slack service, about 4,540 lines | Provider and use-case responsibilities are conflated.                   |
| Largest repository query file                      | Stories queries, about 3,515 lines | SQL needs named, typed, use-case-oriented ownership.                    |
| Current migration pairs                            |   147 in the reviewed working tree | Schema validation must be automated.                                    |
| Transaction starts across current patterns         |                          About 108 | Transaction ownership and external side effects require explicit rules. |

The highest-volume production modules are Slack, stories, feedback, messaging, Maya, calendar, reports, GitHub, users, notifications, objectives, and workspaces. File size is a signal, not the sole reason to refactor: a large file should be split when it contains multiple reasons to change, hides security boundaries, or makes isolated testing difficult.

### 2.2 Boundary and consistency evidence

The repository-wide inventory also found:

| Evidence                                                  |                             Reviewed snapshot | Target implication                                                                    |
| --------------------------------------------------------- | --------------------------------------------: | ------------------------------------------------------------------------------------- |
| HTTP modules with a consistent `routes.go`/`Routes` entry |                                            33 | Preserve this useful convention.                                                      |
| Route config structs carrying a raw `*sqlx.DB`            |                                      30 of 32 | Construct workspace/auth middleware once; handlers should receive ports, not DB.      |
| Service/repository files importing `pkg/web`              |                                      About 70 | Move transport-neutral tracing/context out of the HTTP package and enforce direction. |
| Cross-module dependency edges                             |                                      About 83 | Replace broad concrete imports with small caller-owned capabilities.                  |
| Strongly connected module components                      |              One: subscriptions ↔ workspaces | Break the cycle through an application use case/port.                                 |
| Production `web.GetFilters` call sites                    |                                            17 | Replace reflection and DB-tag maps with typed query objects.                          |
| Module-local pagination implementations                   | Seven, including five byte-identical variants | Adopt one cursor/query parsing primitive without centralizing domain semantics.       |
| HTTP fields carrying validator tags                       |                                            75 | Select and execute one validation mechanism; inert tags must disappear.               |
| HTTP request types with explicit `Validate() error`       |                                      About 10 | Validation coverage is currently inconsistent.                                        |
| HTTP/service `db` tag leakage                             |                              Hundreds of tags | Separate transport, domain, and persistence models.                                   |
| Raw SQL in Email Reply service                            |                                 16 call sites | Give the module an owned repository even if it has no HTTP-facing CRUD.               |

Specific dependency smells include stories HTTP importing links HTTP for response conversion; stories repository importing sibling concrete repositories; repositories deriving user identity from HTTP auth context; and services leaking raw `*sqlx.Tx` across invitations, teams, users, and workspaces. The target rules below address these observed paths rather than imposing theoretical layering.

## 3. What should remain

The modernization should preserve and strengthen these existing choices:

1. **A modular monolith.** FortyOne has many cross-domain workflows. In-process calls are appropriate while ownership is being clarified.
2. **Explicit dependency injection.** Constructors and bootstrap wiring are preferable to a service locator or mutable global registry.
3. **Standard `net/http`.** The current router approach can support a high-quality API; a framework rewrite is not required.
4. **Domain-oriented modules.** `stories`, `objectives`, `calendar`, and similar package names express product concepts better than technical layer-wide packages.
5. **Durable integration patterns.** Messaging and Slack already demonstrate encrypted inboxes, deduplication, generation fencing, asynchronous processing, and outboxes.
6. **Capability-sized provider interfaces.** Calendar's Google/Microsoft abstraction is closer to the intended direction than one universal integration interface.
7. **Explicit errors and request tracing.** Existing response sanitization and OpenTelemetry hooks are foundations to standardize, not discard.

## 4. Problems the target architecture must solve

### 4.1 Discoverability

Today, an engineer may find SQL in repositories, services, middleware, job handlers, task handlers, or HTTP code. Some modules use `commands.go` and `queries.go`; others put thousands of lines in a single file. Route composition is centralized in large bootstrap files, and module constructors expose different shapes.

The target must answer these questions without a full-text search:

- Where is this route registered?
- Which use case authorizes it?
- Which SQL query persists it?
- Which transaction owns the change?
- Which event or task is emitted?
- Which test proves tenant isolation and failure behavior?
- Which public API contract documents it?

### 4.2 Boundary leakage

Several handlers receive the database directly. Some services expose broad concrete dependencies. Provider SDK concepts and untyped maps appear outside adapter boundaries. This makes code easy to start but difficult to reason about once workflows cross modules.

### 4.3 Implicit policy

Authentication, workspace membership, role checks, team scope, token scope, and resource ownership are not one concern. Middleware can establish context, but a service must still authorize the requested operation against the actual resource. The target makes this policy explicit and testable.

### 4.4 Accidental reuse

Copying a validation helper or pagination parser is bad, but an over-general `utils` package is also bad. Reuse must preserve meaning. Syntax-level UUID parsing can be shared; whether a user may assign a story belongs to stories policy.

### 4.5 Scale without evidence

Scalability is not achieved by adding caches and goroutines everywhere. The target establishes query observability, bounded concurrency, cursor pagination, short transactions, and durable asynchronous work before adding complexity.

## 5. Architectural principles

Every future API change should follow these rules:

1. **Make invalid states difficult to represent.** Use named domain types, typed commands, and typed query parameters.
2. **Authorize the resource, not only the route.** Every tenant-owned read or write proves workspace and, where applicable, team ownership.
3. **Keep interfaces small and consumer-owned.** Define an interface where it is consumed, normally beside the service using it.
4. **Prefer concrete types within a package.** Do not create an interface merely because a struct exists.
5. **Keep transport, domain, and persistence models separate.** Mapping is deliberate, local, and tested.
6. **Put transaction boundaries around business invariants.** Do not let repository methods silently open nested or unrelated transactions.
7. **Do not perform network calls inside database transactions.** Persist an outbox intent, commit, then deliver asynchronously.
8. **Make time, randomness, and IDs injectable where behavior depends on them.** This improves determinism without abstracting every standard-library call.
9. **Prefer boring code.** A clear repeated three-line mapper is better than reflection that hides behavior.
10. **Measure before optimizing.** Use traces, `pg_stat_statements`, query plans, profiles, and load tests.
11. **Generated code is an adapter detail.** sqlc and OpenAPI generated types do not become the domain model.
12. **A new abstraction needs an owner and a stability argument.** Shared packages must not become dumping grounds.

These choices align with Go's guidance on simple package names, intentional interface placement, explicit goroutine lifetimes, and the `internal` visibility boundary in the [Go module layout guide](https://go.dev/doc/modules/layout) and [Go code review guidance](https://go.dev/wiki/CodeReviewComments).

## 6. Target repository layout

The following is the default shape for a substantial module. Small modules may omit files they do not need; they must not invent a different location for the same responsibility.

```text
apps/server/
├── api/
│   └── openapi/
│       └── v1/
├── cmd/
│   ├── api/
│   └── worker/
├── docs/
│   ├── adr/
│   ├── architecture/
│   ├── runbooks/
│   └── standards/
├── internal/
│   ├── bootstrap/
│   │   ├── api/
│   │   ├── modules/
│   │   └── worker/
│   ├── migrations/
│   ├── modules/
│   │   └── stories/
│   │       ├── commands.go
│   │       ├── errors.go
│   │       ├── model.go
│   │       ├── policy.go
│   │       ├── queries.go
│   │       ├── http/
│   │       │   ├── routes.go
│   │       │   ├── request.go
│   │       │   ├── response.go
│   │       │   ├── create.go
│   │       │   └── list.go
│   │       ├── repository/
│   │       │   ├── repository.go
│   │       │   ├── mapper.go
│   │       │   ├── queries/
│   │       │   │   ├── create.sql
│   │       │   │   ├── get.sql
│   │       │   │   └── list.sql
│   │       │   └── sqlc/
│   │       │       └── *.sql.go
│   │       ├── service/
│   │       │   ├── service.go
│   │       │   ├── create.go
│   │       │   ├── update.go
│   │       │   └── list.go
│   │       └── worker/
│   │           ├── tasks.go
│   │           ├── handler.go
│   │           └── schedule.go
│   ├── platform/
│   │   ├── authn/
│   │   ├── authz/
│   │   ├── database/
│   │   ├── httpapi/
│   │   ├── idempotency/
│   │   ├── integrations/
│   │   ├── pagination/
│   │   ├── secrets/
│   │   └── validation/
│   ├── testkit/
│   └── worker/
└── go.mod
```

This is a destination, not a request to rename every package in one pull request. Existing packages should move incrementally when a module is migrated.

### 6.1 Where to look

| Question                                               | Canonical location                                              |
| ------------------------------------------------------ | --------------------------------------------------------------- |
| What routes does a module expose?                      | `internal/modules/<module>/http/routes.go`                      |
| How is JSON/path/query input shaped?                   | `http/request.go` and a use-case handler file                   |
| What response is public?                               | `http/response.go` or generated OpenAPI adapter                 |
| What business operation runs?                          | `service/<use-case>.go`                                         |
| What invariants apply?                                 | module root `policy.go`, `commands.go`, and named domain types  |
| What errors can the module return?                     | module root `errors.go`                                         |
| What SQL executes?                                     | `repository/queries/<use-case>.sql`                             |
| What code did sqlc generate?                           | `repository/sqlc/`                                              |
| How are rows mapped to domain values?                  | `repository/mapper.go`                                          |
| Where is the module composed?                          | `internal/bootstrap/modules/<module>.go`                        |
| Where are a module's task types and producer defaults? | `internal/modules/<module>/worker/tasks.go`                     |
| Where is a task/stream/email handler implemented?      | owning module's `worker/handler.go` or a named capability file  |
| Where are module schedules declared?                   | owning module's `worker/schedule.go`                            |
| How is background work registered?                     | module worker registration plus `internal/bootstrap/worker`     |
| How is behavior tested?                                | adjacent `*_test.go`; database contracts under repository tests |
| How is the public contract described?                  | `api/openapi/v1` and `apps/docs/content/docs/api`               |

### 6.2 Module archetypes

The default full resource module is not mandatory. Classify each module so its required artifacts are predictable:

| Archetype             | Typical shape                                                | Examples in the current API          | Required evidence                                                   |
| --------------------- | ------------------------------------------------------------ | ------------------------------------ | ------------------------------------------------------------------- |
| Resource module       | domain + service + repository + optional HTTP/worker         | stories, objectives, teams           | policy, persistence, transport/worker contracts that actually exist |
| Orchestration module  | domain/service consuming narrow ports; no owned SQL          | emailagent                           | use-case and provider/task tests; no invented repository            |
| Static domain/catalog | pure data/rules and no infrastructure                        | workflowtemplates                    | deterministic unit tests and documented owner                       |
| Transport/gateway     | HTTP/protocol adapter over existing services/platform        | agentreadiness                       | protocol/auth/contract tests; persistence only if truly owned       |
| Adapter/integration   | provider adapter plus service/repository/worker capabilities | Slack, GitHub, calendar              | provider contract, credential, inbox/outbox, and recovery tests     |
| Platform/runtime      | lifecycle or cross-cutting infrastructure                    | health, SSE, bootstrap, task runtime | lifecycle/readiness/race/recovery evidence                          |

Do not add empty `repository`, `http`, `worker`, or sqlc packages to make an archetype look uniform. If a module changes archetype, document why and update its owner/contract.

## 7. Dependency rules

### 7.1 Allowed direction

| Package            | May depend on                                                     | Must not depend on                                      |
| ------------------ | ----------------------------------------------------------------- | ------------------------------------------------------- |
| Module root/domain | standard library and deliberately chosen domain-safe packages     | HTTP, database drivers, sqlc output, provider SDKs      |
| Module service     | module domain, consumer-owned ports, selected platform primitives | module HTTP, concrete provider SDKs, generated SQL rows |
| Module repository  | module domain, module sqlc output, database platform              | HTTP, other modules' handlers, network clients          |
| Module HTTP        | module service ports, domain values, HTTP platform                | concrete repositories, raw database handle              |
| Provider adapter   | provider SDK, integration capability contracts                    | unrelated module internals                              |
| Bootstrap          | all constructors needed for composition                           | business logic and SQL                                  |
| Worker handler     | service use cases and task envelope                               | ad hoc cross-domain SQL                                 |

Add an architecture test or static dependency rule that fails when imports violate these directions. Start with the strictest, clearest constraints rather than trying to encode every preference at once.

`pkg` is not a second home for application internals. Retain it only for packages that are coherent and reusable without importing FortyOne modules. The current `pkg/jobs`, `pkg/consumer`, and `pkg/emailthread` import application internals and should move under `internal` or their owning modules during migration. Do not move everything mechanically; decide ownership first and preserve public package stability only if an actual external consumer exists.

### 7.2 Cross-module calls

A module may call another module through a narrow interface owned by the caller:

```go
// Defined by the package that needs the behavior.
type StoryCreator interface {
	CreateFromFeedback(ctx context.Context, actor Actor, cmd CreateStoryFromFeedback) (StoryID, error)
}
```

The interface should describe the business capability, not mirror every method on `stories.Service`. An adapter is acceptable when the existing concrete service has a wider or differently named method. Avoid import cycles by moving stable shared domain concepts into an explicit platform or kernel package only when the concept is genuinely shared.

### 7.3 No database handles in handlers

Handlers should not receive `*sqlx.DB`, `*sql.DB`, or `*pgxpool.Pool`. The reviewed route configs often pass a database directly because middleware or helper code performs queries. Migrate those operations behind authorization, session, or module ports. This creates one auditable persistence boundary and makes handler tests cheap.

## 8. Module design standard

### 8.1 Domain types

Use named values where they prevent category mistakes or centralize stable parsing:

- `WorkspaceID`, `TeamID`, `StoryID`, and `PrincipalID` where crossing boundaries is otherwise ambiguous;
- enums with explicit parsers for roles, states, providers, sort fields, directions, and scopes;
- value objects for cursor, email, webhook delivery ID, external provider identity, and idempotency key;
- commands that distinguish omitted, null, and provided values for patch semantics.

Do not wrap every integer or string. Introduce a type when it protects an invariant, improves an interface, or prevents mixing values.

### 8.2 Commands and queries

Service entry points should reflect intent:

- `CreateStory`
- `AssignStory`
- `MoveStory`
- `ArchiveStory`
- `ListStories`

Avoid a universal `Update(id, map[string]any)` operation. It obscures allowed fields, authorization, null behavior, emitted events, and optimistic concurrency. A command may still contain multiple optional fields when those fields form one coherent user action.

### 8.3 Services

A module's `service/service.go` owns the constructor and stable dependency set. Put individual use cases in named files. A method should usually:

1. validate domain input;
2. authorize the actor and resource scope;
3. invoke a repository or transaction callback;
4. map persistence conflicts into domain errors;
5. persist an event/outbox record inside the same transaction when needed;
6. return a domain result.

It should not parse HTTP, know status codes, or return provider SDK models.

### 8.4 Repositories

Repositories are persistence adapters, not a second service layer. They should:

- expose methods in domain terms;
- accept workspace scope on tenant-owned operations;
- call sqlc-generated queries;
- map generated rows and PostgreSQL errors to domain values/errors;
- participate in a supplied transaction when required;
- contain no email, provider API, task publishing, or permission policy.

Repository interfaces live beside the consuming service. Large repositories may have multiple concrete stores by cohesive capability, for example `StoryStore`, `StorySearchStore`, and `StoryScheduleStore`. Do not create one interface per file merely to make files smaller.

### 8.5 HTTP handlers

Handlers are adapters with a deliberately small lifecycle:

1. retrieve the authenticated `Actor` established by middleware;
2. parse path and query values through shared syntax helpers;
3. decode exactly one bounded JSON object where a body is expected;
4. map the transport request into a domain command;
5. call one use case;
6. map domain errors through the central error registry;
7. write the documented response.

Transport structs should implement explicit validation or use the one standardized validator path. Validation tags that are never invoked are misleading and must be removed or made effective.

### 8.6 Background jobs

Background behavior belongs to the module whose business state it changes. Its `worker` package owns versioned envelopes, producer defaults, handlers, and schedules; `internal/bootstrap/worker` only composes registrations, while `internal/worker` or a small platform package owns generic Asynq/Redis execution infrastructure. A cross-module job calls an application use case through narrow ports rather than issuing its own SQL. Move current `pkg/jobs` files to their owning modules capability by capability, move the Redis consumer to the domain/runtime that owns its stream contract, and move email-thread processing under emailreply/messaging ownership after mapping callers. Do not create one new giant `internal/jobs` package.

Every task should have:

- a versioned task type;
- a typed payload including tenant and idempotency identity;
- a producer helper that sets queue, timeout, uniqueness, and retry policy;
- a handler that loads current state and calls a service use case;
- retry classification that distinguishes transient, rate-limited, revoked, invalid, and terminal failures;
- trace and correlation propagation;
- a dead-letter or operator recovery story.

Queue payloads should normally carry IDs, not mutable records or secrets.

## 9. File and package size policy

Line limits are review signals, not compiler rules:

- **Target:** fewer than 400 handwritten lines per file.
- **Review warning:** 400–700 lines requires a clear cohesive reason.
- **Refactor gate:** above 700 handwritten lines requires an explicit exception or a split before substantial new behavior is added.
- Generated sqlc/OpenAPI files and data-only fixtures are exempt.

Split by use case or capability, never by arbitrary suffixes such as `service_part2.go`. Good examples are `oauth_install.go`, `webhook_ingress.go`, `delivery_worker.go`, `story_mutation.go`, and `repository_installations.go`.

Prioritize files where size intersects risk:

| Priority | Current hotspot               | Intended split                                                                                       |
| -------- | ----------------------------- | ---------------------------------------------------------------------------------------------------- |
| 1        | Slack service                 | install/auth, identity linking, webhook intake, event processing, rendering, delivery, cleanup       |
| 1        | Stories query repository      | create/get/list, mutation commands, scheduling, relationships, activities, search/report projections |
| 1        | GitHub service                | app installation, user linking, repository sync, webhook normalization, delivery/reconciliation      |
| 1        | Feedback repository/service   | public ingress, contributor identity, triage, delivery, analytics                                    |
| 1        | Maya HTTP/service             | conversation transport, planning, approvals, tool execution, scheduling, recovery                    |
| 2        | Messaging mutation/tool files | tool catalog, authorization, confirmation, execution, provider delivery                              |
| 2        | Reports queries               | named report projections or database views by report                                                 |
| 2        | Notifications task handler    | enqueue, fan-out, provider send, retry/recovery                                                      |
| 2        | Calendar service              | provider orchestration, sync/watch lifecycle, scheduling, outbox                                     |

Before splitting, add characterization tests around existing public behavior. A file move should not silently change authorization, SQL semantics, event order, or response shape.

## 10. Reusable platform primitives

Create these packages only with narrow contracts and owners.

### 10.1 Authentication and actor context

Replace a context containing only `userID` with an immutable actor value:

```go
type Actor struct {
	PrincipalID  uuid.UUID
	Kind         PrincipalKind // user, service account, OAuth app, system
	WorkspaceID  uuid.UUID
	CredentialID uuid.UUID
	Scopes       ScopeSet
	TeamAccess   TeamAccess
}
```

The actual representation may evolve, but it must preserve who acted. A GitHub installation, service account, or system process must not be silently attributed to the human who installed it.

### 10.2 Authorization

Authorization has two layers:

- middleware authenticates credentials, selects the tenant, and performs cheap route-level scope checks;
- a service policy authorizes the operation against the loaded resource, role, team membership, and actor type.

Policies return typed decisions/errors and are unit tested with a permission matrix. Repository queries still scope by workspace as defense in depth.

### 10.3 Request decoding

One HTTP primitive should provide:

- a default maximum body size and smaller route-specific overrides;
- `Content-Type` enforcement for JSON endpoints;
- unknown-field rejection;
- exactly one JSON value and no trailing input;
- clear syntax, type, empty-body, and too-large errors;
- one consistently executed validation path;
- safe form/multipart limits.

Raw webhook bodies use a separate bounded verifier path because signatures must cover untouched bytes.

### 10.4 Validation

Share only stable syntax-level rules:

- UUID and external-ID parsing;
- bounded string and collection lengths;
- email/URL syntax with an explicit threat model;
- cursor decoding;
- enum parsing.

Keep business rules in modules. For example, `validation` may establish that a role string is known; invitations policy decides which actor may invite that role.

### 10.5 Pagination

Adopt cursor pagination for externally visible lists and high-cardinality internal lists:

- opaque, versioned, integrity-protected cursors;
- stable ordering with a unique tie-breaker;
- `first`/`after` or a consistent equivalent;
- bounded default and maximum page sizes;
- typed sort fields and directions;
- filters included in or validated against cursor state;
- a response page with items, next cursor, and `hasMore`.

Offset pagination can remain for small, admin-only reports when its consistency and cost are acceptable, but it must be explicit.

### 10.6 Transactions

Provide a transaction runner over pgx/sqlc that:

- accepts isolation/access options;
- supplies module query sets bound with `WithTx`;
- rolls back on error or panic and commits once;
- does not hide retry semantics;
- emits duration and outcome telemetry;
- prevents service code from depending on driver transaction types where practical.

Transactions belong to use cases spanning a business invariant. Repository methods that only execute one statement should not start their own transaction.

### 10.7 Clock, IDs, and randomness

Use small concrete dependencies such as `Clock` and `IDGenerator` only in code whose behavior depends on them. Security tokens always use `crypto/rand`. Tests use fixed clocks and deterministic generators, eliminating hard-coded calendar dates and sleeps.

### 10.8 Error model

Define stable domain error categories such as:

- invalid argument;
- unauthenticated;
- forbidden;
- not found;
- conflict/version mismatch;
- rate limited;
- dependency unavailable;
- internal.

The HTTP adapter turns these into one documented error envelope with a stable machine code, safe message, request ID, field violations, and optional retry metadata. PostgreSQL and provider errors are classified at their adapter boundaries; raw errors are never exposed.

### 10.9 Idempotency

Create one idempotency service for externally retried writes:

- scope keys by principal, workspace, method, and route operation;
- hash and compare the request body;
- store in-progress/completed state and the safe response result;
- reject reuse with a different payload;
- expire records according to documented policy;
- coordinate with the use-case transaction or outbox.

This does not replace domain uniqueness constraints or provider delivery IDs.

## 11. Composition and registration

The current API service and route bootstrap files are large because every module is constructed centrally. Preserve explicit wiring while moving cohesive construction into module bundles:

```go
type Module struct {
	Service *service.Service
	HTTP    *modulehttp.Handler
	Worker  worker.Registrar
}

func Build(deps Dependencies) (Module, error)
```

This is illustrative, not a required universal interface. Each `internal/bootstrap/modules/<name>.go` should construct repositories, adapters, and services for one module or a tightly coupled family. The top-level API bootstrap composes module outputs and registers routes. Startup should return errors rather than panic where recovery/reporting is possible.

Provider registration should use compile-time descriptors and capability registries, not runtime Go plugins or reflection. Details are in the companion integration plan.

## 12. Concurrency and lifecycle

Every goroutine must have an answer to four questions:

1. Who starts it?
2. Which context cancels it?
3. Who waits for it?
4. Where does its terminal error go?

Use `errgroup` or an equivalent explicit supervisor for API background consumers, worker services, schedulers, metrics endpoints, and SSE infrastructure. Shutdown should:

- stop accepting new HTTP traffic;
- mark readiness false;
- drain bounded in-flight requests;
- stop schedulers from creating new work;
- gracefully stop consumers/workers;
- flush telemetry;
- close database and Redis pools once.

Operator consoles such as Asynqmon must not bind publicly without authentication and network controls.

## 13. Scalability and performance standard

### 13.1 Database

- Size `pgxpool` from database capacity, API task count, worker concurrency, and deployment replica count.
- Set statement, lock, and request deadlines appropriate to the operation.
- Use explicit columns; prohibit `SELECT *` in application queries.
- Eliminate N+1 paths with set-based queries, joins, or deliberate batching.
- Use cursor pagination on high-cardinality lists.
- Record slow query fingerprints through tracing and `pg_stat_statements`.
- Require `EXPLAIN (ANALYZE, BUFFERS)` evidence for critical new or regressed queries in a representative dataset.
- Add indexes from measured access paths; verify write cost and unused indexes.
- Introduce read replicas, partitioning, materialized views, or caches only after evidence and an ADR.

### 13.2 HTTP

- Apply server read-header, read, write, and idle timeouts.
- Bound headers, bodies, multipart data, list sizes, and concurrent expensive operations.
- Return request IDs and rate-limit metadata.
- Compress only appropriate response types and sizes.
- Stream intentionally; do not buffer arbitrarily large provider payloads or exports.

### 13.3 Async work

- Make inbox/outbox operations idempotent.
- Bound retries with jitter and respect provider `Retry-After`.
- Separate queues by latency and failure domain where evidence supports it.
- Expose queue age, retry count, dead-letter count, and terminal failure reasons.
- Prevent one tenant or provider from exhausting global concurrency.

### 13.4 Profiling

Use benchmarks and profiles for proven hotspots. Go's official [profiling guidance](https://go.dev/blog/pprof) should inform investigation. A refactor is not a performance improvement until measurements show it.

## 14. Initial module priority map

| Group                   | Modules                                                                                                          | First architectural objective                                                        |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Security boundary       | users, workspaces, teams, teamsettings, invitations, admin, subscriptions                                        | Central actor/policy model, immediate revocation, safe auth and billing webhooks.    |
| Core work graph         | stories, comments, links, labels, states, epics, sprints, objectives, objectivestatus, keyresults, okractivities | Typed commands, tenant-scoped sqlc, optimistic concurrency, consistent events.       |
| Integration runtime     | messaging, Slack, GitHub, calendar, Figma, integrationrequests, emailreply                                       | Capability boundaries, credential vault, common webhook inbox/outbox.                |
| AI and automation       | Maya, chatsessions, agentreadiness, workflowtemplates                                                            | Explicit tool authorization, approval state, OAuth scopes, deterministic scheduling. |
| Reporting and discovery | reports, search, activities, notifications                                                                       | Typed projections, cursor/filter standards, query performance budgets.               |
| Ingress and content     | feedback, attachments, documents, mentions, emailagent                                                           | Bounded input, identity semantics, storage safety, background processing.            |
| Platform plumbing       | health, SSE, taskhandlers, bootstrap, shared `pkg`                                                               | Lifecycle supervision, dependency readiness, architecture enforcement.               |

The data migration order for every module is specified in the companion sqlc plan.

## 15. Architecture work packages

### A0 — Freeze the standard

Deliverables:

- approve ADRs for modular-monolith boundaries, sqlc/pgx, actor model, error format, pagination, and public API versioning;
- add an architecture glossary and canonical module template;
- record baseline package dependencies, route inventory, file hotspots, and test status;
- define owners for platform packages.

Acceptance:

- a new engineer can locate route, use case, SQL, mapper, and test for a sample operation from documentation alone;
- no architecture ADR contradicts the other two plans.

### A1 — Establish platform contracts

Deliverables:

- `Actor` and credential-independent auth context;
- policy decision/error contract;
- bounded decoder and parameter parsers;
- stable error envelope;
- cursor package;
- clock, ID generator, transaction runner, and idempotency contracts.

Acceptance:

- each primitive has focused unit/fuzz tests;
- no primitive embeds stories-, Slack-, or other domain-specific behavior;
- adoption can occur route by route.

### A2 — Add architecture enforcement

Deliverables:

- static import-boundary checks;
- lint rule or CI script for direct database use outside repository/database platform packages;
- generated-code exclusions and handwritten file-size reporting;
- route inventory containing method, path, auth mode, workspace mode, role/scope, handler, and OpenAPI operation.

Acceptance:

- violations fail CI with an actionable file and rule;
- existing debt is baselined explicitly, and the baseline cannot grow.

### A3 — Split bootstrap by module

Deliverables:

- module-focused composition files;
- explicit provider registries;
- startup error aggregation and validation;
- supervised lifecycle for API and worker processes.

Acceptance:

- adding a normal module does not require editing a 400+ line constructor;
- API and worker dependencies are visible without runtime reflection;
- startup and shutdown have integration tests.

### A4 — Refactor hotspots during sqlc migration

Deliverables:

- characterization tests before moves;
- split large handlers/services/repositories by use case;
- replace broad interfaces and map updates;
- move SQL to module query files.

Acceptance:

- no migrated handwritten file exceeds 700 lines without an approved exception;
- service and HTTP behavior remains contract-compatible unless versioned deliberately;
- every moved operation has ownership and tests.

### A5 — Remove legacy paths

Deliverables:

- remove direct handler/service SQL;
- remove obsolete helper/validation/pagination variants;
- remove query-string bearer authentication and legacy credential fallbacks;
- remove old bootstrap baselines once all modules comply.

Acceptance:

- architecture checks run with no debt allowlist for the target rules;
- repository and docs searches return one canonical solution for each shared concern.

## 16. Required ADRs

Before broad implementation, record at least:

1. modular monolith and module dependency direction;
2. sqlc plus native pgx target and SQLx exception policy;
3. transaction ownership and outbox usage;
4. actor/principal model and authorization caching/revocation;
5. public API versioning and OpenAPI code generation;
6. cursor pagination and compatibility policy;
7. API key, service-account, and OAuth token storage/lifecycle;
8. integration capability registry and external extension boundary;
9. webhook signing, replay protection, inbox/outbox retention;
10. observability fields, SLO ownership, and sensitive-data rules.

An ADR is not a substitute for tests or implementation. It records the choice, alternatives, consequences, owner, and conditions under which the decision should be revisited.

## 17. Architecture definition of done

A module is architecture-complete when all requirements applicable to its declared archetype pass; “not applicable” must name the missing capability rather than create an empty layer:

- if it exposes routes, its route inventory and public/internal classification are explicit;
- if it has handlers, they have no direct database or provider SDK dependency;
- if it executes actor-initiated use cases, services/policies authorize a typed actor against the resource;
- if it owns persistence, repositories are tenant-scoped and sqlc-backed or the exact operation follows the near-zero exceptional-query ADR policy;
- if it owns multi-step state change or external delivery, transaction and event/outbox ownership are visible;
- generated types do not leak into domain or transport contracts;
- large files have been split by cohesive use case;
- shared validation and pagination primitives are used by applicable transports/lists;
- transport errors map to the standard envelope where HTTP exists;
- the applicable unit, repository, HTTP, provider, worker, authorization, and failure-path tests pass for its archetype;
- traces, logs, and metrics contain correlation and tenant-safe metadata without secrets;
- internal architecture and public API docs are current;
- the module has no unexplained dead code, duplicate implementation, or compatibility path.

## 18. 10/10 architecture scorecard

| Dimension       | 10/10 gate                                                                                                                                  |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Discoverability | Canonical locations are followed across every migrated module; route-to-SQL tracing is documented and mechanically inventoryable.           |
| Boundaries      | Transport, domain, persistence, and provider models do not leak across layers.                                                              |
| Authorization   | Every operation has explicit actor, scope, tenant, and resource policy tests.                                                               |
| Readability     | No unexplained giant handwritten files; names express use cases; abstractions have stable meaning.                                          |
| Reuse           | One canonical primitive exists for request decoding, cursors, errors, transactions, actors, idempotency, secrets, and integration delivery. |
| Scalability     | Critical paths have budgets, traces, query plans, bounded concurrency, and representative load evidence.                                    |
| Reliability     | Transactions, outboxes, retries, shutdown, and recovery behavior are explicit and tested.                                                   |
| Extensibility   | New providers implement capability contracts without changing core business rules.                                                          |
| Onboarding      | A new engineer can make and test a small endpoint change using one documented workflow.                                                     |
| Enforcement     | CI prevents boundary, generated-code, API-contract, security, migration, and test regressions.                                              |

No score is awarded merely for moving files. The target is simpler reasoning, stronger invariants, and faster safe change.
