# FortyOne API Modernization: Delivery, Testing, and Documentation Roadmap

**Status:** Active implementation plan
**Scope:** `apps/server`, its CI/release path, and public API documentation in `apps/docs`  
**Companion plans:** [target Go architecture](./01-target-go-architecture.md) and [typed data, security, and integrations](./02-typed-data-security-and-integration-platform.md)
**Implementation status:** [evidence and crash-recovery ledger](./00-implementation-status.md)

## 1. Outcome

This document turns the target architecture into a safe sequence of work. It is designed to avoid three common modernization failures:

1. a long refactor branch that cannot be released incrementally;
2. a mechanical sqlc migration that preserves insecure or unclear behavior;
3. a large test rewrite that produces coverage numbers without trustworthy failure detection.

The program finishes when FortyOne can make API changes quickly with strong compile-time, security, behavioral, database, contract, and operational feedback. “All SQL uses sqlc” is a major milestone, but it is not the whole definition of success.

### 1.1 Deployment and distribution scope

FortyOne is an internally operated managed service. This program does not
maintain a public source distribution, self-hosting installer, Compose stack,
public container registry, or community support contract. Production API and
worker images remain internal deployment artifacts and are published only to a
private registry through the reviewed release workflow.

The public developer API remains in scope because it is the supported boundary
for customer and partner integrations; it is not a self-hosting mechanism.

## 2. Baseline and interpretation

The repository was actively changing during this review. Test metrics were therefore measured from clean commit `b71f7cb3d4a678b56e858d414ca1a71d2ececc73`, while architectural and persistence inventories also considered the current working tree. Re-baseline immediately before implementation.

### 2.1 Clean baseline

| Measure                                      |              Result | Meaning                                                                   |
| -------------------------------------------- | ------------------: | ------------------------------------------------------------------------- |
| Production Go files / LOC                    | 542 / about 154,600 | Substantial modular monolith.                                             |
| Test files / test functions                  |         223 / 1,321 | Many tests exist, but count does not imply broad confidence.              |
| Production package directories without tests |                  70 | Important repository and runtime paths are uncovered.                     |
| Directional whole-package coverage           |         About 28.6% | Use as a starting signal, not an absolute truth.                          |
| Fuzz tests                                   |                   0 | Parser, webhook, token, and cursor boundaries need fuzzing.               |
| Meaningful benchmarks                        |                   0 | The sole benchmark is compiler-optimizable and not representative.        |
| Default PostgreSQL integration tests         | 0 reliably executed | Existing environment-gated tests silently skip and use inconsistent DSNs. |
| Redis/Asynq integration tests                |                   0 | Queue and consumer recovery are not validated as a system.                |
| Pull-request CI workflows                    |                   0 | There is no required pre-merge API quality gate.                          |

### 2.2 Current command evidence

On the reviewed clean commit:

- `go test -count=1 ./...` failed eight Maya tests because fixed August 17–18, 2026 dates were compared with production `time.Now()`;
- `go test -run '^$' ./...` passed, proving packages compiled;
- `go vet ./...` passed;
- `go build ./...` passed;
- a focused race command covering `cmd/api`, worker bootstrap, consumer, SSE, and the health package passed; health had no test file and therefore only compiled/instrumented in that run.

The active working tree later had a separate in-progress Maya fake/interface mismatch. Neither failure should be “fixed” by weakening assertions. Introduce a clock seam and keep the branch-specific work isolated.

### 2.3 Existing strengths

- Many service tests exercise non-trivial domain behavior.
- Slack and messaging have substantial service and provider-fake coverage, but their real PostgreSQL tests still skip without environment DSNs and there is no Redis/Asynq/system runtime suite.
- `t.Parallel` is widely used.
- Provider simulations already use `httptest` in several places.
- Worker handler and scheduler registration have explicit tests.
- The AST-based architecture boundary test is a good foundation to expand.
- Migrations are embedded and have complete up/down pairs in the reviewed snapshots.

### 2.4 Highest-confidence gaps

- Tests are concentrated in services: repository and HTTP behavior is much less covered.
- The ten largest test files contain roughly a third of all test LOC; Slack alone has a roughly 4,600-line test file.
- More than 100 handwritten test doubles use no common naming or ownership convention.
- Existing PostgreSQL integration tests use seven different environment variable names, do not apply migrations, and skip silently.
- There is no end-to-end API → database/outbox → worker → provider/SSE test.
- Time is often taken from the wall clock; the clean suite already demonstrated date-based decay.
- CI deploys images from `main` without testing, vetting, migration validation, vulnerability checks, or service-stability waiting.
- API shutdown does not supervise all consumers/SSE lifecycles, and worker monitoring is not production-hardened.
- README and `go.mod` disagree on the required Go version.

## 3. Program rules

### 3.1 Security before structure

Fix confirmed exploitable/high-risk behavior with focused regression tests before moving the affected code. A file move that makes a vulnerability harder to recognize is negative progress.

### 3.2 Incremental, releasable slices

Each pull request should normally migrate one coherent use case or a small module:

- tests characterize behavior;
- new platform/sqlc path is introduced;
- callers switch;
- old path is deleted;
- documentation and inventory update;
- the result can deploy independently.

Avoid a repository-wide package rename followed by weeks of behavior changes.

### 3.3 One source of truth

Do not create competing old/new helpers indefinitely. Compatibility adapters may exist during a short migration, but they need an owner, telemetry if behavior is compared, and an explicit deletion task.

### 3.4 Baseline-aware gates

CI may initially carry an explicit debt baseline so modernization can start. New violations are prohibited immediately, and the baseline only decreases. Coverage gates should detect regressions while targeted critical-path thresholds rise.

### 3.5 Evidence over completion labels

A work item is complete only with the acceptance evidence stated here: tests, query plans, migration verification, contract diff, security review, load result, or runbook exercise. A checklist entry such as “add tests” is not enough.

## 4. Program structure

Run the work as six coordinated streams:

| Stream                             | Owner responsibility                                                        | Primary artifacts                                       |
| ---------------------------------- | --------------------------------------------------------------------------- | ------------------------------------------------------- |
| Architecture                       | Module boundaries, composition, platform primitives, dependency enforcement | ADRs, module template, architecture tests               |
| Data                               | pgx/sqlc, query migration, transactions, schema, performance                | sqlc config, query files, migration matrix, query plans |
| Security/API platform              | principals, authorization, credentials, API v1, idempotency, rate limits    | policy matrices, OpenAPI, key/OAuth services, audit     |
| Integrations                       | provider registry, credential vault, inbox/outbox, adapters                 | capability contracts, provider tests, runbooks          |
| Quality/operations                 | tests, CI, observability, lifecycle, release                                | testkit, workflows, dashboards, deployment gates        |
| Documentation/developer experience | architecture, onboarding, API docs, SDK examples                            | internal docs, public docs, tutorials, changelog        |

One technical program owner maintains dependency order and the scorecard. Each work package still has a code owner; “platform team” must not become ownerless shorthand.

## 5. Phase 0 — Stabilize and secure the baseline

**Goal:** establish a green, reproducible starting point and close known security gaps before broad refactoring.

### P0.1 Reproduce and record

Tasks:

- select a clean baseline commit and record the active feature-branch delta;
- capture `go env`, Go/sqlc/tool versions, PostgreSQL version, and Docker image bases;
- generate module, route, SQL, transaction, credential, webhook, and test inventories;
- capture current coverage, test duration, race status, binary/image size, startup/readiness time, and representative endpoint/query latency;
- publish unresolved failures separately from plan work.

