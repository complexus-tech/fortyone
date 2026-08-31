# FortyOne API engineering guide

This directory is the internal operating manual for the managed FortyOne API.
The three modernization plans remain the program source of truth; the pages
here describe the code and operational contracts that are already enforced.

## Start here

- New engineer: [`onboarding/first-hour.md`](onboarding/first-hour.md)
- Making a module change: [`onboarding/change-a-module.md`](onboarding/change-a-module.md)
- Finding a route, guard, handler, or persistence owner:
  [`inventory/api.md`](inventory/api.md)
- Reviewing dependency and implementation rules:
  [`architecture/standards.md`](architecture/standards.md)
- Understanding a cross-cutting decision:
  [`architecture/decisions/README.md`](architecture/decisions/README.md)
- Looking up API or worker configuration: [`configuration.md`](configuration.md)

## Where code belongs

| Concern                                              | Canonical location                                           | Primary guide                                                                    |
| ---------------------------------------------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| Route and protocol mapping                           | `internal/modules/<module>/http`                             | [`architecture/standards.md`](architecture/standards.md)                         |
| Use case and caller-owned ports                      | `internal/modules/<module>/service`                          | [`onboarding/change-a-module.md`](onboarding/change-a-module.md)                 |
| Stable domain values and errors shared with adapters | `internal/modules/<module>/domain`                           | [`architecture/standards.md`](architecture/standards.md)                         |
| Reviewed SQL                                         | `internal/modules/<module>/repository/queries`               | [`database/sqlc.md`](database/sqlc.md)                                           |
| Handwritten row/domain mapping                       | `internal/modules/<module>/repository`                       | [`database/sqlc.md`](database/sqlc.md)                                           |
| Generated SQLC code                                  | `internal/modules/<module>/repository/sqlc`                  | [`database/sqlc.md`](database/sqlc.md)                                           |
| Shared actor and authorization infrastructure        | `internal/platform/actors`, `auth`, and `authorization`      | [`security/authorization.md`](security/authorization.md)                         |
| Provider-neutral integration capabilities            | `internal/platform/integrations`, `codehost`, and `webhooks` | [`integrations/providers.md`](integrations/providers.md)                         |
| Process composition and lifecycle                    | `internal/bootstrap/api` and `internal/bootstrap/worker`     | [`architecture/api-lifecycle.md`](architecture/api-lifecycle.md)                 |
| Reusable test infrastructure                         | `internal/testkit`                                           | [`testing/integration-infrastructure.md`](testing/integration-infrastructure.md) |

Generated packages are outputs, not extension points. Domain and service code
must not import SQLC, OpenAPI-generated, HTTP, SQLx, pgx, queue, or provider SDK
types. Repositories implement service-owned ports using transport-neutral
domain values; compile-time assignment at composition proves compatibility
without reversing the package dependency.

## Security and integrations

- Actor, scope, role, team, resource, and revocation rules:
  [`security/authorization.md`](security/authorization.md)
- Secrets and provider credentials:
  [`security/purpose-specific-keys.md`](security/purpose-specific-keys.md) and
  [`security/provider-credential-vault.md`](security/provider-credential-vault.md)
- Developer credentials and OAuth:
  [`security/developer-credentials.md`](security/developer-credentials.md) and
  [`security/developer-oauth.md`](security/developer-oauth.md)
- Provider registry and capabilities:
  [`integrations/providers.md`](integrations/providers.md)
- Durable inbound and outbound webhooks:
  [`integrations/webhook-gateway.md`](integrations/webhook-gateway.md) and
  [`integrations/outbound-webhooks.md`](integrations/outbound-webhooks.md)
- Public API and SDK ownership:
  [`integrations/public-api-v1.md`](integrations/public-api-v1.md) and
  [`integrations/public-sdks.md`](integrations/public-sdks.md)

## Operations

- ECS release, migration ordering, immutable images, and rollback:
  [`operations/ecs-release.md`](operations/ecs-release.md)
- AWS workload identity and permissions:
  [`operations/aws-identity.md`](operations/aws-identity.md)
- Traces, logs, safe fields, and incident correlation:
  [`operations/observability.md`](operations/observability.md)
- Database migration policy and recovery:
  [`database/migration-operations.md`](database/migration-operations.md)
- Provider and developer credential rotation:
  [`operations/provider-credential-rotation.md`](operations/provider-credential-rotation.md)
  and
  [`operations/developer-credential-rotation.md`](operations/developer-credential-rotation.md)

The repository intentionally does not provide a public self-hosting bundle,
Compose stack, or community support contract. The retained multi-target
Dockerfile, release workflow, generation scripts, and migration tooling serve
the managed internal application and its documented developer platform.
