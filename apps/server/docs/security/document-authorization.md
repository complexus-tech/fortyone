# Document authorization and tenant isolation

Documents contain workspace-authored HTML and can reference private work and
stored media. Treat a document ID as an untrusted identifier: UUID shape does
not prove tenancy, membership, visibility, or edit rights.

## Required identity for every operation

Every repository call receives all of the following from its caller:

1. the authoritative `workspace_id` resolved from the request;
2. the current human actor's user ID;
3. the resource ID being accessed.

The repository does not read authentication state from an HTTP context. SQL
revalidates that the actor is an active user and a current member of the exact
workspace. Mutation queries additionally reject the `guest` role. This is
defence in depth: route middleware remains useful, but cannot be the only
resource authorization layer.

## Policy sequence

For a read, PostgreSQL proves these facts in order:

```text
workspace is active
  -> actor is an active current workspace member
  -> document belongs to that workspace and is not archived
  -> visibility permits this actor
  -> related work/media also belongs to the same workspace
```

For a mutation, the query also proves that the actor is not a guest and is the
creator, a restricted editor, or (for a workspace-visible document) a current
member. Owner-only operations are access replacement, archive, and permanent
delete.

Private documents are creator-only. A `document_members` grant is effective
only while the document is `restricted`; changing the document to private
removes grants transactionally, and the SQL still ignores any stale member row
left by legacy or interrupted data. To share a private document, the creator
must explicitly change it to restricted and provide a new access list.

Returning `not found` for hidden resources is deliberate. Callers must not
distinguish another tenant's UUID from a nonexistent UUID.

## Relationship isolation

A document relationship is returned only if:

- the relationship and document have the requested workspace;
- the Story/Objective has the same workspace and still exists;
- the Story/Objective's team is owned by that same workspace;
- the actor is a current member of the related entity's team.

Someone who can read a workspace document but cannot access an associated team
sees neither the relationship details nor its contribution to
`related_work_count`. Add, remove, and reverse-list operations resolve the
target through the same team policy before changing or returning associations.

## Media isolation

Media authorization proves the exact four-part tuple:

```text
(workspace_id, actor_id, document_id, attachment_id)
```

The attachment must have the same workspace as the document. A duplicate link
is idempotent only after edit authorization succeeds. Unlink removes only the
exact relation and cannot be used to probe or detach another workspace's media.

Object URLs are resolved by the attachment service only after document media
authorization. They remain short-lived, private, no-store responses. Database
code never stores or logs a signed access URL.

## Access-list replacement

Only the document creator can replace access. Targets are validated twice:

- the service rejects zero IDs, unsupported roles, duplicates, and the owner;
- SQL inserts only active current members of the document's workspace and
  downgrades guests to viewer.

The repository compares the inserted count with the requested count inside the
transaction. Any mismatch rolls back visibility and member changes. Never
convert this to a best-effort loop: partial access replacement is a security
failure.

## HTML and content responsibility

This repository stores `content_html` and `content_text`; it does not declare
stored HTML safe for rendering. Clients and transport adapters must continue to
apply the product's content sanitization/rendering rules. SQL search parameters
are bound values, so search text cannot become executable SQL, but that does not
make HTML trusted.

## Review checklist

Before merging a document change, verify:

- every SQL resource predicate contains the authoritative workspace;
- active membership is revalidated for the actor;
- guests cannot mutate through a direct service/repository call;
- private documents ignore stale member grants;
- hidden and cross-tenant IDs have indistinguishable errors;
- team visibility is enforced for relationships;
- document and attachment workspaces must match;
- multi-table changes roll back as a unit;
- no object-storage or provider call occurs inside a transaction;
- generated SQLC values do not escape the repository;
- PostgreSQL 18 integration tests cover both allowed and denied paths.

Related standards: [authorization](./authorization.md),
[attachment authorization](./attachment-authorization.md), and
[SQLC data access](../database/sqlc.md).
