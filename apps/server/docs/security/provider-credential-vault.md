# Provider credential vault

The shared credential vault is the at-rest boundary for recoverable GitHub,
Slack, and Figma OAuth credentials. New provider credential writes persist a versioned
authenticated envelope; they never persist a plaintext fallback. Migration
`000154_provider_credential_vault` adds the GitHub envelope metadata needed to
bind and safely backfill existing tokens. Slack reuses its existing credential
payload, version, and installation-generation columns. Pending Slack uninstall
outbox credentials are migrated with their owning installation context before
the worker can claim them. Migration
`000169_figma_vault_installation_generation` adds Figma's envelope version and
immutable installation generation before its bounded legacy backfill.

## Security contract

Every call to `Seal` creates a random 256-bit data-encryption key (DEK). The
credential is encrypted with AES-256-GCM and a fresh `crypto/rand` nonce. A
configured 256-bit key-encryption key (KEK) wraps that DEK with a second
AES-256-GCM operation and independent nonce. The stored `vault.v2` envelope
identifies the algorithm, AAD format, KEK ID/version, nonces, ciphertext, and
wrapped DEK. It does not carry the authenticated provider binding.

Callers must reconstruct the complete associated-data context from trusted
database identity:

| Provider | Tenant binding          | Subject binding       | Credential type           | Generation binding                          |
| -------- | ----------------------- | --------------------- | ------------------------- | ------------------------------------------- |
| GitHub   | `account:<user UUID>`   | user UUID             | `user-oauth-access-token` | `users.github_access_token_generation`      |
| Slack    | FortyOne workspace UUID | Slack team ID         | `bot-oauth`               | `slack_workspaces.installation_generation`  |
| Figma    | FortyOne workspace UUID | Figma connection UUID | `oauth-token`             | `figma_connections.installation_generation` |

Changing the provider, tenant, subject, type, or generation causes AEAD
authentication to fail. An envelope cannot be copied to another user,
workspace, Slack team, Figma connection, or installation generation and still decrypt. Unknown
envelope versions and missing KEK generations fail closed. The returned secret
holder redacts ordinary, Go-syntax, and JSON formatting; application logs must
still record only safe IDs, key IDs, versions, counts, and outcomes.

`Inspect` exposes only the envelope format and KEK reference for diagnostics;
it does not authenticate an envelope. `Open` authenticates and returns a
redacted, explicitly destroyable secret holder. `Rewrap` unwraps the DEK with
the referenced old KEK, authenticates both the wrapped DEK and payload against
the caller-supplied binding, clears its temporary plaintext buffer, and wraps
the same DEK with the active KEK. The provider ciphertext, payload nonce, AAD,
and provider generation do not change. A current-key rewrap is an authenticated
exact no-op.

The old Slack secret-box primitive remains only inside the bounded startup
cutover. It can open pre-vault Slack credentials and legacy Slack webhook inbox
payloads long enough to authenticate and replace them. Normal Slack credential
reads and webhook processing never receive that compatibility behavior and fail
closed on a legacy envelope. New Slack webhook payloads use the independent
`APP_SLACK_WEBHOOK_PAYLOAD_SECRET`; see
[Slack webhook security](slack-webhooks.md).

## Configuration

The API and worker must receive the same keyring:

```text
APP_CREDENTIAL_VAULT_ACTIVE_KEY_ID=provider-credentials
APP_CREDENTIAL_VAULT_ACTIVE_KEY_VERSION=1
APP_CREDENTIAL_VAULT_KEYS={"provider-credentials@1":"<base64 of exactly 32 random bytes>"}
# Worker only. Keep false outside one controlled rotation worker.
APP_CREDENTIAL_VAULT_REWRAP_ON_STARTUP=false
```

Figma webhook bodies use a separate shared secret:

```text
APP_FIGMA_WEBHOOK_PAYLOAD_SECRET=<at least 32 independent random bytes>
```

That secret is not a vault KEK and never protects retained OAuth tokens. The
API and worker validate that it does not reuse the auth, feedback, invitation,
OAuth, developer-credential, or credential-vault material.

Slack webhook bodies have the same independent-key requirement:

```text
APP_SLACK_WEBHOOK_PAYLOAD_SECRET=<at least 32 independent random bytes>
```

This key is shared only by the Slack API ingress and Slack worker. It is not an
OAuth credential KEK, Slack signing secret, client secret, or application auth
secret. Production validation rejects key reuse across those domains.

The keyring is a JSON object from `<key-id>@<positive-version>` to standard
base64. Key IDs contain only letters, digits, `.`, `_`, and `-`. Generate key
material in the production secret manager or KMS workflow; never commit it,
paste it into a ticket, place it in shell history, or reuse an authentication,
HMAC, webhook, or action-link key. Production rejects the public development
key semantically, including JSON formatting changes and a renamed all-zero key.
Malformed keyrings, unknown active references, duplicate references, and keys
that do not decode to exactly 32 bytes prevent startup.

The current backend wraps DEKs with KEKs loaded from the managed deployment
secret into the process. The bounded worker rewrap operation is implemented;
it never persists or returns provider plaintext. Moving the KEK operation into
a managed KMS/HSM remains a future backend change behind the same typed vault
contract. Do not remove an old KEK until the rotation proof in the operator
runbook succeeds.

## Migration and rollout

This is an expand/backfill/cutover rollout. Old processes cannot read `vault.v2`
credentials, while new processes deliberately reject legacy credentials during
normal provider operations. Use a short controlled integration-maintenance
window rather than running old and new provider code concurrently.

1. Back up PostgreSQL and apply migrations through
   `000169_figma_vault_installation_generation` before deploying the new
   binaries. Earlier vault migration `000154` remains part of that immutable
   chain.
