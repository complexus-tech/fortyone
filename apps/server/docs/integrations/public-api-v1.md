# Developer API v1 implementation guide

`/api/v1` is the only developer-facing HTTP boundary. The unversioned routes
remain first-party web/mobile contracts and must never mount machine bearer
authentication.

## Source of truth

- Split OpenAPI 3.1 source: `api/openapi/v1/openapi.yaml`
- Generated models and strict std-http adapters: `internal/generated/openapi/v1/api.gen.go`
- Handwritten service adapters: `internal/modules/apiv1/http`
- Developer documentation: `apps/docs/content/docs/api`

Change the split OpenAPI source first. Never hand-edit generated code. From
`apps/server`, run:

```sh
make oapi-bootstrap
make openapi-generate
make openapi-check
```

The bootstrap target installs the pinned `oapi-codegen` v2.8.0 binary under
`.tools/bin`. The check target bundles the split contract in a temporary
directory, generates into another temporary file, and fails on checked-in
drift.

## Request pipeline

Every route follows the same fail-closed order:

1. Resolve exactly one `Authorization: Bearer` credential as a PAT,
   service-account key, delegated OAuth token, or installed OAuth application
   token for the exact public-API resource.
2. Enforce the per-credential fixed-window rate limit.
3. Match a PAT, service-account, or installed-application path workspace to the
   credential workspace, or select an OAuth user's workspace only after
   rechecking current membership.
4. Enforce the operation scope.
5. Bound JSON mutation bodies to 64 KiB. An idempotent route also captures the
   exact bytes once and gives an independent reader to the generated handler.
6. Validate path, query, content type, and body against OpenAPI.
7. For an adopting write, validate and begin the scoped idempotency receipt
   before invoking the domain mutation.
8. Invoke a narrow service interface; HTTP code never imports repositories.
9. Complete an adopting receipt with only the reviewed status, JSON body, and
   content type. An identical retry replays that safe result.

Handlers repeat workspace, scope, team, membership, and role checks through
the service layer. Transport checks are an early rejection boundary, not the
authorization source of truth.

Workspace, team, story, comment, label, workflow-state, sprint, objective, and
key-result reads accept a PAT or a user-authorized OAuth token and repeat the
user's current membership/resource checks. Story creation additionally accepts
a service-account key or an installed OAuth application with `stories:write`.
Webhook management accepts a PAT or user-authorized OAuth token and still
requires the delegated user to be a current workspace administrator.
Non-human actors fail with `principal_not_supported` on user-only operations;
never attribute one to the human who created or installed it.

## OAuth audience boundary

The public API is an exact OAuth resource, not a scope alias:

```text
<APP_API_PUBLIC_URL>/api/v1
```

For production documentation this is
`https://api.fortyone.app/api/v1`. Discovery is available at
`/.well-known/oauth-protected-resource/api/v1`. Authorization, code exchange,
refresh, and access verification must all carry that exact resource. A token
whose audience is `<APP_API_PUBLIC_URL>/mcp` cannot authenticate here, and an
API-audience token cannot authenticate to MCP or an unversioned product route.

The delegated public-API OAuth policy supports `offline_access`, `workspaces:read`,
`teams:read`, `stories:read`, `stories:write`, `comments:read`, `labels:read`,
`sprints:read`, `objectives:read`, and `webhooks:manage`. Every grant includes
`offline_access` and must request at least one API capability. The authorization
server implements authorization code with S256 PKCE for public clients.

Managed confidential applications may also use `client_credentials`, but only
for an exact active workspace installation and an explicit subset of the
installation scopes. This release permits only `stories:write`; there is no
`offline_access` or refresh token. The installation has a dedicated
`oauth_application` principal. Its administrator is historical lifecycle
metadata and must never become the story reporter, actor, receipt identity, or
rate-limit identity. See [`developer-oauth.md`](../security/developer-oauth.md).

## Story-create idempotency contract

`POST /api/v1/workspaces/{workspaceId}/stories` is the first adopting mutation.
It requires `stories:write` and a 16–255-byte visible-ASCII `Idempotency-Key`.
The receipt scope is:

```text
principal kind + stable credential identity + workspace ID
+ POST + stories.create + SHA-256(idempotency key)
```

For PAT, service-account, and delegated OAuth callers, existing actor rules
continue to apply. For `oauth_application`, the stable credential identity is
the installation UUID, never the short-lived access-token `jti` or installer
user ID. The one-way external creation key uses the same installation identity,
so obtaining a fresh access token cannot create a second story for the same
logical retry.

The request hash covers the exact captured JSON bytes. Do not decode and
re-encode before hashing: member order, whitespace, numeric spelling, and
trailing bytes are contract-significant. Raw keys and bodies are not stored.

The route returns:

- `201` for a newly created story or a completed exact replay;
- `409 idempotency_in_progress` plus `Retry-After` while the current lease is
  live;
- `409 idempotency_key_reused` when an unexpired scoped key is paired with
  different request bytes; and
- a retryable `503` with `Retry-After` if receipt coordination cannot safely
  proceed. Other terminal `5xx` responses are not automatic retry signals.

The current service uses a two-minute lease and retains completed receipts for
24 hours. Clients must never intentionally recycle keys after that window.
They generate one random key per logical operation, persist the key and exact
request before sending, and retain both until recovery is complete.

Crash safety does not rely on the receipt alone. Story creation derives a
tenant/principal-scoped, one-way external creation identity from the key and
uses the story repository's uniqueness contract. A crash after the story
transaction commits but before receipt completion therefore converges on the
same story after lease recovery. The transactional story mutation event also
prevents an idempotent retry from publishing a duplicate `story.created`
event. See [`idempotency-receipts.md`](../security/idempotency-receipts.md) for
the shared receipt lifecycle and adoption checklist.

## Compatibility rules

- Keep `{ "error": { "code", "message", "requestId" } }` stable.
- Keep cursors opaque, signed, purpose-separated, and bound to principal,
  workspace, page size, and filters.
- Add fields compatibly; do not silently rename or repurpose existing fields.
- Keep signing secrets show-once and out of logs, errors, lists, and audit data.
- Keep idempotency keys and exact request bodies out of logs; document every
  adopting mutation's crash-recovery invariant before exposing the header.
- Declare all HTTP outcomes in OpenAPI, including `413`, `415`, `429`, and
  dependency `503` responses.
- Add a contract/security test before exposing another generated operation.

The first story/team cursor implementation adapts stable service ordering to a
signed offset cursor. A future keyset migration may change cursor internals
without changing the wire contract; clients must never inspect cursor values.
