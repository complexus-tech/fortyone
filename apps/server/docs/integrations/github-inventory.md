# GitHub integration implementation inventory

This is the current code-level source of truth for the GitHub integration. It
distinguishes behavior that is running today from the remaining
provider-lifecycle work. Read it with the
[public API contract](../api/github.md), [code-host architecture](code-hosts.md),
and [webhook gateway](webhook-gateway.md).

## Status legend

- **Implemented**: wired into the API or worker and covered by focused tests.
- **Not exposed**: a typed capability exists for internal composition or as a
  proof, but there is no public route selecting it.

## Ownership and package boundaries

| Package                              | Owns                                                                                                        | Must not own                                                                 |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `internal/modules/github/http`       | Route authentication, workspace extraction, bounded decoding, request validation, response/status mapping   | SQL, GitHub SDK calls, OAuth secrets, webhook effects                        |
| `internal/modules/github/service`    | GitHub use cases, provider translation, OAuth and credential policy, synchronization and workflow decisions | SQLC/pgx or generated database types, sibling service types, HTTP middleware |
| `internal/modules/github/shared`     | Stable GitHub records shared by the service and repository adapter                                          | Database execution, provider clients, orchestration                          |
| `internal/modules/github/repository` | GitHub persistence implementation and row-to-contract mapping                                               | HTTP responses, provider network calls, story rules                          |
| `internal/bootstrap/githubadapter`   | Translation between GitHub-owned ports and existing Stories/Integration Requests use cases                  | Product policy or persistence                                                |
| `internal/platform/codehost`         | Provider-neutral code-host capabilities and reconciliation primitives                                       | GitHub-specific fields or product synchronization rules                      |
| `internal/platform/webhooks`         | Durable inbox, payload binding, deduplication, leases, recovery, safe outcomes                              | Provider signature algorithms and GitHub event semantics                     |

The GitHub service imports neither its concrete repository adapter nor the
Stories or Integration Requests service packages. Its repository contract is
split into authorization, settings, installation, story-link, comment-link,
identity, credential, and webhook-grant capabilities. Bootstrap performs the
only sibling-module translations.

All handwritten GitHub Go files are below the 700-line architecture limit. The
largest files are kept capability-specific (comments, code-host adapter, story
sync, credentials, OAuth/workflow, and HTTP models) rather than growing a
single integration service.

## HTTP route inventory

Every `/workspaces/{workspaceSlug}/...` route uses authenticated workspace
resolution. Middleware rechecks current membership; path workspace slugs are
not accepted as authorization by themselves.