Acceptance:

- one documented command reproduces every baseline result;
- generated inventories are committed or produced deterministically;
- unrelated dirty-tree work is neither overwritten nor counted as modernization completion.

### P0.2 Make the suite deterministic

Tasks:

- inject a clock into Maya reconciliation and other time-sensitive decisions;
- replace hard-coded soon-to-expire dates with relative/fixed-clock fixtures;
- eliminate arbitrary sleeps through synchronization/eventual assertions;
- make IDs/randomness injectable only where assertions depend on them;
- classify and resolve every `t.Skip` or attach an issue/condition;
- remove or replace the meaningless invitation benchmark.

Acceptance:

- `go test -count=1 ./...` passes repeatedly across local and CI time zones;
- a multiple-count run of sensitive packages has no flaky failures;
- tests never require the actual current date unless explicitly testing wall-clock integration.

### P0.3 Close security blockers

Execute `DS1` from the typed-data/security plan:

- OTP and public-auth abuse protection;
- invitation/admin authorization;
- cross-tenant comment/link/private-memory operations;
- GitHub integration administration policy;
- GitHub token encryption and Slack plaintext removal;
- Stripe webhook state/retry correctness;
- request/webhook body limits;
- query-token removal and strict credential verification;
- revocation-safe authorization cache;
- Asynqmon network/auth controls.

Acceptance:

- abuse and cross-tenant regression suites pass;
- a secret-data migration report confirms no prohibited plaintext;
- all fixes can deploy before broad sqlc/file refactors.

### P0.4 Establish minimum PR CI

Required jobs:

1. format/import/tidy drift;
2. compile, vet, static analysis;
3. fast unit tests;
4. architecture debt-growth check;
5. generated-artifact drift check;
6. dependency vulnerability and secret scan.

Acceptance:

- protected branches require the jobs;
- failed jobs block merge;
- unwaived high/critical vulnerability, secret, or image findings block; every temporary exception has a named owner, justification, compensating control, and expiry;
- release workflow depends on validated artifacts rather than rebuilding untested source.

### P0.5 Establish minimum runtime release signals

Before the first modernization release relies on automated rollout gates:

- supervise API HTTP, SSE, and Redis consumers under one cancelable lifecycle;
- supervise worker server, scheduler, monitoring endpoint, and telemetry;
- expose truthful, authenticated/network-safe API and worker liveness/readiness;
- validate a declared **compatible schema range**, not exact equality with the latest migration;
- emit real build/service/environment metadata, request/error/latency signals, queue health/age, and deployment correlation;
- make readiness false during drain and test bounded graceful shutdown;
- establish a minimal deploy dashboard/alert and smoke check.

Acceptance:

- an API/worker dependency failure or dead supervisor changes readiness visibly;
- an expand migration does not evict an older instance that declares compatibility with the new schema;
- deployment automation never waits on a signal the process cannot truthfully provide.

## 6. Phase 1 — Foundations

**Goal:** provide the small shared components needed for safe incremental migration.

### P1.1 ADRs and standards

Approve the ADR list in the architecture plan. Add short standards for:

- module structure and imports;
- sqlc query/mapping conventions;
- transaction/outbox ownership;
- request decoding and validation;
- actor/authorization policy;
- error and cursor contracts;
- testing and fixture ownership;
- logs/traces/metrics and sensitive data;
- public API compatibility.

Acceptance: the standards do not merely describe future ideals; each has a lint/test/review mechanism and an adoption path.

### P1.2 Test infrastructure

Create `internal/testkit` for infrastructure only:

```text
internal/testkit/
  postgres.go       create DB, apply migrations, reset isolated schema
  redis.go          start/connect, namespace keys, cleanup
  api.go            build test application and authenticated actors
  clock.go          manual/fixed clock
  ids.go            deterministic ID source
  eventually.go     bounded polling with useful diagnostics
  provider.go       signed request/server primitives, no domain fixtures
```

Rules:

- domain builders remain beside their module;
- repository/Redis/Asynq integration files use one documented `integration` build tag, and their CI job runs `go test -tags=integration ./...`;
- the default `go test ./...` suite remains hermetic and requires no external service;
- tests cannot silently skip required integration infrastructure;
- one DSN/environment contract is used;
- parallel database tests receive isolated databases/schemas;
- cleanup is registered with `t.Cleanup`;
- real migrations construct schemas;
- secrets in fixtures are obviously fake and never logged.

Acceptance:

- one repository, one Redis consumer, and one HTTP slice use the harness in CI;
- a missing dependency fails the integration job with a clear message;
- parallel tests do not share mutable tenant data.

### P1.3 pgx/sqlc foundation

Deliver `DS2` from the typed-data plan:

- checksum-pinned sqlc v1.31.1 tooling and independently pinned pgxpool;
- module-local generation;
- one root config with global UUID/time/date override policy and explicit per-module blocks;
- a generated-type contract fixture covering nullability, enum, numeric, JSONB, arrays, stable params structs, and empty-list behavior;
- transaction runner;
- SQLSTATE mapping;
- non-mutating clean generation/compile/drift checks plus database-backed vet/prepare against the exact migrated schema;
- direct-SQL/SQLx debt inventory.

Pilot with a small tenant-sensitive module, then a read-heavy complex module. The pilot must prove both CRUD and aggregate/null/filter shapes before the convention is frozen.

Acceptance also requires that generated directories contain no hand-written files, a sqlc upgrade produces a reviewable generated-contract diff, an empty `SQLC_DATABASE_URL` fails before vet can connect to a default local database, and SQLC Cloud is not required unless separately approved by ADR.

### P1.4 HTTP platform foundation

Implement and test:

- bounded exactly-one-value JSON decoding;
- path/query typed parsers;
- one validation mechanism;
- standard error envelope;
- actor context;
- cursor codec;
- rate-limit response metadata;
- idempotency receipt service.

Acceptance:

- table and fuzz tests cover malformed, oversized, trailing, unknown-field, Unicode, numeric, and cancellation cases;
- existing routes can adopt components independently.

### P1.5 Expanded architecture checks

Fail on new:

- SQL outside repository/platform database packages;
- SQLx imports;
- repository/service imports of HTTP or `pkg/web`;
- cross-module HTTP imports;
- repository imports of service packages or sibling concrete repositories;
- service imports of concrete repository adapter packages;
- cross-module concrete service-to-service imports where a caller-owned capability port should be injected;
- repository-to-repository dependencies;
- cycles in the module dependency graph, including reintroduction of the current subscriptions/workspaces cycle;
- auth-context reads from repositories;
- database fields in route configs;
- generated sqlc/OpenAPI type leakage into service/domain packages;
- unbounded raw webhook reads;
- handwritten files over the configured warning/gate without exception.

Baseline existing debt and report it by module.

## 7. Phase 2 — Tenant and identity core

**Goal:** establish the actor, authorization, tenant, membership, and credential model before core work and public APIs depend on it.

Sequence:

1. comments and links as small tenant-isolation pilots;
2. users and system actors;
3. workspaces and memberships;
4. teams and team settings;
5. invitations;
6. subscriptions/billing webhook state;
7. admin separation;
8. attachments/documents ownership.

For each module:

