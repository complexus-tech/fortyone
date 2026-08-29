# Calendar module map

The Calendar module owns account connections, provider synchronization,
availability, schedule blocks, and the durable provider-write outbox. Google
and Microsoft adapters share FortyOne capability contracts while retaining
their native token, watch, payload, and error semantics.

## Where behavior lives

| Area                  | Primary files                                                     | Responsibility                                                                                                            |
| --------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Service boundary      | `service/calendar.go`, `service/provider.go`, `service/models.go` | Errors, narrow ports, configuration, and shared domain models.                                                            |
| Connection lifecycle  | `service/connections.go`                                          | OAuth connection, callback, primary connection, revoke, and notification ingress.                                         |
| Calendar reads        | `service/schedule_reads.go`                                       | Availability, calendar view, and schedule preference queries.                                                             |
| Schedule mutations    | `service/schedule_mutations.go`                                   | Manual blocks and Maya reconciliation entry points.                                                                       |
| Provider-write outbox | `service/schedule_outbox.go`                                      | Locked dispatch, retry classification, cleanup, and terminal failure handling.                                            |
| Synchronization       | `service/synchronization.go`                                      | Full/incremental sync, watches, generation fencing, and sync failure state.                                               |
| Credential helpers    | `service/credentials.go`                                          | Signed OAuth state, encrypted token payloads, provider selection, and token refresh.                                      |
| Validation            | `service/schedule_validation.go`                                  | Time-range and schedule-block normalization.                                                                              |
| Google adapter        | `service/google.go`, `service/google_mapping.go`                  | Google API operations and provider-to-domain mapping.                                                                     |
| Microsoft adapter     | `service/microsoft.go`                                            | Microsoft Graph operations and provider-to-domain mapping.                                                                |
| Persistence           | `repository`, `repository/queries`, `repository/sqlc`             | Native pgx/SQLC connection, event, schedule, watch, and outbox persistence; generated types remain inside the repository. |

Provider SDK values stop at the provider adapter. Business rules consume the
provider-neutral interfaces in `service/provider.go`; adding another calendar
provider should not duplicate connection, scheduling, outbox, or authorization
logic.

The calendar adapter reuses the process-owned native pgx pool. SQLx is a
prohibited production dependency; static SQL belongs in the module-owned query
directory and generated code is wrapped by the handwritten repository.

## Invariants

- OAuth state is signed, short-lived, actor-bound, and single-use where the
  provider flow requires replay protection.
- Connection credential generations fence refresh, sync, watch replacement,
  disconnect, and delayed worker operations.
- A schedule change and its provider-write outbox record are committed
  together. External provider calls never run inside a database transaction.
- Provider-write retries are idempotent. Permanent authorization or validation
  failures are classified separately from rate limits and transient failures.
- Reads and mutations bind the current actor, workspace, account owner, and
  story; a UUID alone never grants access.
- Token payloads, OAuth state, provider bodies, and webhook proofs are never
  logged or returned in operational errors.

## Change workflow

Run focused service/provider tests with the race detector. Repository changes
also require PostgreSQL 18 integration tests for tenant isolation, credential
generation races, transaction rollback, outbox replay, and stable query plans.
Static SQL belongs in a module-owned `repository/queries` directory and generated
SQLC values must remain inside the repository adapter.

The shared provider model and adoption checklist are documented in
[`docs/integrations/providers.md`](../../../docs/integrations/providers.md).
Repository-wide rules are in
[`docs/architecture/standards.md`](../../../docs/architecture/standards.md).
