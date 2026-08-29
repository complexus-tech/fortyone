# Figma module

The Figma module adds design context to stories without making the stories
module depend on Figma models. It owns Figma OAuth, encrypted credentials,
design-link metadata, provider webhooks, and design-to-story orchestration.

Start here when changing Figma behavior. The security details live in
[`docs/security/figma-webhooks.md`](../../../docs/security/figma-webhooks.md)
and
[`docs/security/provider-credential-vault.md`](../../../docs/security/provider-credential-vault.md).

## Package map

| Path                   | Responsibility                                                                                                  | Must not contain                                         |
| ---------------------- | --------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| `domain/`              | Persistence-independent Figma values and errors                                                                 | HTTP, SQLC, pgx, story-service, or SDK types             |
| `http/`                | Route registration, bounded input, request/response mapping, status mapping                                     | SQL, provider business rules, token encryption           |
| `service/`             | OAuth and design-link use cases, provider calls, webhook verification/processing, credential refresh and rewrap | SQL strings, pgx rows, concrete story repository imports |
| `repository/queries/`  | Reviewed named PostgreSQL statements                                                                            | Dynamic SQL or HTTP/provider behavior                    |
| `repository/sqlc/`     | Generated SQLC code                                                                                             | Hand edits                                               |
| `repository/`          | SQLC-to-domain mapping and native pgx transactions                                                              | HTTP models, provider API calls, story business logic    |
| `credentialmigration/` | The only reader for retired provider-local Figma ciphertext                                                     | Normal runtime reads or a plaintext fallback             |
| `provider.go`          | Stable provider descriptor and capabilities                                                                     | Runtime construction or credentials                      |

Cross-module mapping belongs in `internal/bootstrap/providers`. API and worker
both use the same `FigmaStoryAdapter`, which implements Figma's narrow
`StoryService` port without exposing story-service models to this module.

## Core invariants

1. Every repository read or mutation is tenant-scoped directly or is reached
   through a globally unique, authenticated provider webhook ID that resolves
   its owning active connection.
2. OAuth state is random, hashed at rest, workspace-and-user bound, expiring,
   and consumed exactly once in one SQL statement.
3. Retained OAuth tokens are `vault.v2` envelopes. Normal service code never
   falls back to the old Figma cipher.
4. Vault associated data is reconstructed from trusted database identity:
   provider `figma`, workspace UUID, connection UUID, credential type, and
   immutable installation generation.
5. Credential refresh and KEK rewrap use compare-and-swap predicates containing
   the connection, installation generation, and exact previous envelope. A
   disconnect, reconnect, refresh, or concurrent rotation always wins over a
   stale write.
6. Reconnection deactivates the previous connection and its webhooks in the
   same native pgx transaction before creating the new generation.
7. Creating or deleting a Figma story link changes both `story_links` and
   `story_figma_links` in one transaction. A partial generic link is never
   committed.
8. Webhook ingress persists encrypted exact bytes before queueing. Redis carries
   only the shared inbox UUID. The API and worker share a dedicated
   `APP_FIGMA_WEBHOOK_PAYLOAD_SECRET`; auth, passcode, OAuth-token, and
   credential-vault keys are never used for inbox encryption.
9. The worker rechecks the current connection generation before performing any
   event effect. Stale work becomes a terminal cancellation.
10. Tokens, passcodes, request bodies, vault envelopes, and raw provider errors
    never enter logs, queue payloads, response bodies, or activity text.

## OAuth connection flow

```text
authenticated workspace member
        |
        v
CreateInstallSession
  - random state + PKCE verifier
  - persist state hash, workspace, user, slug, expiry
        |
        v
Figma authorization callback
        |
        v
atomic ConsumeOAuthState
        |
        v
exchange code + load Figma user
        |
        v
seal token with workspace/connection/generation AAD
        |
        v
transactional UpsertConnection
  - deactivate old webhooks
  - deactivate old connection
  - create new active generation
```

`CompleteOAuth` cleans up old provider webhooks on a best-effort basis, but the
local database generation is the authorization boundary. Remote cleanup
failure cannot keep an old local installation authoritative.

