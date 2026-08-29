# ADR 0009: Webhook delivery and replay safety

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and integrations

## Context

Webhooks cross an untrusted, retrying network. Slow synchronous processing,
signature verification after parsing, or a single `processed` boolean creates
replay, data-loss, and ambiguous-retry failures.

## Decision

Inbound delivery:

1. bound and retain the exact raw body needed for verification;
2. authenticate signature, timestamp, provider/account, and supported algorithm
   before trusting parsed fields;
3. reject stale timestamps and record a unique provider delivery ID/digest;
4. atomically claim an inbox row with `pending`, `processing`, `completed`,
   `ignored`, `failed`, or `cancelled` state and a bounded processing lease;
5. acknowledge quickly after durable acceptance and process asynchronously;
6. retry failed/stale work idempotently and quarantine poison deliveries.

Outbound delivery signs `delivery_id.timestamp.raw_body` with a per-endpoint
versioned HMAC secret. Delivery IDs remain stable across attempts. The
`Webhook-Id`, `Webhook-Timestamp`, and `Webhook-Signature` headers follow the
[Standard Webhooks](https://www.standardwebhooks.com/) signing shape. The dispatcher uses bounded
exponential backoff with jitter, respects safe retry hints, records sanitized
status/error metadata, and disables or alerts on terminal endpoints.

Raw payload retention is minimized and documented. Secrets and raw customer or
provider payloads are not logged. Inbox/outbox cleanup preserves audit metadata
and idempotency for at least the maximum retry/replay window.

The concrete inbound lifecycle, SQLC repository, recovery policy, and provider
adoption checklist are documented in the
[`inbound webhook gateway`](../../integrations/webhook-gateway.md) runbook.
The separate customer-facing delivery lifecycle is documented in the
[`outbound webhook delivery`](../../integrations/outbound-webhooks.md) runbook.

## Enforcement and adoption

- Shared contract tests cover invalid signatures, algorithm confusion, stale
  timestamp, duplicate, concurrent claim, crash/lease recovery, retry, and poison delivery.
- Provider adapters supply verification/parsing; the durable gateway owns lifecycle.
- Metrics cover receive, verification failure, queue delay, attempts, terminal
  failure, duplicate rate, and oldest pending age.

## Consequences

Webhook handlers become reliable and fast, at the cost of durable state and a
worker. Payload access is intentionally narrower than ordinary event data.

## Revisit when

A managed gateway proves equivalent signature, tenant, idempotency, retention,
and audit guarantees without weakening provider-specific verification.
