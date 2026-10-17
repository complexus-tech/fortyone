# Preview caching

External public-page previews use the same service from Chat and story links.
Internal task/document permissions and summaries are handled separately; this
cache never stores private internal records or forwards session credentials to
an external host.

## Request behavior

- Browser query keys normalize host/default port/fragments while preserving query
  string values and order. Repeated links and virtualized row remounts share one
  request. Successful metadata is fresh for one hour; failures retry after 30
  seconds. Inactive metadata is retained for up to 24 hours in the QueryClient.
- Every metadata and image route authenticates before reading shared caches.
  Metadata responses remain `private, no-store`; image responses are
  `private, max-age=86400` and have sandbox/nosniff/same-origin headers.
- Redis stores small public metadata records for 24 hours. One-hour-old records
  render immediately and schedule revalidation with Next `after`. A stale client
  briefly polls every two seconds until the refreshed record is available. A
  transient failure keeps the last-good card and applies a 30-second retry delay.
- Metadata misses coalesce in-process and across instances with a 12-second lease.
  There are at most 32 distinct in-flight upstream HTML fetches per process and
  1,000 local metadata entries. The browser starts at most four metadata requests
  at once. Missing/unavailable Redis falls back to this bounded local cache.
- Refresh preview bypasses metadata freshness, returns errors instead of replacing
  good content, and updates every mounted observer of the canonical URL. It is
  limited to one refresh per user/link every five seconds and 20 per user/minute.
- Each successful metadata fetch has a new `version` and numeric `fetchedAt`.
  Image URLs carry that version, so refreshing fetches new bytes even if a website
  reuses an unchanged image URL.
- Public raster images are downloaded through the existing pinned-DNS fetcher,
  limited to 2 MiB, and stored privately in S3 under `previews/external/v1/` using
  hashed URL/version keys. No external cookies or authorization are forwarded.
  SVGs and HTML are not accepted as images. One process retains at most 32 MiB/256
  images and performs at most eight distinct image loads concurrently. Concurrent
  callers share work; a canceled subscriber does not abort other clients. Failures
  receive a short 30-second local negative cache. Redis leases coalesce image
  creation across replicas when shared object storage is configured.

## Deployment

Set server-only `PREVIEW_REDIS_URL`, or reuse `APP_REDIS_HOST`,
`APP_REDIS_PORT`, `APP_REDIS_PASSWORD`, `APP_REDIS_DB`, and
`APP_REDIS_DISABLE_TLS`. TLS is enabled unless explicitly disabled for an approved
local service.

Set `PREVIEW_S3_BUCKET`, `PREVIEW_S3_REGION`, `PREVIEW_S3_ENDPOINT`,
`PREVIEW_S3_ACCESS_KEY_ID`, `PREVIEW_S3_SECRET_ACCESS_KEY`, and
`PREVIEW_S3_FORCE_PATH_STYLE` as needed. Each has an existing `APP_AWS_*` fallback
(`APP_AWS_SECRET_ACCESS_KEY` is the secret-key fallback). Standard AWS credential
resolution is supported when explicit keys are absent. The bucket must already
exist; the app neither creates buckets nor changes bucket policies.

**Production prerequisite:** add a lifecycle rule expiring only
`previews/external/v1/` objects after seven days. In a versioned bucket, also expire
noncurrent versions under that prefix. The application ignores objects older than
seven days, but application expiry alone does not reclaim storage. Do not apply
this rule to uploaded attachments or any other prefix. Keep Block Public Access
enabled and grant the application only `GetObject`/`PutObject` for this prefix.

The rule is [preview-cache-lifecycle-rule.json](./preview-cache-lifecycle-rule.json).
This file is **one rule**, not a replacement bucket lifecycle configuration.
Use an administrator identity to read the current lifecycle configuration and
bucket versioning, merge this rule by its ID, and preserve every other rule and
`TransitionDefaultMinimumObjectSize`. Re-read immediately before writing to
detect concurrent changes, then read back the result to verify the exact prefix
and that unrelated rules remain unchanged. Never submit a new configuration
without first reading the existing one: S3 replaces the entire configuration.

