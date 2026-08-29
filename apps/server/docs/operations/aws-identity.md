# AWS deployment and runtime identity

FortyOne currently uses separate credentials for deployment and runtime AWS
access.

| Identity | Obtained by | Purpose |
| --- | --- | --- |
| Deployment identity | GitHub Actions secrets `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` | Update the existing API and worker ECS services |
| Runtime task role | ECS supplies short-lived credentials through the AWS SDK default chain | Access only the S3 buckets and object prefixes required by the API or worker |

The ECS **execution role** that pulls images and writes platform logs is a third
AWS-managed concern. Do not reuse it as the application task role merely
because both identities belong to the same task definition.

## Deployment credential contract

`.github/workflows/ecs-fargate-release.yml` uses the repository's existing AWS
access-key secrets together with `AWS_REGION`. The identity needs only the ECS
and task-definition permissions exercised by the workflow, including the exact
`iam:PassRole` permissions required by the existing task definitions.

The workflow does not run database migrations and does not require ECR
repositories or `AWS_DEPLOY_ROLE_ARN`. Images are built and published with the
existing Docker Hub credentials documented in
[`ecs-release.md`](ecs-release.md).

Long-lived deployment credentials have greater operational risk than short-lived
workload identity. Keep the IAM policy least-privileged, rotate the key pair,
remove unused keys, restrict repository administration, and never print secret
values in logs or task overrides.

Every third-party workflow action is pinned to a reviewed 40-character commit.
The release comment beside each pin records its human-readable version. Update
the commit and version together after reviewing upstream release notes,
permissions, inputs, and output behavior.

## Runtime S3 contract

In staging and production, leave `APP_AWS_ACCESS_KEY_ID` and
`APP_AWS_SECRET_ACCESS_KEY` unset. `pkg/aws` then uses the AWS SDK default
credential chain, which resolves the ECS task-role credentials and refreshes
them before expiry. The task role grants only the required object actions on
the configured profile, logo, and attachment bucket prefixes; it does not gain
ECS deployment permissions.

Production startup rejects static `APP_AWS_*` credentials even when both are
present. All environments reject a partial pair. Development may set a complete
pair for an approved local S3-compatible endpoint, together with
`APP_AWS_ENDPOINT` and `APP_AWS_FORCE_PATH_STYLE` when required.

## Verification and recovery

Run these checks after changing workflow or AWS credential behavior:

```bash
make workflow-check
go test -race ./internal/bootstrap/architecture ./internal/platform/deployment ./pkg/aws
```

If deployment authentication fails, verify that both repository secrets exist,
the access key is active, the region is correct, and the IAM policy covers only
the expected ECS resources. If runtime S3 access fails, inspect the ECS task
role, bucket policy, region, endpoint, and object-prefix permissions. Do not put
temporary credentials in logs, task overrides, tickets, or shell history.
