# Admin module

The admin module is the internal support surface for inspecting users and
workspaces, extending trials, changing account state, recording private notes,
and requesting a Stripe subscription refresh. It is a platform-administration
surface, not workspace administration and not part of public API v1.

## Where to find things

| Question                                             | Location                                                            |
| ---------------------------------------------------- | ------------------------------------------------------------------- |
| Which routes exist?                                  | `internal/modules/admin/http/routes.go`                             |
| How is HTTP input decoded?                           | `internal/modules/admin/http/requests.go` and the use-case handlers |
| Which commands and finite filter values are allowed? | `internal/modules/admin/domain`                                     |
| Where are use cases orchestrated?                    | `internal/modules/admin/service`                                    |
| Which SQL runs?                                      | `internal/modules/admin/repository/queries`                         |
| Where are SQLC rows converted to domain values?      | `internal/modules/admin/repository/mapping.go`                      |
| Where do transactions and authorization locks live?  | `internal/modules/admin/repository`                                 |
| How is the module constructed?                       | `internal/bootstrap/api/services.go`                                |
| Which database and security rules apply?             | `docs/database/admin.md` and `docs/security/admin-authorization.md` |

Generated files under `repository/sqlc` are adapters. Do not edit them and do
not return their types from the repository.

## Request flow

```text
authenticated HTTP request
        |
        v
strict request/query parsing
        |
        v
typed admin command or query
        |
        v
repository transaction locks and rechecks active internal actor
        |
        +---- read: execute stable SQLC query and map rows
        |
        `---- write: lock target -> change state -> append audit -> commit
```

Middleware proves that a browser credential is valid. It does not prove that
the user is still an internal administrator. Every repository operation checks
the current `users.is_active` and `users.is_internal` values again. This is why
new admin operations must accept an actor ID all the way to persistence.

## Adding an admin operation

1. Add a named command/query and finite values in `domain`; do not accept a
   generic field map or SQL fragment.
2. Add one named SQLC query in the owning repository query file. Select explicit
   columns, use `CAST(value AS type)`, and add a unique ordering tie-breaker.
3. Put a multi-statement invariant in a repository-owned transaction. Lock and
   recheck the actor before reading or changing the target.
4. For a state change, append its immutable `admin_audit_logs` row before the
   transaction commits.
5. Map generated rows in `mapping.go`; keep JSON decoding behavior explicit.
6. Add service validation, strict HTTP mapping, and negative authorization,
   rollback, concurrency, filter, and PostgreSQL query-plan tests.
7. Run the SQLC drift/vet, race, static/security, architecture, and focused
   integration gates documented in `docs/security/quality-gates.md`.

Do not call Stripe or another network provider while a database transaction is
open. The subscription-sync exception is described below and in the database
guide.

## Subscription synchronization

Subscription synchronization is intentionally a three-step workflow:

1. commit an immutable `subscription.sync_requested` audit entry;
2. call the subscriptions capability, which reads Stripe and persists its own
   subscription transaction;
3. commit either `subscription.synced` or `subscription.sync_failed`, linked to
   the request audit ID.

These steps are not one atomic transaction. A process crash or actor revocation
between them can leave a request without a result entry, and Stripe may already
have been read before finalization fails. The request entry is the durable fact
operators use to identify and reconcile that condition. Never describe this
workflow as atomic across Stripe.

The session-revocation endpoint atomically advances the target account's
`auth_session_version` and records the durable audit. Authentication compares
every Redis session record with that PostgreSQL epoch, so all earlier cookies
fail closed immediately without scanning or deleting Redis keys.

Administrator disable and enable transitions also advance the epoch and update
the login-reactivation policy in the same transaction. Disable sets
`admin_only`; enable restores `verified_sign_in`. Ambiguous legacy inactive
accounts remain `legacy_admin_review` until an administrator explicitly
enables them. Neither enable nor ordinary sign-in can resurrect a pre-transition
browser session.
