# Native desktop protocol

Desktop uses the existing opaque `fortyone_session` session through Rust HTTP.
The browser PKCE handoff uses `POST /auth/desktop/authorize` and
`POST /auth/desktop/exchange`, with the exact allowlisted redirect
`fortyone-desktop://login`. Codes are client-bound, single use and expire after two
minutes. Native requests retain the trusted hosted `Origin` header required by
the existing cookie CSRF policy. Session revocation version, original
`AuthenticatedAt`, and verified workspace SSO proof survive the handoff; the
new session has an independent session ID.

## Downloads

`GET /workspaces/{slug}/desktop/sync` returns the normal `{data: ...}` envelope.
The data contains `schemaVersion`, `workspaceId`, `userId`, `role`, `syncedAt`,
`stories`, `documents`, `comments`, `teams`, `statuses`, `complete`, `removed`,
`commentsCompleteFor`, `nextStoryOffset`, `nextDocumentOffset`, and `scope`.

- `storyLimit` is 1–100 (default 100); `documentLimit` is 1–50 (default 50).
- `storyOffset` / `documentOffset` traverse pages independently. Each nullable
  `next…Offset` is the next position, not a durable synchronization cursor.
- `myWork=true` limits stories to the current actor's assignments. It never
  claims the whole workspace story collection is complete.
- `scope=selected` skips recent lists. `storyIds` / `documentIds` are comma-
  separated UUIDs, at most 100 each. Revalidate **every** stored identity in
  bounded batches, including items beyond the recent page.
- Explicit `removed` IDs mean the item is no longer readable, including deletion
  and permission loss. Evict its server cache and related cached children; keep
  dirty local work separately for deliberate recovery.
- `complete.<collection>=true` permits replacing that whole collection. It is
  never true for a partial story/document page, My Work story selection or
  selected-ID story/document scope. Reference teams and statuses remain complete.
  Offset traversal is not an atomic snapshot: do not infer deletion from absence
  across a paginated traversal. Use explicit known-ID revalidation.
- `commentsCompleteFor` identifies fully included story thread collections.
  Fetch `GET /desktop/stories/{id}/comments?page=N` (under the workspace prefix)
  for remaining pages; response is `{storyId, comments, hasMore, page}`.
- Collaborative documents include `collaborationState` as standard base64. The
  access check, epoch and CRDT state share one repeatable-read transaction.
  Never initialize an existing CRDT from an independent HTML snapshot.
- Snapshots are bounded to 20 MiB; reduce page sizes / selected-ID batches after
  HTTP 413. A single oversized resource is an explicit failed download.

`GET /workspaces/{slug}/desktop/settings` returns
`{profile, workspace, teams, members, preferences, downloadedAt}` using current
membership and explicit safe field allowlists. Authentication credentials,
integration secrets and billing credentials are excluded.

Use the existing `POST /workspaces/{slug}/documents/{id}/collaboration-session`
for a short-lived document-scoped `{token, name}` ticket. `name` binds document
ID and collaboration epoch. The account session remains in native storage.

## Durable operations

`POST /workspaces/{slug}/desktop/operations` takes:

```json
{"id":"UUID","kind":"story.update","resourceId":"UUID","payload":{"title":"Edited offline"},"baseVersion":"original updatedAt"}
```

Supported kinds and payloads:

| Kind | Payload | Base version |
| --- | --- | --- |
| `story.create` | Full Projects `AppNewStory` DTO, including estimates, scheduling, dates, parent/objective/sprint/key result, labels and custom-field values | none |
| `story.update` | Same scalar field allowlist and null semantics as Projects PATCH, including dates, estimates, scheduling and references | original RFC3339Nano `updatedAt` |
| `comment.create` | `storyId`, `comment` (or `body`), optional `parentId`, `mentions` | none |
| `document.create` | `title`, optional `contentHtml`, `contentText`, `visibility` (defaults private) | none |
| `document.update` | optional `title`, `contentHtml`, `contentText`, `collaborationEpoch` | original revision as a decimal string |
| `attachment.upload` | `fileName`, `contentType`, `data` (standard base64), `targetKind` (`story` / `document`) | none |

`resourceId` is the client-generated provisional ID for creations, the existing
resource for updates, and the target story/document for uploads. Story/document
creations return server IDs; comments retain their client ID. Attachment uploads
return the attachment ID, which **must not** replace the target entity ID.
Document media follows the existing byte-validated image / MP4 policy. Decoded
uploads are limited to 10 MiB; other mutation payloads are limited to 4 MiB.

Success returns `{resourceId, operationVersion, resource?, replayed}`. The
immutable `operationVersion` is the exact story timestamp or document revision
produced by this operation, captured and persisted in its transaction. The
optional `resource` is a current permission-checked refresh and may already be
newer. **Only `operationVersion` may advance a dependent local operation's base.**
Never replace it with the current resource version: that would silently rebase
past somebody else's concurrent edit after a lost response.

The `(actor, workspace, operation ID)` receipt and domain writes commit in one
serializable transaction. Requests with the same ID but different canonical
commands fail with 409. Receipts have no TTL and store only command hashes,
resource identity and operation version. A transport failure has an unknown
outcome: replay the same ID and unchanged payload. A user-reviewed conflict
resolution is a new operation ID. Stories and documents retain their existing
domain validation, permission and expected-version checks. HTML replacement of
an existing collaborative document is rejected with 409.

Story create and update reuse the Projects transport conversion and typed PATCH validator. Target IDs never belong in update payloads. Authoritative `reconcileDescriptionMedia` snapshots reuse the normal exact story/workspace media-reference parser. Orphan attachment metadata and the existing durable object-deletion outbox commit with the story and desktop receipt; the existing worker deletes storage objects after commit, including crash/replay recovery. The desktop receipt ID owns creation idempotency even when the normal form includes its own `idempotencyKey`.

Database mutation events remain in their domain transactional outboxes.
Best-effort notifications, scheduling and image optimization execute only after
commit. Storage objects are cleaned up on known rollbacks; an ambiguous COMMIT
never deletes a potentially committed object. A rare aborted ambiguous commit
may leave an unreferenced storage object requiring operational cleanup.

Migration `000218_desktop_operation_receipts` is required before enabling writes.
Do not prune receipts while any client can replay their operation IDs. Its down
migration refuses to drop a nonempty receipt table. Apply it explicitly through
the repository migration tooling; source changes and ordinary tests do not apply it.

Unit/race tests cover desktop PKCE client separation, SSO proof preservation,
transaction/savepoint outcomes, changed-payload reuse, lost-response replay,
immutable committed versions, snapshot scopes, limits and conflict bases. The
PostgreSQL integration suite additionally exercises real document commits and
rollbacks; run it only against an explicitly approved disposable database. It
was compile-checked without executing migrations.
