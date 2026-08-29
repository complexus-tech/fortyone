# Workspace invitation token security

This document is the implementation and deployment contract for workspace
invitation bearer links. Migration `000155_workspace_invitation_token_digests`
is an expand migration: it preserves bounded reads of pre-migration plaintext
tokens while every new invitation is written without a plaintext bearer.

## Security invariants

- Issuance uses `crypto/rand` to generate 32 bytes of entropy. The external
  token has the versioned form `wi1.<key-id>.<nonce>.<signature>`, using raw
  URL-safe base64 for the nonce and HMAC-SHA256 signature.
- Signature and persistence digest use different domain-separation labels. A
  token is authenticated locally before a fixed-size database lookup occurs.
  Unknown key IDs, invalid base64, wrong lengths, unsupported versions, and
  invalid signatures share the same public not-found behavior.
- PostgreSQL stores only a 32-byte keyed digest, the non-secret random nonce,
  key ID, and protocol version for new invitations. `token` is always `NULL`
  for new writes. The database alone cannot recover a bearer; compromise of
  both the database and the configured HMAC key remains a live-token
  compromise and requires key rotation and invitation invalidation.
- The raw token exists at issuance only in memory, is discarded before the
  repository call, and is reconstructed only by the worker immediately before
  rendering the invitation email. It is not stored in an outbox payload and is
  not placed on Redis or another general-purpose event transport.
- Request logging and HTTP tracing record the matched route template, never the
  concrete invitation-token path or query string. Application logs, errors,
  metrics, and tracing attributes must contain only safe IDs, key IDs, versions,
  counts, and outcomes.
- Creation, superseding an older invitation, team assignment, and the token-free
  email outbox record commit in one pgx transaction. Acceptance locks the
  invitation row and atomically validates the active recipient, creates
  workspace/team memberships, updates the user's last workspace, consumes the
  invitation, and records the acceptance notification.
- Invitation team IDs are constrained to the invitation workspace in SQL.
  Revocation includes both workspace and invitation ID. A recipient email
  mismatch returns the same not-found result as a missing token.
- Acceptance is single-use under concurrency. Expired, used, revoked, malformed,
  unknown, and wrong-recipient credentials never create partial membership.
- The worker claims outbox rows with `FOR UPDATE SKIP LOCKED`, reclaims stale
  leases, fences completion/retry with a random claim token, and applies bounded
  retry with a terminal failure state. Invitation email delivery rechecks
  expiry/use state before reconstructing the bearer.
- Notification email uses a deterministic, digest-derived `Message-ID` from the
  durable outbox idempotency key. Delivery is at-least-once: the stable message
  identity gives SMTP receivers a deduplication signal, but operators must still
  treat a duplicate email after a send/ack crash window as possible.
- Creation and acceptance have authenticated, per-user Redis rate limits.
  Bulk requests are capped at 50 normalized, unique recipient emails and
  duplicate team IDs are removed before persistence.

## Configuration and key separation

The API issues and verifies tokens. The worker reconstructs outstanding tokens
for email delivery. Both processes therefore need the same current and previous
keyring:

```text
APP_INVITATION_TOKEN_HMAC_KEY_ID=2026-08-v1
APP_INVITATION_TOKEN_HMAC_KEY=<independent secret containing at least 32 random bytes>
APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS=
```

Previous keys use comma-separated `key-id=secret` entries. Key IDs contain only
letters, numbers, `_`, and `-`, with a maximum length of 64 characters.
Production startup rejects missing/weak/default keys, duplicate key IDs, and
reuse with browser-session authentication or provider-vault key material. Keep
key values in the managed secret store, never source control, shell history,
tickets, logs, or dashboards.

Rotation is read-old/write-new:

1. Add the next key as a previous-capable generation to every API and worker
   task while the old key remains current.
2. Verify every task can start with both generations.
3. Change the current ID/key everywhere. New invitations immediately use it;
   outstanding old invitations remain readable and deliverable.
4. Keep the old entry for at least the seven-day invitation lifetime after the
   last old-key issuance, and until the active-row query for that key returns
   zero. Then remove it from API and worker configuration together.

Removing a key early makes its outstanding emails impossible to reconstruct and
its presented tokens impossible to verify. Recover by restoring the exact key
generation or revoking and reissuing affected invitations; never replace a key
under an existing key ID.

## Migration 000155 schema

The workspace invitation storage constraint permits exactly one shape:

1. Legacy compatibility: plaintext `token` is present and every digest metadata
   column is `NULL`.
2. Current storage: plaintext `token` is `NULL`; digest and nonce are 32 bytes;
   key ID is non-empty; version is positive.

The partial unique index on `(token_key_id, token_version, token_digest)` keeps
the fixed-size lookup unique. The old plaintext unique index remains only for
legacy compatibility and should be removed by the later contract migration.

`workspace_invitation_outbox` stores token-free JSON snapshots and foreign keys
to the owning invitation, workspace, and actor. A database check explicitly
rejects a top-level `token` field. Lifecycle constraints reject contradictory
claim/completion state, and partial indexes support ready, stale-claim, and
retention scans.

## Expand and controlled rollout

Migration `000155` is additive, but invitation routes are not safely served by
mixed application versions after the first digest-only write. An old API can
continue reading and writing legacy rows after the expansion, but it cannot
accept a new digest-only invitation. An old worker also does not dispatch the
new database outbox. Use a controlled route cutover:

1. Back up PostgreSQL and provision one independent invitation HMAC keyring in
   every replacement API and worker task. Confirm API and worker configuration
   are identical without printing key values.
