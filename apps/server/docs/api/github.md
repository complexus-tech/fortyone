# GitHub API contract

This document describes the current legacy JSON API for GitHub integration
management and collaboration. It is intended for FortyOne clients and future
out-of-process integrations while the endpoints are moved into the versioned
OpenAPI surface.

## Common conventions

- Workspace routes require a valid FortyOne session and current membership in
  `{workspaceSlug}`. Admin routes also recheck the actor's current workspace
  role inside the service before performing side effects.
- JSON requests require `Content-Type: application/json`.
- Successful JSON responses use `{ "data": ... }`. A 204 response has no body.
- Errors use `{ "error": { "code", "message", "hint", "requestId" } }`.
  Server and dependency failures return a generic message; database/provider
  causes, authorization codes, OAuth state, tokens, and signed bodies are not
  returned.
- UUID path parameters must be canonical UUIDs. Unknown workspace resources do
  not authorize cross-workspace access.

These routes currently use browser-session authentication. They are not yet a
public machine-credential surface. External integrations should not scrape or
reuse browser cookies; a versioned, scoped API-key/OAuth contract must be added
before third-party use.

## Integration administration

| Method   | Path                                                                        | Role   | Success                                                                           |
| -------- | --------------------------------------------------------------------------- | ------ | --------------------------------------------------------------------------------- |
| `GET`    | `/workspaces/{workspaceSlug}/integrations/github`                           | member | 200 integration aggregate                                                         |
| `POST`   | `/workspaces/{workspaceSlug}/integrations/github/install-session`           | admin  | 200 `{ "installUrl": "https://github.com/apps/.../installations/new?state=..." }` |
| `POST`   | `/workspaces/{workspaceSlug}/integrations/github/repositories/resync`       | admin  | 200                                                                               |
| `GET`    | `/workspaces/{workspaceSlug}/integrations/github/settings`                  | member | 200 workspace settings                                                            |
| `PUT`    | `/workspaces/{workspaceSlug}/integrations/github/settings`                  | admin  | 200 workspace settings                                                            |
| `POST`   | `/workspaces/{workspaceSlug}/integrations/github/issue-sync-links`          | admin  | 201 issue-sync link                                                               |
| `PUT`    | `/workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}` | admin  | 200 issue-sync link                                                               |
| `DELETE` | `/workspaces/{workspaceSlug}/integrations/github/issue-sync-links/{linkId}` | admin  | 204                                                                               |

### Workspace settings

Updates are partial. Omitted properties retain their current value.

```json
{
  "branchFormat": "username/identifier-title",
  "linkCommitsByMagicWords": true,
  "syncAssignees": true,
  "syncLabels": true,
  "autoPopulatePrBody": true,
  "closeOnCommitKeywords": true
}
```

`branchFormat` must be one of `username/identifier-title`, `identifier-title`,
or `identifier/title`.

### Issue-sync links

Create request:

```json
{
  "repositoryId": "3de4f3f3-21a7-4c45-96b7-e37a7df94cd8",
  "teamId": "96e0945b-1fd0-4471-a275-aebc998544aa",
  "syncDirection": "bidirectional"
}
```

`syncDirection` is `inbound_only` or `bidirectional`. Update accepts optional
`syncDirection` and `isActive` properties. Repository and team identity are
validated inside the current workspace.

## Team workflow settings

| Method | Path                                                         | Role   | Success                  |
| ------ | ------------------------------------------------------------ | ------ | ------------------------ |
| `GET`  | `/workspaces/{workspaceSlug}/teams/{teamId}/settings/github` | member | 200 settings and rules   |
| `PUT`  | `/workspaces/{workspaceSlug}/teams/{teamId}/settings/github` | admin  | 200 replacement settings |

The update body replaces the complete rule set and accepts at most 64 rules:

```json
{
  "rules": [
    {
      "eventKey": "pr_merge",
      "targetStatusId": "64bdb36e-aa5e-4710-a8dd-6359e9c96a52",
      "baseBranchPattern": "release/*",
      "isActive": true
    }
  ]
}
```

Supported event keys are `draft_pr_open`, `pr_open`, `pr_review_activity`,
`pr_ready_for_merge`, `pr_merge`, `issue_open`, `issue_reopen`, `issue_close`,
and `commit_close`. A branch pattern is either an exact branch or a prefix
ending in `/*`; it is not an arbitrary regular expression.

## Story links and comments

| Method   | Path                                                                           | Role   | Success                |
| -------- | ------------------------------------------------------------------------------ | ------ | ---------------------- |
| `GET`    | `/workspaces/{workspaceSlug}/stories/{storyId}/github-links`                   | member | 200 list               |
| `DELETE` | `/workspaces/{workspaceSlug}/stories/{storyId}/github-links/{linkId}`          | member | 204                    |
| `GET`    | `/workspaces/{workspaceSlug}/stories/{storyId}/github-comments`                | member | 200 flattened comments |
| `POST`   | `/workspaces/{workspaceSlug}/stories/{storyId}/github-comments`                | member | 200                    |
| `GET`    | `/workspaces/{workspaceSlug}/integration-requests/{requestId}/github-comments` | member | 200 comments           |
| `POST`   | `/workspaces/{workspaceSlug}/integration-requests/{requestId}/github-comments` | member | 200                    |

Comment writes accept:

```json
{
  "body": "The fix is ready for another review."
}
```

