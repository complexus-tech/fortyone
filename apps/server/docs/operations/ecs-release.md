# API and worker ECS release

FortyOne builds the API and worker as commit-tagged Docker Hub images and
deploys them to the existing ECS services. The production workflow is
`.github/workflows/ecs-fargate-release.yml`; it is an internal deployment
contract, not a self-hosting distribution path.

## Release sequence

Every release uses the triggering Git commit SHA for both images. Pull-request
checks are the merge gate; the release does not repeat them after merge:

1. build and publish `fortyoneapp/server:<commit-sha>` and
   `fortyoneapp/worker:<commit-sha>` through one shared Docker builder;
2. scan both published runtime images;
3. submit the API rollout without waiting for service stability;
4. submit the worker rollout and wait for worker service stability.

The workflow does not launch a standalone schema task, invoke the release helper,
or verify that either service is running the candidate image. It does not run
production database migrations. Weekly assurance
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

Normal deployments use the existing service deployment permissions. The workflow
does not require the standalone-task or cleanup permissions used by the optional
operator helper below. This repository does not configure or change the deployment
identity's IAM policy.

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
the API rollout was submitted and the worker stability wait completed. It does not
prove that the API stabilized or that either candidate image remained active after
an automatic rollback. Inspect those outcomes through the existing ECS operations
process.

## Optional operator tooling

`.github/scripts/ecs_release.py` is available for explicitly chosen manual schema
checks, reviewed migrations, and candidate-image verification. The normal release
workflow does not call it. See the
[reviewed migration procedure](../../../../docs/chat-release-readiness.md#reviewed-migration-procedure)
for one optional way to run a migration; existing reviewed operator procedures
remain valid.

Running the helper's schema commands requires its own operator identity permissions:
`ecs:DescribeServices`, `ecs:DescribeTaskDefinition`,
`ecs:RegisterTaskDefinition`, `ecs:RunTask`, `ecs:DescribeTasks`, `ecs:StopTask`,
and `ecs:DeregisterTaskDefinition`, with `iam:PassRole` for the API task and
execution roles. These are permissions of the identity invoking the helper, not
the container's task role. They are not additional requirements for ordinary
GitHub deployments.

The helper registers a temporary family named `<server-family>-schema` (the base
family is truncated to 240 characters). Scope `RunTask` to this family and the
intended cluster. Review each action's supported resources in the
[AWS ECS authorization reference](https://docs.aws.amazon.com/service-authorization/latest/reference/list_ecs.html):
`DeregisterTaskDefinition` does not support task-definition resource scoping and
requires `Resource: "*"` for that specific action. A grant limited to the normal
server task-definition ARN does not authorize every helper operation.

If a manual helper command fails, read the original `Release stopped` error and
each separate `Schema cleanup failed` diagnostic. Cleanup errors cannot replace
the original launch/check failure, and cleanup failure alone still makes the
command fail. Inspect the temporary task and revision before retrying if stopping
or deregistration failed. The helper does not deploy either service.

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