The filter includes the trailing slash and covers only generated external
OG/favicon cache objects. Uploaded originals use UUID filenames at the bucket
root; private file thumbnails use `thumbnails/`. Neither matches this rule.
Do not broaden it to `previews/` or an empty prefix. Current cache objects become
eligible for expiry after seven days; noncurrent versions become eligible seven
days after becoming noncurrent, so versioned bytes can remain for approximately
14 days overall. S3 performs expiry asynchronously. The age-based rule also
eventually removes expired delete markers; no bucket-wide cleanup rule is needed.

**Applied and verified on 2026-10-17** in the AWS console for account
`352154385712`, bucket `fortyone`, region `us-east-1`. The bucket had no lifecycle
rules and versioning was disabled. Created and read back
`fortyone-external-preview-cache-7-days` as **Enabled**, filtered by exactly
`previews/external/v1/`, with seven-day current and noncurrent expiration and no
retained-version minimum. The noncurrent action will apply if versioning is
enabled later; bucket versioning was not changed. There are no transitions,
size/tag filters, or multipart-upload actions. No uploaded-file paths match the
filter, and no existing rules were replaced. This verifies the saved
configuration, not a completed future expiry cycle.

The application's storage identity cannot administer lifecycle settings; setup
used the user's authenticated administrator console session. Keep
`s3:GetLifecycleConfiguration`, `s3:PutLifecycleConfiguration`, and
`s3:GetBucketVersioning` administrative permissions out of runtime credentials.

AWS references: [lifecycle configuration replacement](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html)
and [versioned-object expiration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/lifecycle-configuration-examples.html).

Missing or unavailable shared services degrade to bounded local caching and emit
credential-free operational warnings. Monitor these warnings; a fallback is not
proof that cross-instance caching is working. Cold misses still run the full
SSRF policy (validated target, public DNS addresses pinned to the socket, validated
redirects, response byte caps and deadline). Successful cache hits contain only
content previously fetched under this policy.

S3 reads have a 1.5-second deadline and writes a five-second deadline. Their
30-second failure cooldowns are independent: a slow cold read must not prevent
publishing the subsequently fetched image for other instances to reuse.

## Internal task, objective, and document summaries

`GET /workspaces/{workspaceSlug}/previews/native/{kind}/{id}` returns a small presentation summary rather than hydrating the full task, objective, or document. Task references and UUIDs are supported; objective links additionally supply the team ID. Document text is capped at 400 characters, and summaries exclude full HTML, relationships, comments, attachments, and member lists.

Every API request runs authenticated workspace middleware and then checks live resource visibility in PostgreSQL **before** consulting Redis. Task reads require the current active actor's workspace and team membership and obey credential team restrictions. Objective reads also require the member/admin role and the linked team to match. Document reads enforce current workspace/restricted/private visibility, explicit grants, ownership, and archive state. Cached content never stores or substitutes for an authorization decision.

The shared Redis cache is scoped by workspace, kind, resource UUID, and the current database version. Task versions include the team code; documents include their revision. A changed resource version cannot reuse an older snapshot. Entries expire after one hour. Redis reads/writes have a 100 ms budget each and fail open only to a new **authorized database read**, never to unverified cached private content. Refresh adds `refresh=true`, loads the current authorized summary directly, and replaces its cached snapshot. HTTP responses remain `private, no-store`; protected content is not put in a public CDN or persistent browser cache.

The browser keeps summaries in React Query under the account identity, destination workspace, kind, and target. Duplicate cards coalesce requests, including explicit refreshes. A fresh authorization result can be reused for 30 seconds; inactive summaries are garbage-collected after five minutes. An expired cached grant on a newly mounted card remains hidden until reauthorization succeeds. An IntersectionObserver controls revalidation independently of rendering, so offscreen cards retain their layout but do not keep polling. Visible cards revalidate every 30 seconds and on window focus/reconnect. Workspace realtime events invalidate affected task summaries immediately; external document edits currently become visible within the next visible-card check (up to approximately 30 seconds, excluding network latency), rather than through a dedicated document SSE event.

