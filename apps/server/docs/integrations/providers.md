# First-party provider registry

FortyOne is privately operated. Provider adapters compile into the API/worker;
third-party extensions run out of process and will use the versioned API,
scoped machine credentials, and signed webhooks. Arbitrary Go plugins and a
self-hosted plugin ABI are not supported.

`internal/bootstrap/providers` is the explicit catalog of compiled providers.
Each descriptor has a stable provider key, family, versioned capabilities,
authentication strategies, configuration names (never values), disconnect
policy, and this runbook. Registry construction fails on duplicates, invalid
keys, zero/duplicate capability versions, or ambiguous configuration metadata.

The registry is discovery metadata—not a universal provider interface. A
consumer owns a small port such as repository catalog, comment writer, message
delivery, or calendar watch. Bootstrap selects a provider only after verifying
the exact capability major version and then injects the typed adapter factory.
Provider SDK types stay inside that adapter.

Signed inbound providers use the shared
[`webhook gateway`](webhook-gateway.md) for canonical encrypted receipts,
deduplication, queue-by-ID, processing leases, recovery, and payload expiry.
Provider adapters still own their signature algorithms, replay guarantees,
payload schemas, installation lookup, and provider-native acknowledgements.

## GitHub

- Family: code host.
- Auth: GitHub App installation plus user OAuth linking where attribution needs it.
- Capabilities: repository catalog, work-item/comment writes, external identity,
  and webhook verification. GitHub user OAuth tokens are not advertised as a
  refresh capability because the current flow does not persist or rotate a
  refresh token.
- Webhook ingress reads exact signed bytes under a deliberately conservative 1
  MiB application limit and rejects overflow before verification. Accepted
  deliveries use the shared encrypted inbox, deduplication, queue-by-ID, and a
  current installation-generation worker fence.
- Credentials use the envelope vault; installation scope is preferred over
  long-lived user access. User credentials carry a generation in authenticated
  context, and rewrap compare-and-swaps that generation plus the exact original
  envelope so a relink or unlink cannot be overwritten.
- User unlink revokes the OAuth token remotely before clearing the local vault
  envelope. A complete GitHub App installation disconnect/reconciliation use
  case is not yet exposed; installation deletion, suspension, and repository
  lifecycle events remain explicit SQLC-wave work. The app webhook is
  application-owned rather than a per-installation webhook, so disconnect must
  not claim to delete it.

See the detailed [code-host integration contract](code-hosts.md) and the
[exact GitHub implementation inventory](github-inventory.md).

## GitLab

- Family: code host.
- Status: narrow adapter and durable signed-webhook proof; not API/worker wired.
- Proven capabilities: installation authentication boundary, repository
  catalog, issue creation, issue comments, and issue/note normalization.
- Webhooks require GitLab Standard Webhooks HMAC signing tokens; the weaker
  legacy plaintext secret header is intentionally unsupported.
- Merge requests, OAuth lifecycle, persistence, production worker fencing, and
  disconnect reconciliation remain explicit prerequisites, not implied support.

See [GitLab proof boundaries](code-hosts.md#gitlab-proof).

## Slack

- Family: messaging.
- Auth: workspace OAuth installation and separate account linking.
- Capabilities: messages, threads, commands, external identity, and signed
  ingress. Token refresh is not advertised: the current OAuth flow does not
  persist and atomically rotate Slack refresh tokens yet.
- Bot credentials use the envelope vault. Plaintext compatibility is a bounded
  migration path, not a supported steady state.
- Installation generation is a lifecycle fence shared by refresh, event work,
  disconnect, provider revoke, uninstall delivery, and key rewrap. Adapter code
  must never update an envelope without that generation and original-value CAS.
- Slack owns the remote event subscription, so disconnect revokes OAuth and
  removes local credentials rather than claiming a separately deletable webhook.
- Verified webhook bodies use the internally derived Slack payload key, are
  bound to the durable receipt and installation generation, and are never
  queued as raw bytes. See
  [Slack webhook security](../security/slack-webhooks.md).

## Figma

- Family: design context; it is not forced into feedback or code-host semantics.
- Capabilities: file context, OAuth refresh, and webhook verification.
- OAuth credentials use the shared context-bound vault with workspace,
  connection, and immutable installation-generation AAD. Refresh, legacy
  migration, and KEK rewrap are exact compare-and-swap operations.
- Webhook ingress is bounded to 1 MiB and rejects overflow rather than silently
  verifying or decoding a truncated payload.
- The provider passcode is verified in constant time. Accepted exact bytes are
  context-bound encrypted in the shared durable inbox; Redis receives only the
  inbox UUID. The historical event table is no longer on the runtime path.
- Worker processing reacquires an inbox lease and rechecks the active connection
  generation. Reconnect or disconnect cancels stale deliveries before any
  design/story effect.
- Disconnect deactivates the connection and webhooks while retaining only safe
  story/file mappings. See the
  [Figma module guide](../../internal/modules/figma/README.md) and
  [Figma webhook security boundary](../security/figma-webhooks.md).

## Google Calendar

- Family: calendar.
- Capabilities: events, availability, watch subscriptions, refresh, and webhook verification.
- OAuth state is single-use. Disconnect stops watches, revokes the grant, and
  deletes recoverable credentials.

## Microsoft Calendar

- Family: calendar, with the same FortyOne capability contracts as Google where
  behavior truly overlaps.
- Native Microsoft subscription validation, resource identifiers, token claims,
  and provider payloads remain adapter-specific.
- Notification JSON is bounded to 1 MiB before decoding; validation-token echo
  remains separately bounded to 2 KiB.
- Disconnect stops subscriptions, revokes the grant, and deletes recoverable credentials.

## Adding a provider

1. Identify the real family and consuming capability ports.
2. Add an adapter without importing its SDK into service/domain packages.
3. Add a descriptor and explicit bootstrap registration.
4. Pass the shared contract suite for every declared capability and negative
   tests for every intentionally unsupported capability.
5. Add configuration validation, vault context, OAuth state/replay behavior,
   webhook verification/inbox behavior, rate-limit classification, health, and
   disconnect cleanup.
6. Define credential generation, refresh/rewrap compare-and-swap, revoke, and
   retained-outbox semantics, then pass the shared rotation contract.
7. Update this runbook and the configuration reference.

Adding another code host should change registration/adapters and only extend
core orchestration when the provider exposes a genuinely new capability. It
must not copy GitHub business rules or pretend provider-native concepts are
identical.
