# ADR 0008: Integration capability registry

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering and integrations

## Context

GitHub, GitLab, Slack, and future providers overlap in capability but differ in
authentication, identifiers, events, and limits. Copying whole provider flows
duplicates business rules; one enormous provider interface creates fake
portability and leaks SDK details.

## Decision

Define small capability contracts owned by the consuming use case, such as code
host repositories, issue synchronization, identity linking, messaging delivery,
and webhook installation. A provider registers a descriptor containing stable
provider ID, supported capability versions, required configuration, health, and
factory functions for only the capabilities it implements.

Domain commands/results use FortyOne types. Provider SDK types, pagination,
errors, rate-limit headers, and webhook payloads stay inside the adapter. Common
orchestration—authorization, mapping policy, idempotency, audit, retry, and state
transitions—lives above adapters. Provider-specific behavior remains explicit;
the registry does not force the least-common denominator.

Because the application is privately operated, custom integrations use the
versioned API, OAuth/service accounts, and signed webhooks. Loading untrusted
in-process plugins and self-host installation packages is out of scope until a
separate support/security decision is funded.

## Enforcement and adoption

- Contract suites run against each adapter capability.
- Compile-time assertions prove registrations and capability versions.
- Architecture tests reject provider SDK imports outside adapters/bootstrap.
- Adding a provider changes registration/config/adapters, not core business rules.

Start by extracting contracts from proven GitHub/Slack flows, then implement
GitLab or another provider to validate that the boundary is real.

## Consequences

Capabilities remain composable and externally extensible without exposing
internal Go interfaces as an unsupported plugin ABI. Some providers implement
only a subset, which callers must discover and handle deliberately.

## Revisit when

There is a resourced commitment to a sandboxed plugin runtime or supported
self-hosting distribution, including compatibility and incident ownership.