## Credential formats and migration

Migration `000169_figma_vault_installation_generation` adds:

- `credential_key_version`, where `1` is the retired provider-local envelope
  and `2` is the shared credential vault;
- an immutable `installation_generation`; and
- the partial active workspace/generation index used by lifecycle checks.

The worker runs `credentialmigration.Migrator` before consuming tasks. It scans
stable UUID pages, authenticates legacy AES-GCM, strictly decodes one token JSON
document, seals it with the shared vault context, clears plaintext buffers, and
compare-and-swaps the exact old row. One failure aborts worker startup; there is
no silent skip and no normal-operation compatibility fallback.

The down migration succeeds only while every row is still version 1. After any
credential is written or migrated to version 2 it fails closed, because removing
the installation generation would make the authenticated context unrecoverable.

## SQLC repository

The module has one SQLC package configured in `sqlc.yaml`. Queries are grouped
by behavior:

| File              | Named operations                                                                 |
| ----------------- | -------------------------------------------------------------------------------- |
| `oauth.sql`       | save and atomically consume OAuth state                                          |
| `connections.sql` | reconnect transaction, active read, refresh CAS, legacy migration, KEK rewrap    |
| `links.sql`       | tenant-scoped link lists, handoff summaries, transactional upsert/delete, update |
| `webhooks.sql`    | active webhook grant, generation-fenced current grant, list/save/deactivate      |

Repository files mirror these groups. Do not put all queries back into one Go
file and do not add handwritten `db.Query` or `db.Exec` calls. Add a named query,
run the canonical generator, map generated rows at the repository edge, and
return domain errors.

The repository maps:

- no row to `domain.ErrNotFound`;
- unique, serialization, and deadlock failures to `domain.ErrConflict`; and
- foreign-key, not-null, and check failures to `domain.ErrForbidden`.

Callers must still interpret those errors according to the use case. For
example, an unknown webhook is ignored, while a database outage is a retryable
verification failure.

## Durable webhook flow

The API and worker construct the same provider runtime around the shared
webhook platform:

```text
bounded POST
  -> Figma verifier
  -> context-bound exact-body encryption
  -> shared SQLC inbox
  -> Asynq {inboxId}
  -> worker lease
  -> decrypt and reparse
  -> current-generation check
  -> idempotent design/story work
  -> bounded terminal outcome
```

The recovery scheduler runs every minute. At-least-once queue delivery is
intentional; the inbox lease and idempotent repository mutations, not Redis
exactly-once assumptions, provide correctness.

The API and worker must receive the same independent, randomly generated
`APP_FIGMA_WEBHOOK_PAYLOAD_SECRET`. Production startup rejects a missing,
undersized, development-default, or reused value. Rotating this key requires a
drained Figma inbox because pending ciphertext must remain decryptable; do not
silently fall back to `APP_AUTH_SECRET_KEY` or a credential-vault KEK.

## Adding behavior

For a new design-context operation:

1. Put provider-independent values in `domain` or a small service-owned input.
2. Add the smallest consumer-owned port; do not expand one universal Figma or
   integration interface.
3. Keep Figma HTTP/SDK types in the adapter or service provider client.
4. Add named SQLC queries with tenant and lifecycle predicates visible in SQL.
5. Bind all multi-table state changes in one native pgx transaction.
6. Add unit tests for validation/error mapping and PostgreSQL 18 tests for
   authorization, concurrency, rollback, and the critical query plan.
7. Update this file and the relevant integration/security runbook.

For another design provider, implement the same narrow design-context
capability only where the semantics match. Do not copy Figma passcode rules,
OAuth scopes, IDs, or payload models into a supposedly generic interface.

## Verification

Fast checks:

```bash
go test -race ./internal/modules/figma/...
.tools/bin/sqlc compile -f sqlc.yaml
make sqlc-check
make architecture-check
```

Database checks:

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/figma/repository
```

The database role must point at a disposable PostgreSQL 18 control database and
have `CREATEDB`. The test kit creates and drops only prefixed, test-owned
databases and applies the complete real migration chain.
