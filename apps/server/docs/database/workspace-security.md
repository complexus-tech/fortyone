# Workspace security and OIDC single sign-on

`internal/modules/workspacesecurity` owns tenant access policies, session metadata,
revocation epochs, and the unified administrative audit view. Migration `000203`
adds these tables. `internal/modules/enterprisesso` owns OpenID Connect connections,
subject-to-account bindings and immutable SSO audit records; migration `000206`
adds that slice. Apply migrations through the existing release process before
activating these screens. The shared Administration navigation opens Workspace
security with Policies, Single sign-on, Sessions and Audit tabs; SCIM provisioning
is a separate module and migration `000208`.

## Access policies and sessions

Policies require a current workspace administrator and optimistic
`expectedVersion`. Allowed email domains use exact normalized domain matches; an
empty list allows every domain. The guest switch controls current guest access.
Maximum session age accepts 0–720 hours, where zero turns the age restriction off.
Age is measured from the real authentication time rather than cookie renewal.
The administrator's current email must remain allowed when saving a policy.

The workspace middleware resolves live membership first, then checks these
policies and tenant revocation state on each interactive request. Session IDs are
opaque metadata, never raw cookie tokens. Session activity updates at most once
per minute. Revoking a session or a member's sessions blocks that workspace until
fresh authentication; it does not revoke access to their other workspaces.
Existing versioned sessions without newer metadata continue under default policy.
A configured age limit or revocation epoch requires those sessions to sign in
again, without inventing a new authentication time during renewal.

The session list returns up to 500 rows with `hasMore`; filter by member to narrow
it. Revocation requires a nonempty reason of at most 240 characters. All write
paths recheck the actor, current membership, and workspace lifecycle in the same
transaction as the mutation and audit record.

## Audit API

All security endpoints are under `/workspaces/{workspaceSlug}/security` and are
restricted to live workspace administrators:

| Method and path | Behavior |
| --- | --- |
| `GET`, `PUT /policy` | Read or update allowed domains, guests and session age. |
| `GET /sessions` | Filter with `userId` and `includeRevoked`. |
| `DELETE /sessions/{sessionId}` | Revoke one session with `{reason}`. |
| `POST /members/{userId}/revoke-sessions` | Revoke that member's tenant sessions with `{reason}`. |
| `GET /audit` | Filter with `actorId`, `resourceType`, `resourceId`, `from`, `to`; use signed `cursor` and `limit` up to 100. |
| `GET /audit/export` | Export matching events as CSV, bounded at 10,000 events and 20 MiB. |

Dates are RFC3339 timestamps; `from` is inclusive and `to` exclusive. Cursors
expire after 24 hours and are bound to workspace, administrator and filters. The
unified view includes existing administrative, API credential, webhook, custom
field, session, export, OIDC and SCIM events. Each source exposes an explicit safe
metadata allowlist. It does not export provider credentials, webhook bodies or
raw custom field values. Security, SSO and SCIM audit records have immutable
database guards. Successful work exports and audit exports write content-free
audit facts with exported counts.

## Configure an identity provider

1. Register a confidential OIDC web client at the provider. Use the callback URL
   displayed in Workspace security → Single sign-on, which is
   `{APP_API_PUBLIC_URL}/auth/sso/callback`.
2. Enable authorization-code flow, PKCE S256, and `openid email profile` scopes.
   The provider must expose discovery and JWKS, issue RS256 or ES256 ID tokens,
   return a verified email and fresh `auth_time`, and honor `prompt=login` with
   `max_age=0`.
3. Enter its HTTPS issuer URL, client ID and client secret. Issuer, authorization,
   token and JWKS network endpoints must be public HTTPS on port 443. Private
   issuers, IP literals, redirects and arbitrary egress targets are rejected.
4. Test SSO while signed into the administrator's existing FortyOne account.
   The initial verified provider email must match that account's current email.
   After successful return, the UI enables Require SSO for the current connection
   generation. Enabling it before this proof is rejected by the server.

The API uses the existing context-bound credential vault, derived from
`APP_AUTH_SECRET_KEY`, to encrypt the client secret. Preserve that application key
according to the existing integration-key rotation process. The secret is never
returned by the API. Set `APP_API_PUBLIC_URL`, `APP_WEBSITE_URL`, allowed origins
and the existing cookie domain consistently so browser redirects and cookies
reach the intended API and application. Provider configuration is per workspace;
there is no global IdP secret or required vendor SDK.

`GET`, `POST`, `PUT` and `DELETE /security/sso` read, create, configure and archive
the tenant connection. Updates supply `enabled`, `requireSSO`, `expectedVersion`
and an optional replacement `clientSecret`. Issuer and client ID are immutable;
archive and create a new connection to change them. Secret rotation increments
the connection generation and requires a new verified test before enforcement.
Archiving ends linked sign-ins and enforcement while retaining audit history.

## Account linking and recovery

An unlinked provider identity cannot claim an existing account solely because it
has the same email. A member first authenticates their existing FortyOne account
and starts `/auth/sso/{workspaceSlug}`. The one-time state binds that current
account, nonce and PKCE verifier to the same browser and connection generation.
The callback requires the same account to remain authenticated and a matching
verified provider email. Later sign-ins resolve the stable issuer connection and
provider subject, preserving account identity when the provider email changes.
Deleting tenant membership also removes that tenant's subject binding.

Required SSO uses a server-issued assertion in the existing opaque browser
session. It contains tenant, connection, generation and real authentication time;
it does not store ID, access or refresh tokens. Cookie renewal and native session
handoff preserve that assertion. Normal sign-in to another workspace does not
grant SSO access to this one. The session currently carries one tenant SSO
assertion; switching between tenants that require different providers can require
another provider sign-in.

After live membership and all other session policies pass, an absent or stale SSO
assertion returns `403` with `error.code=workspace_sso_required`. The Projects
workspace layout checks this before hydrating tenant data and opens
`/auth/sso/{workspaceSlug}` outside the protected layout. That page retains the
existing account session and continues to the public API sign-in endpoint.
Other session-policy denials remain separate. A signed-out member returns through
ordinary account sign-in with the recovery path preserved as the callback. SSO
errors return to the existing sign-in page with a safe error code, without a
forced logout or authenticated redirect loop.

## Verification and scope

Local tests use a signed RSA discovery/JWKS/token fixture to exercise real OIDC
verification, PKCE, nonce, issuer, audience, expiry, authorized party, verified
email and fresh authentication-time rejection. HTTP tests prove browser-bound,
one-time state, account continuity, safe callbacks and replay rejection.
Disposable PostgreSQL integration tests prove current membership/admin fences,
account-linking ownership, cross-tenant isolation, secret-generation changes,
stale version rejection, retained audit history and immutable audit guards.

These proofs do not activate an external identity provider. A deployment still
needs its real issuer/client credentials and a successful provider sign-in. This
slice implements OIDC, browser session policies and tenant audit administration;
it does not implement SAML, provider-side MFA enrollment, device fingerprinting,
or a separate native IdP login protocol. Native handoff preserves an assertion
created by browser OIDC. Programmatic developer credentials retain their existing
scope and revocation controls; browser SSO enforcement does not turn them into
IdP bearer tokens.
