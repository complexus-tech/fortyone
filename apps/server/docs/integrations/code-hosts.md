# Code-host integration architecture

This document is the implementation contract for first-party code-host
adapters. It covers the provider-neutral capability boundary, GitHub's durable
production ingress, and the deliberately narrow GitLab proof.

## Delivery status

| Area                        | GitHub                                    | GitLab                                                                |
| --------------------------- | ----------------------------------------- | --------------------------------------------------------------------- |
| Installation authentication | GitHub App installation token             | Injected OAuth or private-token source                                |
| Repository catalog          | Production                                | Proof, keyset-paginated                                               |
| Work-item write             | GitHub issue                              | GitLab issue proof                                                    |
| Comment write               | Issue comment                             | Issue-note proof                                                      |
| Webhook normalization       | Existing supported GitHub events          | Issue and issue-note only                                             |
| Durable ingress             | API and worker wired                      | Reusable runtime proof, not API/worker wired                          |
| Current-grant worker fence  | Production                                | Required before production wiring                                     |
| Provider persistence        | GitHub installation/repository SQLC fence | None; do not invent persistence until installation design is approved |

“Proof” is intentional. The GitLab adapter can be exercised through its typed
contracts and tests, but there is no public route, OAuth callback, scheduler,
or worker registration. Describing it as production-enabled would be unsafe.

## Package map

- `internal/platform/codehost` contains handwritten provider-neutral values and
  five narrow ports. It never imports a provider SDK.
- `internal/platform/webhooks` owns the encrypted durable inbox, immutable
  receipt identity, deduplication, queue-by-ID, processing leases, recovery,
  retention, and bound payload codec.
- `internal/modules/github/service` owns GitHub translation and orchestration.
  Capability files are named by reason to change: installation, settings,
  comments, story sync, pull requests, reviews, workflow automation, API client,
  webhook ingress, and webhook processing.
- `internal/modules/github/service` also owns the narrow persistence, story,
  and integration-request ports it consumes. It imports neither SQL adapters
  nor sibling module services. `internal/bootstrap/githubadapter` performs the
  sibling-type translations at the composition root.
- `internal/modules/github/repository` owns all GitHub persistence. Named SQLC
  queries execute through the process-owned native pgx pool; generated
  parameters and rows remain inside the repository adapter.
- `internal/modules/gitlab` is the narrow adapter proof. Provider JSON and HTTP
  details never cross into `internal/platform/codehost`.
- `internal/bootstrap/providers` is discovery metadata. It is not runtime
  wiring and must not be treated as proof that an adapter is publicly enabled.

## Capability contracts

The consumer asks for the smallest behavior it needs:

| Port                        | Responsibility                                                | Does not own                                 |
| --------------------------- | ------------------------------------------------------------- | -------------------------------------------- |
| `InstallationAuthenticator` | Prove that a current installation credential can authenticate | OAuth UI or credential persistence           |
| `RepositoryCatalog`         | Return neutral repositories and an opaque next cursor         | Provider-specific pagination structs         |
| `WorkItemWriter`            | Create one issue-like work item                               | Story business rules                         |
| `CommentWriter`             | Add one comment to an existing work item                      | FortyOne comment synchronization policy      |
| `WebhookNormalizer`         | Convert one authenticated provider payload to a neutral event | Signature verification or durable processing |

Reconciliation shares provider-neutral `WorkItemSnapshot`, `SyncOrigin`, and
revision/echo-suppression primitives. GitHub now uses that canonical newline-
stable revision instead of owning a provider-local hash. The persistence of
mappings, source revisions, and conflict policy remains in the product module;
it is deliberately not pushed into provider adapters or a god interface.

Every call includes an `InstallationRef` with provider, workspace, internal
installation ID, external installation ID, and authorization generation. A
provider mismatch, zero tenant identity, missing generation, invalid cursor, or
malformed resource is rejected before network I/O.

Capabilities are explicit. Call `Capabilities().Require(...)` before selecting
an adapter. Unsupported behavior returns `codehost.ErrCapabilityUnsupported`;
callers must not silently emulate it with a different provider concept.

## GitHub durable ingress

The GitHub HTTP endpoint has one job: read exact raw bytes under a 1 MiB hard
limit, copy only the signed headers, invoke the gateway, and return `202
Accepted`. It does not parse or execute provider business logic.

The gateway flow is:

1. Reject empty, oversized, or header-amplified requests.
2. Verify `X-Hub-Signature-256` with HMAC-SHA256 over the exact bytes.
3. Read the stable `X-GitHub-Delivery` identity and supported event name.
4. Parse only the installation and repository identity after authentication.
5. Resolve an active installation/repository grant through generated SQLC.
6. Encrypt the exact payload together with provider, delivery, workspace,
   installation, and installation-generation binding.
7. Insert or read the deduplicated durable inbox receipt.
8. Queue only `{inboxId}`; raw payloads, provider strings, tokens, and credentials never
   enter Redis/Asynq.
9. A worker leases the inbox record, decrypts it, re-resolves the current grant
   and generation, executes existing GitHub behavior, and completes the receipt.

Signature failure happens before JSON parsing or repository lookup. Duplicate
terminal deliveries return the existing receipt without dispatch. Dispatch
failure leaves a recoverable durable row; the minute recovery schedule claims
eligible receipts using bounded exponential backoff and dispatches by ID.

### Authorization fencing

Migration `000160_code_host_webhook_fencing` adds an opaque generation to each
GitHub installation. A database trigger rotates it when the installation's
tenant, provider identity, repository selection, permissions, subscribed
events, installing identities, active state, suspension, or disconnection
changes.