- write role × team × principal × resource matrix;
- add negative tests using IDs from a second tenant;
- migrate queries to sqlc;
- remove raw transactions and side effects before commit;
- add immutable audit events for privileged changes;
- update route inventory and docs;
- delete old paths.

### P2 exit gate

- membership removal and role demotion take effect immediately for sensitive operations;
- all identity/tenant modules are SQLx-free;
- no handler or middleware queries the database directly;
- invitation, credential, membership, and billing state machines pass concurrent retry tests;
- public API work can depend on one stable `Actor` and policy model.

## 8. Phase 3 — Core work graph and reusable query behavior

**Goal:** migrate the product's central collaboration model while eliminating generic update/filter duplication.

### P3.1 Reference-data modules

Migrate activities, labels, states, objectivestatus, epics, okractivities, and workflowtemplates. Use them to finish:

- shared cursor parsing;
- typed enum filters/sorts;
- repository mapper convention;
- audit actor convention;
- error mapping.

### P3.2 Objectives and sprints

Migrate objectives, key results, sprints, and their automation/activity flows:

- intent-specific update commands;
- optimistic concurrency where user edits can collide;
- transactionally consistent activity/outbox events;
- schedule/automation retry tests;
- no reflection-derived DB field maps.

### P3.3 Stories

Migrate stories in vertical slices rather than one rewrite:

1. get by ID and visibility;
2. list/filter/pagination;
3. create;
4. title/description/detail mutation;
5. state/position transition;
6. assignment and relationships;
7. dates/scheduling;
8. comments/links/attachments projections;
9. activity/event/outbox;
10. automation/provider entry points.

For every slice:

- preserve route compatibility unless intentionally versioned;
- prove cross-workspace/team isolation;
- record representative query plan;
- test expected-version conflict;
- test event/outbox exactly-once intent;
- delete the corresponding dynamic map/SQLx path.

Split the giant files by those capabilities as behavior moves.

### P3.4 Reports, search, and notifications

- name and type every report projection;
- cap date windows/page sizes and add statement budgets;
- make ranked search cursors stable and visibility-safe;
- make notification fan-out and delivery retryable/idempotent;
- add representative-data plan/load tests.

### P3 exit gate

- all core work modules are SQLx-free;
- `web.GetFilters`, copied pagination helpers, and DB-tagged transport/service patch models are removed;
- stories' critical create/read/update/list flows pass repository, HTTP, authz, concurrency, and performance suites;
- no migrated handwritten file exceeds the architecture gate without an approved reason.

## 9. Phase 4 — Integration runtime consolidation

**Goal:** migrate provider persistence and make future providers additive rather than invasive.

### P4.1 Preserve the reference runtime

Extract common contracts from the proven Slack/messaging paths without rewriting their behavior:

- signed bounded ingress;
- installation-generation fencing;
- encrypted/deduplicated durable inbox;
- quick acknowledgement;
- async idempotent processing;
- outbox and message mapping;
- classified retry/revoke behavior;
- hashed one-time linking nonce.

Contract tests run against each provider adapter.

### P4.2 Credential vault and installation model

- migrate Slack, GitHub, calendar, and Figma credential storage;
- remove provider-local plaintext/crypto fallbacks;
- bind AAD to tenant/provider/install/generation;
- add atomic refresh/rotation tests;
- audit install, refresh, revoke, disconnect, and access class;
- document recovery and key rotation.

### P4.3 GitHub hardening and decomposition

- admin-only installation/settings mutations;
- migrate user tokens to vault;
- move webhook to the common durable gateway;
- split install/OAuth, repository catalog, sync mapping, webhook normalization, and reconciliation;
- extract code-host capability ports.

### P4.4 GitLab proof

Implement a narrow GitLab slice—installation/auth, repository catalog, one work-item mapping, comment, and webhook—using the code-host ports. This is the proof that the abstraction is reusable.

The proof fails if GitLab requires changes to stories business rules or provider SDK types escape the adapter.

### P4.5 Messaging/calendar/feedback families

- retain native provider capabilities rather than lowest-common-denominator UI;
- keep Slack account linking separate from broad bot installation scopes;
- preserve Google/Microsoft calendar capability-sized interfaces;
- keep external feedback identities explicit;
- keep Figma as design context rather than forcing it into feedback semantics.

### P4 exit gate

- every provider ingress is bounded, signed, durable, deduplicated, and asynchronous where the provider permits;
- every retained credential is encrypted through the shared vault;
- provider contracts pass duplicate, replay, out-of-order, rate-limit, revoke, refresh race, disconnect, and redelivery cases;
- adding GitLab has not duplicated GitHub business logic.

## 10. Phase 5 — Public developer platform

**Goal:** expose a stable, documented integration surface for out-of-process applications.

### P5.1 OpenAPI v1

- select and pin the supported OpenAPI/generator versions;
- create one split source under `apps/server/api/openapi/v1`;
- generate strict server adapters and validation;
- expose selected workspace/team/story/comment/label/sprint/objective/webhook resources;
- add schema lint and breaking-change comparison;
- consolidate the separate MCP OpenAPI copy as an explicitly separate contract.

### P5.2 Credentials and consent

- personal tokens;
- service accounts and API keys;
- OAuth applications with PKCE, exact redirects, app/user actors, scoped consent, rotating refresh/reuse detection;
- expiry, revoke, rotate, last-used, and audit;
- credential-specific rate budgets.

### P5.3 API behavior

- consistent cursors;
- stable errors/request IDs;
- idempotent writes;
- optimistic concurrency/ETags where needed;
- rate-limit headers and retry rules;
- version/deprecation/changelog policy;
- fair resource limits.

### P5.4 Outbound webhooks and SDKs

- subscription/event catalog;
- signed versioned deliveries;
- retry/redelivery/operator UI;
- TypeScript and Go SDKs;
- webhook verification helpers;
- runnable examples validated in CI.

### P5 exit gate

- every external operation has contract, authorization, repository, example, and abuse tests;
- one sample third-party app completes OAuth or service-account setup, paginates stories, performs an idempotent write, verifies a webhook, and handles rate limiting using only public documentation;
- a breaking OpenAPI change cannot merge unnoticed.

## 11. Phase 6 — Remove bridges and harden operations

**Goal:** delete migration scaffolding and prove production behavior.

Tasks:

- migrate remaining feedback, Maya, chatsessions, agentreadiness, email automation, health, jobs, consumers, and task-handler SQL;
- remove SQLx from `go.mod`;
- remove direct DB fields from HTTP configs;
- move application-internal `pkg/jobs`, `pkg/consumer`, and `pkg/emailthread` to clear internal ownership;
- remove old validation, pagination, auth, secret, and webhook helpers;
- eliminate debt allowlists in architecture checks;
- complete lifecycle supervision, observability, deployment, and recovery work;
- run representative soak/load/chaos and incident exercises.

### Program exit gate

- all three plans' definitions of done pass;
- no legacy SQLx/plaintext/query-token/synchronous-unbounded-webhook path remains;
- public API and operator runbooks are current;
- release and rollback have been exercised in staging;
- the scorecard in Section 26 is evidence-backed.

## 12. Canonical testing strategy

Percentages are portfolio guidance, not a quota to game:

- roughly **70% fast unit/domain tests**;
- roughly **25% component/adapter tests** using real PostgreSQL/Redis or controlled provider HTTP;
- roughly **5% system tests** spanning built processes.

