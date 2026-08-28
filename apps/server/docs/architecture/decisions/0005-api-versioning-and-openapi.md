# ADR 0005: API versioning and OpenAPI

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and product

## Context

The current routes primarily serve first-party clients. External integrations
need a deliberate stable contract, but FortyOne is now a privately operated
application and does not promise source distribution or self-hosting support.
Those are independent concerns.

## Decision

Only routes explicitly placed under `/api/v1` and included in the canonical
OpenAPI document are external developer contracts. Existing first-party routes
remain internal and may not be inferred as public because they are reachable.

OpenAPI is the source for transport schemas, examples, authentication/scopes,
error envelopes, pagination, idempotency, rate limits, and deprecation metadata.
Generated transport types are mapped at the HTTP boundary and never become
domain or repository models. The server remains the behavioral authority; code
generation does not replace policy or integration tests.

Compatible changes may add optional fields or endpoints. Removing/renaming a
field, tightening a previously accepted value, changing meaning, or changing an
error/status contract requires a new API version or an announced deprecation
window with measured usage. Security fixes may override compatibility when the
risk is documented and rollout is explicit.

## Enforcement and adoption

- Generation and lint must be reproducible and drift-gated.
- A breaking-change comparison runs against the accepted base contract.
- Contract tests exercise examples, authorization, errors, and idempotency.
- The docs site publishes only explicit integration contracts; internal routes
  and deployment configuration stay private.

The initial preview begins with scoped workspace/team/story reads and outbound
webhook management after the actor and cursor contracts. Broader mutations and
SDK previews remain deliberate follow-up work.

## Consequences

First-party refactoring remains possible while external integrators receive a
small, intentional surface. Maintaining versions and deprecations has a cost,
so endpoints are promoted only for a real integration use case.

## Revisit when

Product strategy changes source distribution/self-hosting support or a protocol
other than HTTP/OpenAPI becomes a supported external contract.
