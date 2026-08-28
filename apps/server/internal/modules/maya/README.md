# Maya module map

Maya turns workspace data into reviewed work plans, assignments, schedules, and
realtime tool operations. Transport, orchestration, deterministic policy, and
persistence are separate so an engineer can trace one use case without loading
the entire agent stack.

## Current persistence checkpoint

The Maya production path is fully cut over from SQLx and handwritten Go query
strings to pgx and SQLC:

- six files in `repository/queries` define 30 named SQLC operations;
- the generated `repository/sqlc.Querier` exposes the same 30 methods;
- `repository.Repo` is the only Maya PostgreSQL adapter and maps generated rows
  to types from `domain`;
- handwritten production Go under this module has no SQLx import, SQL literal,
  or direct database execution call;
- `internal/taskhandlers/maya.go` and `pkg/jobs/maya_work_focus.go` depend on
  caller-owned interfaces and contain no SQLx or application SQL; and
- API and worker bootstraps construct a pgx-backed Maya repository. The worker
  shares that repository with the Maya service, assignment handler, and
  work-focus job.

The server production path has no SQLx runtime or compatibility view. Each API
or worker process owns one native pgx pool, and the Maya repository reuses that
pool alongside the other module repositories. SQLx is a prohibited production
dependency.

The query files are organized by capability rather than by CRUD verb:

| Query file         | Owned behavior                                                                         |
| ------------------ | -------------------------------------------------------------------------------------- |
| `access.sql`       | Workspace entitlement checks.                                                          |
| `runs.sql`         | Work-plan runs, actions, reads, and guarded action outcomes.                           |
| `scheduling.sql`   | Schedule locks, recovery leases, interrupted-run recovery, ownership, and eligibility. |
| `realtime.sql`     | Voice quota reservations, session lifecycle, and idempotent tool-call claims.          |
| `work_focus.sql`   | Work-focus candidates, evidence, and guarded inferred-role writes.                     |
| `worker_reads.sql` | Bounded assignment and workspace scheduling candidates.                                |

Generated SQLC files are repository implementation details. HTTP, service,
task-handler, and job code must use handwritten Maya types and caller-owned
ports; they must not import `repository/sqlc` or execute SQL directly.

## Where behavior lives

| Area                                 | Primary files                                                                                                              | Responsibility                                                                                           |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| HTTP routes and shared dependencies  | `http/routes.go`, `http/maya.go`                                                                                           | Route registration, handler construction, and transport dependencies.                                    |
| Work-plan HTTP endpoints             | `http/work_plans.go`                                                                                                       | Decode and map manual work-plan requests.                                                                |
| Realtime lifecycle                   | `http/realtime_handlers.go`, `http/realtime_session.go`                                                                    | Start/end sessions, provider client-secret setup, and session instructions.                              |
| Realtime tool dispatch               | `http/realtime_tool_execution.go`, `http/realtime_capabilities.go`                                                         | Validate and execute the supported tool catalog.                                                         |
| Realtime resolution and mapping      | `http/realtime_capability_resolution.go`, `http/realtime_resolution.go`, `http/realtime_mapping.go`                        | Resolve user phrases to tenant-scoped resources and map safe responses.                                  |
| Persistence-neutral types and errors | `domain/persistence.go`, `domain/realtime.go`                                                                              | Run, action, scheduling, work-focus, realtime, and persistence error contracts shared across boundaries. |
| Service composition                  | `service/service.go`, `service/ports.go`, `service/models.go`                                                              | Narrow capability ports and the service dependency boundary.                                             |
| Work-plan orchestration              | `service/work_plans.go`, `service/action_application.go`                                                                   | Create/apply plans, preserve durable mutations, and persist action outcomes.                             |
| Assignment policy                    | `service/assignment_batch.go`, `service/assignment_recommendations.go`                                                     | Candidate selection and batch assignment recommendations.                                                |
| Schedule planning                    | `service/planning.go`, `service/schedule_planning.go`                                                                      | Pure planning rules, capacity windows, focus blocks, and preemption.                                     |
| Durable reconciliation               | `service/reconciliation_flow.go`, `service/reconciliation_plan.go`, `service/reconciliation_transition.go`                 | Recover ownership, converge schedules, and emit meaningful transitions.                                  |
| Repository adapter                   | `repository/maya.go`, `repository/scheduling.go`, `repository/realtime.go`, `repository/worker.go`, `repository/models.go` | Transactions, locks, generated-query calls, validation, and domain row mapping.                          |
| SQLC source and output               | `repository/queries`, `repository/sqlc`                                                                                    | Reviewed application SQL and generated pgx contracts.                                                    |
| Assignment and recovery tasks        | `internal/taskhandlers/maya.go`                                                                                            | Page through typed candidate reads and invoke Maya service use cases.                                    |
| Work-focus inference job             | `pkg/jobs/maya_work_focus.go`                                                                                              | Apply deterministic inference policy through a narrow `MayaWorkFocusStore`.                              |
| Process composition                  | `internal/bootstrap/api/services.go`, `internal/bootstrap/worker/maya.go`, `internal/bootstrap/worker/handlers.go`         | Construct the repository once per process boundary and inject it into services and workers.              |