2. Apply migrations through `000155` before deploying a new binary. The schema
   remains compatible with existing legacy invitation rows.
3. Temporarily pause invitation create/get/accept/revoke traffic, or route all
   invitation paths exclusively to the replacement pool. Other API routes may
   continue a rolling deployment, but no invitation request may reach an old
   instance after a digest-only row can be created.
4. Start at least one replacement worker and verify its readiness, invitation
   outbox schedule registration, PostgreSQL connectivity, SMTP connectivity,
   and keyring. Starting it before the new API is safe because no new rows exist.
5. Replace every invitation-serving API instance, then reopen invitation routes.
   Do not restore an old API instance to that route pool.
6. Create and accept one canary invitation. Verify the email arrives, membership
   and team assignment are correct, and both outbox rows reach `completed`.
7. Monitor oldest ready/processing row age, terminal failures, retry counts,
   invitation error rates, SMTP errors, and unknown-key/invalid-metadata counts.
   Never log the canary token.

Post-deployment storage checks:

```sql
-- New writes must not use plaintext. Replace the timestamp with the rollout start.
SELECT COUNT(*) AS new_plaintext_rows
FROM public.workspace_invitations
WHERE created_at >= TIMESTAMPTZ '2026-08-28 00:00:00+00'
  AND token IS NOT NULL;

-- This must always be zero; the CHECK constraint is the primary enforcement.
SELECT COUNT(*) AS incomplete_digest_rows
FROM public.workspace_invitations
WHERE token IS NULL
  AND (
    octet_length(token_digest) <> 32
    OR octet_length(token_nonce) <> 32
    OR NULLIF(token_key_id, '') IS NULL
    OR token_version IS NULL
    OR token_version <= 0
  );

SELECT status, COUNT(*) AS rows, MIN(created_at) AS oldest
FROM public.workspace_invitation_outbox
GROUP BY status
ORDER BY status;
```

## Legacy compatibility and contract cleanup

The application deliberately dual-reads old and new token shapes but writes
only the new shape. Legacy parsing accepts only the historical padded URL-safe
base64 encoding of exactly 32 bytes; arbitrary plaintext is not treated as a
legacy credential.

Wait at least seven days after the final old API issuance, then verify there are
no active legacy invitations:

```sql
SELECT COUNT(*) AS active_legacy_invitations
FROM public.workspace_invitations
WHERE token IS NOT NULL
  AND used_at IS NULL
  AND expires_at > now();

SELECT token_key_id, token_version, COUNT(*) AS active_invitations
FROM public.workspace_invitations
WHERE token_digest IS NOT NULL
  AND used_at IS NULL
  AND expires_at > now()
GROUP BY token_key_id, token_version
ORDER BY token_key_id, token_version;
```

Do not edit migration `000155`. Create a separately reviewed contract migration
that, in this order:

1. Refuses to proceed while an active legacy row exists.
2. Deletes or otherwise applies the approved retention policy to expired/used
   legacy rows; it must not invent digest metadata for unknown plaintext tokens.
3. Drops the storage-shape check and old `workspace_invitations_token_key`
   plaintext index.
4. Drops `token`, makes digest/nonce/key/version non-null for retained rows, and
   installs the digest-only shape constraint.
5. Removes the legacy SQLC predicate, legacy parser branch, and compatibility
   tests in the same release.

Outbox retention must also be explicit: retain failed rows for incident review,
alert before pruning them, and delete completed/cancelled rows only after the
operations retention window.

## Rollback and forward-fix

Before any digest-only invitation is issued, the down migration can remove the
expansion and restore the old non-null plaintext column. Once a digest-only row
exists, the down migration intentionally raises an exception because it cannot
reconstruct plaintext and would destroy the only lookup material.

After new writes, recovery is forward-only:

1. Pause invitation routes and leave the replacement worker stopped only if its
   failure mode could send incorrect mail. Preserve the complete keyring.
2. Diagnose with safe IDs/status/counts. Do not copy token, nonce, digest, or key
   material into incident channels.
3. Deploy a forward code/schema fix and resume the worker before reopening
   invitation routes.
4. If business recovery requires invalidating digest-only invitations, make that
   an explicit, audited operator decision and reissue them after the fix. Only
   after no digest-only rows remain can an application/schema rollback be
   considered.

The forward-only guard is a safety control, not an inconvenience to bypass with
`migrate force`.

## Verification commands

Run from `apps/server`:

```bash
make sqlc-generate
make sqlc-check

# Against a disposable database migrated exactly to repository head.
SQLC_DATABASE_URL='postgresql://<disposable-role>@<host>/<database>?sslmode=disable' \
  make sqlc-vet

go test -race ./internal/modules/invitations/... \
  ./internal/bootstrap/worker ./internal/taskhandlers \
  ./internal/platform/http/middleware ./pkg/web

TEST_DATABASE_URL='postgresql://<disposable-createdb-role>@<host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration -count=1 \
  ./internal/modules/invitations/repository

go vet ./internal/modules/invitations/... ./internal/bootstrap/worker \
  ./internal/taskhandlers ./internal/platform/http/middleware ./pkg/web
go test ./internal/bootstrap/architecture
```

The tagged tests apply the real migration chain to isolated PostgreSQL 18
databases and prove digest-only storage, token-free outbox payloads, bounded
legacy reads, transaction rollback on cross-tenant team IDs, scoped revocation,
wrong-recipient non-enumeration, and concurrent single-use acceptance.
