# Developer OAuth persistence

The `developeroauth` module owns the durable OAuth authorization-code state used
by the exact remote MCP and `/api/v1` resources. PostgreSQL is authoritative.
Redis stores only the five-minute browser consent handoff tied to the
authenticated session; it never stores an OAuth application, authorization
code, grant, or refresh token.

## Tables and ownership

| Table                                   | Purpose                                            | Important invariant                                                                                                 |
| --------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `oauth_applications`                    | Governed public or confidential client metadata    | Managed confidential rows have immutable owner workspace/user metadata; dynamic public clients remain separate.     |
| `oauth_application_redirect_uris`       | Exact registered callbacks                         | Authorization compares the complete URI; there is no prefix, wildcard, or sibling-domain match.                     |
| `oauth_client_secrets`                  | Show-once confidential credential lifecycle        | Only a prefix, keyed digest, key ID, immutable rotation link, expiry, overlap cutoff, and revocation state persist. |
| `oauth_application_installations`       | Admin-approved application access to one workspace | Each installation owns one dedicated `oauth_application` principal and one exact resource.                          |
| `oauth_application_installation_scopes` | Explicit application capability grants             | The first release admits only `stories:write`; scope presence never implies installer membership.                   |
| `oauth_grants`                          | User consent for one application and resource      | Delegated grants remain `oauth_user` and recheck that the user remains active.                                      |
| `oauth_grant_scopes`                    | Normalized delegated least-privilege scopes        | An empty delegated grant cannot authenticate.                                                                       |
| `oauth_authorization_codes`             | Five-minute, single-use PKCE code                  | Only a lookup prefix, keyed digest, key ID, bindings, and lifecycle timestamps are stored.                          |
| `oauth_refresh_token_families`          | Rotation and family-wide revocation                | A replay revokes the complete family in the same transaction that records the audit event.                          |
| `oauth_refresh_tokens`                  | One generation in a refresh family                 | `parent_token_id` is unique, so a generation can have only one replacement.                                         |
| `oauth_audit_events`                    | Immutable security ledger                          | Actor, stable credential, subject, request, and lifecycle UUID facts have no mutable foreign keys.                  |

Delegated OAuth is introduced by migration `000163_developer_oauth`; application
actors are added by forward-only migration
`000170_oauth_application_actors`. Application SQL
is defined in `internal/modules/developeroauth/repository/queries` and generated
through the `developer_oauth` SQLC package. The repository uses native pgx and
owns every transaction; HTTP code has no database dependency.

## Transaction boundaries

Consent first invalidates every outstanding code for the `(application, user,
resource)` subject, then reactivates or creates the grant, replaces its scopes,
revokes superseded refresh families, creates the new one-time authorization
code, and appends the audit event atomically. The application and user must
still be active when the repository writes the grant; an earlier HTTP-layer
check is not treated as durable authorization.

Code exchange locks the code before reading its grant and scope projection,
verifies the stored keyed digest, client, redirect URI, resource, actor, and
PKCE challenge, consumes the code, creates the family and first refresh
generation, and appends the audit event in one transaction. Reauthorization
uses the same code-before-grant lock order, preventing both deadlock and scope
replacement while an older code is exchanging. A failed verifier or binding
check rolls back without consuming the code.

Refresh exchange locks the token, family, and grant. The first caller consumes
the generation and inserts its uniquely parented replacement. A later caller
presenting the consumed generation is treated as replay: the family revocation
and `refresh_token.reuse_detected` audit event are committed before the public
`invalid_grant` response is returned.

Managed application creation rechecks the current active workspace admin in
SQL, inserts the confidential application, exact redirects, initial digest-only
secret, and attributed audit event in one transaction. Rotation locks the
application and current secret-chain head, inserts exactly one replacement,
and stores the old secret's explicit `overlap_expires_at`. Concurrent rotations
serialize into one chain. Effective old-secret expiry is
`min(expires_at, overlap_expires_at)` because authentication requires both
predicates to remain active.

Installation creation rechecks the current admin, locks the active
confidential application, inserts a distinct `oauth_application` principal,
inserts the workspace/resource installation and exact scopes, then appends the
audit event atomically. The principal may carry bounded `member` schema
metadata, but no application authorization query derives access from that role.
Installation update/revocation locks both installation and principal.
Revocation changes the installation to `revoked` and principal to `disabled` in
the same transaction.

Client-credentials exchange locks the exact `(lookup_prefix, installation_id)`
candidate and validates current secret expiry/overlap/revocation, confidential
application status/expiry, installation resource/status, principal status, and
installed scopes. The service performs the keyed constant-time digest check
and signs the token inside that transaction. Secret/installation last-use
updates and the immutable `client_credentials.exchanged` audit append must then
commit before plaintext token bytes are returned. An audit or commit failure
rolls back and fails closed.

## Query review checklist

- Lookup predicates use the random 12-hex-character prefix only to locate a
  candidate; service verification always performs the keyed constant-time
  digest comparison.
- Code and refresh queries repeat current application, grant, user, resource,
  expiry, and revocation checks.
- Code-to-grant application and resource identity, and refresh-family-to-grant
  resource identity, are repeated in joins even though normal writes create
  consistent rows.
- Token text, signing keys, digest keys, PKCE verifiers, and browser state are
  never selected into general projections or audit metadata.
- New scopes require coordinated audience-specific domain validation, database
  constraint, consent copy, and documentation changes. Never add a scope to the
  global catalog without adding it to exactly the resource policies that may
  issue it.
- Delegated grants remain `actor_kind = 'oauth_user'`. Application actors use
  installation rows, not `oauth_grants`, and never receive refresh-token
  families.
- Application authentication is keyed by both secret prefix and installation
  ID. A client ID alone never selects a workspace or principal.
- The stored installer is lifecycle metadata only. Runtime and audit actor ID
  must equal the dedicated installation principal; stable actor credential ID
  must equal the installation ID.
- JWT `jti` is an `access_token` audit subject only. It must not be used as a
  rate-limit, idempotency, or durable credential identity.
- Story SQL admits `oauth_application` only in `AuthorizeStoryCreate` and the
  dedicated idempotent creation replay query. General mutation snapshots,
  updates, deletes, reads, and webhook administration remain denied.
