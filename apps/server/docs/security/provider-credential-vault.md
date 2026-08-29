# Provider credential vault

The credential vault protects recoverable GitHub, Slack, and Figma OAuth
credentials at rest. Provider tokens are stored as versioned authenticated
envelopes, never as plaintext fallbacks.

## Simple configuration model

The API and worker both receive the same stable `APP_AUTH_SECRET_KEY`. FortyOne
uses HKDF-SHA256 to derive independent versioned keys for:

- provider credential encryption;
- GitHub webhook inbox encryption;
- Slack webhook inbox encryption; and
- Figma webhook inbox encryption.

The purpose label is part of each derivation. A key derived for Slack cannot
open a GitHub payload or a provider credential, even though every key starts
from the same application root. The application root must be at least 32 bytes
and must come from the production secret store.

There are no `APP_CREDENTIAL_VAULT_*` or
`APP_<PROVIDER>_WEBHOOK_PAYLOAD_SECRET` variables to provision. Keep
`APP_AUTH_SECRET_KEY` identical across API and worker tasks.

## Security contract

Every `Seal` call creates a random 256-bit data-encryption key (DEK). The
credential is encrypted with AES-256-GCM and a fresh random nonce. The derived
provider-vault key wraps that DEK with a second AES-256-GCM operation and an
independent nonce.

The stored `vault.v2` envelope records the algorithm, associated-data format,
key ID and version, nonces, ciphertext, and wrapped DEK. It does not store the
authenticated provider binding. Callers reconstruct that binding from trusted
database identity:

| Provider | Tenant binding | Subject binding | Credential type | Generation binding |
| --- | --- | --- | --- | --- |
| GitHub | `account:<user UUID>` | user UUID | `user-oauth-access-token` | `users.github_access_token_generation` |
| Slack | workspace UUID | Slack team ID | `bot-oauth` | `slack_workspaces.installation_generation` |
| Figma | workspace UUID | Figma connection UUID | `oauth-token` | `figma_connections.installation_generation` |

Changing the provider, tenant, subject, credential type, or generation causes
authentication to fail. An encrypted token cannot be copied to another user,
workspace, installation, or connection and still decrypt.

`Open` returns a redacted, explicitly destroyable secret holder. Application
logs contain safe identifiers and bounded outcomes only. They must never
contain tokens, derived keys, vault envelopes, or decrypted payloads.

## Legacy migration

The worker has a bounded compatibility path for credentials written before the
shared vault existed. During startup it can open those legacy values with the
application root, authenticate their database identity, and replace them with
the derived `vault.v2` envelope using compare-and-swap writes.

Normal provider traffic does not use that compatibility reader. New writes use
the derived vault key, and a missed migration fails closed instead of silently
extending the legacy format.

Apply migrations before deploying the replacement API and worker. Start one
worker first and let its bounded backfill complete before scaling the remaining
workers. Then verify GitHub link operations, Slack installation and event
processing, and Figma connection and webhook processing.

## Root-secret changes

Do not casually rotate `APP_AUTH_SECRET_KEY`. It is the root for the derived
integration keys and still supports bounded legacy protocols. Replacing it
without a migration makes existing provider credentials and retained webhook
receipts unreadable by design.

If the root is compromised, follow the
[root-secret recovery runbook](../operations/provider-credential-rotation.md).
The safe recovery may require reconnecting provider installations. A future
managed keyring or KMS backend can add independent online rotation when the
operational need justifies that complexity.

## Verification

```bash
go test -race ./internal/platform/appkeys ./internal/platform/credentialvault
go test -race ./internal/modules/github/... ./internal/modules/slack/... ./internal/modules/figma/...
go test ./internal/bootstrap/architecture
make security-check
```