The trimmed body must not be empty and is limited to 65,536 characters by
request validation. Reads page through GitHub instead of silently returning only the
first page and stop at the service's 1,000-comment safety limit.

### Safe retries

Clients that may retry a comment write must send exactly one stable header:

```http
Idempotency-Key: <stable-operation-id>
```

The key is scoped to operation, workspace, actor, and story/request. Reusing
the same key and body is successful and does not intentionally create another
GitHub comment. Reusing it with different content returns 409. A malformed or
duplicated header returns 400. Omitting the header opts out of retry
deduplication.

Story comment writes target every linked GitHub issue and are fail-fast. If a
later target fails, retry with the same key and body: completed targets are
recognized by their opaque marker and skipped.

## User OAuth linking

| Method   | Path                                     | Access                  | Success          |
| -------- | ---------------------------------------- | ----------------------- | ---------------- |
| `POST`   | `/user/integrations/github/link-session` | authenticated user      | 200 opaque state |
| `POST`   | `/user/integrations/github/link`         | same authenticated user | 200              |
| `DELETE` | `/user/integrations/github/link`         | authenticated user      | 204              |

Create a link session with an allowed application return destination:

```json
{
  "returnTo": "https://app.fortyone.example/settings/integrations"
}
```

The response contains an opaque `state` string. It is a 256-bit one-time bearer
value, not a JSON object or signed return payload. The client sends it to the
GitHub authorization request and then posts the provider result:

```json
{
  "code": "provider-authorization-code",
  "state": "opaque-43-character-base64url-state"
}
```

State is consumed before code exchange. It expires after 15 minutes and is
bound to the authenticated FortyOne user. Reuse, user mismatch, malformed
encoding, expiry, or state-store failure returns 400 and requires a fresh
session.

Unlink remotely revokes the OAuth token before clearing the local encrypted
credential. A retryable revocation/configuration outage returns 503 and keeps
the local credential so the caller can retry.

### Frontend compatibility seam

The API no longer embeds `returnTo`, workspace identity, or any JSON payload in
the browser-visible state. A client that still tries to split/decode state as a
signed payload will not recover the return path and must use a safe application
fallback. The API intentionally does not restore stateless state for backward
compatibility. Frontend remediation belongs to the frontend project and is not
part of this API-only change.

## GitHub App setup callback

`GET /integrations/github/setup` is provider-facing and requires exactly one
positive `installation_id` and exactly one opaque `state` query parameter.

The state is consumed atomically, then the API rechecks that the initiating
user is still an administrator in the stored workspace. It fetches the
installation and complete repository catalog, persists the ownership-fenced
grant, and returns a 307 redirect constructed from the configured FortyOne
website origin and stored workspace slug. Callback query strings and state are
never logged.

There is no supported `returnTo` query parameter or signed state payload in this
callback contract.

## GitHub webhook ingress

`POST /webhooks/github` is not session-authenticated. It requires exactly one
value for:

- `X-Hub-Signature-256: sha256=<hex HMAC>`;
- `X-GitHub-Delivery: <stable delivery id>`; and
- `X-GitHub-Event: <supported event>`.

The body limit is 1 MiB. The signature covers the exact bytes; intermediaries
must not rewrite JSON or line endings. On durable acceptance:

```json
{
  "data": {
    "accepted": true,
    "duplicate": false
  }
}
```

The status is 202. A terminal duplicate returns 202 with `duplicate: true` and
does not intentionally redispatch. Authentication failures return 401.
Malformed/unsupported deliveries use a safe 4xx response. Durable persistence,
authorization lookup, or dispatch availability failures return 503 so GitHub
can retry. Internal causes and signed content are never echoed.

Supported `X-GitHub-Event` values are `issues`, `pull_request`,
`pull_request_review`, `issue_comment`, `check_run`, `create`, and `push`.

## Stable status semantics

| Status | Meaning for these routes                                                                              |
| -----: | ----------------------------------------------------------------------------------------------------- |
|    400 | malformed path/body/query, invalid OAuth state/code, or invalid idempotency key                       |
|    401 | missing FortyOne authentication or invalid webhook signature                                          |
|    403 | current workspace role is insufficient                                                                |
|    404 | requested collaboration resource has no linked GitHub issue                                           |
|    409 | idempotency key reused with different comment content or canonical webhook identity conflict          |
|    413 | request body exceeds the endpoint limit                                                               |
|    431 | signed header set exceeds the gateway bound                                                           |
|    503 | provider, credential, state, verification-store, durable inbox, or dispatch dependency is unavailable |
|    500 | unexpected internal failure; response remains sanitized                                               |

Retry 503 and transport failures with exponential backoff only when the
operation is idempotent. For comment writes, that means supplying the same
`Idempotency-Key` and request body.

## Current external-integration boundary

The internal `codehost` contracts prove repository catalog, work-item write,
comment write, and webhook normalization without leaking GitHub SDK types. They
are not yet a public API. Before exposing them to third parties, add:

1. versioned OpenAPI endpoints;
2. scoped, expiring machine credentials;
3. tenant and installation-generation authorization on every call;
4. per-credential rate limits and audit events;
5. idempotency receipts for mutations; and
6. signed outbound webhooks for asynchronous results.

Do not allow arbitrary in-process Go plugins. Third-party integrations run out
of process against that versioned boundary.
