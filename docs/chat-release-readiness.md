# Messages release readiness

This is a release runbook, not a production deployment record. Local browser and
database verification is recorded in [preview-cache-e2e.md](preview-cache-e2e.md).
Production promotion remains a separate action after the gates below pass.

## Release order and database policy

Production migrations remain an explicitly reviewed operator procedure. A push to
`main` does **not** apply migrations. The release workflow builds the API and worker
images and retains the existing mandatory runtime vulnerability gate unchanged.
It then:

1. Submits the API rollout without waiting for service stability.
2. Submits the worker rollout and waits for worker service stability.
3. Leaves web promotion to the existing Vercel project. Before promoting Messages,
   operators verify schema compatibility and the health of the intended API and
   worker images through their existing release procedure.

The normal workflow does not launch a standalone schema task or invoke
`.github/scripts/ecs_release.py` for schema or candidate-image verification. It
uses the existing deployment permissions; no standalone-task or cleanup
permissions are required for ordinary releases. Successful workflow completion
does not prove API stability or detect a rollback to an older image. See the
[ECS release contract](../apps/server/docs/operations/ecs-release.md) for the
workflow's exact scope.

## Optional ECS operator tooling

The Python helper remains available for manually chosen schema checks and reviewed
migrations. It is not required by the normal release workflow and does not replace
an existing reviewed operator procedure. A manual `check-schema` invocation starts
the candidate API image once using the API service's private network configuration,
task role, execution role, and environment. It requires a clean migration version
that exactly matches the candidate's embedded migration head. The check executes
a read-only transaction; it cannot create a migration table, apply SQL, repair a
dirty migration, or advance a version.

The helper's default task polling budget is five minutes, with separately bounded
AWS calls and a 30-second database check. Its HTTP health check is removed because
the schema command does not start an HTTP server. The helper stops its own task on
timeout/cancellation and deregisters its temporary task-definition revision. It
never changes an ECS service or desired count. A failed check or cleanup makes the
manual command fail; the normal deployment workflow does not invoke this gate.

