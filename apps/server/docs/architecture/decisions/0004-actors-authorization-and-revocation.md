# ADR 0004: Actors, authorization, and revocation

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and security

## Context

A user identifier alone cannot represent human sessions, service accounts,
OAuth applications, API keys, or internal jobs. Middleware-only role checks also
cannot safely authorize a loaded resource or react consistently to membership
revocation.

## Decision

Every authenticated use case receives an immutable typed actor containing:

- principal and actor kind (`user`, `service_account`, `oauth_app`, or `system`);
- authenticated credential/session identifier and expiry;
- granted scopes;
- workspace and team constraints when present;
- authentication time and an authorization-version/freshness marker.

Transport authentication establishes credential validity and constructs the
actor. Service policy authorizes the actor against the current workspace,
membership, team, resource, and operation. Repositories still include tenant
predicates as defense in depth. Unknown actor kinds, roles, scopes, or membership
states fail closed.

Authorization caches store short-lived derived facts, never bearer credentials.
Privilege changes increment or invalidate a versioned subject/workspace entry.
Privileged mutations, credential management, and billing revalidate current
authority rather than accepting stale cache state. Revoked sessions and machine
credentials are rejected immediately through the credential record/version.

The existing `userID` context is a migration adapter, not the final contract.

## Enforcement and adoption

- Central policy packages have table-driven actor/role/scope/team matrices.
- Each protected repository has two-workspace negative tests.
- Tests cover membership removal, role downgrade, credential revoke, stale cache,
  and unknown enum states.
- HTTP middleware cannot be the only authorization evidence for a use case.

Adopt first on workspace administration and credentials, then carry the actor
through each module's SQLC migration.

## Consequences

Policies become explicit and reusable across HTTP, workers, integrations, and
future API clients. Reads may use bounded caches; sensitive operations incur an
intentional current-state lookup.

## Revisit when

Measured authorization load requires a different consistency mechanism that can
still prove bounded staleness and immediate revocation for sensitive paths.
