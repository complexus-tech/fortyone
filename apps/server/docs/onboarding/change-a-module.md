# Changing an API module

Use this checklist for an endpoint, business rule, repository query, or worker
behavior. It explains the current production shape; architecture standards and
the owning module's current database/security documents are the review contract.

## Locate ownership

Start in [`docs/inventory/api.md`](../inventory/api.md), not with a broad search
through the monorepo. A domain module normally owns:

```text
internal/modules/<domain>/
├── http/        routes, protocol DTOs, parsing, response/error mapping
├── domain/      stable values/errors shared by services and adapters
├── service/     commands, queries, policies, and caller-owned ports
├── repository/  pgx/SQLC adapter, mapping, transaction implementation
│   ├── queries/ reviewed named SQL
│   └── sqlc/    generated only
└── worker/      versioned tasks and handlers when the domain owns async work
```

Create files by use case or capability—`oauth_install.go`,
`story_mutation.go`, `delivery_worker.go`—not by arbitrary suffixes such as
`service_part2.go`. A handwritten file approaching the enforced size threshold
is a signal to separate coherent behavior before adding more code.

## Define the use case first

The service contract should make mistakes difficult:

- accept a typed actor and explicit workspace/team/resource identifiers;
- use a command for a mutation and a query for a read;
- distinguish omitted, null, and supplied patch values;
- return domain results and sentinel/typed errors through service contracts;
- inject narrow caller-owned ports for cross-module capabilities;
- keep HTTP, SQLC, pgx, Redis, Asynq, and provider SDK types out of the
  service contract.

Prefer one function whose name describes the user intent over a universal
`Update(id, map[string]any)`. Reusable platform behavior belongs in the existing
decoder, validation, actor/policy, pagination, transaction, idempotency,
credential, or integration packages; similar domain rules stay with the
domain rather than becoming a vague utility package.

## Make authorization explicit

For each operation, write down the matrix before implementation:

| Input             | Questions to answer                                                                                              |
| ----------------- | ---------------------------------------------------------------------------------------------------------------- |
| Principal         | Which human, personal token, service account, OAuth user/app, system, or external contributor kinds are allowed? |
| Credential        | Which resource/action scopes and optional team restrictions apply?                                               |
| Tenant            | Does the actor's selected workspace match the resource workspace?                                                |
| Current authority | What current workspace role and team membership are required?                                                    |
| Resource          | Is ownership, visibility, status, or optimistic version relevant?                                                |

Load role and membership from authoritative current state for sensitive work.
Credential scopes narrow product authority; they never create membership or
ownership. Return the same safe not-found result when exposing existence would
create a cross-tenant oracle.

Tests include an allowed case, each denied principal/role/scope case that
matters, an identifier copied from a second workspace, a removed/demoted member,
and concurrent revoke-versus-write behavior for privileged mutations.

## Add or change SQL

Follow [`docs/database/sqlc.md`](../database/sqlc.md):

1. put named SQL in the owning `repository/queries` directory;
2. list columns explicitly and bind every value;
3. include the tenant predicate in every tenant-owned read and mutation;
4. use stable unique ordering and a bounded cursor for lists;
5. select an annotation whose cardinality/affected-row behavior matches the
   service invariant;
6. generate into `repository/sqlc` and keep generated types inside the adapter;
7. map no-row, constraint, affected-row, and SQLSTATE outcomes to service-owned
   errors;
8. document a complex query's business question, input/output grain, CTE
   purpose, null/timezone semantics, indexes, and plan budget.

Use the shared pgx transaction runner. All statements that preserve one
invariant bind to the same `pgx.Tx`. Network calls, email, provider requests,
and queue publication occur after commit through a durable outbox. SQLx is a
prohibited production dependency. The `database/sql` connection API is reserved
for the golang-migrate driver and must not be injected into repositories.

## Implement bounded background work

For a scheduled or queued use case, define the narrow store capability next to
the job and implement it in the domain-owning repository. Reuse an existing
repository instance at composition boundaries instead of opening another pool or
constructing a parallel adapter.

A recurring scanner or automation must:

- capture one explicit application UTC `as_of` value for the complete run;
- use a stable unique keyset cursor or a transition that removes rows from its
  own eligibility set;
- declare a page size and maximum page count, validate repository result counts,
  and return an explicit backlog result when the bound is exhausted;
- check `context.Context` cancellation between batches and before external work;
- keep a state transition and its required activity, audit, or outbox records in
  one pgx transaction; and
- perform provider calls only after commit, with durable claims, leases,
  idempotency keys, or claim-token fencing appropriate to the side effect.

Story/sprint automation and attachment-object deletion are the concrete
reference implementations in
[`docs/database/stories-mutations.md`](../database/stories-mutations.md).

## Implement the HTTP contract

- Use `web.Decode` for JSON. It rejects wrong media types, unknown fields,
  oversized bodies, trailing values, and structural validation failures.
- Put structural constraints in `validate` tags. Implement `Validate() error`
  only for cross-field or protocol-domain rules.
- Use `web.ParseMultipartForm` with an explicit hard limit for multipart input.
- Use typed path/query primitives and reject zero UUIDs, unsupported enums,
  unstable sorts, and unbounded limits.
- Map errors to the shared envelope. Field violations never echo the rejected
  value; internal database/provider errors never enter the response.
- Keep raw signed webhook intake on its provider-specific bounded verifier path.
- Use the shared idempotency receipt service for externally retried writes in
  addition to domain uniqueness constraints.

## Test by risk

Minimum layers for a persisted authenticated mutation are:

- pure unit tests for validation, policy, mapping, and state transitions;
- adapter tests for SQLC parameter and error mapping;
- real PostgreSQL integration tests for tenant scope, transaction rollback,
  null/edge values, and relevant concurrency;
- HTTP tests for body/parameter parsing, actor/role/scope behavior, status,
  error code, request ID, and sensitive-field absence;
- worker/provider contract tests when the mutation emits async work.

Do not hide infrastructure absence with `t.Skip`, use arbitrary sleeps, share
mutable tenant fixtures across parallel tests, or make live provider sandboxes a
required deterministic test.

## Update discoverability

After moving a route, query, test, config value, or migration:

```bash
make sqlc-generate      # only when SQLC source changed
make config-generate    # only when config source changed
make inventory-generate
make migration-docs     # only when post-baseline migrations changed
make generated-check
```

Update the owning architecture/security/operations document and any external
OpenAPI contract in the same change. The inventory must make the route-to-use-
case-to-query path findable without relying on the original author.
