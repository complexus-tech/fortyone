# Messaging Integration Runtime Contract

## Ownership

`apps/server` is the only messaging integration runtime. It owns provider
webhooks, OAuth installations, encrypted credentials, identity links, product
permissions, inbox/outbox delivery, story operations, audit logs, and calls to
provider APIs. Slack is the first adapter; Microsoft Teams, WhatsApp, and future
providers must plug into the same Go domain contract.

The integration is provider-neutral at the domain boundary, not at the user
experience boundary. A provider adapter may expose richer native capabilities
such as Slack shortcuts and modals while still producing the same normalized
commands, conversations, identities, and delivery outcomes.

## Stable Provider Identity

- `provider`: stable adapter key such as `slack`, `teams`, or `whatsapp`
- `externalWorkspaceId`: Slack team/enterprise installation ID, Teams tenant
  ID, WhatsApp business account ID, or provider equivalent
- `externalUserId`: user ID scoped to that provider installation
- `externalConversationId`: channel/chat/conversation ID
- `externalThreadId`: provider thread/root message ID when available
- `externalMessageId`: immutable or best available provider message identity
- `workspaceId`: FortyOne workspace ID resolved from the installation
- `userId`: FortyOne user ID resolved from a verified link

For Slack, `team_id` is the installed Slack workspace identifier. It is never a
FortyOne team ID.

## Adapter Boundary

Every provider adapter implements equivalent operations while retaining its
native payload types internally:

- verify the untouched inbound request and normalize its delivery identity;
- produce an immediate provider acknowledgement;
- normalize mentions, direct messages, commands, actions, option requests,
  modal submissions, lifecycle events, and delivery receipts;
- post/update messages and replies;
- open/update provider-native forms where supported;
- classify errors as retryable, rate-limited, unauthorized/revoked, invalid,
  or terminal;
- surface the provider request ID and retry-after duration without exposing
  credentials.

The domain layer must not import Slack Block Kit models, Teams Adaptive Card
models, WhatsApp webhook shapes, or provider SDK types. Adapters translate
between those types and normalized domain commands. Provider-specific rendering
belongs behind capabilities such as `SupportsModalUpdate`, `SupportsThreads`,
`SupportsEphemeralReply`, and `SupportsStreaming`; do not force all providers
into Slack's feature set.

## Inbound Delivery Contract

1. Capture the exact body and verification headers.
2. Resolve the adapter by route/provider.
3. Verify authenticity and replay window before parsing normalized content.
4. Upsert a durable inbox record using provider, external workspace/account,
   and delivery/event identity.
5. Return the provider acknowledgement within its deadline.
6. Process the accepted inbox record asynchronously and idempotently.
7. Persist the terminal outcome, retry state, and correlation IDs.

Duplicate deliveries return success after the inbox record is found. Ordering
must not be assumed across lifecycle and message events. When replay requires a
provider payload, store it encrypted in the inbox; queue only the scoped inbox
identity. Provider payloads and normalized conversation/outbox content expire
after 30 days, while request logs retain no raw payload.

The concrete Go and SQLC implementation is documented in the API
[`inbound webhook gateway`](../apps/server/docs/integrations/webhook-gateway.md)
runbook. Queue envelopes contain only the stable provider key and inbox UUID;
workers load and lease the canonical receipt from PostgreSQL before decrypting
provider content.

## Outbound Delivery Contract

Domain events and assistant results create durable outbox entries. A worker
selects the installation, resolves the current credential, renders through the
provider adapter, and records the provider message identity. Retry policy uses
bounded exponential backoff with jitter, respects provider `Retry-After`, and
never retries invalid authorization or revoked installations forever.

An idempotency key represents the product intent, for example
`story-created:{storyId}:{routeId}`. Provider message mappings allow later
updates or thread replies without relying on message text.

## Installation and Credential Contract

- One active installation maps one provider workspace/account to one FortyOne
  workspace unless that provider explicitly supports an organization install
  model that is represented separately.
- OAuth state is opaque, high entropy, short-lived, single-use, stored only as a
  hash, and bound to the initiating FortyOne user and workspace.
- Credentials are encrypted at rest with key versioning; plaintext exists only
  at the provider call boundary and is never logged.
- Persist granted scopes/capabilities from the provider response. Feature gates
  use granted capabilities, not the requested manifest alone.
- Disconnect, uninstall, and token-revocation events disable delivery and make
  queued work terminal or explicitly recoverable through reauthorization.
- Credential rotation is enabled only after refresh-token persistence,
  concurrency control, and atomic replacement are implemented and tested.

## User Linking and Authorization

Provider identity does not grant product access by itself. Resolution is scoped
by provider, provider installation, external user ID, and FortyOne workspace.

1. Use an existing verified link when present.
2. Link the OAuth installer only from the provider's authenticated installer
   identity and the initiating active FortyOne workspace member.
3. Otherwise issue a short-lived opaque connect URL bound to the provider user,
   provider installation, and FortyOne workspace.
4. Complete the link only in an authenticated FortyOne session and reject
   cross-workspace or replayed links.

Every product read or write executes as the linked FortyOne user. Team choices
come only from that user's memberships in the installed workspace. Status,
assignee, label, objective, and story selections are revalidated against the
selected team at submission, even if the adapter rendered them earlier.

## Assistant Contract

Mentions and direct messages become normalized assistant requests containing
the authenticated FortyOne actor, workspace, provider conversation/thread
context, and a bounded message context. The existing Maya/tool authorization
boundary remains authoritative. Provider prompts cannot select a different
workspace or bypass tool-level permission checks.

The Go runtime may call an internal AI service, but that service is not a
provider webhook runtime and receives no Slack signing secret or installation
token. Assistant output returns through the durable outbox and provider adapter.

## Adding Microsoft Teams or WhatsApp

A new adapter must provide:

- verified ingress and delivery deduplication;
- installation/account mapping and encrypted credential lifecycle;
- identity link/fallback flow;
- capability declaration and native message/card rendering;
- normalized assistant and work-item commands;
- outbound error classification, rate-limit handling, and message mappings;
- contract tests plus provider-specific live-installation tests.

It should not add a second database authority, duplicate FortyOne permission
rules, or expose provider SDK types to domain services.
