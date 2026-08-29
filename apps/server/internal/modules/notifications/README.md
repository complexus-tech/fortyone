# Notifications module

This module turns internal product events into durable, tenant-scoped
notifications. It owns the workspace inbox, public-feedback portal inbox,
channel preferences, email-delivery reads, and the key-result audience lookup
used by the event consumer.

## Where to find things

| Question                                           | Location                                                                                 |
| -------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Which notification/entity/preference values exist? | `domain/types.go` and `domain/preferences.go`                                            |
| Which commands carry actor and tenant scope?       | `domain/commands.go` and `domain/delivery.go`                                            |
| Where are use cases and caller-owned ports?        | `service`                                                                                |
| Where are HTTP routes and response models?         | `http/routes.go` and the focused handler files in `http`                                 |
| Which reviewed SQL runs?                           | `repository/queries`                                                                     |
| Where do SQLC rows become domain models?           | `repository/mapping.go` and the focused repository files                                 |
| Which files are generated?                         | `repository/sqlc`; never edit or import these above the repository                       |
| Where is the module assembled?                     | `internal/bootstrap/api/services.go` and `internal/bootstrap/worker/handlers.go`         |
| Which non-HTTP callers use it?                     | `internal/eventconsumer` and the focused `internal/taskhandlers/notification_*.go` files |
| Which database and security rules apply?           | `docs/database/notifications.md` and `docs/security/notification-authorization.md`       |

## Dependency flow

```text
product event consumer                 authenticated HTTP request
          |                                      |
          v                                      v
notification rules                    notification HTTP handler
          |                                      |
          +------------------+-------------------+
                             v
                    notifications service
                    (caller-owned ports)
                             |
                             v
                     pgx/SQLC repository
                             |
                             v
                         PostgreSQL

email task handler -> notifications delivery port -> the same repository
```

The consumer and email task handler do not own notification SQL or a raw
database handle. `ListKeyResultAudience`, `GetEmailDelivery`,
`ListEmailDigest`, `ListDeliveryTeamIDs`, and `MarkEmailSent` are deliberately
narrow use cases. This keeps event and job code independent of the notification
schema and makes another delivery adapter possible without copying SQL.

## Adding a notification type

1. Add the PostgreSQL enum value in a new migration. Never edit an applied
   migration.
2. Add the finite domain constant, parser support, entity compatibility, and
   preference default. Email-only categories belong in `PreferenceType`; they
   do not automatically belong in the persisted `NotificationType` enum.
3. Add or update a rule that produces `NewNotification` with a stable event and
   recipient-specific dedupe key.
4. Extend every resource-authorization branch that must recognize the new
   entity. Creation, inbox reads/mutations, portal reads/mutations, and email
   delivery must agree.
5. Add typed SQLC queries or parameters. Do not accept a SQL fragment, a
   free-form sort identifier, or `map[string]any` update data.
6. Add domain, service, HTTP, PostgreSQL 18 tenant-negative, concurrency,
   rollback, and plan tests.
7. Update the database and authorization guides before running the normal
   SQLC, race, static/security, and architecture gates.

## Dedupe and delivery semantics

`dedupe_key` identifies one logical event for one recipient. An exact replay
returns the original notification and does not change its read or email state.
The same key with different typed content in the same recipient/workspace/actor
scope is a conflict; a collision from another scope is forbidden to avoid
cross-tenant disclosure. Concurrent exact replays use a bounded fresh-snapshot
retry so PostgreSQL's conflict visibility cannot turn a valid replay into a
false conflict.

The notification row is also the durable pending-email intent:
`email_sent_at IS NULL` means the worker can still deliver it. The service
enqueues a unique recipient/workspace digest wake-up after persistence. If the
queue write fails, the caller receives an error; replaying the same event finds
the durable row and re-enqueues the wake-up. Realtime publication is only for a
new, in-app-enabled, non-feedback row and is not the source of truth.

The database row and an external Redis/Asynq operation cannot share a
transaction. Do not describe that boundary as globally atomic. The durable row
plus idempotent replay is the recovery contract.

## Pagination and privacy

Workspace and portal lists use bounded look-ahead pagination. SQL orders by
`created_at DESC NULLS LAST, notification_id DESC`, so equal timestamps remain
deterministic. The workspace endpoint retains legacy `limit`/`offset` input for
client compatibility while routing bounds through the shared pagination
primitive.

Strategy notification snapshots can contain internal planning detail. Domain
and HTTP mapping call `Public()` before returning them, replacing the rich
snapshot with a safe message. Do not return a repository or generated SQLC row
directly from a transport.