An operator choosing this helper needs ECS `DescribeServices`,
`DescribeTaskDefinition`, `RegisterTaskDefinition`, `RunTask`, `DescribeTasks`,
`StopTask`, and `DeregisterTaskDefinition`, plus `iam:PassRole` for the API task and
execution roles. These permissions belong to the identity running the helper and
are separate from the requirements for ordinary deployment. Review the helper's
[resource scope requirements](../apps/server/docs/operations/ecs-release.md#optional-operator-tooling)
before using it. Its task uses on-demand Fargate. Confirm private subnets can reach
PostgreSQL, image storage, secrets, and logging services.

## Reviewed migration procedure

Before applying SQL, record the current database version and dirty flag, verify a
recoverable database backup, and review every pending entry in the
[migration operations contract](../apps/server/docs/database/migration-operations.md).
Never use a force-version command to bypass a failure.

Messages uses migrations 229–231; the combined branch also contains 232 for Maya
reply-seen notifications, 233 for attachment thumbnails, and 234 for the default-on
workspace Chat setting. Those entries declare
schema-first rollout. Verify the actual pending range: a production database behind
228 may require an earlier coordinated cutover and must follow that entry's
specific procedure. The candidate binary determines the required version; this
document does not substitute for inspecting the current migration head.

The optional one-off helper below can apply reviewed migrations without deploying
either service. Use it only when it is the chosen operator procedure and its
permissions and network access have been prepared.
Use the same commit-tagged API image that passed the existing image gate. Prepare a
private copy of the current API task definition and replace **only** the named API
container's image with that candidate image. The file contains environment settings
and secret references; do not upload it as a workflow artifact or paste it in logs.

```sh
# Set these to the reviewed deployment's existing values. The image must identify
# the approved commit; never use a floating latest tag.
umask 077
aws ecs describe-task-definition \
  --task-definition "$ECS_SERVER_TASK_DEFINITION" \
  --query taskDefinition > candidate-server.json

python3 - "$ECS_SERVER_CONTAINER" "$CANDIDATE_API_IMAGE" <<'PY'
import json, pathlib, sys
path = pathlib.Path("candidate-server.json")
definition = json.loads(path.read_text())
containers = [c for c in definition["containerDefinitions"] if c["name"] == sys.argv[1]]
if len(containers) != 1:
    raise SystemExit("Expected exactly one API container")
containers[0]["image"] = sys.argv[2]
path.write_text(json.dumps(definition))
PY

python3 .github/scripts/ecs_release.py migrate \
  --cluster "$ECS_CLUSTER" --service "$ECS_SERVER_SERVICE" \
  --container "$ECS_SERVER_CONTAINER" \
  --task-definition candidate-server.json --run-id "$RELEASE_RUN_ID" \
  --timeout 2100 --execute-reviewed-migrations

python3 .github/scripts/ecs_release.py check-schema \
  --cluster "$ECS_CLUSTER" --service "$ECS_SERVER_SERVICE" \
  --container "$ECS_SERVER_CONTAINER" \
  --task-definition candidate-server.json --run-id "$RELEASE_RUN_ID-check"
```

Delete the local candidate JSON after recording the task ARN and reviewed outcome.
The helper's task ARN points to CloudWatch logs for diagnosis. If a network failure
makes task launch or stop uncertain, inspect ECS before retrying a migration. Do
not start a second migrator blindly. Do not deploy either service after a failed
migration; repair forward under the migration contract. Disabling Messages is the
first feature rollback action; preserve conversations, originals, and queued work.
An older application image against a newer schema requires explicit compatibility
review, so the helper's read-only check deliberately rejects that mismatch.

## Web and worker packaging

Vercel remains the production web-build authority, as specified in
[Projects README](../apps/projects/README.md). Do not add a second deployment or
publish a standalone web image. Its Node runtime must satisfy PDF.js 6.4.299
(Node 22.13 or newer); the repository CI uses Node 22.23.2. The Projects build
prepares its matching worker, CMaps, fonts, WASM, and license in
`public/pdfjs/6.4.299/`. Verify the generated public assets exist in the Vercel
deployment and the matching PDF worker URL is served before promotion.

A clean local build also needs the existing shared UI outputs: run
`pnpm --filter ui build:styles` and `pnpm --filter ui build:components` before
`pnpm --filter projects build`. Existing local `dist` files can otherwise hide this
dependency. Confirm the managed hosting build configuration prepares these outputs.
The worker Docker target includes `poppler-utils`; check `pdftoppm` in that actual
image. API, worker, Redis, and object storage settings must agree across replicas.
The cache-only S3 lifecycle rule does not expire private attachment originals or
private thumbnails; their authorized retirement continues through the worker.

## Staging and pilot acceptance

Messages is enabled by default for both existing and new workspaces. Follow the
[workspace settings contract](../apps/server/docs/chat-contract.md#workspace-rollout):
workspace administrators can turn Chat off or on alongside Objectives in
**Workspace Settings → General → Features**. The API and worker enforce the
database setting without a process restart. The previous environment switch and
workspace allowlist are no longer used. Do not assume the first release is
restricted to a pilot: explicitly disable non-pilot workspaces if a restricted
rollout is intended. Turning Chat off preserves conversations and pending work.

Verify two
independently authenticated users and at least two API replicas through the actual
load balancer: sending, reconnecting SSE, Direct Message membership isolation,
mention/reply notifications, editing, deleting, and revoked access. Exercise worker
restart/retry and inspect outbox age, failed work, and cleanup outcomes.

Follow [Chat upload validation](../apps/server/docs/chat-file-scanning.md):
attachments and voice notes require no scanner service or environment variable.
Verify normal file and recorded-audio uploads, file-type/size/duration rejection,
private storage access, and conversation permissions through the deployed API.
New uploads remain `unscanned`; existing `pending`, `infected`, and `failed` states
remain blocked. No malware scan runs as part of uploading.

Test upload/download authorization, PDF thumbnails/viewer, image galleries,
repeated-link cache reuse, Refresh preview, and task/document edit invalidation
through the deployed web/API/worker. Check real microphone capture, playback, and
explicit transcription on supported desktop and mobile browsers, including Safari.
Verify actual notification delivery with the intended provider and user preferences.

Run the representative capacity scenario against the intended staging resources,
starting at the planned 1,000 concurrent users. Record SSE reconnects, API latency,
database pools, Redis load, queue age, memory, storage bandwidth, and model demand.
A local passing functional test or synthetic load result is not a production
capacity guarantee. Keep the pilot limited until these checks and budget/alert
thresholds are accepted.

## Local validation of release controls

### Scanner dependency removal

The later upload-policy change removes ClamAV initialization, its configuration
variable, and the clean-scan prerequisite. Chat now shares the existing attachment
policy: `unscanned` and `clean` are available, while `pending`, `infected`, and
`failed` stay blocked. File validation, authorization, private storage, transcription,
and cleanup remain in place. No database migration or scan-state backfill is needed.

The full Go suite, configuration and architecture checks, and disposable PostgreSQL
attachment/transcription regressions pass. Ten focused frontend suites pass all 63
tests, with Projects type checking and scoped lint also passing. A live local
browser check uploaded a text file and generated audio while ClamAV was stopped
and the API had no Chat environment variables. Both records remained `unscanned`
with no scan timestamp. Existing voice-note play/pause worked. The temporary upload
draft was cleared without sending a message; the development API remains running
without a scanner. Real microphone capture was not repeated for this change.

The capacity reports below identify the earlier measured candidates, before this
upload-policy change. Their source fingerprints are retained as historical
evidence rather than represented as the current working tree.

### Workspace setting follow-up

The Chat rollout environment variables were removed after the earlier release
checks below. Migration 234 now adds `workspace_settings.chat_enabled`, enabled
by default, using the existing settings API, table, and Features UI.

The updated full Go test suite, configuration/migration/architecture checks, and
real PostgreSQL regressions pass. Database tests cover default-on legacy and new
settings, read-only membership lookups, live disable/re-enable, queued work,
existing SSE streams, Maya authorization, and concurrent settings changes. Stale
settings saves return HTTP 409 instead of accidentally overwriting a newer Chat
toggle. Six focused frontend suites pass all 42 tests, with scoped lint passing.
The final full Projects type check passes. An earlier check encountered an
unrelated document-test type error that was subsequently resolved in the parallel
document work.

Migration 234 was applied only to the local database. A live browser check verified
Chat enabled beside Objectives and Key results, disabled navigation and direct
route access after switching it off, then restored access after switching it on.
Chat is left enabled. No production configuration, schema, or service was changed.

The fresh five-minute load regression passed against the database-setting and
optimistic-save implementation with unchanged budgets. It held 1,000 authenticated
SSE users, including 200 active Messages viewers: 5,999 periodic history reads,
1,499 top-level messages, 59 replies, 299 reactions and 136,720 reactive HTTP reads
during the hold. All 1,000 streams stayed connected with zero unexpected errors.
Create p95/p99 was 200/441ms, history 128/284ms and steady SSE delivery 854/1,102ms.
Every final recipient probe passed, and deliberate outbox retry recovered in 7.5s.
Both API replicas immediately enforced database disable/re-enable without restart;
the shared platform streams remained connected. Peak Redis usage was 1,073 clients
and 27.5MB, with 239.2MB combined API/generator Go heap. The development watcher and
scanner were restored afterward.

The [new result and exact source fingerprints](../apps/server/docs/performance/results/messages-1000-workspace-features.md)
were verified against the uncommitted working tree at the time of that run. The test used isolated
loopback PostgreSQL/Redis and two API instances in one Go process, with 20% active
Messages viewers. It does not certify deployed capacity or revalidate the earlier
candidate's synchronized-start diagnostic. All older reports remain preserved.

### Earlier release-hardening candidate

The following checks passed locally on October 17, 2026. They establish build and
release-control readiness; they do not establish deployed infrastructure capacity
or production readiness on their own.

The expanded load scenario exposed contention in request activity writes and
exhausted immediate serialization retries. After addressing those paths, the
complete Go quality pipeline and both production images were rebuilt successfully,
including container smoke checks. The final five-minute local load test and a
separate synchronized-start diagnostic then passed against the same uncommitted
working-tree candidate before the switch moved into workspace settings. Their
runtime and harness SHA256 fingerprints match each other, but do not identify the
later default-on database-setting implementation;
see the [measured results and retained failures](../apps/server/docs/performance/results/messages-1000-final.md).

The representative test held 1,000 authenticated SSE users, of whom 200 actively
viewed Messages, for five minutes across two API instances in one Go process.
All 1,000 remained connected, with zero unexpected errors: 5,999 periodic history
reads, 1,499 top-level messages, 59 replies, 299 reactions and 137,840 reactive HTTP
reads during the hold. Create p95/p99 was 150/263ms, history 117/213ms and steady
SSE delivery 862/1,077ms. Every final recipient probe passed; a deliberately failed
outbox publication retried and recovered in 9s. The final 30-second synchronized
800-GET startup diagnostic also passed without transport resets or HTTP errors.
The unchanged acceptance budgets and exact operation counts are in the reports.

This was a macOS/arm64 loopback test with disposable PostgreSQL/Redis, not the
deployed Linux images or production proxy. Peak Redis usage was 1,073 clients
and 27.6MB; combined API/generator Go heap peaked at 240.3MB. It does not establish
capacity for 1,000 simultaneously active chat viewers, larger rooms, attachments,
voice/AI jobs, browser rendering or production failover. The staging and limited
workspace-pilot checks above remain necessary.

- The production Projects build completed compilation, TypeScript validation, and
  page generation using an isolated checkout snapshot. The generated PDF.js worker
  matches the pinned dependency; CMaps, fonts, WASM, and license assets are present.
- Thirty focused frontend suites passed all 193 tests, with Projects type checking
  and scoped lint passing. Local browser checks verified hidden navigation and a
  blocked direct route when Messages is disabled, an enabled pilot workspace, and
  hidden attachment/recording controls when the scanner is not configured. After
  both final load runs, the development API and private scanner were restored;
  a final live browser check confirmed the enabled pilot's Chat and
  attachment/recording controls without further runtime source changes.
- A local ClamAV 1.5.4 instance using freshly verified official signature databases
  accepted a clean sample and rejected the harmless EICAR test sample through the
  real INSTREAM integration. This validates the local scanner component, not an
  end-to-end deployed upload, object-storage, or microphone workflow.
- Linux/amd64 API and worker compilation and both production Docker targets passed.
  Runtime smoke checks confirmed the API exposes the schema-check command and the
  worker runs as user 10001 with `pdftoppm` available.
- The Python release helper's 12 offline tests cover successful preflight, nonzero
  or missing task exit, launch failure, timeout/cancellation cleanup, failed-stop
  cleanup, migration opt-in, zero desired tasks, incomplete rollouts, and completed
  rollback detection.
- Go unit tests verify exact clean schema matching and embedded migration discovery.
- PostgreSQL integration tests execute against a temporary PostgreSQL 18 database,
  checking the current head plus pending, dirty, newer, and missing-table cases.
  Failed preflight leaves the migration state unchanged and never creates a table.
- The actual mandatory server `make test` passed across all packages, including
  the complete architecture gate. Earlier architecture-debt failures recorded in
  the implementation notes were resolved with shared story request contracts and
  cohesive file splits; the debt baseline was not relaxed.
- Formatting, module verification, tidy-diff, full Go vet, migration-manifest
  validation, and workflow action validation passed. No AWS deployment or
  production migration was performed by these local validations.

AWS references: [RunTask command overrides and network configuration](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_RunTask.html),
[task completion waiter](https://docs.aws.amazon.com/cli/latest/reference/ecs/wait/tasks-stopped.html).
