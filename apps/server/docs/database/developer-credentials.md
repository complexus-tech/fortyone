# Developer credential data model and SQLC contract

Migration `000158_developer_credentials` and
`internal/modules/developercredentials/repository` own this storage. Handwritten
repositories consume only the module-local SQLC interface and map generated
rows into domain types before returning. Generated SQLC types must not cross
the repository boundary.

## Relationship map

```text
workspaces ─┬─ workspace_members ─ users
            │          │
            │          └─ human principal (one per workspace/user)
            │
            ├─ service-account principal (guest/member only)
            │          │
            │          └─ api_credentials
            │                ├─ api_credential_scopes
            │                └─ api_credential_team_restrictions ─ teams
            │
            └─ developer_credential_audit_events (immutable UUID facts)
```

`principals` is the durable identity registry. Human principal IDs are database
record IDs, while PAT actors intentionally expose the underlying user ID to
keep `IsUserActor` and `UserID` truthful. Callers needing a principal foreign
key use the service capability described in the security document.

## Table invariants

### `principals`

- Human shape: `subject_user_id` present, `workspace_role` absent, and current
  `(workspace_id, user_id)` membership required.
- Service shape: no subject user and role restricted to `guest` or `member`.
- Identity fields are immutable; disable changes only status and disable facts.
- Deleting a membership cascades its human principal and credentials, making
  removal effective without a cache invalidation path.

### `api_credentials`

- Credential and principal workspace are tied by a composite foreign key.
- A trigger requires PAT-to-human and service-key-to-service-account mapping.
- Credential identity, secret digest metadata, and rotation lineage are
  immutable; lifecycle updates are limited to expiry, usage, rotation time, and
  revocation facts.
- Prefix is globally unique canonical lower-case hex; digest is exactly 32
  bytes; version is explicit; expiry is after creation.
- Rotation is a one-to-one lineage through `rotated_from_id`.
- Revocation is all-or-none for time/reason attribution.
- There is no plaintext column.

### Grants

Scopes are rows rather than arrays so uniqueness, catalog validation, copying,
and authorization queries remain explicit. A trigger prevents
`service_accounts:manage` on service-account keys even if a caller bypasses the
service. Team restrictions use composite credential/workspace and
team/workspace foreign keys, preventing cross-tenant IDs.

### Audit ledger

The audit table intentionally retains identifiers without foreign keys. The
attribution constraint requires human events to have no credential ID and PAT
or service-account events to include one. Update and delete always raise an
error. Application code records only bounded typed metadata, never arbitrary
request JSON.

## Important query semantics

`EnsureHumanPrincipal` is an idempotent insert-or-select. Its partial unique
index makes concurrent first use converge on one row. `ResolveHumanPrincipal`
is read-only and rechecks active user, active principal, and current membership.

Credential creation is a serializable transaction:

1. validate team IDs in the target workspace (and PAT team membership);
2. resolve/provision the correct principal;
3. insert credential metadata and digest;
4. insert normalized scope/team grants;
5. append the immutable audit event;
6. commit or return no credential at all.

Rotation first locks the current credential `FOR UPDATE`, inserts exactly one
replacement, copies normalized grants, marks the old row rotated/revoked or
overlap-expiring, and appends audit. The unique lineage constraint is a second
defense. Concurrent callers produce one replacement; the loser receives a
typed conflict/concurrent-state error.

Authentication is deliberately two-phase. The lookup query filters by
prefix/kind/version and current principal state before HMAC comparison. After
constant-time verification, `ConfirmCredentialActiveAndTouch` repeats active
checks and coalesces `last_used_at` writes to a 15-minute interval. Data-changing
CTE semantics guarantee the touch runs even though the query returns the
eligible credential ID.

## Index intent

| Index                                       | Workload                              |
| ------------------------------------------- | ------------------------------------- |
| unique lookup prefix                        | one-row authentication lookup         |
| principal + descending creation             | credential administration lists       |
| workspace + active expiry                   | workspace lifecycle/revocation scans  |
| workspace service accounts                  | admin account list                    |
| workspace/team/credential restriction       | team-scoped authorization checks      |
| audit workspace/created and subject/created | incident timeline and subject history |

Do not add speculative indexes. Validate new access paths with representative
`EXPLAIN (ANALYZE, BUFFERS)` data and document the read/write tradeoff here.

## Change procedure

1. Add a new migration; never edit `000158` after deployment.
2. Change SQL in `repository/queries`, using `CAST(value AS type)` rather than
   PostgreSQL shorthand in application query files.
3. Run `make sqlc-generate` only after coordinating active config blocks.
4. Map generated values into domain models and fail closed on corrupt enum,
   scope, role, digest, or identity state.
5. Run SQLC compile/check, unit and race tests, plus the tagged isolated
   PostgreSQL suite on PostgreSQL 18.
