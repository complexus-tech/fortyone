# Developer OAuth security contract

FortyOne implements authorization code with PKCE for delegated remote MCP and
public-API clients. This developer authorization server is deliberately
separate from provider-install OAuth (GitHub, Slack, Figma). PATs and
service-account keys remain alternative credential families for `/api/v1`.

## Current release boundary

The authorization server exposes two exact, non-interchangeable resources:

| Audience        | Exact resource                | Protected-resource metadata                                                       | OAuth scopes                                                 |
| --------------- | ----------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Remote MCP      | `<APP_API_PUBLIC_URL>/mcp`    | `/.well-known/oauth-protected-resource/mcp` (the root metadata route is an alias) | `mcp:access offline_access`                                  |
| Public REST API | `<APP_API_PUBLIC_URL>/api/v1` | `/.well-known/oauth-protected-resource/api/v1`                                    | `offline_access` plus one or more published API capabilities |

In production, the API resource is exactly
`https://api.fortyone.app/api/v1`. A token for one resource cannot authenticate
to the other resource or to an unversioned web route, even when the same client
and user have grants for both audiences.

FortyOne releases two deliberately separate actor models:

- authorization code plus S256 PKCE creates an `oauth_user` token acting for
  the consenting user; every use case still checks the user's current
  workspace, membership, role, team, and resource access; and
- `client_credentials` creates an `oauth_application` token for one active,
  admin-approved workspace installation. The runtime principal is the
  installation's dedicated principal, never the administrator who installed
  it.

The application-actor release is intentionally narrow: it supports only
idempotent story creation with the explicit `stories:write` installation
scope. Reads, updates, deletes, comments, and webhook administration remain
user-only. Adding a scope to the OAuth catalog does not release it for an
application actor; it requires a new installation-aware product policy and
database authorization branch.

Dynamic registration accepts public clients only. Registrations have an
operator-configured lifetime from one hour through 90 days (30 days by
default), at most ten exact redirect URIs, and no client secret. This preserves
remote MCP interoperability without creating an ungoverned permanent client
catalog. Managed confidential applications are created only through the
workspace-administrator management boundary; confidential status is never
inferred from dynamic-registration input. A current active workspace admin with
`integrations:manage` must explicitly install the application for the exact
`/api/v1` resource and grant `stories:write`.

## Secret handling

Authorization codes, refresh tokens, and confidential client secrets contain
an explicit format/version, a random 12-character lookup prefix, and 32 random
secret bytes. Plaintext is shown only in the protocol or management response
that creates or rotates it. Lists expose only the prefix and lifecycle
metadata. PostgreSQL stores a domain-separated HMAC-SHA-256 digest and
digest-key ID, never plaintext.

Client-secret rotation requires an explicit overlap from one minute through 24
hours. The old secret remains usable only until the earlier of its normal
expiry or stored `overlap_expires_at`; the cutoff is immutable once set.
Revocation ends new exchanges immediately. Application access tokens are
short-lived and remain valid after a single secret is revoked, but every token
verification reloads the current application, installation, principal, exact
resource, and installed scopes. Revoking an installation and disabling its
principal therefore invalidates already-issued tokens immediately.

`APP_OAUTH_TOKEN_HMAC_KEYS` is a JSON keyring supporting read-old/write-new
rotation. `APP_OAUTH_TOKEN_HMAC_ACTIVE_KEY_ID` selects the write generation.
`APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY` signs short-lived HS256 access tokens and
must be independent from the digest keyring, browser auth secret, verification
and invitation keys, API-credential digest keyring, and provider vault keys.
Production startup rejects the public development values, malformed or weak
keys, missing active IDs, duplicate material, and detected cross-purpose reuse.

Delegated access tokens contain the user subject and grant ID. Application
access tokens instead contain the dedicated principal subject, application ID,
installation ID, workspace ID, client ID, exact audience, and explicit scopes;
they never contain or resolve through the installer. The JWT `jti` is retained
only for audit correlation. The stable credential, rate-limit, and idempotency
identity is the installation UUID.

## Authorization and PKCE

- Redirect URI comparison is exact after registration-time validation. HTTPS is
  mandatory except for literal loopback development hosts.
- The client and exact redirect are validated before any authorization error is
  sent through a callback. An invalid client/redirect receives a direct error,
  so a wrong `resource` cannot be turned into an open redirect.
- Only `S256` PKCE is accepted. Challenges must decode to 32 bytes; verifiers
  use the RFC unreserved character set and 43-128 character length.
- The requested resource must exactly equal one configured audience at
  authorization, code exchange, refresh, and access verification. Resource-
  specific verifiers prevent an MCP token from being replayed at `/api/v1` and
  vice versa.
