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

`APP_API_CORS_ALLOWED_ORIGINS` is the comma-separated allowlist for credentialed
browser requests and SSE. It accepts exact origins and controlled leading
subdomain patterns such as `https://*.fortyone.app`. A subdomain pattern matches
only the configured scheme and child hostnames; it does not match the apex,
explicit ports, or lookalike domains. Global wildcards, opaque `null` origins,
userinfo, paths, queries, fragments, and non-HTTP schemes are rejected at
startup. Production additionally requires HTTPS for every configured entry.

The FortyOne application uses the `https://*.fortyone.app` pattern because each
workspace has its own subdomain. CORS does not replace authentication,
authorization, SameSite cookies, or request validation; it is an additional
browser boundary.

Unsafe requests that carry `fortyone_session` also pass the global browser
origin middleware. A configured `Origin` is required; when older same-origin
browser behavior omits it, only `Sec-Fetch-Site: same-origin` is accepted.
`same-site` metadata without an `Origin` is deliberately denied; a permitted
workspace request must carry an origin that matches the configured policy.
Requests without the browser cookie (provider webhooks and versioned API
credentials) continue through their dedicated authentication and signature
checks.

## Native mobile sessions

The mobile app reuses opaque first-party sessions without reviving legacy JWT
bearers. `/auth/mobile` on the Projects authentication host carries the app's
random state and S256 PKCE challenge through provider/email sign-in and workspace
onboarding. After an explicit browser confirmation, the cookie-authenticated
`POST /auth/mobile/authorize` issues a random 256-bit code. Codes are indexed by
SHA-256 digest in Redis, expire after two minutes, and bind state, PKCE challenge,
the fixed `fortyone://login` redirect, and the browser session's account epoch.

`POST /auth/mobile/exchange` requires the code, matching state, verifier and
redirect URI. It validates the binding before atomically taking the code,
revalidates account activation/epoch, and issues an independent 30-day cookie.
The epoch captured by browser authorization is retained, so revocation during
the exchange cannot mint a session at a newer epoch. Neither the browser cookie
nor a session credential is included in a redirect or JSON response. The old
unbound `/users/session/code` handoff is removed; unlaunched legacy mobile builds
must update. Ordinary browser authentication is unchanged.

On Expo SDK 57, the app deliberately uses `expo/fetch` with credentials omitted
and redirects rejected. Its transport sends only `fortyone_session`, persisted
in SecureStore with its expiry and API origin, to the configured API origin and
path. This avoids relying on browser/native cookie-store sharing or residual
native cookies after logout. Cookie changes are read from the direct exchange
response. The configured `EXPO_PUBLIC_APP_URL` (default
`https://cloud.fortyone.app`) supplies the Origin required by the existing
server policy. That origin must be present in the API's allowlist. It is request
metadata, not an app identity assertion; missing or untrusted browser origins
remain forbidden. HTTPS is required outside local development.

Local logout removes the credential, account caches and drafts before attempting
remote revocation. If revocation is unavailable, the UI states that only local
sign-out is confirmed. Timeouts, offline operation and API failures do not
otherwise erase a valid saved session. Expiry and authoritative 401 responses
invalidate it. Sessions have a fixed 30-day lifetime; no silent renewal or
refresh-token family is introduced.

Protocol/unit tests do not establish device behavior. Before release, exercise
cookie capture and requests, app termination/relaunch, browser cancellation,
PKCE callback delivery, sign-out online/offline, account switching, expiry and
revocation in actual iOS and Android development/release builds.