Risk determines exact allocation. Credential, authorization, concurrency, migration, and integration runtime code needs more adapter/system evidence than a pure formatter.

### 12.1 Pure domain tests

Use for:

- policies and permission matrices;
- state transitions;
- typed filter/sort/cursor decisions;
- validation and value objects;
- retry/error classification;
- scheduling calculations;
- idempotency request-hash behavior;
- mapping where semantics are non-trivial.

Requirements:

- no database, network, sleep, or wall clock;
- table tests for finite matrices;
- property/fuzz tests where the input space is large;
- parallel by default when no shared mutable state exists;
- assertions focus on public behavior, not implementation order.

### 12.2 Service/use-case tests

Use hand-written fakes that implement the narrow consumer-owned ports. A fake should model observable behavior, not reproduce the SQL or service algorithm.

Test:

- authorization before mutation;
- transaction callback behavior;
- conflict/not-found/dependency mapping;
- outbox/event intent;
- external side effect after commit;
- cancellation/deadline propagation;
- retry classification;
- actor attribution.

Keep fakes beside the consuming module unless an infrastructure fake is stable and broadly shared.

### 12.3 Repository contract tests

Run against real PostgreSQL built from all migrations. Do not use `sqlmock` as the primary proof of SQL correctness.

Every sqlc operation needs coverage through a repository scenario that validates, as applicable:

- mapping and null behavior;
- tenant/team predicate;
- zero/one/many behavior;
- ordering and cursor stability;
- uniqueness/foreign key/check errors;
- affected-row optimistic concurrency;
- transaction rollback/commit;
- concurrent claim/lease/update behavior;
- query plan on critical paths.

One contract suite may be shared across adapters only when their semantics genuinely match.

### 12.4 HTTP adapter tests

Use `httptest` around module routes or generated v1 adapters. Cover:

- method/path and content type;
- missing, malformed, oversized, unknown, and trailing JSON;
- path/query parsing;
- authentication credential kinds;
- coarse scope plus service authorization mapping;
- stable status/error code/envelope/request ID;
- cursor and rate headers;
- idempotency and version preconditions;
- response schema and sensitive-field absence.

Avoid asserting incidental JSON field order or internal function calls.

### 12.5 Module slice tests

For critical modules, construct real handler + service + repository + PostgreSQL, with controlled clocks/providers. These tests prove layer mappings that isolated unit tests can miss.

Initial slices:

- invitation creation and acceptance;
- membership removal/demotion;
- story create/update/list across two workspaces;
- Stripe webhook failure/retry;
- API-key creation/use/revoke;
- GitHub install and webhook intake;
- Slack duplicate delivery and outbox;
- Maya approval and execution lease.

### 12.6 System tests

Run built API and worker with PostgreSQL and Redis. Cover:

- health/readiness and migrations;
- authenticated API write → database/outbox → worker → controlled provider;
- Redis stream consumer reclaim/restart;
- Asynq retry/dead-letter/restart;
- SSE/client disconnect and graceful shutdown;
- the declared compatibility matrix across old/new API × old/new worker × pre/post-migration schema, including the window where an old worker sees the expanded schema;
- task-envelope forward/backward compatibility when a new API producer and old worker consumer overlap, plus rollback behavior for already-enqueued work;
- one public API/OAuth/webhook developer flow.

Keep the suite small, deterministic, and high-value.

### 12.7 Provider contract tests

Every provider adapter must pass a common suite for supported capabilities:

- valid/invalid signature;
- stale signed timestamp/replay where the provider supplies one, and delivery-ID deduplication where it does not;
- duplicate delivery;
- out-of-order event;
- unknown installation and stale generation;
- revoked/expired credential;
- refresh race and atomic replacement;
- rate limit with `Retry-After`;
- transient 5xx/network failure;
- terminal 4xx;
- redacted logging;
- disconnect/uninstall;
- outbound idempotency and echo suppression.

Store sanitized provider payloads in module `testdata` with a source/version note. Add a small, opt-in live sandbox suite for provider contract drift; it is not a substitute for deterministic CI.

### 12.8 Fuzz testing

Start with boundaries likely to receive hostile or irregular input, following Go's [fuzzing guidance](https://go.dev/doc/tutorial/fuzz):

- JSON decoder/error humanizer;
- cursor encode/decode and tamper detection;
- UUID/list/filter parsers;
- API key/token prefix parser;
- webhook signature/header parser;
- OAuth redirect/scope/state parsing;
- email thread/reply parsing;
- provider event normalizers;
- idempotency request hashing.

Every discovered crash or invariant failure becomes a permanent corpus entry.

### 12.9 Concurrency and race tests

Required scenarios:

- simultaneous membership revoke and authorized write;
- two optimistic story updates;
- duplicate webhook workers;
- refresh-token reuse and provider credential refresh race;
- inbox/outbox claim lease expiry/reclaim;
- idempotency key same/different body;
- shutdown while HTTP/SSE/consumer/task is in flight;
- scheduler/consumer dependency loss and recovery.

Run focused race tests on PRs and broader `-race` nightly. Race-free does not prove business atomicity; database concurrency assertions are also required.

### 12.10 Benchmarks and load tests

Meaningful Go benchmarks should exercise:

- cursor/signature/token verification;
- provider event normalization;
- large but bounded story mapping/rendering;
- hot policy evaluation only if profiles show it matters.

Database/API load tests should measure:

- story list/filter at representative tenant size;
- concurrent story transitions;
- reports/search windows;
- webhook bursts and duplicate rate;
- notification/outbox worker drain;
- API key verification/cache behavior;
- SSE connection lifecycle.

Use `benchstat` or statistically responsible comparisons. Prevent compiler elimination by consuming results and include allocation reporting.

## 13. Test data and double conventions

### 13.1 Builders

Use valid-by-default builders local to a domain:

```text
stories/testfixtures_test.go
  newWorkspace(...)
  newActor(...)
  newStory(...)
```

Overrides should be explicit. Do not create a universal factory with hundreds of optional fields.

### 13.2 Test doubles

Names communicate behavior:

- `stubStoryStore`: returns configured values;
- `fakeOutbox`: records writes and can model state;
- `spyPublisher`: records calls only;
- `failingCredentialVault`: produces a named failure;
- avoid `MockEverything`.

Implement only narrow ports. Generated mocks are acceptable for mechanical interfaces if their usage stays readable, but are not a program requirement.

### 13.3 Time and IDs

- fixed/manual clock for decisions and expiry;
- explicit timezone in fixtures;
- deterministic IDs when output is asserted;
- real `crypto/rand` in security integration tests plus injectable reader for rare failure tests;
- no `time.Sleep` for coordination.

### 13.4 Golden files

Use golden files only for stable, reviewable large output such as OpenAPI examples or provider message rendering. Updates must be explicit, and secrets/dynamic timestamps normalized. Do not hide dozens of business assertions in opaque snapshots.

## 14. Coverage policy

Coverage is a diagnostic and regression guard, not the goal.

### Initial gates

- publish reliable all-package coverage from the full green suite;
- no global coverage regression;
- changed-line coverage at least 80% after the baseline phase;
- every security fix has direct regression coverage;
- every new sqlc query has repository integration coverage through its behavior;
- new external API operations have HTTP contract and authorization tests.

### Progressive targets