- The consent handoff is random, five-minute, single-use, stored under a digest
  key in Redis, and bound to the browser session user.
- Codes are single-use. A client, redirect, resource, digest, or PKCE failure
  does not consume a valid code.
- A new consent decision invalidates every outstanding code for the same
  application, user, and resource before replacing the grant scopes. An older
  code therefore cannot inherit permissions approved by a later decision.
- Refresh tokens rotate on every successful exchange. Reuse commits family-wide
  revocation and an immutable audit event before returning `invalid_grant`.
- Protocol error logs contain operation and reason class only; they never
  contain a code, refresh token, access token, state, verifier, header, or body.
- Client credentials authenticate only through exactly one bounded
  `Authorization: Basic` header. `client_id` or `client_secret` in a form body
  or query string fails with `invalid_client`; there is no fallback channel.
- A client-credentials request includes the exact installation UUID, resource,
  and an explicit scope subset. The MCP resource returns
  `unauthorized_client` for application actors.
- Successful client authentication, coarse last-used updates, and the immutable
  issuance audit append commit in one database transaction. If the audit write
  or commit fails, no token response is released.

Dynamic registration accepts exactly one bounded, strict JSON object; unknown
fields and trailing JSON values fail closed. OAuth endpoints also have
Redis-backed global fixed-window limits, while authenticated MCP traffic is
limited per durable grant to 600 requests per minute. Redis failure fails
closed. These application limits are the final safety net, not a replacement
for per-source limits and anomaly controls at the managed edge.

## Resource-specific scopes

The MCP resource requires both `mcp:access` and `offline_access`.

The `/api/v1` resource supports:

```text
offline_access
workspaces:read
teams:read
stories:read
stories:write
comments:read
labels:read
sprints:read
objectives:read
webhooks:manage
```

`offline_access` is mandatory and every API grant must also contain at least
one explicit API capability. An API request still passes route scope checks and
authoritative product authorization. `comments:read` routes additionally
require `stories:read`; workflow-state reads use `stories:read`; key-result
reads use `objectives:read`; and webhook management requires the delegated user
to remain a workspace administrator.

Do not put MCP and API scopes in one authorization request. Unknown scopes,
MCP scopes requested for the API resource, and API scopes requested for the MCP
resource fail closed.

Those rules describe delegated `oauth_user` grants. An application installation
does not receive `offline_access` or a refresh token and currently accepts only
the exact explicit scope `stories:write`. The token endpoint returns an access
token only; the application obtains another one with its client secret when
needed.

## Choosing another credential

- Use a PAT for a personal script or a simple user-owned integration that does
  not need an authorization redirect. PAT access ends with expiry, revocation,
  membership loss, or user deactivation.
- Use a service account for workspace-owned automation where the published
  operation explicitly supports a non-human principal. The current `/api/v1`
  preview supports service accounts for idempotent story creation; it never
  borrows the creator's identity.
- Use delegated OAuth for a third-party client that asks each user for
  least-privilege access across the workspaces that user can currently access.
- Use a managed OAuth application installation when a third-party backend must
  create stories without a user session and a workspace admin has explicitly
  approved that exact workspace and scope.

## Incident response

For one leaked refresh token, call `/oauth/revoke`; the complete family becomes
unusable. For replay detection, require a new consent flow. For a signing-key
exposure, replace `APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY` and restart all API
instances; all existing access JWTs become invalid. For a digest-key exposure,
add a new active key, retain the old key only while live code/refresh records
need verification, and require reauthorization if confidentiality is uncertain.
Never paste the suspected credential into logs, tickets, shell history, or SQL.

For a suspected client-secret leak, rotate it with the shortest safe bounded
overlap, update the integration, then revoke the old secret. Revoke the complete
installation when application access itself is no longer trusted; this also
disables the dedicated principal and invalidates live application tokens on
their next verification.

## Required evidence

Focused tests cover secret redaction, keyed digest verification, read-old/
write-new rotation, exact redirect/resource/client binding, PKCE failure without
code consumption, current-user validation, access-token revocation, concurrent
authorization versus reauthorization, scope-replacement fencing, refresh
rotation, family replay revocation, and audit immutability. Repository tests run
against migrated PostgreSQL 18 with the race detector.

Application-actor evidence additionally covers Basic-only authentication,
alternate secret-channel rejection, digest-only persistence, explicit overlap
cutoffs, concurrent rotation chains, idempotent concurrent revocation, exact
resource and installed-scope checks, installer-attribution denial, stable
installation rate/idempotency identity, fail-closed audit persistence, live
token invalidation, cross-workspace denial, and the story-create-only SQL
authorization boundary.
