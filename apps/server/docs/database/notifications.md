# Notification persistence and delivery

This guide documents the SQLC/pgx boundary for product notifications. It is
written for engineers who are new to Go, SQLC, or the notification pipeline.

## Data model

```text
users ───────────────┐
workspaces ──────────┼─> notifications <─ unique dedupe_key
workspace_members ──┤         |
team_members ────────┤         +─ read_at       inbox state
resource table ──────┘         +─ email_sent_at delivery state
                               +─ in_app_enabled channel snapshot

users + workspaces ─> notification_preferences (one row per pair)

feedback_portal + feedback_contributor + feedback_item
                  └──────────────────────> portal feedback notifications
```

A notification stores the recipient, workspace, finite notification and entity
types, resource ID, actor, title, structured message, and independent in-app,
read, and email-delivery state. Foreign keys prove that users and workspaces
exist; the SQLC queries prove current authorization to the referenced resource.

## Query ownership

| File                                     | Responsibility                                     |
| ---------------------------------------- | -------------------------------------------------- |
| `repository/queries/create.sql`          | authorized, preference-aware, idempotent create    |
| `repository/queries/inbox_reads.sql`     | workspace inbox, unread count, actor authorization |
| `repository/queries/inbox_mutations.sql` | finite read/unread/delete and bulk mutations       |
| `repository/queries/preferences.sql`     | atomic default creation and typed channel patch    |
| `repository/queries/portal.sql`          | public-portal contributor inbox and read state     |
| `repository/queries/contexts.sql`        | key-result/objective context and audience          |
| `repository/queries/delivery.sql`        | pending email, digest, team scope, sent marker     |

SQL under these files is handwritten and reviewed. `repository/sqlc` is
generated. The adapter accepts and returns domain types, validates enum and
JSON conversion, checks integer narrowing, and maps database classes to domain
errors.

## Create and idempotency

`CreateNotification` is one PostgreSQL statement:

1. resolve the active actor and live workspace;
2. prove the active recipient still has workspace/team/resource access, or is
   an unblocked account contributor to the exact public feedback portal;
3. resolve the recipient's typed in-app preference in the same snapshot;
4. insert with `ON CONFLICT (dedupe_key) DO NOTHING`; and
5. return either the new row or an exactly matching existing row.

The dedupe key must include the immutable source event ID and recipient ID. An
exact replay returns `inserted=false`. Within the same recipient, workspace,
and actor scope, reusing the key with a different type, entity, title, or
message returns `ErrConflict` rather than silently accepting different intent.
A collision owned by another tenant, recipient, or actor is classified as
forbidden so the API does not disclose that another scope already uses the key.

PostgreSQL can observe an uncommitted unique-key conflict while the remainder
of that statement still uses its older snapshot. The adapter performs one
bounded retry only for `pgx.ErrNoRows`; the second statement sees the committed
row and deterministically classifies exact replay, conflict, or forbidden.

## Preferences

Preferences are JSONB because each user/workspace row contains a small finite
matrix of notification types and channels. JSONB does not make the API
untyped:

- `PreferenceType` rejects unknown keys;
- `ChannelPatch` has independent presence-aware email and in-app fields;
- omitted channels remain unchanged;
- missing historical keys are backfilled from `DefaultPreferences`; and
- strategy, reminders, and weekly digest are forced to email-only behavior.

The update query applies `jsonb_set` to the current locked row. Concurrent
patches to different channels therefore merge instead of performing a
read-modify-write in Go and losing one writer.

## Reads and deterministic order

Every workspace read starts from an active user, live workspace, and current
workspace membership. Each notification then rechecks its resource:

- story/comment: live story in the workspace and current story-team access;
- objective/key result: resource in the workspace and current objective-team
  access;
- strategy: workspace entity plus the documented admin/weekly-check-in rule;
- feedback: excluded from the internal inbox and handled by portal queries.

Portal queries require a public portal and an active, unblocked `account`
contributor for that exact portal. Feedback items must still be live and belong
to the notification workspace and portal.

Both inboxes order by `created_at DESC NULLS LAST, notification_id DESC`.
Digest delivery orders oldest first with the notification ID as the tie-breaker
so a retry processes the same sequence. Limits and offsets are checked before
conversion to PostgreSQL `int32`; counts are checked before conversion to Go
`int`.

## Email delivery boundary

`GetEmailDelivery` and `ListEmailDigest` recheck the recipient's active account,
live workspace, current membership/team/resource access, public portal
contributor state, pending read/email state, and typed email preference. A
revoked user disappears from delivery even if a task was already queued.

`MarkEmailSent` is scoped by recipient, workspace, and notification IDs and is
idempotent through `COALESCE(email_sent_at, sent_at)`. It deliberately records
completion after a provider send even if access changes between the final read
and acknowledgement; otherwise a successfully sent email could retry forever.

The task payload never grants access. It contains IDs that are revalidated by
the repository.

## Transactions and failure recovery

Notification creation and its pending-email intent are the same row and commit
atomically. Preference changes and read/unread/delete changes are single static
statements. A statement failure persists no partial state.

Queue and realtime systems are external to PostgreSQL. The create service uses
this recovery sequence:

```text
persist durable row
  -> publish realtime only for a new visible inbox row
  -> enqueue unique recipient/workspace digest wake-up

retry after queue failure
  -> exact dedupe replay returns durable row
  -> no duplicate realtime event
  -> enqueue the digest wake-up again
```

If notification delivery later requires a strict database event history beyond
the row itself, add a notification-owned outbox in the same database statement
or transaction. Do not hold a pgx transaction open across Redis, Asynq, email,
or another network provider.

## Index contract

The migration chain already provides:

- unique `idx_notifications_dedupe_key` for idempotency;
- `idx_notifications_in_app_recipient_workspace_created` for inbox selection;
- recipient/entity created and unread indexes for portal/resource filtering;
- `idx_notifications_pending_email_digest` for pending digest scans; and
- the unique user/workspace preference index.

PostgreSQL 18 integration tests disable sequential scans for focused plan
assertions and prove that the inbox, dedupe, and preference access shapes can
use their intended indexes. Add a new migration only when an actual production
query shape cannot use an adequate index; never retrofit an applied migration.

## Verification

Run from `apps/server`:

```bash
go test -race ./internal/modules/notifications/...

TEST_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/test_control?sslmode=disable' \
  go test -race -tags=integration \
  ./internal/modules/notifications/repository

make sqlc-check

SQLC_DATABASE_URL='postgresql://test_user:fake_password@127.0.0.1:5432/fortyone_sqlc?sslmode=disable' \
  make sqlc-vet
```

Integration tests require disposable PostgreSQL 18 and create/drop an isolated,
fully migrated database through `internal/testkit.NewPostgres`. They cover
tenant isolation, inactive and revoked users, guest/team scope, portal privacy
and blocking, exact and concurrent replay, preference and read-state
concurrency, deterministic equal-timestamp pagination, forced rollback, email
delivery scope, key-result audience scope, and query plans.