2. Provision one independent 32-byte KEK and the same active ID/version/keyring
   on every replacement API and worker task. Provision a different
   `APP_SLACK_WEBHOOK_PAYLOAD_SECRET` on the Slack API and worker. Validate both
   configurations in staging. Keep the existing `APP_AUTH_SECRET_KEY` available
   only to the bounded cutover because it is required to decrypt pre-vault Slack
   and Figma credentials and old Slack inbox payloads; rotate that key only after
   every legacy-count query is zero.
3. Pause Slack installation/account-link callbacks, GitHub user-link writes,
   and Figma OAuth callbacks; drain provider requests, then stop every old API
   and worker task. Do not let an old task write legacy ciphertext after the
   one-time backfill.
4. Start one new worker. Worker construction backfills GitHub, active Slack
   installations, retained Slack uninstall credentials, retryable Slack webhook
   inbox payloads, and Figma connections before it reports ready or starts
   consuming jobs. Any decrypt, strict decode, seal, or compare-and-swap failure
   aborts startup; logs contain only aggregate counts and safe IDs.
5. Verify every legacy count is zero:

   ```sql
   SELECT COUNT(*) AS github_legacy_credentials
   FROM public.users
   WHERE github_access_token_envelope_version = 0
     AND github_access_token_generation IS NULL
     AND NULLIF(github_access_token, '') IS NOT NULL;

   SELECT COUNT(*) AS slack_legacy_credentials
   FROM public.slack_workspaces
   WHERE NULLIF(COALESCE(NULLIF(credential_payload, ''), bot_access_token), '') IS NOT NULL
     AND (
       credential_key_version < 2
       OR NULLIF(bot_access_token, '') IS NOT NULL
     );

   SELECT COUNT(*) AS slack_legacy_uninstall_credentials
   FROM public.slack_uninstall_outbox
   WHERE status <> 'completed'
     AND credential_key_version < 2
     AND NULLIF(credential_payload, '') IS NOT NULL;

   SELECT COUNT(*) AS figma_legacy_credentials
   FROM public.figma_connections
   WHERE credential_key_version = 1;
   ```

6. Keep `APP_CREDENTIAL_VAULT_REWRAP_ON_STARTUP=false`, start the remaining new
   workers and then the new API tasks. Exercise GitHub
   link/comment attribution, Slack install/event processing, Figma link and
   authenticated webhook processing, disconnect, and uninstall retry before
   reopening integration mutations.
7. Monitor vault authentication/unknown-key failures and provider errors. Never
   restore an old binary after vault envelopes exist; roll forward with the
   complete keyring.

The `000154` down migration refuses to remove GitHub metadata while encrypted
rows exist. The `000169` down migration refuses to remove Figma generation
metadata after any row reaches shared-vault version 2. Treat rollback after
backfill as a credential-aware recovery operation, not a schema-only downgrade.

## Rotation and revoke lifecycle

The keyring supports read-old/write-new and authenticated DEK rewrap. GitHub,
Slack, and Figma repositories scan stable, bounded UUID pages and compare-and-swap the
exact original envelope, envelope version, and provider generation. A
concurrent OAuth refresh, relink, reinstall, disconnect, provider revoke, or
uninstall completion therefore wins over maintenance. A partially completed or
repeated run is safe and deterministic; `current`, `rewrapped`, and `stale`
counts make the result explicit.

Run rotation only through
[`docs/operations/provider-credential-rotation.md`](../operations/provider-credential-rotation.md).
In summary: expand every process to old+new keys, activate the new generation
everywhere, run exactly one controlled worker with
`APP_CREDENTIAL_VAULT_REWRAP_ON_STARTUP=true`, rerun until the proof pass reports
only current envelopes and no stale writes, then disable the flag before the old
key is retired. Keeping the previous key in the ring also supports a deliberate
active-key rollback and rewrap in the other direction.

GitHub unlink clears its envelope and generation. Figma reconnect creates a new
connection generation and deactivates the previous connection/webhooks in one
transaction; disconnect deactivates that grant. Slack disconnect atomically
moves the still-encrypted credential to the uninstall outbox before deleting
the active installation; provider revoke deletes the active installation, and
uninstall completion clears the retained ciphertext. A stale maintenance write
cannot recreate any of those records.

If a KEK is suspected compromised, pause provider operations, preserve the
keyring only for a controlled credential-rotation window, revoke/reinstall the
affected provider credentials, and investigate secret-manager and runtime
access. Rewrapping alone does not revoke a provider token an attacker already
obtained.

## Migration acceptance contract

Provider repositories must retain the exact legacy-value backfill predicates
and envelope-plus-generation compare-and-swap predicates when persistence code
is generated with SQLC. PostgreSQL tests for envelope-only writes, associated
data binding, atomic backfill, residual-column scrub, concurrent refresh and
rewrap, revoke fencing, and rollback are mandatory. A compatibility reader in
normal provider traffic is not an acceptable substitute for completing the
cutover.

## Verification

```bash
go test -race ./internal/platform/credentialvault \
  ./internal/modules/github/repository ./internal/modules/github/service \
  ./internal/modules/slack/repository ./internal/modules/slack/service \
  ./internal/modules/figma/repository ./internal/modules/figma/service

TEST_DATABASE_URL='postgresql://<disposable-createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -tags=integration -count=1 \
  ./internal/modules/github/service ./internal/modules/slack/service \
  ./internal/modules/figma/repository

go test ./internal/bootstrap/architecture
make security-check
make gitleaks-check
```

The tagged suite applies the real migration chain to isolated PostgreSQL
databases and never silently skips when the database contract is unavailable.
