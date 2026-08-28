# Documents data access

This guide explains how workspace documents are stored and why the repository
is structured the way it is. It is intended to be readable by an engineer who
is new to Go, SQLC, or FortyOne's authorization model.

## Where the code lives

```text
internal/modules/documents/
├── domain/                  transport-neutral values and errors
├── service/                 validation and document use cases
├── http/                    existing web routes and JSON mapping
└── repository/
    ├── repository.go        pgx pool, transactions, error mapping
    ├── documents.go         create, get, list, update
    ├── lifecycle.go         duplicate, archive, permanent delete
    ├── access.go            visibility and member replacement
    ├── relationships.go     Story and Objective associations
    ├── media.go             attachment relation authorization
    ├── mapper.go            generated row to domain mapping
    ├── queries/             handwritten, named SQL
    └── sqlc/                generated code; never edit by hand
```

The repository imports `domain`, but never `service` or `http`. SQLC-generated
types stop at the repository boundary. The service keeps compatibility aliases
for the existing routes, so this separation does not change the API payloads.

## Tables and ownership

- `documents` owns the title, HTML/text content, visibility, creator, updater,
  and archive timestamp. `workspace_id` is the tenant boundary.
- `document_members` grants `viewer` or `editor` access to a restricted
  document. The document creator does not need a row here. Private documents
  ignore member rows and are creator-only.
- `document_relationships` links a document to a Story or Objective. The
  relationship carries `workspace_id` so every lookup can prove tenant scope.
- `document_attachments` links stored attachments to a document. The attachment
  row must belong to the same workspace as the document.

Migration `000103_workspace_documents` created documents and relationships;
`000104_document_attachments` added media relations. The SQLC migration did not
need a new schema migration: the existing keys and checks already represent the
required invariants.

## Visibility is not edit permission

The SQL repeats one intentional policy rather than relying on a route check:

| Document state | Active workspace actor can read | Active non-guest actor can edit |
| -------------- | ------------------------------- | ------------------------------- |
| `workspace`    | yes                             | yes                             |
| `restricted`   | creator or listed member        | creator or listed `editor`      |
| `private`      | creator only                    | creator only                    |

Guests may read workspace-visible documents and may be shared a restricted
document, but they cannot mutate documents. If a guest is requested as an
`editor`, persistence stores `viewer`. Inactive users, removed members, deleted
workspaces, cross-workspace IDs, and non-creator private IDs fail closed. SQL
checks that a member grant applies only while the document is `restricted`, so
a stale row cannot accidentally expose a document after it becomes private.

For protected single-resource operations, a hidden or cross-tenant identifier
maps to `documentdomain.ErrNotFound`. This avoids confirming that another
workspace's document exists. Creation without a current mutation-capable
membership maps to `ErrForbidden` because no existing resource is disclosed.

## Named queries

The handwritten SQL is grouped by use case rather than by generic CRUD verbs:

- `documents.sql`: `ListAccessibleDocuments`, `GetAccessibleDocument`,
  `CreateDocument`, and `UpdateEditableDocument`.
- `access.sql`: owner-scoped visibility update, member removal, and typed bulk
  member insertion.
- `lifecycle.sql`: duplicate source locking, media copying, archive, permanent
  delete, and orphan-candidate discovery.
- `relationships.sql`: target, team, and document workspace resolution before
  association or disclosure.
- `media.sql`: exact document/attachment authorization, idempotent link, unlink,
  and final-reference calculation.

Each query accepts explicit `workspace_id` and `actor_id` parameters. There are
no reflection-built filters, parameter maps, string-concatenated predicates, or
SQLx fallbacks. Search and scope remain data parameters, not SQL fragments.

`ListAccessibleDocuments` keeps the existing optional limit contract. SQLC
generates `*int32` for `row_limit`; PostgreSQL treats a null `LIMIT` as no
limit. Public API v1 uses its separately versioned cursor contract.

## Transaction boundaries

Multi-query invariants use native pgx transactions:

- `Get` uses repeatable-read/read-only so the document, member list, and visible
  relationships come from one snapshot.
- `Update` is serializable so a returned hydrated document describes the exact
  committed update.
- `SetAccess` updates visibility, removes old members, inserts all valid current
  members, and re-reads the document atomically. If one requested user is
  invalid, the entire operation rolls back and the previous access list remains.
- `Duplicate` locks the readable source, creates a private copy, rewrites stable
  document-media paths, and copies media links in one transaction.
- relationship add/remove and media unlink are serializable.
- permanent delete locks the owner-scoped document and candidate attachment
  rows before deleting. Cascading relation removal and orphan calculation are
  one transaction.

Serializable/deadlock failures are retried a bounded four times. Cancellation
is always respected. A persistent conflict is returned to the caller; it is
never silently accepted.

## Attachment cleanup boundary

Document persistence owns database relationships, not object storage. Unlink
and delete return **orphan candidates** only after checking document, generic
Story, and inline Story relations in the same transaction. The attachment
service then performs an atomic `DELETE ... WHERE NOT EXISTS` recheck before it
removes the stored object. This second check closes the race where another
feature links the attachment immediately after the document transaction commits.

Do not add network/object-storage calls inside a document transaction.

## Generate and verify

From `apps/server`:

```bash
make sqlc-generate
make sqlc-check
go test -race ./internal/modules/documents/...
TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -tags=integration ./internal/modules/documents/repository
```

The integration suite creates a disposable database, applies the real migration
chain, and covers tenant isolation, visibility/edit policy, access rollback,
team-scoped relationships, media orphan rules, duplication, permanent delete,
and concurrent partial updates.

When adding a query, put it in the cohesive query file, regenerate SQLC, map the
row locally, add a real PostgreSQL contract test, and update this guide if an
invariant changed.