| Milestone             | Target                                                                                                       |
| --------------------- | ------------------------------------------------------------------------------------------------------------ |
| Foundation            | Trustworthy all-package coverage at or above 35%; zero silent integration skips.                             |
| Tenant/core migration | At or above 50% global; critical actor/authz/credential/story paths at or above 80%.                         |
| Platform completion   | At or above 60% global if useful tests justify it; critical security/state-machine packages at or above 90%. |

Do not add trivial getter tests solely to hit a percentage. Review missed branch reports for security, error, and concurrency paths.

## 15. Critical workflow matrix

| Workflow                        | Unit/policy | Repository | HTTP/contract | Component/system |                Abuse/concurrency                |
| ------------------------------- | :---------: | :--------: | :-----------: | :--------------: | :---------------------------------------------: |
| OTP request/verify/session      |      ✓      |     ✓      |       ✓       |        ✓         |        brute force, replay, enumeration         |
| Invite admin/member/guest       |      ✓      |     ✓      |       ✓       |        ✓         | cross-tenant, role escalation, duplicate accept |
| Membership remove/demote        |      ✓      |     ✓      |       ✓       |        ✓         |               cached access race                |
| Story create/read/update/list   |      ✓      |     ✓      |       ✓       |        ✓         |      tenant/team, lost update, pagination       |
| Stripe webhook/subscription     |      ✓      |     ✓      |       ✓       |        ✓         |       duplicate, failure/retry, ordering        |
| PAT/service API key             |      ✓      |     ✓      |       ✓       |        ✓         |      leak, expiry, revoke, scope confusion      |
| OAuth app grant/refresh         |      ✓      |     ✓      |       ✓       |        ✓         |     redirect, PKCE, reuse, actor confusion      |
| Provider install/refresh/revoke |      ✓      |     ✓      |       ✓       |        ✓         |         refresh race, stale generation          |
| Inbound provider webhook        |      ✓      |     ✓      |       ✓       |        ✓         |    signature, size, replay, duplicate, burst    |
| Outbound app webhook            |      ✓      |     ✓      |       ✓       |        ✓         |    SSRF, retry, redelivery, secret rotation     |
| Maya mutation approval          |      ✓      |     ✓      |       ✓       |        ✓         |          expiry, lease race, tool auth          |
| API/worker shutdown             |      —      |     —      |       ✓       |        ✓         |         active SSE/task/dependency loss         |

## 16. CI design

### 16.1 Pull-request pipeline

Run independent jobs in parallel where possible:

1. **Source hygiene (target ≤3 minutes)**

   - `gofmt`/`goimports` check;
   - `go mod tidy -diff` and `go mod verify`;
   - pinned-tool assertion plus non-mutating `make sqlc-check`;
   - generated sqlc/OpenAPI/SDK drift through `make generated-check`;
   - generated API/worker configuration reference and `.env.example` drift;
   - architecture rules;
   - OpenAPI lint/breaking diff.

2. **Static/security**

   - `go vet`;
   - one pinned comprehensive linter configuration;
   - `govulncheck`;
   - secret scan;
   - dependency/license policy as approved;
   - block unwaived high/critical results; waivers require owner, justification, compensating control, and expiry.

3. **Fast tests**

   - unit/service/HTTP tests without external services;
   - coverage and changed-line gate;
   - deterministic randomized order/count where useful.

4. **PostgreSQL/Redis adapters**

   - start pinned service versions;
   - migrate empty database;
   - verify migration head, then run `make sqlc-vet` with an explicit ephemeral `SQLC_DATABASE_URL`;
   - repository and Redis/Asynq component tests;
   - migration-manifest validation, reversible-only up/down/up, and N-1 → N checks.

5. **Focused race**

   - actor/authz, integration inbox/outbox, credential refresh, consumer/SSE/lifecycle.

6. **Build/package**

   - reproducible API/worker binaries;
   - container build after tests;
   - non-root runtime, required CA certificates/timezone data, pinned base-image digest, health metadata, stop behavior, and explicit build version/commit;
   - image scan, SBOM, provenance/signature according to deployment policy;
   - deterministic checks of repository-owned or resolved deployment configuration.

Target total required PR latency: at most 10 minutes after stabilization. Optimize caching and job partitioning before weakening coverage.

### 16.2 Nightly pipeline

- full race suite;
- fuzz jobs with bounded time and corpus persistence;
- provider fixture drift and opt-in sandbox smoke where credentials exist;
- broad vulnerability/container scan;
- representative query-plan regression suite;
- load/soak on a schedule appropriate to cost;
- migration from a recent production-like schema snapshot with sanitized data shape.

### 16.3 Release pipeline

Release only a previously validated immutable image digest:

1. acquire deployment concurrency lock;
2. verify the release's API/worker/task-envelope/schema compatibility declaration against the currently deployed versions and queued task versions;
3. apply a one-shot migration job and verify the compatible schema version/range;
4. select and enforce the binary order declared by the migration compatibility
   manifest: API-first is the normal additive path, while a fail-closed
   credential or queued-payload cutover may require a replacement worker to
   finish before the vault-only API starts;
5. deploy the first binary with stability/readiness wait and run its smoke or
   cutover verification;
6. deploy the second binary only after the declared compatibility gate passes,
   then run API smoke/canary checks;
7. verify queue, webhook, error-rate, latency, and old/new task processing signals;
8. promote or roll back according to runbook, including durable queued work;
9. record artifact, migration, commit, actor, and outcome.

The order is an explicit release property, not a convention inferred from job
names. The current provider-vault and Figma migrations require schema, then one
replacement worker completing its fail-closed cutovers, then the replacement
API. An API-first release is unsafe for that window because the vault-only API
cannot use unconverted legacy credentials. Once those compatibility paths are
deleted, the manifest and workflow must be updated together and re-exercised in
staging.

Use cloud workload identity/OIDC with an assumed deployment role rather than long-lived AWS access-key inputs. Pin third-party GitHub Actions to reviewed immutable commit SHAs, with a documented update process; a moving major tag is not a reproducible release dependency. Do not publish `latest` as the sole operational identity.

Keep deploy configuration in versioned IaC or enforce a deterministic repository policy against resolved task definitions. Validate health checks, stop timeout, CPU/memory, IAM, logs, autoscaling, deployment circuit breaker/rollback, network exposure, and Asynqmon isolation before release. A live ECS task definition downloaded during deployment is not the only reviewable source of those controls.

## 17. Migration verification suite

Maintain a machine-readable migration manifest that classifies each migration as `reversible` or `forward-only`, names compatible API/worker versions or schema range, and links the backout/forward-fix procedure. A `.down.sql` file executing successfully on an empty ephemeral database does not prove that destructive production rollback is safe.

The migration command is schema-only. Seeders, credential creation, backfills, and post-migration provisioning run as separately named, idempotent, tested operations with their own deployment evidence. A force-version command is a break-glass recovery tool with explicit operator confirmation and runbook context, not part of normal onboarding or deployment.

For every migration:

- empty database up to latest;
- previous release schema to new schema;
- down then up only when the manifest claims the migration is reversible;
- sqlc generation/compile against resulting schema;
- constraints/index definitions asserted through catalog queries where critical;
- backfill counts/checksums/invariants;
- application N-1 compatibility during expand/contract window;
- representative lock duration for risky DDL;
- no migration modification after production application.

For forward-only or destructive changes, require point-in-time recovery/restore preflight appropriate to the environment, a tested forward-fix, an application rollback path that remains compatible with the expanded schema, and explicit approval before the contract/drop phase.

For data backfills:

- idempotent/resumable batches;
- checkpoint and error table/metric;
- per-tenant fairness if applicable;
- rate control;
- before/after invariants;
- stop/rollback procedure;
- deletion of compatibility columns only after verified adoption.

## 18. Observability plan

### 18.1 Correlation

Propagate W3C trace context through:

- inbound HTTP;
- service/repository spans;
- provider client calls;
- inbox/outbox and Asynq task envelopes;
- worker processing;
- outbound webhooks.

Use real service name, version/commit, environment, instance, and deployment metadata rather than hard-coded development values.

### 18.2 Structured logs

Standard safe fields:

```text
request_id, trace_id, operation, route, actor_kind, actor_id,
workspace_id, resource_type, resource_id, credential_id,
provider, installation_id, delivery_id, task_id, attempt,
duration_ms, outcome, error_code
```

Hash or omit sensitive/high-cardinality external values when needed. Never log tokens, OTPs, invitation secrets, raw authorization headers, full signed URLs, provider payloads, email bodies, or unredacted AI prompts.

### 18.3 Metrics

At minimum:

- HTTP rate, latency, status/error code, body rejection, and rate limiting;
- database pool saturation, query duration/fingerprint, timeout, transaction outcome;
- auth failure by safe reason, token expiry/revoke, authorization denial;
- webhook received/duplicate/replay/invalid/queue age/retry/terminal;
- outbox backlog/oldest age/delivery latency;
- Asynq queue depth/age/retry/dead letter by type/provider;
- credential refresh success/failure/race/revoke;
- Redis consumer lag/reclaim/failure;
- SSE connection count/duration/error;
- external provider latency/rate-limit/5xx;
- API version/client and deprecation usage.

Avoid unbounded labels such as raw workspace/resource ID in metrics; those belong in traces/logs.

### 18.4 Readiness and health

Liveness means the process can run. Readiness means it can safely serve its role. Include, as appropriate:

- running binary declares and satisfies a compatible schema range; readiness does not require exact latest-version equality during expand/contract rollout;
- PostgreSQL reachable and pool usable;
- Redis reachable for routes requiring it;
- consumer/scheduler supervision healthy;
- worker can reserve queues;
- required configuration/keys loaded;
- shutdown state not draining.

Dependency degradation policy must distinguish fail-open optional features from fail-closed security/state operations.

## 19. Initial performance and reliability budgets

These are **provisional targets to ratify after production baselining**, not claims about current performance:

| Signal                        | Initial target                                                                                                                    |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Required PR pipeline          | ≤10 minutes; fast hygiene/unit stage ≤3 minutes                                                                                   |
| Non-AI API reads              | p95 ≤250 ms under representative load                                                                                             |
| Non-AI API writes             | p95 ≤500 ms under representative load                                                                                             |
| API 5xx                       | <0.5% over the user-facing window; attribute dependency-caused failures separately without excluding them from the SLO            |
| Critical queue start delay    | p95 <5 seconds                                                                                                                    |
| Default queue start delay     | p95 <30 seconds                                                                                                                   |
| Readiness after process start | <15 seconds under healthy dependencies                                                                                            |
| Graceful shutdown             | <25 seconds with zero lost work; unfinished tasks are durably requeued/reclaimable and external effects are idempotently recorded |
| Benchmark regression          | investigate/block statistically significant >10% on governed cases                                                                |
| Thirty-minute soak            | <10% unexplained heap/goroutine growth after warm-up                                                                              |

Per-operation SLOs should be derived from actual user journeys. AI/provider latency needs separate budgets and dependency attribution.

## 20. Lifecycle and recovery work

### API process

- one root cancelable context and supervisor;
- HTTP, SSE, Redis consumer, and other long-lived components register with it;
- stop readiness before drain;
- avoid raw socket hijacking where it prevents normal shutdown tracking, or explicitly track/halt connections;
- consumer group creation failure affects health appropriately;
- no fast `NOGROUP` loop;
- wait and report all terminal errors;
- flush telemetry and close dependencies once.

### Worker process

- supervise Asynq server, scheduler, monitoring endpoint, and telemetry;
- authenticated/private monitoring with HTTP timeouts;
- startup failure is fatal when the worker cannot perform its role;
- versioned typed task envelope;
- default timeout/retry/retention/uniqueness policy with task overrides;
- readiness/health endpoint;
- graceful stop and active-task deadline behavior;
- provider-aware backoff and tenant fairness.

### Recovery exercises

- PostgreSQL restart during read/write/transaction;
- Redis restart and stream/queue reclaim;
- worker crash after external effect but before acknowledgement;
- duplicate provider delivery;
- provider 429/5xx outage;
- encryption key rotation/revoke;
- failed migration and API rollback;
- stuck outbox/dead-letter redelivery;
- membership revoke during an active session/API key request.

## 21. Internal documentation architecture

Create a concise entry point under `apps/server/docs`:

```text
apps/server/docs/
├── README.md
├── architecture/
│   ├── overview.md
│   ├── module-standard.md
│   ├── request-lifecycle.md
│   ├── worker-lifecycle.md
│   ├── authorization.md
│   └── integrations.md
├── standards/
│   ├── data-access-sqlc.md
│   ├── migrations.md
│   ├── testing.md
│   ├── errors-and-http.md
│   ├── configuration.md
│   ├── observability.md
│   └── security-and-secrets.md
├── runbooks/
│   ├── deploy-and-rollback.md
│   ├── migrations-and-backfills.md
│   ├── provider-credential-rotation.md
│   ├── webhook-redelivery.md
│   ├── queue-recovery.md
│   └── security-incident.md
└── adr/
```

The root server README should become a fast setup/navigation document and link into these pages rather than growing into an encyclopedia.

The configuration guide must be generated or mechanically checked against typed API and worker configuration. It lists every environment key, owning process, type, required/optional status, safe default, deployment-mode rule, reload/restart behavior, and secret classification. CI compares it and `.env.example` with both process configs so the current 92-versus-71 key drift and API/worker default differences cannot recur. Startup validates the same schema and fails clearly on missing or unsafe production values.

### 21.1 Architecture overview

Explain:

- process topology (API, worker, PostgreSQL, Redis, provider APIs);
- modular-monolith dependency direction;
- actor/authz lifecycle;
- request → service → sqlc → outbox → worker flow;
- where to find routes, use cases, SQL, tests, and docs;
- first-party adapter versus public integration extension.

### 21.2 Module standard

Provide one real, maintained example module and templates for:

- domain command/query/error;
- consumer-owned repository port;
- sqlc query and mapper;
- handler/request/response;
- route registration;
- unit/repository/HTTP tests;
- module bootstrap.

Avoid abstract pseudo-projects that drift from FortyOne conventions.

### 21.3 sqlc guide

Document:

- installation/pinned version and generation commands;
- why FortyOne adopted the useful UUID/time/null patterns from the reference sqlc project but rejected its central generated package and unpinned cleanup workflow;
- schema/query/output organization;
- every non-default config option, global override, nullable/non-nullable pair, and the v1.31 nullable-enum decision;
- the difference between offline migration-driven generation and database-backed `sqlc/db-prepare` vetting;
- query naming and annotations;
- UUID/null/type overrides;
- mapping boundaries;
- tenant-scoping checklist;
- typed filters/sorts/patch alternatives;
- transactions and `WithTx`;
- error mapping;
- integration test and query-plan workflow;
- exception ADR process;
- common errors and remediation.

