# API and worker ECS release

FortyOne builds the API and worker as commit-tagged Docker Hub images and
deploys them to the existing ECS services. The production workflow is
`.github/workflows/ecs-fargate-release.yml`; it is an internal deployment
contract, not a self-hosting distribution path.

## Release sequence

Every release uses the triggering Git commit SHA for both images:

1. run the required server quality and SQLC workflows;
2. build and scan the API and worker images;
3. publish `fortyoneapp/server:<commit-sha>` and
   `fortyoneapp/worker:<commit-sha>` to Docker Hub;
4. deploy the API task definition and wait for service stability;
5. deploy the worker task definition and wait for service stability.

The workflow does not run production database migrations. The typed-database
workflow still applies the migration chain to its disposable PostgreSQL
container so SQLC generation and queries are validated without changing
production data.

Apply required production migrations separately before deploying code that
depends on a new schema. Never edit an applied migration; add a forward
migration and follow the compatibility instructions in
`internal/migrations/manifest.json`.

## Existing GitHub configuration

The release uses these repository secrets:

- `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`;
- `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

It uses these repository variables:

- `AWS_REGION` and `ECS_CLUSTER`;
- `ECS_SERVER_SERVICE`, `ECS_SERVER_TASK_DEFINITION`, and
  `ECS_SERVER_CONTAINER`;
- `ECS_WORKER_SERVICE`, `ECS_WORKER_TASK_DEFINITION`, and
  `ECS_WORKER_CONTAINER`.

The referenced Docker Hub repositories, ECS services, task definitions, and
container names must already exist. The workflow never publishes a `latest`
tag; task definitions receive the exact commit-tagged image.

## Failure and recovery

- **Quality, SQLC, build, or scan fails:** no service changes.
- **API deployment fails:** the worker is not deployed. Repair the API release
  or restore a schema-compatible task definition.
- **Worker deployment fails:** the stable API remains deployed. Repair the
  worker using an image compatible with the deployed API and schema.

The workflow uses `cancel-in-progress: false`, so two production releases do
not interleave their ECS deployments.

## Verification

Run from `apps/server` after workflow changes:

```bash
make workflow-check
go test ./internal/bootstrap/architecture
```

Static checks cannot prove Docker Hub access, AWS credential validity, ECS
permissions, task health, or production schema compatibility. Verify those
boundaries in the deployment environment.