| Method and route                                                                    | Minimum access     | Behavior                                                                            | Status      |
| ----------------------------------------------------------------------------------- | ------------------ | ----------------------------------------------------------------------------------- | ----------- |
| `GET /workspaces/{workspaceSlug}/integrations/github`                               | workspace member   | Read settings, installations, repositories, and issue-sync links                    | Implemented |
| `POST /workspaces/{workspaceSlug}/integrations/github/install-session`              | workspace admin    | Create opaque, ten-minute GitHub App installation state and return the provider URL | Implemented |
| `POST /workspaces/{workspaceSlug}/integrations/github/repositories/resync`          | workspace admin    | Refresh active installation repository catalogs                                     | Implemented |
| `GET /workspaces/{workspaceSlug}/integrations/github/settings`                      | workspace member   | Read workspace GitHub settings                                                      | Implemented |
| `PUT /workspaces/{workspaceSlug}/integrations/github/settings`                      | workspace admin    | Validate and patch supplied workspace settings                                      | Implemented |
| `POST /workspaces/{workspaceSlug}/integrations/github/issue-sync-links`             | workspace admin    | Link one active repository to one team                                              | Implemented |
| `PUT /workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}`     | workspace admin    | Change direction or active state                                                    | Implemented |
| `DELETE /workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}`  | workspace admin    | Delete the scoped link                                                              | Implemented |
| `GET /workspaces/{workspaceSlug}/teams/{teamId}/settings/github`                    | workspace member   | Read workflow rules, seeding defaults when absent                                   | Implemented |
| `PUT /workspaces/{workspaceSlug}/teams/{teamId}/settings/github`                    | workspace admin    | Replace at most 64 validated workflow rules                                         | Implemented |
| `GET /workspaces/{workspaceSlug}/stories/{storyId}/github-links`                    | workspace member   | List issue, pull request, branch, and commit links                                  | Implemented |
| `DELETE /workspaces/{workspaceSlug}/stories/{storyId}/github-links/{linkId}`        | workspace member   | Delete a workspace-scoped story link                                                | Implemented |
| `GET /workspaces/{workspaceSlug}/stories/{storyId}/github-comments`                 | workspace member   | Read linked-issue comments with bounded pagination                                  | Implemented |
| `POST /workspaces/{workspaceSlug}/stories/{storyId}/github-comments`                | workspace member   | Post to all linked issues, with optional idempotency                                | Implemented |
| `GET /workspaces/{workspaceSlug}/integration-requests/{requestId}/github-comments`  | workspace member   | Read comments for the request's linked GitHub issue                                 | Implemented |
| `POST /workspaces/{workspaceSlug}/integration-requests/{requestId}/github-comments` | workspace member   | Post to the request's linked GitHub issue, with optional idempotency                | Implemented |
| `POST /user/integrations/github/link-session`                                       | authenticated user | Create opaque, 15-minute user-link state bound to that user and a safe return URL   | Implemented |
| `POST /user/integrations/github/link`                                               | authenticated user | Consume state, exchange code, identify user, vault token, and persist link          | Implemented |
| `DELETE /user/integrations/github/link`                                             | authenticated user | Revoke the OAuth token remotely, then clear local identity and envelope             | Implemented |
| `GET /integrations/github/setup`                                                    | provider callback  | Consume install state, reauthorize current admin, catalog repositories, redirect    | Implemented |
| `POST /webhooks/github`                                                             | GitHub signature   | Verify and durably accept one supported delivery                                    | Implemented |

There is currently no public GitHub App installation disconnect endpoint. Do
not infer one from the user OAuth unlink endpoint or provider descriptor.

## Asynchronous runtime inventory

| Task type                 | Queue contract                 | Retry/lease behavior                                                                           | Effect owner                 |
| ------------------------- | ------------------------------ | ---------------------------------------------------------------------------------------------- | ---------------------------- |
| `github:webhook:process`  | `{ "inboxId": "uuid" }` only   | Asynq max retry 8, 90-second timeout; durable inbox owns dedupe and lease                      | GitHub webhook processor     |
| `github:webhook:recovery` | no provider payload            | scheduled every minute; shared gateway claims eligible receipts with backoff                   | webhook gateway              |
| `github:story:sync`       | story and workspace UUIDs only | integration queue, max retry 10; handler re-reads story through a workspace-bound system actor | GitHub story synchronization |

The webhook task never contains raw provider bytes, a provider token, a
workspace selected by the caller, or an installation credential. The story
sync handler refuses missing system/workspace identity and rechecks the story in
the requested workspace before mapping the small sync input.

## Supported webhook event matrix

| GitHub event          | Accepted actions/effects                                                                                      | Echo/idempotency control                                          |
| --------------------- | ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `issues`              | import pending request; update linked story title/description/priority; apply open/reopen/close workflow rule | configured app-bot sender suppression plus origin/revision ledger |
| `pull_request`        | link referenced stories, managed link comment, optional PR body, workflow movement                            | stable link identity and strict supported-action mapping          |
| `pull_request_review` | recalculate latest review state per reviewer, update link counters, record activity, apply rule               | inbox delivery dedupe; deterministic state update                 |
| `issue_comment`       | create one FortyOne comment for a linked issue                                                                | outbound ledger check plus atomic inbound reservation/completion  |
| `check_run`           | update linked PR check state                                                                                  | deterministic state update                                        |
| `create`              | attach matching branch references                                                                             | story-link identity                                               |
| `push`                | attach branch/commit references and apply commit-close behavior                                               | canonical newline-stable revision and link identity               |

