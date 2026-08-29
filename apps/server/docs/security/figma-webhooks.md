# Figma webhook security and delivery

Figma sends a webhook passcode inside the JSON event envelope rather than a
detached HTTP signature header. FortyOne treats that value as a bearer
credential even though it is structurally part of the request body. Figma does
not provide the same signed-timestamp replay contract as GitHub or Slack, so
the adapter combines constant-time passcode verification, exact-body delivery
digests, a durable encrypted inbox, and an installation-generation fence.

## Ownership

| Concern                                               | Owner                                                   |
| ----------------------------------------------------- | ------------------------------------------------------- |
| 1 MiB request limit and HTTP response                 | `internal/modules/figma/http`                           |
| Passcode, event, file, and installation verification  | `internal/modules/figma/service/webhook_runtime.go`     |
| Exact-body encryption and canonical receipt lifecycle | `internal/platform/webhooks`                            |
| SQLC inbox persistence, lease, recovery, and expiry   | `internal/platform/webhooks/repository`                 |
| Figma connection, link, and webhook persistence       | `internal/modules/figma/repository`                     |
| Queue decoding and recovery scheduling                | `internal/taskhandlers` and `internal/bootstrap/worker` |

The Figma service owns provider semantics. The shared gateway does not know
Figma payload fields or passcode rules.

## Payload-encryption key

The API and worker derive the same Figma payload key from the stable
`APP_AUTH_SECRET_KEY` using HKDF-SHA256 and a versioned Figma purpose label. It
encrypts and authenticates exact Figma request bodies in the durable inbox
without requiring another environment variable.

The derived key is not the provider passcode and does not encrypt retained
Figma OAuth tokens. Passcodes are one-way digested; OAuth tokens use a separate
derived credential-vault purpose. Before changing the application root or
derivation version, pause Figma ingress and drain every pending or processing
Figma inbox receipt; old inbox ciphertext is intentionally not opened with a
cross-purpose fallback key.

## Ingress sequence

1. The HTTP adapter reads at most 1 MiB. It rejects an unreadable or oversized
   body before JSON decoding or provider lookup.
2. The verifier parses only the fields needed for authentication and routing,
   requires a positive provider webhook ID, and loads the active webhook plus
   its active connection through module-owned SQLC queries.
3. A missing or inactive installation is intentionally ignored. A database or
   infrastructure failure is surfaced as a safe verification failure; it is
   never disguised as a harmless unknown webhook.
4. The presented passcode is SHA-256 digested and compared with the stored
   digest in constant time. Plaintext passcodes are never stored.
5. The verifier requires the configured event type and file key. `PING` is the
   only event allowed to bypass the file-specific event match.
6. SHA-256 of the exact request body becomes the provider delivery identity.
   This deduplicates exact retries. The passcode authenticates the body, while
   the inbox uniqueness constraint prevents the same exact delivery from
   creating another receipt.
7. The shared payload codec encrypts the exact body with the dedicated payload
   key and authenticated context
   containing provider, delivery ID, FortyOne workspace, connection ID, and
   installation generation.
8. PostgreSQL records the canonical inbox row before queue dispatch. Redis
   receives only `{inboxId}`; it never receives the body, passcode, token, file
   metadata, workspace ID, or provider credential.
9. The API returns `202 Accepted` after durable persistence and queue handoff.
   If dispatch fails, the pending row remains recoverable and the provider gets
   a retryable response.

## Worker sequence

1. The worker validates the task UUID and acquires a bounded inbox processing
   lease. Duplicate tasks are expected under at-least-once delivery.
2. It opens the encrypted body only with the receipt's trusted provider,
   delivery, tenant, connection, and generation binding.
3. It parses the event again at the provider boundary and immediately clears
   the decoded passcode field.
4. Before any product effect, it reloads the active Figma webhook using the
   receipt's connection ID, installation generation, and provider webhook ID.
5. A disconnect or reconnect makes the receipt stale; the worker completes it
   as `cancelled` and performs no story or design mutation.
6. Current events execute provider-specific, idempotent work and complete with
   a bounded outcome code. Failures are marked retryable without persisting raw
   database or provider error text.

Recovery runs once per minute. It uses the shared provider-scoped claim policy,
`FOR UPDATE SKIP LOCKED`, bounded exponential backoff, and the same generation
check as an ordinary delivery.

## Replay boundary

Figma's passcode authenticates possession but does not cryptographically bind a
detached request timestamp. FortyOne therefore cannot claim the signed replay
window available from providers that sign timestamp plus body. Exact-body
retries converge on one inbox identity; a byte-different authenticated body is
a different delivery. Product writes must remain idempotent, and operators
should rotate a passcode if it may have escaped.

Do not invent a timestamp rejection rule from the untrusted JSON `timestamp`
field. Such a rule would reject delayed legitimate provider retries without
adding cryptographic proof.

## Logging and retention

Never log the incoming body, decoded event, passcode, passcode digest, encrypted
payload, OAuth token, Figma URL containing sensitive query data, or raw provider
error. Safe fields are the inbox UUID, internal connection UUID, installation
generation, provider webhook ID, normalized event type, safe outcome code, and
request correlation ID.

The shared inbox retains encrypted body content for the configured bounded
period and then clears only that ciphertext. Provider, delivery identity,
installation fence, timestamps, status, and attempt counters remain as minimal
audit and deduplication facts.

## Historical compatibility table

Migration `000165_redact_figma_webhook_passcodes` removed `passcode` from the
legacy `figma_webhook_events.payload` column and installed a database check that
rejects future credential-bearing JSON. The current runtime does not write or
process that legacy event table; all new deliveries use the shared encrypted
inbox. Keep the redaction constraint during the expand/cutover period so an old
binary cannot resume storing passcodes.

If the constraint rejects a production write, an obsolete ingress path is
still running. Pause Figma ingress, stop that binary, preserve sanitized audit
rows, deploy the current gateway-backed path, and allow the provider to retry.
Do not remove the constraint or restore redacted values.

## Verification

```bash
go test -race ./internal/modules/figma/... ./internal/platform/webhooks/...

TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/figma/repository \
    ./internal/platform/webhooks/repository

.tools/bin/sqlc compile -f sqlc.yaml
make architecture-check
make security-check
```

The PostgreSQL suite proves migration backfill and guarded rollback, exact
concurrent deduplication, encrypted storage, tenant and generation binding,
transaction rollback, cross-workspace link rejection, credential refresh CAS,
and the critical active-generation index plan.
