# ADR 0006: Cursor pagination

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering

## Context

Copied offset parsing has inconsistent bounds and becomes unstable as rows are
inserted or deleted. External list contracts need deterministic continuation
without exposing mutable database implementation details.

## Decision

New or migrated unbounded lists use keyset pagination with the shared signed
opaque cursor codec. Every query defines a deterministic total order ending in a
unique immutable tie-breaker. The cursor contains only the version, resource
kind, normalized filter/order fingerprint, boundary values, and optional expiry.
It is authenticated with a rotatable HMAC key and is never treated as authority.

Clients provide a bounded `limit`; the server fetches `limit + 1`, returns at
most `limit`, and emits the next cursor only when another row exists. Invalid,
tampered, expired, wrong-resource, or filter-mismatched cursors return the same
safe request error.

Offset pagination may remain temporarily for a documented first-party route.
Its parsing uses the shared bounded primitive. Public v1 does not expose both
models for the same operation unless compatibility requires it.

## Enforcement and adoption

- Unit/fuzz tests cover encoding, tamper, expiry, version, bounds, and filter mismatch.
- Repository tests cover equal sort values, concurrent insert/delete, forward
  traversal, tenant isolation, and empty/final pages.
- SQL review requires a matching composite index or query-plan evidence.

## Consequences

Lists are stable and scale without increasing offset scans. Cursors are not
human-readable and sort/filter changes require an explicit cursor version.

## Revisit when

A bounded dataset demonstrably benefits from a simpler complete response or a
product requirement needs random page access and accepts its consistency/cost.