Unsupported event names are rejected before persistence. Events signed for an
unknown, inactive, suspended, disconnected, or repository-mismatched grant are
ignored without revealing whether an installation exists.

## Webhook security lifecycle

1. The HTTP adapter reads the exact body under a 1 MiB limit.
2. The verifier requires `POST` and exactly one value for each of
   `X-Hub-Signature-256`, `X-GitHub-Delivery`, and `X-GitHub-Event`.
3. HMAC-SHA256 authentication happens before JSON decoding or database access.
4. Only installation and repository identity are decoded for authorization.
5. The active tenant/repository grant and opaque installation generation are
   resolved with generated SQLC.
6. Exact bytes are encrypted with the dedicated GitHub payload key derived
   from the application root; the vault, OAuth, and GitHub signing keys remain
   separate cryptographic domains.
7. The durable inbox deduplicates the stable GitHub delivery ID and queues only
   its UUID.
8. The worker leases the row, opens the bound payload, and rechecks workspace,
   internal installation, external installation, external repository, and
   generation before any effect.
9. Authentication failures return 401. Verification infrastructure failures
   return retryable 503. Both use sanitized error bodies.

GitHub does not provide a signed timestamp in this webhook contract. Replay is
therefore controlled by stable delivery-ID deduplication, terminal receipt
state, and worker leases rather than an invented time window.

## OAuth and credential lifecycle

### GitHub App installation

- State is 32 random bytes encoded as 43-character unpadded base64url.
- Redis stores only a digest-derived key and the short-lived bound context.
- The record binds provider, `app-install` purpose, workspace, initiating user,
  workspace slug, version, and expiry.
- Callback consumption is atomic and precedes GitHub API work.
- The callback rechecks the initiating user's current admin role.
- Duplicate or missing query values, invalid installation IDs, replayed state,
  expiry, binding mismatch, and store outage fail closed.
- An installation external ID cannot be reassigned to another workspace by an
  upsert. An empty repository catalog deactivates all previously known repos.

### User OAuth link

- State uses the same opaque primitive, a distinct `user-link` purpose, a
  15-minute TTL, the authenticated user, and a validated return destination.
- OAuth code exchange is a bounded form POST. Credentials never appear in the
  URL; redirects are disabled; responses are capped at 64 KiB and must contain
  exactly one JSON value.
- A newly issued token is revoked if user lookup, vault encryption, or local
  persistence fails.
- Stored tokens are authenticated credential-vault envelopes bound to provider,
  account/user, credential type, and an opaque generation.
- Legacy plaintext upgrade and key rewrap use compare-and-swap. A concurrent
  relink or unlink wins; maintenance cannot restore stale credentials.
- Unlink revokes remotely first. A transient remote failure keeps the local
  credential so the operation can be retried safely; successful revocation is
  followed by one local update clearing identity, envelope, version, and
  generation.

## Comment idempotency contract

Both comment-write routes accept an optional `Idempotency-Key`. When present:

- the common key parser rejects empty, duplicate, or malformed values;
- a deterministic UUID is derived from operation, workspace, actor, resource,
  and key, so reuse in another scope cannot collide;
- only that UUID marker is sent to GitHub; the raw key is not exposed;
- retries search bounded paginated comments before posting;
- the same marker and normalized body is a successful replay;
- the same marker with different content returns 409;
- multi-issue story posting is fail-fast, and a retry skips effects already
  completed with the same marker.

Without the header, each attempt gets a new marker and the API cannot promise
deduplication across client retries. Clients that retry comment writes must send
the same key and body.

## Persistence inventory

All GitHub-owned application persistence uses native `pgx/v5` and the
module-local SQLC block in `sqlc.yaml`. Handwritten SQL is grouped by capability:

| Query file          | Owned operations                                                                                                              |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `authorization.sql` | current workspace-role lookup                                                                                                 |
| `installations.sql` | ownership-fenced installation upsert and complete repository-set reconciliation                                               |
| `settings.sql`      | atomic partial settings updates and integration listings                                                                      |
| `sync_links.sql`    | tenant-scoped issue-sync links, locked workflow replacement, status validation                                                |
| `story_links.sql`   | current-grant repository routing, story lookup, indexed link upsert, review/check state, generic URL and label reconciliation |
| `comments.sql`      | atomic inbound reservation and outbound comment ledger                                                                        |
| `identities.sql`    | bounded user resolution by GitHub identity, name, or system email                                                             |
| `credentials.sql`   | vault-envelope link/unlink, bounded maintenance pages, and compare-and-swap upgrade/rewrap                                    |
| `webhooks.sql`      | receipt-time authorization and execution-time tenant/generation fencing                                                       |

Generated rows are mapped inside `internal/modules/github/repository`; neither
service nor HTTP packages import SQLC output. Transactions use the shared pgx
runner and generated `WithTx` binding. PostgreSQL no-row results are translated
at the repository boundary while the service retains its current stable error
contract.

The expression index already present on `github_story_links` is the conflict
target for atomic story-link upserts, so no schema migration was required.
Legacy `story_links` and `labels` tables do not have a safe uniqueness contract;
their GitHub reconciliation paths therefore take transaction-scoped advisory
locks before typed lookup/insert operations. This prevents duplicate effects
without changing historical migrations or imposing a risky global constraint
on data owned by other modules.

API and worker composition pass the Integration Requests repository and Maya
workspace-entitlement capability to GitHub through consumer-owned interfaces.
Both adapters use generated SQLC against the process-owned native pgx pool; they
do not create compatibility database handles or duplicate the owning module's
business logic. The GitHub repository itself also receives only that native
pool.

## Provider-lifecycle gaps

The completed pgx/SQLC cutover does not imply that these product and delivery
lifecycle capabilities exist:

1. Add GitHub App lifecycle handlers for installation deletion, suspension,
   unsuspension, and repository add/remove events, including generation
   rotation and reconciliation.
2. Add an admin disconnect/reconcile use case if the product intends to manage
   installation removal from FortyOne.
3. Add a durable per-delivery effect ledger for non-idempotent downstream
   effects. Inbox leasing prevents normal concurrent duplicates, but a crash
   after a remote/product effect and before receipt completion can replay that
   effect.
4. Fence provider-neutral code-host installation calls against current local
   tenant/generation data before exposing them to external API clients.
5. Add effect-replay integration coverage when the durable effect ledger is
   implemented.

## Verification map

Fast focused checks:

```bash
go test -count=1 ./internal/modules/github/...
go test -count=1 ./internal/platform/webhooks ./internal/bootstrap/githubadapter
go test -race -count=1 ./internal/modules/github/... ./internal/platform/webhooks
go vet ./internal/modules/github/... ./internal/platform/webhooks
./scripts/check-sqlc.sh
SQLC_DATABASE_URL='postgresql://<role>@<postgres-18-host>/<migrated-db>?sslmode=disable' ./scripts/vet-sqlc.sh
```

The focused suite covers opaque OAuth state/replay/binding, OAuth response
bounds, redirect refusal, credential AAD/tamper/rotation, remote revoke
redaction, signature-before-parse, duplicate headers, payload tampering,
repository outage classification, current-grant fencing, queue-by-ID, provider
capability truth, comment markers, safe redirect origins, admin role checks,
typed query generation drift, and API/worker native-pool composition.

Tagged persistence checks use a disposable PostgreSQL 18 control database:

```bash
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -tags=integration -count=1 ./internal/modules/github/repository ./internal/modules/github/service
```

The repository contracts additionally cover installation ownership conflict,
empty and partial repository reconciliation, atomic story-link update, concurrent
generic-link and label deduplication, inbound-comment reservation, credential
CAS, and webhook generation rotation. Never point this suite at a shared
development or production database.
