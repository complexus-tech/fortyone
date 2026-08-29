# Attachments persistence

The attachments module owns metadata for uploaded objects and the relations that make those objects visible from stories and documents. Object bytes remain in the configured managed storage provider; PostgreSQL is the authority for tenant ownership, references, processing state, and deletion eligibility.

## Where to find the code

| Concern                                   | Location                                                                 |
| ----------------------------------------- | ------------------------------------------------------------------------ |
| Domain states and availability policy     | `internal/modules/attachments/domain/attachment.go`                      |
| Application workflows                     | `internal/modules/attachments/service/`                                  |
| Handwritten SQL                           | `internal/modules/attachments/repository/queries/attachments.sql`        |
| SQLC adapter and mappings                 | `internal/modules/attachments/repository/`                               |
| Generated SQLC code                       | `internal/modules/attachments/repository/sqlc/`                          |
| Schema evolution                          | `internal/migrations/000162_attachment_processing_state.*.sql`           |
| Background task contract                  | `pkg/tasks/attachments.go`                                               |
| Worker handler                            | `internal/taskhandlers/attachments.go`                                   |
| Real-database isolation/concurrency tests | `internal/modules/attachments/repository/repository_integration_test.go` |

Generated files under `repository/sqlc` must never be edited by hand. Change `queries/attachments.sql`, run `make sqlc-generate`, and review the generated contract.

## Ownership invariant

An attachment identity is the pair `(workspace_id, attachment_id)`. An attachment ID or storage object name by itself is not an authorization key.

Every externally reachable read, relation, processing transition, and delete includes `workspace_id` in its repository call and SQL predicate. Story relations additionally join the story and attachment through the same workspace:

```text
request workspace
      │
      ├── story.workspace_id
      │          │
      │          └── relation.story_id
      │
      └── attachment.workspace_id ── relation.attachment_id
```

If any edge does not match, the adapter returns `attachmentdomain.ErrNotFound`. Returning not-found instead of a cross-tenant authorization detail prevents resource enumeration.

Storage object names are deliberately absent from the public repository contract. The optimization worker receives both `attachmentId` and `workspaceId`, claims the database record, and only then obtains the stored object name.

## Processing state

`scan_status` records the security state of an uploaded object:

| State       | Download policy           | Meaning                                                                    |
| ----------- | ------------------------- | -------------------------------------------------------------------------- |
| `unscanned` | allowed for compatibility | No scanner has made a claim. This is not evidence that the object is safe. |
| `pending`   | blocked                   | A configured scanner owns the object.                                      |
| `clean`     | allowed                   | A scanner completed successfully and declared the object clean.            |
| `infected`  | blocked                   | Malware or prohibited content was detected.                                |
| `failed`    | blocked                   | Scanning failed and the object must not be exposed as safe.                |

FortyOne does not currently ship a malware-scanner provider. Existing and new uploads therefore use `unscanned`; operators must not describe this state as scanned. A future scanner can move new uploads to `pending` without changing attachment authorization or download code.

`optimization_status` is a separately fenced state machine:

```text
not_requested
queued ──claim──> processing ──success──> succeeded
  │                    └────no gain────> skipped
  └──enqueue error──> failed <──runtime error──┘
                           │
                           └──retry claim──> processing
```

Claims use a bounded lease and a conditional `UPDATE ... RETURNING`. Only one worker can move a queued or failed record to `processing`; a second concurrent claim receives a state conflict. Completion and failure are compare-and-set transitions from `processing`, so a stale task cannot overwrite a newer terminal state.

The persisted failure string is capped at 512 bytes. It must contain an operational category only, never request credentials, object URLs, or provider secrets.

## Relation and deletion rules

- `story_attachments` represents ordinary downloadable story attachments.
- `story_inline_attachments` represents images or MP4 media embedded in story content.
- `document_attachments` represents images or MP4 media embedded in documents.
- Linking a story attachment verifies the story and attachment share the requested workspace.
- Linking inline story media also verifies that `created_by` is a current member of that workspace.
- Inline media unlink and orphan deletion happen in one serializable transaction.
- `DeleteUnreferencedWorkspaceAttachment` deletes only when no story, inline-story, or document relation remains.
- The database row is deleted before best-effort object cleanup. A storage deletion failure is logged without restoring a now-unauthorized database record; operators can reconcile orphaned objects safely.

## Query behavior

Story attachment lists use deterministic ordering by `(created_at DESC, attachment_id DESC)`. Tenant predicates are present on both the story and attachment rows. The primary keys on relation tables serve exact authorization and unlink queries. `idx_attachments_optimization_queue` supports future reconciliation workers over queued, failed, and expired processing records without scanning terminal attachments.

When changing a query:

1. keep workspace ownership inside the SQL rather than filtering returned rows in Go;
2. use a typed SQLC argument for every caller-controlled value;
3. preserve deterministic ordering with a unique tie-breaker;
4. add a second-workspace negative test;
5. add a concurrency test for any processing-state transition;
6. run `make sqlc-check`, `make sqlc-vet`, the attachment integration tests against PostgreSQL 18, and `make architecture-check`.

## Verification

```sh
go test ./internal/modules/attachments/...
TEST_DATABASE_URL='postgres://…/postgres?sslmode=disable' \
  go test -tags=integration -count=1 -race ./internal/modules/attachments/repository
make sqlc-check
make architecture-check
```

The integration database must be disposable and the configured role must have `CREATEDB`. The test kit creates a random database, applies the complete migration chain, and drops it after the test.
