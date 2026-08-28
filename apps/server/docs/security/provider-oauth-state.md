# Provider OAuth and installation state

GitHub, Slack, and Figma authorization callbacks use opaque, single-use state.
The state value is a bearer secret for the short lifetime of the authorization
attempt: it may cross the browser and provider redirect, but it must never be
persisted in plaintext, written to a log, included in an error, or reused.

## Security contract

The shared `internal/platform/oauthstate` primitive generates exactly 32 bytes
(256 bits) with `crypto/rand` and emits canonical unpadded base64url. Callback
parsing rejects malformed, padded, non-canonical, or wrong-length values before
accessing a store. Stores receive only a SHA-256 digest or a key derived from a
SHA-256 digest.

Each record binds the authorization attempt to its provider, purpose, and
FortyOne identity. Consumption is an atomic take/update and happens before an
authorization code is exchanged. A provider denial, malformed payload,
authorization change, or downstream provider failure therefore burns the state
and requires a fresh session. This is intentional fail-closed behavior.

| Provider flow      |   Lifetime | Durable binding                                                                              | One-time store                                                |
| ------------------ | ---------: | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| GitHub App install | 10 minutes | provider, app-install purpose, workspace, initiating user, workspace slug                    | Redis digest-derived key and atomic `Take`                    |
| GitHub user link   | 15 minutes | provider, user-link purpose, authenticated user, validated return path                       | Redis digest-derived key and atomic `Take`                    |
| Slack App install  | 10 minutes | provider, oauth-install purpose, workspace, initiating user, workspace slug                  | `messaging_nonces` digest row and atomic conditional `UPDATE` |
| Slack account link | 15 minutes | provider, account-link purpose, workspace, FortyOne user, Slack team and optional Slack user | `messaging_nonces` digest row and atomic conditional `UPDATE` |
| Figma install      | 10 minutes | dedicated Figma purpose/table, workspace, initiating user, workspace slug, PKCE verifier     | `figma_oauth_states` hash row and atomic conditional `UPDATE` |

GitHub App callbacks re-read the initiating user's current workspace role after
consuming state. Revoking admin access while a browser is at GitHub therefore
prevents installation. Slack validates the returned Slack team against the
bound team for account-link flows. Figma returns the stored workspace/user
binding only after the state row is successfully consumed and uses its stored
PKCE verifier for the code exchange.

The stores intentionally retain non-secret callback context such as UUIDs,
workspace slug, validated relative/allow-listed return path, and Figma's PKCE
verifier. They never retain the raw state bearer. Figma's verifier is a separate
one-time secret and remains in the dedicated short-lived row because the OAuth
exchange requires it.

## Callback and logging rules

- Consume state before exchanging an authorization code or processing a
  provider-reported denial.
- Match the exact provider and purpose. Never try a state under a generic key.
- Match the authenticated FortyOne user/workspace when the callback endpoint
  has that trusted context. Public provider callbacks must use the identity
  stored at issuance and reauthorize privileged mutations against current data.
- Treat missing, malformed, expired, replayed, mismatched, corrupt, and
  store-unavailable state as invalid. There is no fail-open fallback.
- Log only safe provider names, workspace/user IDs already authorized for the
  operation, and outcome categories. Never log callback query strings, state,
  authorization codes, token digests, Redis keys, PKCE verifiers, or provider
  tokens.

## Compatibility and rollout

No schema migration is required. Deploy in this order:

1. Confirm the shared Redis service is healthy for every API replica and the
   existing PostgreSQL migrations for `messaging_nonces` and
   `figma_oauth_states` are present.
2. Deploy the API replacement as one coordinated rollout. Do not run a long
   mixed-version window: new GitHub sessions are opaque and old binaries cannot
   consume them.
3. Expect GitHub authorization pages opened before the rollout to fail once.
   The former GitHub install state had no expiry and was replayable, so accepting
   it after deployment is not a defensible transition. The former user-link
   state was expiring but not one-time; it is also rejected rather than adding a
   replay bypass.
4. Slack sessions remain compatible because token encoding, digest derivation,
   purposes, TTLs, and database rows are unchanged.
5. Figma sessions remain compatible because its canonical state encoding and
   persisted hash derivation are unchanged; callback parsing is stricter.
6. Exercise one fresh install/link flow per provider, then verify a second use
   of each callback state fails before any provider exchange.

If Redis is unavailable, new GitHub sessions and GitHub callbacks fail closed.
Do not temporarily restore signed stateless state. Recover Redis availability
and ask the user to start a new authorization flow.

### GitHub callback contract

GitHub state is always one canonical 43-character opaque base64url value. It
never contains `returnTo`, workspace data, or a signed JSON payload. The App
installation callback accepts exactly one positive `installation_id` and one
`state`; the authenticated user-link completion accepts `code` and `state` in
its JSON body. Duplicate query values are rejected.

The current frontend callback still contains a legacy signed-payload decoding
assumption. With opaque state it must use its safe fallback route until the
frontend is updated. This API-only migration intentionally does not weaken
state or embed navigation context to preserve that behavior.

## Verification

```bash
go test -race ./internal/platform/oauthstate ./pkg/cache \
  ./internal/modules/github/service ./internal/modules/slack/service \
  ./internal/modules/figma/service

go vet ./internal/platform/oauthstate ./pkg/cache \
  ./internal/modules/github/service ./internal/modules/slack/service \
  ./internal/modules/figma/service

TEST_DATABASE_URL='postgresql://<disposable-createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -tags=integration -count=1 \
  ./internal/modules/messaging/repository ./internal/modules/figma/repository
```

The tagged tests apply the real migration chain to isolated databases and prove
that concurrent PostgreSQL consumers produce exactly one winner. The Redis
atomic-take test uses an actual Redis-compatible Lua execution path. This slice
adds no SQL or SQLx call site; the existing parameterized Slack/Figma state
queries remain transitional repository debt for their owning SQLC migration
waves.
