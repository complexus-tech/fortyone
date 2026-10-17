# Preview cache verification — 2026-10-17

Verified the real local application at `http://localhost:3000/second/chat`, with
the local API, PostgreSQL, Redis, and configured private S3 preview cache. The
static chat UI fixture was not used as integration evidence. This is functional
verification, not a production deployment or concurrent-user load test.

## Results

| Scenario                                | Evidence                                                                                                                                                                                                                           | Result |
| --------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| Same external link in two messages      | Both cards used the same Redis metadata version and `fetchedAt`; a full browser reload preserved both.                                                                                                                             | Pass   |
| Concurrent instances and repeated reads | Actual metadata cache code with an isolated Redis namespace: two concurrent cache instances and twelve subsequent reads made one upstream request.                                                                                 | Pass   |
| Refresh preview                         | Replaced the isolated test URL's cached title with a stale marker, verified both cards displayed it, then clicked Refresh preview. Both cards returned to the live upstream title with a new version and timestamp.                | Pass   |
| Failed refresh                          | Controlled cache integration retained the last successful revision when the upstream refresh failed.                                                                                                                               | Pass   |
| Shared OG/favicon storage               | Actual UI refresh wrote both new image objects to `previews/external/v1/`. S3 HEAD confirmed their content types, sizes, timestamps, and ETags.                                                                                    | Pass   |
| Normal reload after refresh             | Metadata version and timestamp remained unchanged; both S3 objects retained the same ETags and modification times.                                                                                                                 | Pass   |
| Task changes                            | Edited TM-267's title and priority through the chat side panel; the preview updated immediately. Editing priority from the card also updated the side panel. Original title and priority were restored and verified in PostgreSQL. | Pass   |
| Document changes                        | Created a local document, posted its link, then changed its title and body in a separate editor tab. The chat card updated its title and excerpt without reloading chat.                                                           | Pass   |
| Permission enforcement                  | Executed isolated PostgreSQL integration tests for current tenant/team/workspace access, document visibility and grants, revocation, deleted tasks, and inactive users.                                                            | Pass   |
| Duplicate native cards and rapid edits  | Hook tests verify one shared refresh per invalidation, a newer edit cancels obsolete work, and an older completion cannot overwrite the latest preview.                                                                            | Pass   |

Document changes made in another tab currently use the visible-card recheck
(30-second interval, plus request time), rather than a dedicated document SSE
event. This browser run confirmed eventual update without reload; it did not
measure a precise propagation time. Local editor updates and task realtime
signals invalidate summaries directly.

## Defects corrected during verification

- A cold S3 read could exceed its short deadline and disable cache writes as well,
  leaving images only in process memory. Read/write cooldowns are now independent
  and writes have a bounded five-second deadline. Live refresh then produced
  persistent S3 objects, which were reused on reload.
- Duplicate native cards each reacted to one canonical invalidation, causing
  duplicate reads. Invalidation is now coalesced by event and shared query.
- An expired cached card could briefly say Preview unavailable while permission
  revalidation was pending. It now shows Loading preview while keeping private
  cached fields hidden until access succeeds.
- Restoring the test task exposed a missing current-priority prop in the task
  side-panel menu. Passing the current value fixes clearing a priority back to
  No Priority; inline preview editing already propagated correctly.

## Repeatable automated checks

```sh
pnpm --filter projects test -- src/app/api/metadata src/lib/hooks/link-metadata.test.tsx src/lib/queries/links/get-metadata.test.ts --runInBand
pnpm --filter projects test -- src/components/ui/link-preview/native-query.test.tsx src/components/ui/link-preview/native-preview.test.tsx src/shell/realtime/workspace-realtime-connection.test.tsx --runInBand
pnpm --filter projects type-check
```

These focused frontend suites passed 114 external-cache tests and 43
native-preview/realtime tests. Scoped ESLint passed. React Doctor reported no
errors and retained existing complexity/set-state warnings.

The priority-menu correction also passed four tests across its new CoreOptions
regression and the existing Options suite. Full project type checking and scoped
lint passed after that correction. The final live chat had no browser console
errors.

From `apps/server`:

```sh
go test ./internal/modules/nativepreviews/...
TEST_DATABASE_URL=<approved-local-control-database> go test -count=1 -tags=integration ./internal/modules/nativepreviews/repository -v
```

The integration test created, migrated, and removed its own isolated test
database. Its 17 subtests executed successfully; this was not a compile-only run.

## Test data

Four labeled local chat messages and one document remain available for visual
review in the Second workspace's Team 1 conversation. TM-267 was restored to
`Verify chat keyboard navigation` / `No Priority`. The isolated external URL is
`https://fortyone.app/?preview-e2e=20261017-cache-check`; its artificial stale marker
was replaced by a real refresh. Temporary Redis test-namespace keys were removed.
Uploaded attachments and unrelated cache records were not modified.
