# Document collaboration

Self-hosted Tiptap + Yjs + Hocuspocus, with PostgreSQL persistence and no paid service dependency. The Go API provides document-scoped session tokens, permissions, history, restoration, and public sharing. This service synchronizes title/body edits, carets, and presence.

## Deploy history and public sharing now

No new environment variables or collaboration service are required for this release:

1. Apply migration **194** through the normal database migration process.
2. Deploy the updated Go API and Projects app using their existing environment configuration.
3. Leave `NEXT_PUBLIC_COLLABORATION_URL` unset. Documents use the existing HTTP autosave with revision checks; version history, restoration, and public link sharing are available.

The schema includes dormant collaboration columns/tables so the feature can be enabled later without deleting or rewriting this implementation. Migration 194 does not activate collaboration on existing documents. The optional service is excluded from the root `pnpm dev` task and has no build task required by Projects. Do not deploy its container for this release.

Public sharing uses the existing API URL and application host. Owners enable view-only links from the document sharing menu; links remain off until explicitly enabled. Existing open document tabs should reload after rollout because saves now include `expectedRevision`.

## Enable collaboration later

1. Apply migration **194** using the existing migration process. It adds document revisions, CRDT state, scoped sessions, public tokens, and history triggers. The migration backfills the current version of existing documents; it cannot reconstruct earlier edits. Take a database backup using the normal production process.
2. Deploy the updated Go API and Projects app together. Document REST updates now require `expectedRevision`. Existing open tabs should reload. Leave `NEXT_PUBLIC_COLLABORATION_URL` unset until the collaboration service is available; version history/public links work independently.
3. Run one collaboration instance against the same database. Supply `DATABASE_URL`, `DATABASE_CA_FILE` (production), and `COLLABORATION_ALLOWED_ORIGINS`. Do not put `sslmode`, `sslcert`, `sslkey`, or `sslrootcert` in the URL: node-postgres URL SSL options can override the explicit verified CA configuration.
4. Expose port 1234 through a TLS proxy with WebSocket upgrade support. Set the proxy's idle timeout above 60 seconds and apply connection/rate limits. `/healthz` returns 200 only when PostgreSQL is reachable. Use `COLLABORATION_ALLOWED_ORIGINS=https://cloud.fortyone.app,https://*.fortyone.app` for the production workspace hosts, or an explicit comma-separated list. Only HTTPS suffix wildcards are supported.
5. Configure the Projects app's `NEXT_PUBLIC_COLLABORATION_URL=wss://<collaboration-host>` through its existing public environment mechanism, then restart/redeploy it. Verify two authenticated workspace users editing the same document and a signed-out public viewer.
6. Before scaling beyond one collaboration process (including overlapping rolling deployments), configure `COLLABORATION_REDIS_HOST`, port, password and TLS against a shared Redis instance. Redis carries synchronization/presence; PostgreSQL remains authoritative. Without Redis, use one replica and a stop/start deployment strategy.

Build from the repository root with `docker build -f apps/collaboration/Dockerfile -t fortyone-collaboration .`. The container runs as the non-root `node` user. Mount database CA files read-only and inject secrets through your deployment environment. Local development: export the variables in `.env.example`, then run `pnpm --filter collaboration dev:collaboration`.

## Behavior and boundaries

- Accepted edits are committed before broadcast/acknowledgement. The editor shows Saved only after synchronization acknowledgement. Reconnects merge unacknowledged updates while the editing session remains valid.
- Permissions are checked for each message, again inside write transactions, and every five seconds for idle connections. Tokens expire after one hour and are refreshed on reconnect; password/session revocation is enforced through `auth_session_version`.
- Every durable content/state change creates an immutable revision. History is paginated. There is no automatic retention purge: budget storage accordingly. Media referenced by historical versions is retained until document deletion.
- Restore creates a new version and advances the document's collaboration epoch. Existing editors pause and must reload; stale/offline updates cannot undo the restore. There is no offline browser persistence: copy unsaved text before closing a disconnected editor.
- Once collaboration is activated on a document, legacy HTML writes are fenced in both API and database. Do not disable the service as a rollback without a deliberate state migration: activated documents become read-only when collaboration is unavailable.
- Public links are **view only**, owner-controlled, disabled by default, and revocable. Re-enabling creates a new unguessable link. No account is required to view. They expose title/body and embedded media, not history, members, related work, or Drive attachments. Search engines receive noindex, and responses are not cached. Recipients can still copy content they have already viewed. Media redirects are signed for one minute, so an already-issued URL can remain usable briefly after revocation.
- Collaborative document state and individual updates are limited to 4 MiB. This includes accumulated CRDT metadata; a document reaching the cap requires recovery/compaction, not repeated retries. Uploaded media stays outside this state.

## Verification

- `pnpm --filter collaboration type-check`
- `pnpm --filter collaboration test`
- `TEST_DATABASE_URL=postgres://...@127.0.0.1:5432/postgres?sslmode=disable pnpm --filter collaboration test:integration`
- From `apps/server`: `TEST_DATABASE_URL=... go test -tags=integration ./internal/modules/documents/repository -count=1`

The integration harness accepts only localhost PostgreSQL, creates a random temporary database, applies all migrations, exercises real WebSocket clients, and drops only its own database. It requires the server's `.tools/bin/migrate` binary. Tests cover concurrent initial loading, simultaneous edits and durable acknowledgement, read-only writes, membership revocation, history, and stale-session rejection after restore. Production deployment, TLS, Redis fanout, and authenticated browser acceptance remain deployment checks.
