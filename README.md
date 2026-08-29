# FortyOne

FortyOne is Complexus's hosted project-management platform. This repository is
the private monorepo for the web, mobile, documentation, administration, API,
and worker applications that power the managed FortyOne service.

This repository contains no public installation or distribution bundle.
Production infrastructure, secrets, release access, and operational procedures
are managed internally.

## Repository structure

| Path                 | Purpose                                     |
| -------------------- | ------------------------------------------- |
| `apps/projects`      | Main Next.js application                    |
| `apps/landing`       | Marketing site                              |
| `apps/docs`          | Customer documentation site                 |
| `apps/admin`         | Internal administration application         |
| `apps/mobile`        | Expo mobile application                     |
| `apps/server`        | Go API and background worker                |
| `packages/ui`        | Shared React components and styles          |
| `packages/lib`       | Shared TypeScript utilities                 |
| `packages/icons`     | Shared icon library                         |
| `deployments/docker` | Production API and worker image definitions |

The JavaScript workspaces are managed with pnpm and Turborepo. The API is a Go
module with domain packages under `apps/server/internal/modules`.

## Toolchain

- Node.js 18 or newer
- pnpm 9.3.0
- Go 1.25.0
- PostgreSQL and Redis endpoints approved for development
- `air` for Go live reload

Docker is only required when building the production API or worker images, or
when using an internally approved container-based dependency environment. The
repository does not provide a supported full-stack Compose installation.

## Internal development

Install JavaScript dependencies from the repository root:

```bash
pnpm install
```

Create local environment files only for the applications you are running:

```bash
cp apps/projects/.env.example apps/projects/.env
cp apps/server/.env.example apps/server/.env
```

Obtain development credentials and service endpoints through the internal
team process. Never commit populated environment files, access tokens, private
keys, customer data, or production credentials.

Run the Projects app and API together:

```bash
make dev-projects
```

Run all JavaScript applications and the API:

```bash
make dev
```

The API and worker can also be run independently from `apps/server`:

```bash
make dev
make worker
```

See [`apps/server/README.md`](apps/server/README.md) for database migrations,
SQLC, seeding, and backend architecture.

## Verification

Run checks from the repository root when changing multiple workspaces:

```bash
pnpm lint
pnpm build
```

For focused application checks, use the commands documented in `AGENTS.md`.
For server changes, run the appropriate targets from `apps/server`:

```bash
make generated-check
make test
go vet ./...
```

## Production delivery

The Projects, landing, docs, and admin applications are deployed through their
managed hosting projects. The API and worker are built from
`deployments/docker/dockerfile.server` and
`deployments/docker/dockerfile.worker`, then deployed to the internal ECS
services by `.github/workflows/ecs-fargate-release.yml`.

The release workflow publishes immutable commit-SHA images to private Amazon
ECR repositories named by the `ECR_SERVER_REPOSITORY` and
`ECR_WORKER_REPOSITORY` GitHub repository variables. Both repositories must
exist in the configured AWS account, and the ECS execution roles must be able
to pull from them. The workflow fails closed when this private registry contract
is incomplete.

GitHub authenticates to AWS with a short-lived OIDC session by assuming the IAM
role named by `AWS_DEPLOY_ROLE_ARN`; repository access-key secrets are neither
required nor accepted by the workflow. Restrict that role's trust policy to this
repository's protected `main` release subject. Third-party workflow actions are
pinned to immutable commits and enforced by the API architecture suite. The
release sequencing and ECS identity contracts are documented in
[`apps/server/docs/operations/ecs-release.md`](apps/server/docs/operations/ecs-release.md)
and
[`apps/server/docs/operations/aws-identity.md`](apps/server/docs/operations/aws-identity.md).

Do not publish application images, source archives, installers, or environment
templates as public release assets. Production changes must use the reviewed
CI/CD path and the deployment platform's secret management.

## Architecture and engineering plans

- [`docs/plans/api-modernization/01-target-go-architecture.md`](docs/plans/api-modernization/01-target-go-architecture.md)
- [`docs/plans/api-modernization/02-typed-data-security-and-integration-platform.md`](docs/plans/api-modernization/02-typed-data-security-and-integration-platform.md)
- [`docs/plans/api-modernization/03-delivery-testing-and-documentation-roadmap.md`](docs/plans/api-modernization/03-delivery-testing-and-documentation-roadmap.md)
- [`apps/server/docs/database/sqlc.md`](apps/server/docs/database/sqlc.md)

## Security

Security concerns, leaked credentials, suspected customer-data exposure, and
production incidents must be handled through the internal process in
[`SECURITY.md`](SECURITY.md). Do not open public issues or paste sensitive
material into chat, logs, test fixtures, or pull requests.
