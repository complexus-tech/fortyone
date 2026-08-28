# Email verification token security

This document is the operational and implementation contract for FortyOne
email verification and one-time login codes. The code is intentionally short
lived and human-readable; its security comes from bounded attempts, atomic
consumption, and keyed storage rather than from treating six digits as a
high-entropy secret.

## Security invariants

- Codes are generated with `crypto/rand`, contain exactly six ASCII digits,
  and expire within 15 minutes. Current handlers issue registration codes for
  10 minutes and authenticated login codes for 5 minutes.
- A newly issued plaintext code exists only long enough to return it to the
  delivery boundary. It is never written to PostgreSQL and is never logged.
- PostgreSQL stores a versioned HMAC-SHA256 digest bound to the normalized
  email, token purpose, and plaintext code. A code for one email or purpose
  cannot be replayed for another.
- The HMAC key is separate from browser-session signing. Production refuses to
  start if `APP_VERIFICATION_TOKEN_HMAC_KEY` is missing, weak, uses the
  development default, or equals `APP_AUTH_SECRET_KEY`.
- Issuance quota checking and insertion run in one PostgreSQL transaction under
  a transaction-scoped advisory lock for the normalized email and purpose.
  The current database-enforced quota is three issues per hour.
- Verification is one atomic database operation: it checks email, approved
  purpose, digest/key/version, expiry, and unused state, then marks the row used
  exactly once. Concurrent consumers cannot both succeed.
- Public send and confirm routes use atomic Redis counters. Cache keys contain
  only a version, non-secret scope metadata, key ID, and an HMAC; raw email
  addresses, codes, and network addresses never enter cache keyspace.
- Redis failure fails closed with `503`. Every successful counter check emits
  `RateLimit-Policy` and `RateLimit` without revealing its quota partition;
  limit exhaustion returns `429` with an authoritative bounded `Retry-After`
  value. Invalid, expired, used, and unknown codes share one public response so
  callers cannot enumerate token state.
- First-party routes accept only the opaque browser session cookie. The removed
  application-secret HS256 user-token fallback cannot be confused with a
  verification code, OAuth access token, PAT, or service-account key. Machine
  credentials are accepted only by the versioned API's dedicated verifier;
  query-string bearer credentials are rejected everywhere.

## Configuration

Generate at least 32 random bytes in the production secret manager and assign
an operational generation ID:

```text
APP_VERIFICATION_TOKEN_HMAC_KEY=<independent random secret>
APP_VERIFICATION_TOKEN_HMAC_KEY_ID=2026-08-v1
```

Do not put the key in source control, logs, dashboards, tickets, or shell
history. Do not reuse `APP_AUTH_SECRET_KEY`. The key ID is not secret and may
be used to identify the generation that produced a digest.

Changing the configured key or key ID invalidates outstanding digest-only
codes; users can safely request a replacement. The service keyring supports a
bounded previous-key overlap, but process configuration currently exposes only
the current generation. Wire a managed previous-key setting before a rotation
that must preserve outstanding codes, then remove it after the maximum token
TTL has elapsed.

## Migration 000152: expand and contract

`000152_harden_verification_tokens.up.sql` is an additive expansion. It makes
the legacy `token` column nullable and adds `token_digest`, `token_key_id`, and
`token_version`, with a storage-shape constraint that permits exactly one of:

1. A pre-deployment legacy plaintext row.
2. A new digest-only row with complete version metadata.

The application dual-reads during the compatibility window but writes only
digest rows. Deploy in this order:

1. Back up the database and apply migration `000152` before the new API binary.
2. Deploy every API instance with the same HMAC key and key ID.
3. Confirm new rows have `token IS NULL` and complete digest metadata.
4. Wait at least the maximum token TTL after the final old API instance stops.
5. Confirm there are no active legacy rows:

   ```sql
   SELECT COUNT(*)
   FROM public.verification_tokens
   WHERE token IS NOT NULL
     AND used_at IS NULL
     AND expires_at > now();
   ```

6. Keep the dual-read until an explicitly reviewed contract migration removes
   the plaintext compatibility path, legacy indexes, and `token` column. The
   existing seven-day purge job bounds retained expired rows.

The down migration cannot reconstruct plaintext from a keyed digest. It first
invalidates digest-only rows with a non-credential sentinel and `used_at`, then
restores the old schema. After rollback, affected users request a fresh code.
This is deliberate fail-closed behavior, not a reversible data conversion.

## Persistence implementation

The users-owned repository uses SQLC v1.31.1 and native pgx behind the narrow
`VerificationTokenRepository` port. Reviewed SQL lives in
`internal/modules/users/repository/queries/verification_tokens.sql`; generated
types remain inside the handwritten adapter. Issuance quota enforcement and
insertion share one repository-owned transaction, while consumption is one
row-locking update statement. The users production persistence layer contains no
SQLx or handwritten Go query strings.

See [`docs/database/users.md`](../database/users.md) for query ownership,
SQLSTATE mapping, tenant scope, and the complete users repository test matrix.

## Verification checklist

- Unit tests cover deterministic generation, digest binding, key rotation,
  unsafe keys/codes, opaque cache keys, strict JWT parsing, and cookie policy.
- HTTP tests cover enumeration-safe failures, send/confirm abuse limits, and
  bounded multipart requests.
- Integration-tag tests apply the real migration chain through
  `internal/testkit`, issue concurrently to prove the database quota, consume
  concurrently to prove single use, verify email/purpose binding, inspect the
  absence of plaintext, and exercise the legacy compatibility read.
- Run integration coverage only against a disposable PostgreSQL control
  database using `TEST_DATABASE_URL`; the test role must have `CREATEDB`.