Complex SQL documentation should state business question, tenant/authorization boundary, inputs, output grain, CTE intent, indexes/plan expectations, null/timezone semantics, and owner. Do not narrate each keyword.

### 21.4 Authorization guide

Include:

- principal/credential/actor definitions;
- session, PAT, service-account, OAuth user/app, system, external contributor flows;
- scope × role × team × resource evaluation;
- revocation/cache behavior;
- audit fields;
- examples of route-only checks that are insufficient;
- test matrix template.

### 21.5 Integration guide

Evolve the existing runtime contract to cover:

- provider registry and capability families;
- install/account-link/grant distinctions;
- credential vault and generation fencing;
- inbound/outbound webhook flow;
- inbox/outbox/idempotency/reconciliation;
- adding a first-party adapter;
- building an external integration through the API;
- sandbox and contract tests;
- provider disconnect/recovery/retention.

## 22. Public API documentation

Create this section in the docs app:

```text
apps/docs/content/docs/api/
├── index.mdx
├── quickstart.mdx
├── authentication.mdx
├── personal-tokens.mdx
├── service-accounts.mdx
├── oauth-apps.mdx
├── scopes-and-permissions.mdx
├── errors.mdx
├── pagination.mdx
├── rate-limits.mdx
├── idempotency.mdx
├── versioning.mdx
├── webhooks/
├── resources/
├── sdks/
└── changelog.mdx
```

### Public-doc standard

Every operation/resource page includes:

- purpose and required scopes;
- path/method and workspace context;
- typed parameters with limits/defaults;
- request and response examples generated/validated against OpenAPI;
- error cases;
- pagination/idempotency/version behavior;
- rate-limit considerations;
- webhook effects;
- cURL plus SDK example;
- “try safely” guidance using non-production credentials where supported.

Never paste real credentials or encourage tokens in query strings. State which credentials represent a user versus an application.

### Tutorials

At minimum:

1. create and rotate a service-account key;
2. list and paginate stories;
3. create a story idempotently;
4. register and verify a webhook;
5. build OAuth authorization code + PKCE flow;
6. process retries and rate limits;
7. build a GitLab-style code-host integration without internal Go access.

Run tutorial snippets in CI or a docs contract test so examples do not decay.

## 23. New-engineer onboarding path

### First hour

- install pinned Go/toolchain and dependencies;
- start PostgreSQL/Redis through one documented command;
- run `make check-fast` and one unit test;
- read architecture overview and trace one request through route, service, query, and test;
- create local data using the supported seeder.

### First day

- run the full local CI-equivalent checks;
- run one PostgreSQL repository test;
- inspect an OpenTelemetry trace;
- make a no-behavior documentation/test improvement;
- learn migration and secret-handling rules.

### First endpoint change tutorial

The maintained exercise should guide the engineer through:

1. update OpenAPI if external;
2. define transport request/response;
3. define domain command/query and policy;
4. add a named SQL query;
5. generate sqlc;
6. map row/domain values;
7. implement service and handler;
8. add unit, repository, HTTP, authorization, and example tests;
9. run formatting/static/generation/migration checks;
10. update docs/changelog.

The tutorial must use actual FortyOne packages and commands.

## 24. Local developer commands

Add portable Make targets or scripts that do not assume tools live under `~/go/bin`:

```text
make bootstrap-tools   install checksum-pinned local tools
make tool-versions     print and assert every supported tool version
make sqlc-generate     intentionally regenerate module-local sqlc packages
make sqlc-check        offline compile + type contract + non-mutating clean drift comparison
make sqlc-vet          migrated-PostgreSQL db-prepare and approved query rules
make generated-check   sqlc + OpenAPI + SDK + config-reference drift
make generate          all intentional code/document generation
make check-fast        format/tidy/vet/unit + offline generated checks
make test              full normal suite
make test-integration  PostgreSQL + Redis/Asynq adapter suite, including sqlc-vet
make test-race         governed race suite
make test-fuzz         bounded local fuzz smoke
make test-system       API + worker system suite
make migration-check   manifest + empty/N-1/N + reversible-only down/up checks
make openapi-check     lint + generated drift + breaking diff
make ci                local required-PR equivalent
```

Pin tool versions and checksums in one checked-in manifest consumed by local bootstrap, CI, and container tooling. Document prerequisites, network behavior, cache location, expected duration, and the safe version-upgrade workflow. Do not install `@latest`, assume `~/go/bin`, parse `.env` with Make's `include`, or rely on an undeclared globally installed binary.

Keep FortyOne in standard Go module mode rather than copying the reference repository's implicit vendoring behavior. If vendoring is later approved for a measured supply-chain or offline-build need, an ADR must add `go mod vendor` drift verification and identical `-mod=vendor` behavior across local development, CI, and container builds.

## 25. Risk register

| Risk                                                 | Likelihood/impact | Mitigation                                                                        | Trigger/rollback                                                              |
| ---------------------------------------------------- | ----------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| sqlc migration changes null/order behavior           | High/high         | characterization + real DB tests + shadow comparison for critical reads           | mismatch telemetry/test; switch caller back within short compatibility window |
| sqlc upgrade changes generated type/signature shape  | Medium/high       | exact pin, config-contract fixture, full regeneration and signature review        | upgrade PR blocked; retain previous pin                                       |
| Tenant scope omitted in typed query                  | Medium/critical   | query templates, policy review, cross-tenant contract suite, compound constraints | security test/audit failure blocks wave                                       |
| Long-running dual persistence paths diverge          | Medium/high       | small vertical slices, deletion in same/next PR, explicit deadline                | owner/escalation if compatibility exceeds deadline                            |
| Migration locks production tables                    | Medium/high       | expand/contract, representative lock tests, resumable backfill                    | abort threshold and DB runbook                                                |
| Actor model breaks first-party clients               | Medium/high       | adapter compatibility, route inventory, staged credential rollout                 | client error/denial metrics; rollback adapter, not new audit data             |
| Credential encryption migration loses access         | Low/critical      | checksum/count verification, dual-read-only short window, rollback key access     | decrypt failure metric; pause scrub/drop                                      |
| Provider abstraction becomes a god interface         | Medium/high       | capability ports and GitLab proof                                                 | design review rejects unrelated/core changes                                  |
| Test suite becomes too slow                          | High/medium       | pyramid, parallel jobs, hermetic fixtures, duration budget                        | profile/split suites; never silently skip required tests                      |
| Coverage target drives low-value tests               | Medium/medium     | risk gates and mutation/failure review                                            | reviewers reject assertion-free/trivial tests                                 |
| Release migration incompatible with worker/API order | Medium/critical   | N-1/N suite, API-before-worker, one-shot migration                                | canary/readiness failure; documented rollback                                 |
| Documentation drifts                                 | High/medium       | generated references, tested examples, docs owner in DoD                          | contract/doc drift CI failure                                                 |
| Existing feature work conflicts with refactor        | High/medium       | module ownership, small PRs, rebase checkpoints, avoid giant rename               | pause module wave; finish/reconcile feature slice                             |

## 26. 10/10 delivery scorecard

