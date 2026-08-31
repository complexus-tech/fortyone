# API and worker ECS release

FortyOne builds the API and worker as commit-tagged Docker Hub images and
deploys them to the existing ECS services. The production workflow is
`.github/workflows/ecs-fargate-release.yml`; it is an internal deployment
contract, not a self-hosting distribution path.

## Release sequence

Every release uses the triggering Git commit SHA for both images. Pull-request
checks are the merge gate; the release does not repeat them after merge:

1. build the API and worker targets through one shared Docker builder;
2. scan both runtime images;
3. publish `fortyoneapp/server:<commit-sha>` and
   `fortyoneapp/worker:<commit-sha>` to Docker Hub;
4. submit the API task definition rollout;
5. submit the worker task definition rollout.

The workflow does not run production database migrations. Weekly assurance
applies the migration chain to disposable PostgreSQL, vets SQLC against that
schema, and runs the tagged database and Redis suite without changing
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

- **Build or scan fails:** no service changes.
- **API rollout submission fails:** the worker is not deployed. Repair the API
  release or restore a schema-compatible task definition.
- **Worker rollout submission fails:** the API rollout has already been
  submitted. Repair the worker using an image compatible with the deployed API
  and schema.

Superseded image builds are cancelled so only the latest commit reaches the
deployment queue. The consolidated deployment job is never cancelled while it
is submitting the ordered API and worker rollouts. A successful workflow means
ECS accepted both rollouts; service stability and automatic rollback remain ECS
operational responsibilities.

## Verification

Run from `apps/server` after workflow changes:

```bash
make workflow-check
go test ./internal/bootstrap/architecture
```

Static checks cannot prove Docker Hub access, AWS credential validity, ECS
permissions, task health, or production schema compatibility. Verify those
boundaries in the deployment environment.