Canonical full-resource caches stay owned by their respective features: previews never write partial summaries into them and never initiate a full-detail request. Existing editor cache writes and invalidations update/refetch the summary, and pending task/objective mutations project their changed fields into previews immediately. Successful inline edits preserve the authorization timestamp instead of extending the access lease. The root realtime bridge also refreshes task summaries when no full-detail cache has been loaded.

An already mounted, previously authorized card remains visible during a transient network failure so refreshing does not flash or erase it. A 401/403/404 or deleted-resource result hides it immediately. This is an availability choice for content the user has already received, not a promise of live authorization while offline: edits/downloads continue to enforce server permissions, visible checks resume when connectivity returns, and an expired card cannot be newly revealed from cache after remount. Identity changes select a separate cache and late requests cannot populate a new identity's full-resource cache.

Native coverage includes PostgreSQL integration checks against an isolated fully migrated database for tenant/team restrictions, objective guest denial, credential team scope, document grants/revocation, private documents with stale grants, deleted tasks, inactive users, and revoked workspace memberships. Hook tests cover remount reuse, duplicate-card refresh, expired-grant concealment, identity changes, in-flight edits, transient failures, explicit denial, document editor updates, and stopping requests offscreen.

## Private file thumbnails

Chat images and inline PDF covers use a shared small-thumbnail endpoint rather than
mounting PDF.js or downloading original image bytes on each virtualized remount.
The full attachment viewer still loads the original only when opened. PDF.js is
retained as the existing fallback renderer for other callers that do not supply a
server-thumbnail URL.

Migration `000233_attachment_thumbnails` tracks private derivatives by workspace,
attachment, source revision, and renderer version. The authenticated thumbnail
route checks the live conversation/attachment relation before every response,
including `304 Not Modified`. Pending work returns `202` with a short retry delay.
The worker generates the first PDF page or a resized image once, stores an
840-pixel JPEG in the existing private attachment bucket, and reuses it across
users and API instances after authorization. Source revisions include the original
attachment's update timestamp, type, and size. A mutation during processing cannot
publish the old derivative; completion checks the original version and scan state.

Work is coalesced with Asynq uniqueness and a PostgreSQL claim lease. Each worker
process allows at most two thumbnail renders; source downloads are bounded to
25 MiB, images to 32 megapixels, rendering to a 25-second deadline, and resulting
JPEGs to 2 MiB. The worker Docker stage installs `poppler-utils`; local workers
must have `pdftoppm` on PATH. A missing renderer or unsupported document shows a
readable preview fallback while the original viewer remains available. Failures
have a five-minute cooldown; **Refresh preview** can explicitly retry them.

The browser retains at most 64 JPEG blobs / 16 MiB in memory, scoped to the signed-in
account and workspace. Duplicate requests share work; unmounting the last consumer
cancels its request/polling. Virtualized remounts reuse a recent authorized result
for up to 30 seconds. Visible thumbnails revalidate with ETag before that lease
expires; if renewal does not complete, expired pixels are hidden. Successful
conditional responses reuse the blob without fetching object storage again.
Identity/workspace changes, observed access failures, and window focus clear the
local cache. Object URLs belong to mounted components and are revoked on cleanup.
No private file bytes or grants are persisted to localStorage, IndexedDB, a service
worker, or a public HTTP cache.

Image and expanded PDF cards expose **Refresh preview** in their hover/focus
options menu. Refresh preserves the immutable server derivative when its source
has not changed and retrieves a new version when it has. Stale derivatives and
original-file deletion enqueue independent durable object cleanup records; the
derivative's UUID identifies its physical object in the existing deletion outbox,
so it cannot replace the original attachment's cleanup record.
