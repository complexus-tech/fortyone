# Permanent account deletion

`DELETE /users/account` is the irreversible self-account deletion endpoint. It uses the existing first-party browser/mobile session; it does not accept API keys or a target account identifier. `DELETE /users/profile` retains its historical deactivation semantics.

A browser may send `{ "expectedUserId": "<account shown in the confirmation>" }`. This is a precondition against a cookie/account change in another tab, never a target selector. A mismatch returns 409 without changing either account. Mobile may omit the body because it sends its explicitly scoped session cookie.

Responses:

- **202** with `{ "data": { "status": "cleanup_pending" } }`: the local identity, login access and private account content are erased. Shared workspace content is retained as described below. Durable connected-service cleanup remains. This is currently the normal successful response because mailing-list contact deletion is always queued.
- **204**: protocol support for a completed deletion with no remaining cleanup.
- **409**: account confirmation changed, or the user is the last active human administrator of a live workspace. The message names workspaces and asks the user to transfer administrator access or delete them before retrying. Already-deleted workspaces do not block deletion.
- **401/404**: no authorized active account; **503**: deletion is not configured; **500**: the transaction failed and no deletion committed.

Both successful responses expire the current cookie. Every browser/mobile session also rechecks authoritative account state, so sessions on other devices immediately fail. External identities, verification tokens, personal API principals, OAuth grants and personal integrations are removed. Owned OAuth applications and their installation principals/secrets are removed, which can disconnect workspaces using those applications; unrelated workspace service accounts remain.

## Erasure and shared content

The serializable transaction locks the user and relevant memberships before checking administrator ownership. It stages provider cleanup, removes private documents, chats, memories, preferences, memberships and personal delivery snapshots, and removes the original user row where possible.

Organization-owned tasks, story comments and replies, feedback prose, story and objective activity history, and shared documents remain. Comment IDs and parent links are preserved so deleting an account does not break another person's reply thread. Historical user references are reassigned to the shared, permanently inactive **Former user** actor (`ffffffff-ffff-4fff-8fff-ffffffffffff`) before removing the original account. This actor has no membership, personal avatar, credentials, or login access; it is not the deleted person's retained profile. Active assignee and lead references become unassigned. Feedback retains its schema-required contributor key with personal identity removed.

Structured account references in retained activity and notification metadata are scrubbed. Notifications belonging to the deleted account are removed; other recipients' notifications retain their workspace context with generic authorship. New assignee snapshots carry an internal account reference so cleanup can distinguish people with the same name; this reference is stripped from public notification responses. Older assignee snapshots contain only a username, which is insufficient to identify the account reliably even when only one current member has that name. Those legacy assignee names remain rather than rewriting an unrelated person's history. Actor labels are still replaced using the notification's exact `actor_id`.

Internal mutation and scheduling delivery snapshots associated with the deleted account are removed under the existing privacy boundary, rather than weakening their immutable payload contracts. This can discard pending external deliveries from those outboxes; user-visible activity history is retained separately.

Orphan uploaded objects and the current managed avatar are queued into the existing durable storage deletion outbox. Files with a surviving story/document/feedback consumer remain available. A profile object shared by another account is not deleted. External avatar URLs are not treated as objects owned by FortyOne.

The old active-to-inactive transition starts the existing Google Drive revocation saga before account removal. Figma secrets and local connection generation are removed; remote Figma webhook removal is not confirmed by this transaction. Already-delivered external copies cannot be recalled. Immutable security audit ledgers and backup retention are not rewritten by this endpoint; access-controlled audit and delivery records can retain opaque account identifiers under their retention controls. Shared prose, including names or email addresses typed into it by any author, is preserved rather than automatically redacted. Public privacy wording must describe these boundaries rather than promise that all retained content is anonymous or that every historical copy is erased.

## Asynchronous cleanup

Calendar cleanup stages deletion of FortyOne-created provider events through the existing dispatcher. Only while it drains, an inactive account FK is retained with its email, name, avatar, personal tokens and preferences erased. A database trigger prevents updates or reactivation of that key. `account_deletion_requests` contains only the key, request time and retry metadata. The periodic finalizer hard-deletes the key once no calendar cleanup connection remains. Transient provider failures must retry rather than destroy the only cleanup credential.

Decryption failures also remain pending: startup cannot prove that the configured
secret matches the key used to encrypt an older credential. Restore the correct
secret and retry rather than dropping the cleanup row. Only positively revoked
grants or missing write scope can automatically end otherwise impossible provider
cleanup; that does not prove the provider's event was removed.

The `account_subscriber_deletions` outbox retains the minimum recipient address needed for confirmed Brevo mailing-list contact deletion. The subscriber cleanup dispatcher purges that address after a successful deletion or confirmed not-found response (204/404). Per-email advisory locking serializes contact updates, account deletion and provider deletion, including previously queued updates and a later account registered with the same email. A disabled or failing provider must leave cleanup pending rather than acknowledge a deletion that never happened.

## Rollout and local validation

Apply forward-only migrations 191, 192 and 193, deploy the account finalizer, subscriber cleanup dispatcher and existing calendar/Drive/object cleanup dispatchers, then expose the new API and client flow. Migration 193 renames the inert actor to Former user; the API transaction performs the content reassignment. Do not serve account deletion through older API instances during this rollout: the earlier implementation deletes comments and history even with the new schema. Rollback must not reactivate erased accounts, resume destructive deletion behavior, or drop pending cleanup. Tests apply migrations only to disposable local databases.

Use a disposable PostgreSQL control database with `CREATEDB` for integration tests. Each test creates and destroys its own fully migrated database:

```sh
go test ./internal/modules/users/... ./internal/modules/attachments/service ./internal/bootstrap/api ./internal/bootstrap/worker
go test -tags=integration ./internal/modules/users/uow
```

The integration suite exercises real erasure and credential cascades, preservation of shared content and reply threads with generic authorship, removal of identity snapshots, sole-admin conflicts, deleted-workspace handling, concurrent admin deletion, inactive-key finalization and protection from reactivation, durable cleanup queues, and complete rollback after a late failure. Provider network calls are tested with controlled fakes; local success is not proof of production provider cleanup.
