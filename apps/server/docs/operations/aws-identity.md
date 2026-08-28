# AWS workload identity

FortyOne uses two separate short-lived AWS identities. Neither identity is an
application secret and neither is represented by a long-lived access-key pair
in GitHub or a production task environment.

| Identity          | Obtained by                                                | Purpose                                                                                                       |
| ----------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Release role      | GitHub Actions OIDC assumes `AWS_DEPLOY_ROLE_ARN`          | Build and publish private ECR images, run the one-shot migration task, and deploy the API/worker ECS services |
| Runtime task role | ECS supplies credentials through the AWS SDK default chain | Access only the S3 buckets and object prefixes required by the API or worker                                  |

The ECS **execution role** that pulls images and writes platform logs is a third
AWS-managed concern. Do not reuse it as the application task role merely because
both identities belong to the same task definition.

## Release-role contract

`.github/workflows/ecs-fargate-release.yml` requests `id-token: write` only on
jobs that call AWS. `configure-aws-credentials` exchanges that token for a
bounded role session. The workflow does not accept `AWS_ACCESS_KEY_ID` or
`AWS_SECRET_ACCESS_KEY` GitHub secrets.

Configure these GitHub repository variables:

- `AWS_DEPLOY_ROLE_ARN`;
- `AWS_REGION`;
- `ECR_SERVER_REPOSITORY` and `ECR_WORKER_REPOSITORY`;
- the existing ECS cluster, service, task-definition, and container variables.

The IAM trust policy must require the GitHub token audience used by AWS STS and
an exact subject for this repository's protected `main` release ref. Do not use
an organization-wide or wildcard-repository subject. The permissions policy is
restricted to the two ECR repositories, the API and worker services, and their
task-definition families. It permits the API family to be registered and run
as a one-shot migration task, and permits describing that task until it stops.
Any `iam:PassRole` permission is restricted to the exact ECS execution and task
roles referenced by those task definitions.

The migration task reuses the server service's `awsvpc` configuration and task
secrets; it overrides only the reviewed image and command. The release role
must not be able to inject arbitrary environment secrets, choose unrelated task
roles, or run task families outside this contract. See
[`ecs-release.md`](ecs-release.md) for sequencing and recovery.

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
ECR or ECS deployment permissions.

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

The architecture suite rejects mutable third-party action references and
long-lived AWS release inputs. Deployment tests enforce the production
task-role rule, and S3 tests prove both default-chain and paired local-static
selection without contacting AWS.

If OIDC assumption fails, repair the trust subject, audience, repository
environment protection, or role ARN; never restore a long-lived repository
secret as a workaround. If runtime S3 access fails, inspect the ECS task role,
bucket policy, region, endpoint, and object-prefix permissions. Roll back the
task definition or IAM policy independently; do not place temporary credentials
in logs, task overrides, tickets, or shell history.
