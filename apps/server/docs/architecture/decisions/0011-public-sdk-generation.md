# ADR 0011: Contract-derived public SDK previews

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering

## Context

External integrators need typed clients, but a handwritten client can silently
invent fields, endpoints, retry behavior, or authentication rules that are not
part of the public API. The versioned OpenAPI document is already the accepted
contract boundary in ADR 0005. SDK generation must preserve that boundary and
remain reproducible on developer machines and in CI.

## Decision

The committed `api/openapi/v1/openapi.yaml` graph is the only input to public
SDK generation. Generation first produces one deterministic bundled document;
both languages consume that same temporary bundle.

- Go models and the response-aware standard-library client are generated with
  `oapi-codegen` v2.8.0. The preview module uses
  `github.com/oapi-codegen/runtime` v1.6.0.
- TypeScript declarations are generated with `openapi-typescript` v7.13.0.
  The preview client uses `openapi-fetch` v0.17.0 and is checked with
  TypeScript v5.9.3. All versions are exact, not semver ranges.
- Generated files are never hand-edited. `make sdk-generate` replaces only the
  four declared generated artifacts. `make sdk-check`, and therefore
  `make generated-check`, regenerate into a temporary directory, diff every
  artifact, and run both SDK and sample tests.

Small handwritten helpers may configure bearer authentication, validate the
base URL, retry safe reads, map the documented error envelope, traverse opaque
cursors, or verify documented webhook signatures. They may not add a path,
field, enum, OAuth flow, idempotency promise, or service-account capability
that is absent from OpenAPI v1. Authentication helpers reject redirects so a
bearer credential cannot be forwarded to another origin. Automatic retries are
limited to `GET` and `HEAD` after network failures or explicit `429`/`503`
responses, with bounded attempts and backoff.

Both packages are preview artifacts in this monorepo. They are not currently
published to npm or the Go module proxy. Publication, compatibility support,
release signing, provenance, and deprecation automation require a separate
release decision before the private markers or local module replacement are
removed.

## Consequences

A contract change produces reviewable server and client drift in one change.
Integrators get useful security defaults without creating a second contract.
The nested Go module and TypeScript workspace package add focused dependency
and test maintenance, and preview users must currently consume them from this
repository.

The first preview intentionally has no story mutations, REST OAuth client, or
idempotency helper because OpenAPI v1 exposes none of those contracts. The
existing OAuth documentation is MCP-specific. Adding any of them starts with a
reviewed OpenAPI operation and server behavior, not an SDK-only convenience.

## References

- [`oapi-codegen` client generation and configuration](https://github.com/oapi-codegen/oapi-codegen)
- [`openapi-typescript` CLI](https://openapi-ts.dev/cli)
- [`openapi-fetch` API](https://openapi-ts.dev/openapi-fetch/api)
- [OpenAPI 3.1 read-only and write-only semantics](https://spec.openapis.org/oas/v3.1.0.html)
