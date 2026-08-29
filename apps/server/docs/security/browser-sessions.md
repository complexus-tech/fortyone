# Browser sessions and OAuth state

Browser sessions and Google/Microsoft OAuth state are random 256-bit bearer
values. The browser receives the raw value, but Redis keys use a versioned
SHA-256 digest so a cache snapshot, key listing, or Redis command log does not
become a bearer-token listing. Cache error logs contain only a short one-way key
fingerprint.

OAuth state is consumed with one Redis Lua operation that reads and deletes the
value atomically, so two concurrent callbacks cannot both use the same state.
Session logout deletes both the current digest key and the temporary legacy key.

## Session revocation and account state

Each Redis session record contains a typed `(user_id, version)` value. Every
authenticated request loads the active account's authoritative
`auth_session_version` from PostgreSQL and requires an exact match. Redis is
therefore only an opaque-token index; it never overrides account activation or
revocation state.

Self-deactivation, scheduled inactivity deactivation, administrator state
changes, and explicit administrator revocation atomically advance the
PostgreSQL version. Existing Redis records can remain until TTL expiry because
their older version immediately fails closed. Reactivation never resets or
decrements the epoch, so an old cookie cannot become valid again.

Legacy string-valued Redis records and raw-token cache keys are deliberately not
accepted: they contain no revocation epoch. Logout still deletes the current
digest key and the temporary legacy key as cleanup, but legacy-key deletion is
not an authentication compatibility path.

## Deployment compatibility

- Apply migration 000171 before starting replacement APIs or workers.
- Drain all old APIs and workers before issuing versioned sessions. Old
  processes neither enforce nor consistently advance the PostgreSQL epoch.
- Replacement APIs write only digest-derived keys containing structured
  versioned records. Pre-cutover browser sessions require a one-time sign-in.
- Replacement APIs may consume legacy Google/Microsoft OAuth state during its
  ten-minute TTL; OAuth state does not grant a browser session by itself.
- Old API instances cannot read sessions or OAuth state created by a
  replacement instance and must not return to service after cutover.

Redis session data needs no backfill because it expires and cannot prove the new
epoch. PostgreSQL migration 000171 initializes the authoritative version and
has the coordinated rollout and guarded recovery contract documented in
[`docs/database/migration-operations.md`](../database/migration-operations.md).

The session store is injected into authentication middleware and authorization
handlers. There is no mutable package-global cache pointer, so constructing two
apps or running parallel handler tests cannot redirect one app's session lookup
to another Redis client.

First-party routes authenticate only this opaque session cookie. They no longer
accept the legacy HS256 user bearer token signed with `APP_AUTH_SECRET_KEY`.
Public API PATs, service-account keys, and developer OAuth access tokens are
verified by the versioned API's machine-authentication boundary, which also
loads scopes, tenant restrictions, expiry, revocation state, and immutable
principal attribution. Supplying a bearer header to a first-party or optional-
authentication route fails with `401`; it is never silently treated as an
anonymous request.

## Browser origin policy

`APP_API_CORS_ALLOWED_ORIGINS` is the exact comma-separated allowlist for
credentialed browser requests and SSE. Wildcards, opaque `null` origins,
userinfo, paths, queries, fragments, and non-HTTP schemes are rejected at
startup. Production additionally requires HTTPS for every configured origin.

The API does not grant access to every `fortyone.app` subdomain. Add a preview
or replacement application origin explicitly, deploy the API policy before the
browser begins using it, and remove the old origin after the transition. CORS
does not replace authentication, authorization, SameSite cookies, or request
validation; it is an additional browser boundary.

Unsafe requests that carry `fortyone_session` also pass the global browser
origin middleware. A configured `Origin` is required; when older same-origin
browser behavior omits it, only `Sec-Fetch-Site: same-origin` is accepted.
`same-site` is deliberately denied so a sibling subdomain cannot submit a
cookie-authenticated form. Requests without the browser cookie (provider
webhooks and versioned API credentials) continue through their dedicated
authentication and signature checks.
