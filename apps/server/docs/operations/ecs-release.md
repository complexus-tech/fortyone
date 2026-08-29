# API and worker ECS release

FortyOne ships the API and worker as private, immutable ECS images. The
production workflow is `.github/workflows/ecs-fargate-release.yml`; it is an
internal deployment contract, not a self-hosting distribution path.

## Release invariants

Every release uses one Git commit SHA for both images and follows this order:

1. run the required server quality and SQLC workflows;
2. publish the API and worker images to private ECR;
3. skip production database migrations unless the optional repository variable
   `RUN_PRODUCTION_MIGRATIONS` is exactly `true`;
4. when enabled, register a one-shot task definition using the new API image,
   run `/app/api -migrate` in the API service's existing VPC configuration, and
   require a zero exit code;
5. deploy the worker and wait for service stability, including its fail-closed
   provider credential and payload cutovers;
6. deploy the API only after the replacement worker is stable.

The worker-first binary order is deliberate for the current coordinated
credential migrations. The previous API can read the bounded legacy shapes,
whereas the replacement API is vault-only. Starting the replacement API before
the worker completes those cutovers could make valid installations unreadable.
When the bounded cutover code is removed, changing this order requires an
updated migration compatibility declaration and a staging exercise.

The workflow never runs a migration from a developer checkout and never uses a
floating image tag. Production migrations are disabled by default. While they
remain disabled, an operator must apply required migrations separately before
deploying code that depends on them. When explicitly enabled, the migration
task inherits the reviewed API task definition's environment, secrets,
execution role, task role, CPU, memory, and network boundary; only its image and
command are replaced.

## Required repository variables

The release requires:

- `AWS_DEPLOY_ROLE_ARN` and `AWS_REGION`;
- `ECR_SERVER_REPOSITORY` and `ECR_WORKER_REPOSITORY`;
- `ECS_CLUSTER`;
- `ECS_SERVER_SERVICE`, `ECS_SERVER_TASK_DEFINITION`, and
  `ECS_SERVER_CONTAINER`;
- `ECS_WORKER_SERVICE`, `ECS_WORKER_TASK_DEFINITION`, and
  `ECS_WORKER_CONTAINER`.

The referenced ECR repositories, ECS services, task definitions, container
names, and VPC configuration must already exist. A missing or ambiguous
container, absent network configuration, failed task start, missing exit code,
non-zero migration exit, or unstable service fails the release.

The `release-config` job validates every required repository variable before
the workflow requests an AWS identity. It reports missing variable names but
never prints their values. If `AWS_DEPLOY_ROLE_ARN` is absent, the OIDC action
cannot obtain credentials; configure the role instead of restoring long-lived
AWS access-key secrets.

## Schema and compatibility review

Before merging a migration, update `internal/migrations/manifest.json` and
regenerate `docs/database/migration-operations.md`. Review the entry's
schema-first or coordinated-cutover instructions against this workflow. A
migration that requires traffic draining, provider ingress suspension, a
pre-deploy backup, or a dedicated data job needs an explicit operator change
window; the generic one-shot schema task does not silently satisfy those
requirements.

Never edit an applied migration. Use a new forward migration and retain all
keys required to read existing credential envelopes or token digests.

## Failure and recovery

- **Image publication fails:** no service changes. Repair the build or registry
  contract and rerun the same commit.
- **Migration task fails when enabled:** no replacement service is deployed.
  Inspect the ECS task's structured logs and stopped reason without printing
  database URLs or secrets. Follow the migration manifest's forward-fix
  procedure.
- **Worker deployment fails:** do not deploy the API. Preserve the migrated
  schema and deploy a compatible worker forward fix; never restore a binary the
  manifest declares incompatible with encrypted or migrated rows.
- **API deployment fails:** keep the stable replacement worker, preserve the
  expanded schema, and redeploy a compatible API. Roll back an application only
  when the manifest explicitly says that version remains compatible.

The workflow uses `cancel-in-progress: false`, so two production releases do
not interleave their migration and service rollout phases.

## Verification

For workflow changes, run from `apps/server`:

```bash
make workflow-check
make migration-check
go test -race ./cmd/api ./internal/migrations/... ./internal/platform/deployment/...
```

Before relying on changed ordering or migration behavior in production,
exercise the exact task definitions and staged schema transition in the staging
AWS account. Local and static checks cannot prove IAM, VPC, secret injection,
load balancer health, or ECS replacement behavior.