| Dimension               | Current directional state             | 10/10 evidence                                                       |
| ----------------------- | ------------------------------------- | -------------------------------------------------------------------- |
| Build correctness       | Build/vet pass, default tests fail    | repeated green deterministic full suite and required CI              |
| Unit behavior           | Strong pockets, uneven                | critical policies/state machines exhaustively covered                |
| Persistence             | Runtime SQLx, weak real-DB suite      | all sqlc queries exercised against migrated PostgreSQL               |
| HTTP/API contract       | Sparse handler/OpenAPI coverage       | every external operation contract-tested; breaking diff gate         |
| Security                | Confirmed critical/high gaps          | threat-model controls plus abuse/regression evidence                 |
| Integration reliability | Strong Slack/messaging pockets        | common contract suite across all provider adapters                   |
| Concurrency             | Focused race pass only                | race + database atomicity + recovery tests for critical flows        |
| Performance             | No meaningful suite                   | representative baselines, budgets, plans, load/soak evidence         |
| CI                      | deploy-only                           | fast required PR gates, nightly depth, immutable validated releases  |
| Migrations              | paired files, not executed in CI      | empty/N-1/up-down-up/lock/backfill validation                        |
| Operability             | partial traces/health, lifecycle gaps | supervised lifecycle, correlated telemetry, actionable SLOs/runbooks |
| Documentation           | basic README/product docs             | current internal architecture + complete public developer docs       |
| Onboarding              | tribal commands/config drift          | documented first-hour/day/change path validated by new engineers     |

## 27. First 31 implementation tickets

These tickets are deliberately ordered; teams can parallelize only when dependencies permit.

| ID  | Ticket                                                                        | Depends on               | Acceptance headline                                                           |
| --- | ----------------------------------------------------------------------------- | ------------------------ | ----------------------------------------------------------------------------- |
| 1   | Re-baseline routes, SQL, tests, credentials, migrations                       | —                        | deterministic inventories committed/generated                                 |
| 2   | Add injected Maya clock and restore green clean suite                         | —                        | repeated `go test ./...` green                                                |
| 3   | Align Go/toolchain docs and generate API/worker configuration reference       | —                        | `go.mod`, AGENTS, README, images, CI, `.env.example`, and config matrix agree |
| 4   | Add required PR hygiene/static/unit workflow                                  | 2, 3                     | protected branch and blocking security policy enforce failures                |
| 5   | Add truthful API/worker readiness, supervised lifecycle, and deploy telemetry | 3, 4                     | release gates observe real health, drain, queue, error, and version signals   |
| 6   | Fix invitation and GitHub admin authorization                                 | 4                        | create/list/revoke and GitHub role matrices green                             |
| 7   | Fix tenant scope for comments/links/private memories                          | 4                        | two-workspace negative suite green                                            |
| 8   | Harden OTP and auth token verification/query-token use                        | 4                        | abuse/token-confusion suite green                                             |
| 9   | Correct Stripe webhook state/retry behavior                                   | 4                        | failure/duplicate/retry suite green                                           |
| 10  | Introduce minimal vault; migrate GitHub/Slack plaintext credentials           | 4                        | envelope/migration proof shows zero prohibited plaintext                      |
| 11  | Introduce shared testkit PostgreSQL/Redis harness                             | 4                        | CI integration job never silently skips                                       |
| 12  | Add migration manifest and empty/N-1/reversible rollback CI                   | 11                       | compatibility and reversible/forward-only policy block unsafe change          |
| 13  | Approve architecture/sqlc/actor/API ADRs                                      | 1                        | coherent signed-off decisions                                                 |
| 14  | Add pinned sqlc config, type contract, and clean generation drift gate        | 12, 13                   | v1.31.1 pilot compiles/vets reproducibly with module-local output             |
| 15  | Add pgxpool platform and tx-bound unit-of-work runner                         | 13, 14                   | single/cross-module rollback and binding tests green                          |
| 16  | Add bounded decoder, typed params, and enforced validation standard           | 4, 13                    | table/fuzz and security-DTO tests green                                       |
| 17  | Expand architecture checks with debt baseline                                 | 1, 13                    | new forbidden dependency, cycle, or SQL fails CI                              |
| 18  | Migrate comments to sqlc                                                      | 7, 14, 15                | SQLx-free + tenant repository tests                                           |
| 19  | Migrate links to sqlc                                                         | 7, 14, 15                | SQLx-free + tenant repository tests                                           |
| 20  | Introduce Actor and policy engine                                             | 6, 8, 13                 | role/scope/team/principal matrix green                                        |
| 21  | Migrate users/workspaces/teams/invitations unit of work                       | 15, 18–20                | no raw tx leak; outbox/rollback green                                         |
| 22  | Add cursor primitive and migrate copied pagination pilots                     | 16                       | tamper/stability/fuzz tests green                                             |
| 23  | Migrate reports/search sqlc projections                                       | Phase 2 exit, 14, 15, 22 | typed filters + plan budgets                                                  |
| 24  | Migrate story read/list slice                                                 | Phase 2 exit, 20, 22, 23 | tenant/team/cursor/plan evidence                                              |
| 25  | Migrate story intent mutations                                                | 24                       | no generic update map; CAS/outbox tests                                       |
| 26  | Complete shared credential vault and rotation generation                      | 10, 13, 20               | refresh/race/revoke/rewrap tests                                              |
| 27  | Generalize capability-aware durable webhook gateway from Slack                | 11, 26                   | signature/delivery/replay/quick-ack contract suite                            |
| 28  | Move GitHub webhook and split code-host capabilities                          | 27                       | async bounded ingress; no SDK leakage                                         |
| 29  | Create API v1 OpenAPI source and initial read resources                       | Phase 2 exit, 20, 22, 24 | strict generated adapter + breaking gate                                      |
| 30  | Add PAT/service accounts/OAuth app grants                                     | 20, 26, 29               | scope/expiry/revoke/reuse/audit suite                                         |
| 31  | Add outbound webhooks, SDK preview, public quickstart                         | 27, 29, 30               | sample external app flow passes system test                                   |

After ticket 31, continue the module waves and operational hardening until no architecture/sqlc debt remains.

## 28. Pull-request definition of done

Every modernization PR states:

- work package/ticket and affected inventory items;
- behavior intentionally preserved or changed;
- security/tenant/actor implications;
- SQL and migration implications;
- test layers run and their results;
- query plan/performance evidence when relevant;
- OpenAPI/docs/changelog effect;
- deployment order, feature flag, compatibility, and rollback when relevant;
- legacy path deleted or named deletion follow-up with deadline.

Reviewers reject:

- generated-code-only “migrations” with old SQL still active;
- new generic maps/query builders for ordinary CRUD;
- interfaces that merely mirror a concrete implementation;
- tests that reproduce production logic;
- security checks only in HTTP middleware;
- network calls inside database transactions;
- public behavior without a contract;
- logs containing secrets or raw provider/customer payloads;
- completion claims without acceptance evidence.

## 29. Final definition of success

The API is a 10/10 engineering system when a new engineer can safely answer and act on all of these without tribal knowledge:

- who is making this request and with what authority;
- which workspace/team/resource policy permits it;
- where the route, use case, SQL, transaction, and event live;
- what compile-time types protect the query;
- what happens on duplicate, retry, conflict, revoke, outage, and shutdown;
- which tests prove those behaviors;
- which metrics and runbooks reveal and recover failures;
- how an internal provider adapter is added;
- how an external developer integrates through documented API/OAuth/webhooks;
- how to deploy, migrate, verify, and roll back the change.

That is the practical meaning of ten out of ten: not architectural fashion, but a codebase whose intent is visible and whose important failures are designed, tested, and operable.
