# Admin authorization and audit

## Trust boundary

An authenticated FortyOne user is not automatically a platform administrator.
The admin module permits an operation only when the current `users` row for the
actor has both:

- `is_active = true`; and
- `is_internal = true`.

The repository checks those fields while executing the operation. A prior
middleware or service lookup is insufficient because another administrator can
revoke access between that lookup and the protected query.

Missing, inactive, and non-internal actors all return the same forbidden domain
error. This avoids revealing whether an arbitrary user ID exists.

## Locking and revocation

Reads take a shared row lock on the authorized actor for their short repository
transaction. Mutations take compatible actor/target locks and commit their audit
entry with the state change. Consequently:

- a request beginning after revocation is denied;
- revocation waits for an already-running authorized database operation;
- a user cannot demote or deactivate themselves between a service check and an
  update;
- two administrators attempting to change each other are locked in a stable
  UUID order rather than deadlocking.

This guarantee covers FortyOne database work. It does not make a Stripe network
call cancellable after its request audit has committed.

## Mutation audit requirements

Every state-changing admin command records:

- actor user ID;
- target type and ID;
- workspace ID when applicable;
- finite action and field names;
- JSON old/new values;
- the required human reason;
- bounded, non-secret metadata;
- database creation time.

Audit metadata may include display names, slugs, subscription identifiers, or a
linked request audit ID. It must not include access tokens, provider secrets,
HTTP authorization headers, raw Stripe objects, or raw provider errors.

Admin notes are private support data. Their body is intentionally present in the
associated audit value, so access to admin audit exports must be treated as
sensitive internal access.

## Self-mutation policy

The existing API rejects every user-state patch where actor and target are the
same. The repository repeats that decision after locking the participating user
rows. This stricter rule includes self-demotion and self-deactivation and
prevents request races from bypassing it.

## Negative test matrix

Every new admin operation must prove at least:

| Actor state                     | Read                                   | Write                                  |
| ------------------------------- | -------------------------------------- | -------------------------------------- |
| active + internal               | allowed                                | allowed when command is valid          |
| inactive + internal             | forbidden                              | forbidden                              |
| active + non-internal           | forbidden                              | forbidden                              |
| missing                         | forbidden                              | forbidden                              |
| revoked while waiting on a lock | forbidden after the revocation commits | forbidden after the revocation commits |

Mutation tests also prove self-mutation rejection, audit-failure rollback,
concurrent update behavior, and that no network operation occurs inside an
admin database transaction.
