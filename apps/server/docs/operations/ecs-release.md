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
4. run the candidate API image once with `-check-schema` and require success;
5. submit the API rollout, wait for stability, and verify the candidate image;
6. submit the worker rollout, wait for stability, and verify its candidate image.

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

The GitHub deployment identity also needs the schema helper's permissions:
`ecs:DescribeServices`, `ecs:DescribeTaskDefinition`,
`ecs:RegisterTaskDefinition`, `ecs:RunTask`, `ecs:DescribeTasks`, `ecs:StopTask`,
and `ecs:DeregisterTaskDefinition`, with `iam:PassRole` for the API task and
execution roles. These permissions belong to the identity used by GitHub Actions,
not to the container's task role. Keep the existing service deployment permissions.

The helper registers a temporary family named `<server-family>-schema` (the base
family is truncated to 240 characters). Scope `RunTask` to this family and the
intended cluster. Review each action's supported resources in the
[AWS ECS authorization reference](https://docs.aws.amazon.com/service-authorization/latest/reference/list_ecs.html):
`DeregisterTaskDefinition` does not support task-definition resource scoping and
requires `Resource: "*"` for that specific action. A grant limited to the normal
server task-definition ARN does not authorize the entire preflight. This
repository does not configure or change the deployment identity's IAM policy.

## Failure and recovery

- **Build or scan fails:** no service changes.
- **Schema check or cleanup fails:** neither service is deployed. Read the
  original `Release stopped` error and each separate `Schema cleanup failed`
  diagnostic. Cleanup errors cannot replace the original launch/check failure,
  and cleanup failure alone still blocks release. Inspect the temporary task and
  revision before retrying if stopping or deregistration failed.
- **API rollout submission fails:** the worker is not deployed. Repair the API
  release or restore a schema-compatible task definition.
- **Worker rollout submission fails:** the API rollout has already been
  submitted. Repair the worker using an image compatible with the deployed API
  and schema.

Superseded image builds are cancelled so only the latest commit reaches the
deployment queue. The consolidated deployment job is never cancelled while it
is submitting the ordered API and worker rollouts. A successful workflow requires
both services to be stable on their candidate images; a rollback to an older
image does not pass verification.

AWS CLI exit `254` means the service returned an error; it does not identify
the cause. The helper reports the AWS error code when available, such as
`AccessDeniedException`, and supplies the relevant IAM action for access errors.
It deliberately excludes raw AWS error messages because they can contain task
environment values. Use CloudTrail for the corresponding operation, timestamp,
and deployment identity to inspect the underlying rejection privately. See
[AWS CLI return codes](https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-returncodes.html).
For an access denial, review the indicated action, `iam:PassRole` where needed,
resource scope, permissions boundaries, and organization policies before retrying.

## Verification

Run from `apps/server` after workflow changes:

```bash
make workflow-check
go test ./internal/bootstrap/architecture
```

Static checks cannot prove Docker Hub access, AWS credential validity, ECS
permissions, task health, or production schema compatibility. Verify those
boundaries in the deployment environment.