The worker must match all of these facts before executing:

- provider and inbox ID;
- signed external installation ID;
- workspace ID;
- internal installation ID;
- installation generation;
- active repository membership.

A missing or stale grant is completed as `cancelled` with the safe outcome
`github.stale_installation`. A transient database failure is marked failed and
remains retryable. No credential or raw provider error is stored as an outcome.

See [GitHub webhook persistence](../database/github-webhooks.md) for the SQLC
queries and schema contract.

For the complete route, task, query, lifecycle, and remaining-gap map, use the
[GitHub implementation inventory](github-inventory.md). In particular, user
OAuth unlink/revocation is implemented, while a public GitHub App installation
disconnect and lifecycle-event reconciler are not yet implemented.

### Outbound story-sync read

The asynchronous GitHub story-sync task depends on a task-owned
`StorySyncReader` port. Bootstrap supplies the stories pgx adapter; the task no
longer constructs a repository or reads application SQL. Before every read the
handler creates an explicit workspace-bound system actor with `stories:write`.
The stories repository then rechecks that the configured actor is still an
active system user and that the story remains in the requested workspace. The
task maps only the story ID, workspace, team, title, description, and status
into the GitHub sync input. Provider credentials remain owned by the GitHub
service and never enter the story reader or task payload.

## GitLab proof

### API authentication boundary

`NewAdapter` requires an HTTPS API base URL and a `TokenSource`. The source is
called for every request with the complete installation reference so bootstrap
can open a vault-bound credential at the last responsible moment. The adapter
supports:

- OAuth bearer tokens through `Authorization: Bearer ...`;
- private/project/group tokens through `PRIVATE-TOKEN: ...`.

Token values are redacted by `String`, never included in errors, and never
persisted by the adapter. Redirects are disabled, request timeout is bounded to
30 seconds (15 seconds by default), and response JSON is bounded to 1 MiB.
Provider 401, 403, 404, and 429 responses map to stable neutral errors.

Supported proof operations are:

- current-user authentication check;
- membership-scoped repository catalog with opaque keyset cursor;
- create one issue;
- add one note to an issue;
- normalize `Issue Hook` and issue-only `Note Hook` payloads.

Merge-request writes, merge-request-note normalization, pipelines, branches,
labels, assignees, installations, OAuth exchange/refresh, disconnect cleanup,
and reconciliation are explicitly unsupported.

### Signed durable webhook adapter

The proof targets GitLab's current Standard Webhooks signing-token contract,
not the weaker legacy `X-Gitlab-Token` secret. GitLab documents `webhook-id` as
stable across retries and equal to `Idempotency-Key`; the signature is
HMAC-SHA256 over `{webhook-id}.{webhook-timestamp}.{raw body}` with a `whsec_`
base64 signing key. See the official [GitLab webhook documentation](https://docs.gitlab.com/user/project/integrations/webhooks/).

The verifier:

1. requires `webhook-id`, `webhook-timestamp`, and `webhook-signature`;
2. authenticates exact bytes before JSON parsing;
3. rejects delivery/idempotency mismatch;
4. enforces a five-minute replay window and one-minute future skew;
5. accepts only issue and note events;
6. requires the instance header to match the adapter's configured HTTPS
   instance before resolving the signed project ID;
7. registers with the shared encrypted/deduplicated gateway; and
8. dispatches only the inbox ID.

GitLab's repository endpoint supports pagination and a maximum page size of 100. The proof uses `order_by=id` keyset pagination, reads the provider's
`id_after` continuation from its `Link` header, validates it as a positive
integer, and exposes that scalar as an opaque cursor that callers only echo
back. It never follows a caller-controlled continuation URL. See the official
[Projects API](https://docs.gitlab.com/api/projects/) and [REST pagination
contract](https://docs.gitlab.com/api/rest/).

Before production wiring, add vault-backed installation persistence, OAuth
state/exchange/refresh, generation rotation, a current-grant worker fence,
per-installation signing-key selection, recovery registration, disconnect
reconciliation, health reporting, and a public OpenAPI route. The proof's
static signing token is for one configured instance and must not become a
shared multi-tenant production secret. Do not reuse GitHub installation rows or
story rules.

## Adding another code host

1. Start from an actual product use case, then select only the proven ports.
2. Keep SDK, HTTP, JSON, auth-header, and provider-error types in the adapter.
3. Make installation generation and tenant identity explicit at every async
   boundary.
4. Use the shared gateway; never queue raw payloads or parse before verifying.
5. Add unit contract tests for supported and unsupported capabilities.
6. Add provider-auth, exact-byte signature, replay, payload-binding, duplicate,
   stale-grant, rate-limit, and disconnect tests.
7. Add PostgreSQL 18 integration coverage for every persistence invariant.
8. Register discovery metadata only after capabilities are truthful; add API
   and worker wiring only when the complete lifecycle is implemented.

## Verification

From `apps/server`:

```sh
go test -count=1 ./internal/platform/codehost ./internal/platform/webhooks
go test -count=1 ./internal/modules/github/... ./internal/modules/gitlab/...
go test -race -count=1 ./internal/modules/github/... ./internal/modules/gitlab/...
go vet ./internal/platform/codehost ./internal/platform/webhooks ./internal/modules/github/... ./internal/modules/gitlab/...
go test -tags=integration -count=1 ./internal/modules/github/repository ./internal/modules/gitlab
```

Integration tests require `TEST_DATABASE_URL` pointing at a disposable
PostgreSQL 18 control database whose role can create and drop test databases.
