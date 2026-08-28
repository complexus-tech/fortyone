# ADR 0007: Machine credentials and provider secrets

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and security

## Context

External developers and provider adapters require credentials with different
security properties. API keys need comparison but not recovery; provider OAuth
tokens must be recovered to call the provider. Plaintext storage, shared master
secrets without key identity, and non-expiring broad credentials are not
acceptable.

## Decision

- Personal access/API keys are high-entropy, shown once, and stored as a
  versioned keyed digest plus a non-secret lookup prefix. They have explicit
  scopes, owner, workspace constraints, expiry, last-used metadata, and revoke time.
- Service accounts are first-class principals. Their credentials follow the
  same one-time display, scope, expiry, rotation, and revocation model.
- OAuth authorization codes are short-lived and single-use. Refresh-token
  families rotate on every use; reuse revokes the family and emits an audit event.
- Recoverable provider tokens use the shared envelope vault: an authenticated
  ciphertext with key ID/version and AAD binding provider, owner, workspace, and
  credential identity. Key rotation supports read-old/write-new and audited rewrap.
- Raw credentials exist only at generation or provider-call boundaries. They do
  not enter logs, traces, analytics, URLs, events, or general domain models.

Production startup fails if required keys are absent, weak, reused across
purposes, or reference an unknown active key ID.

## Enforcement and adoption

- Secret scanning plus tests for plaintext absence in tables, logs, errors, and events.
- Tests cover scope, expiry, revoke, concurrent rotate/refresh, refresh reuse,
  unknown key version, AAD mismatch, and rewrap.
- Provider rewrap scans are bounded and idempotent. Persistence compares the
  exact original envelope and provider generation, so refresh and revoke win.
- The controlled procedure, completion proof, and rollback are maintained in
  [`docs/operations/provider-credential-rotation.md`](../../operations/provider-credential-rotation.md).
- Migrations are expand/backfill/read-old-write-new/contract; destructive
  plaintext removal is verified before contract.

## Consequences

A keyring and rotation procedure become production dependencies. This cost buys
compartmentalization, revocation, and safe provider access without plaintext at rest.

## Revisit when

A managed KMS/HSM can replace local key wrapping while preserving the same
versioned envelope and testable lifecycle contract.
