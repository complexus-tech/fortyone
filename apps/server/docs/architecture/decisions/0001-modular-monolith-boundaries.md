# ADR 0001: Modular monolith boundaries

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering

## Context

FortyOne has one deployable API with many business capabilities. Splitting it
into services now would add distributed failure modes without removing the
current coupling. The immediate problem is that responsibilities and dependency
direction are inconsistent inside the monolith.

## Decision

Keep a modular monolith. A business capability lives under
`internal/modules/<module>` and may contain transport, service/use-case,
repository, worker, and provider-adapter packages only when that capability
needs them.

Dependencies point inward:

1. bootstrap composes concrete dependencies;
2. transports depend on service contracts and shared transport primitives;
3. services own use cases, policy, domain values, and transaction intent;
4. repositories implement persistence and may use only their own generated SQLC package;
5. provider adapters contain third-party SDK types and implement capability contracts;
6. modules do not call another module's handler or generated repository package.

Interfaces are declared at the boundary that consumes a capability. Interfaces
are not created merely to mirror a concrete type. Shared code moves to
`internal/platform` only after at least two real consumers demonstrate the same
stable semantic contract.

## Enforcement and adoption

- AST architecture tests reject forbidden imports and generated-type leakage.
- The debt baseline may preserve existing violations but cannot grow.
- A handwritten production file above 700 lines requires an explicit exception
  or a cohesive split before substantial behavior is added.
- Migration happens one complete behavior slice at a time; empty archetype
  directories are not created.

## Consequences

The API remains operationally simple while module ownership becomes visible.
Some temporary adapters will exist during migration. They must be named as debt
and removed with the migrated slice.

## Revisit when

A module has an independently owned SLO, scaling profile, data boundary, and
deployment need that cannot be met safely inside the process. Code size alone
does not justify service extraction.
