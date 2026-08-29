# Architecture decisions

These records make the API modernization plans executable. The plans remain the
program-level source of truth; an ADR fixes one cross-cutting choice so a module
migration does not redesign it locally.

| ADR                                                 | Decision                                          | Status   |
| --------------------------------------------------- | ------------------------------------------------- | -------- |
| [0001](0001-modular-monolith-boundaries.md)         | Modular monolith and dependency direction         | Accepted |
| [0002](0002-sqlc-and-native-pgx.md)                 | SQLC, native pgx, and exceptional SQL             | Accepted |
| [0003](0003-transactions-and-outbox.md)             | Transaction ownership and outbox delivery         | Accepted |
| [0004](0004-actors-authorization-and-revocation.md) | Actors, authorization, caching, and revocation    | Accepted |
| [0005](0005-api-versioning-and-openapi.md)          | API versioning and OpenAPI generation             | Accepted |
| [0006](0006-cursor-pagination.md)                   | Cursor pagination and compatibility               | Accepted |
| [0007](0007-machine-credentials.md)                 | API keys, service accounts, and OAuth credentials | Accepted |
| [0008](0008-integration-capabilities.md)            | Provider capabilities and extension boundary      | Accepted |
| [0009](0009-webhook-delivery.md)                    | Webhook verification, replay safety, and delivery | Accepted |
| [0010](0010-observability-and-sensitive-data.md)    | Observability, SLO ownership, and sensitive data  | Accepted |
| [0011](0011-public-sdk-generation.md)               | Contract-derived public SDK previews              | Accepted |

## Status meanings

- **Proposed:** review is still required and production code must not depend on
  the decision.
- **Accepted:** new work follows the decision. Existing departures are migration
  debt, not precedent.
- **Superseded:** a later ADR names and replaces this record.

Changing an accepted decision requires a new ADR. Editing history to make the
old choice appear different is not allowed.