## Durable mutation rules

These rules are part of Maya's correctness contract and must survive future
refactors:

1. `CreateActions` writes every proposed action in one transaction and preserves
   planner order.
2. Schedule reconciliation acquires the workspace/story advisory transaction
   lock before reading or changing the schedule.
3. Recovery claims are bounded and use `FOR UPDATE ... SKIP LOCKED`; the retry
   watermark advances only for claimed ownership rows, so failed work becomes
   discoverable again after the retry interval.
4. Interrupted-run recovery marks outstanding actions and terminalizes the run
   in the same transaction. Guards prevent a late worker from overwriting the
   recovered terminal state.
5. Realtime voice quota calculation and session creation share a transaction
   protected by a workspace row lock. Tool-call claims use an idempotency key and
   compare request hashes before returning an existing response.
6. `MarkActionApplied` and `MarkActionFailed` require exactly one proposed row.
   A zero-row result is `domain.ErrActionNotProposed`, not success.
7. Service code never swallows action-outcome or run-terminalization failures.
   It returns or joins persistence errors while retaining any mutation that was
   already durably committed. During auto-apply and reconciliation, an outcome
   persistence failure prevents run terminalization, leaving the in-flight run
   available to interrupted-run recovery.

Provider and queue calls do not belong inside repository transactions. Perform
them after commit, fence them with durable state, and make them retryable.

## Testing and verification

The local cutover checkpoint passed:

- `make sqlc-check`;
- focused unit tests across Maya, task handlers, jobs, and API/worker bootstrap;
- focused race tests across Maya, task handlers, and jobs;
- matching focused `go vet`;
- Staticcheck for `./internal/modules/maya/...`; and
- the repository architecture gate and whitespace diff check.

These results apply to the Maya cutover checkpoint, not to subsequent unrelated
working-tree edits or the full branch acceptance suite.

The Maya module currently has 20 test files with 106 `Test`/`Fuzz` functions;
the focused assignment/work-focus edges add two test files with six test
functions. Coverage includes repository transaction order, guarded action
outcomes, scheduling-query contracts, durable service error propagation,
worker paging, and work-focus mapping.

This is not database-backed acceptance evidence. `TEST_DATABASE_URL`,
`SQLC_DATABASE_URL`, and `DATABASE_URL` were unavailable at the checkpoint, so
the tagged PostgreSQL concurrency suite, database-backed SQLC prepare/vet, and
migrated-schema tests remain required. Redis-backed worker recovery also remains
open while `TEST_REDIS_URL` and `REDIS_URL` are unavailable.

## Change workflow

1. Start from the route or worker entry point and identify the service use case.
2. Put deterministic planning or transition policy in a pure function with a
   table test.
3. Pass explicit actor, workspace, team, and resource identifiers into the use
   case. Never recover identity from a repository context.
4. Put static SQL in the capability-specific file under `repository/queries`,
   regenerate SQLC, and map generated values inside the repository.
5. Keep related database invariants in one repository-owned pgx transaction.
   Treat affected-row counts and compare-and-set predicates as domain outcomes,
   not incidental driver details.
6. Add focused unit/race coverage and PostgreSQL integration coverage for any
   persistence, locking, tenant, quota, recovery, or idempotency change.
7. Run `make sqlc-check`, focused tests/race/vet/staticcheck, the architecture
   gate, and diff checks before handing off the slice.

Realtime persistence and its concurrency guarantees are documented in
[`docs/database/maya-realtime.md`](../../../docs/database/maya-realtime.md).
The repository-wide architecture and testing rules remain authoritative in
[`docs/architecture/standards.md`](../../../docs/architecture/standards.md).
