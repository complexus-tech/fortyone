# Slack database contract

Slack persistence is owned by `internal/modules/slack/repository`. Handwritten
repository code calls the module's generated SQLC interface through native pgx;
SQLC types are mapped immediately into `internal/modules/slack/domain` records.
No Slack handler or service receives a database handle.

## Owned tables

| Table                            | Purpose                                     | Important lifecycle facts                                                                |
| -------------------------------- | ------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `slack_workspaces`               | Active installation and vault envelope      | one active Slack team per FortyOne workspace; immutable generation fences provider work  |
| `slack_channels`                 | Last synchronized Slack channel directory   | rows are deactivated then upserted atomically for one current installation generation    |
| `slack_user_links`               | Slack user to FortyOne actor mapping        | tenant/team/install scoped; target user must remain an active workspace member           |
| `slack_channel_team_access`      | Explicit public-team audience for a channel | private or cross-workspace teams are never inserted                                      |
| `slack_agent_settings`           | Workspace-specific assistant guidance       | administrator-only; database constraint limits guidance to 4,000 characters              |
| `slack_request_logs`             | Sanitized provider request outcome metadata | no signature, raw bearer response URL, raw interaction body, or token is retained        |
| `slack_uninstall_outbox`         | Durable remote-revocation intent            | encrypted credential is cleared only on completion; exhausted work requires intervention |
| `slack_user_onboarding_receipts` | Once-only welcome marker                    | stores only SHA-256 of Slack team/user identity and deliberately survives reconnects     |

Slack also participates in provider-neutral tables owned by messaging/webhooks:
`messaging_nonces`, `messaging_inbound_events`,
`messaging_outbound_deliveries`, conversations, and messages. Slack repository
queries may update inbox/outbound rows only as part of installation lifecycle
transactions. Generic inbox claims and delivery state remain owned by their
platform repository.

## Installation generation

`slack_workspaces.installation_generation` is a random UUID created for every
successful OAuth authorization, including a refresh/reinstall of the same Slack
team. It is copied into encrypted credential context, inbox receipts, outbound
deliveries, provider-thread bindings, and uninstall work.

A query that can cause a provider or product effect must match the current
generation. Matching workspace or Slack team ID alone is insufficient because a
queued task from a disconnected installation would otherwise become valid again
after reinstall.

## Repository transactions

The repository owns transactions; callers never pass `pgx.Tx` or SQLC query
objects across the module boundary. Installation lifecycle transactions take a
PostgreSQL advisory transaction lock before inspecting conflicting active
installations or uninstall work. The lock serializes install, reinstall,
disconnect, and provider revoke across distinct uniqueness predicates.

The following groups are indivisible:

- install/reinstall conflict check, safe uninstall supersession, stale message
  cancellation, installation upsert, installer link, and thread-generation
  rebind;
- disconnect outbox creation, message cancellation, account-link deletion, and
  installation deletion;
- provider revoke message cancellation, account-link deletion, and installation
  deletion;
- channel snapshot deactivation and current-channel upserts;
- assistant configuration marker, old public mappings removal, and validated
  replacement mappings; and
- recoverable uninstall dead-lettering and lease claims.

Every transaction defers rollback and commits once. A SQLC method returning no
row is classified before leaving the adapter. Serialization, deadlock, and
uniqueness conflicts map to the Slack conflict contract where the caller can
retry safely.

## Human and trusted query families

Human API queries include `actor_id` and join all of:

```sql
users actor
JOIN workspace_members membership
  ON membership.user_id = actor.user_id
JOIN workspaces workspace
  ON workspace.workspace_id = membership.workspace_id
```

The predicate requires `actor.is_active = TRUE`, a non-deleted workspace, and
the appropriate current role. Team-scoped queries additionally join
`team_members`; public-channel disclosure also requires `teams.is_private =
FALSE`. This is deliberate duplication at the persistence boundary, not a
replacement for service policy.

Trusted provider/worker queries do not invent a human actor. They require a
verified external Slack identity plus the exact active installation ID and
generation. Maintenance queries require exact record identity, envelope
version, original ciphertext, and stable UUID pagination.

## Stable ordering and bounds

Every list has a deterministic final key:

| List                           | Order                                        |
| ------------------------------ | -------------------------------------------- |
| Channels                       | normalized name, channel row ID              |
| Workspace teams                | personal order, creation time, team ID       |
| Statuses                       | configured order, normalized name, status ID |
| Members                        | display value, user ID                       |
| Labels/objectives              | normalized name, resource ID                 |
| Request logs                   | creation time descending, log ID descending  |
| Uninstall recovery             | next-attempt/creation time, uninstall ID     |
| Credential and webhook cutover | record UUID ascending                        |

Limits are normalized before conversion to SQLC's `int32` parameters using the
shared pagination and safe-cast helpers. Search text is bounded and passed as a
value; it is never interpolated into SQL.

## Credential and webhook cutovers

Normal reads accept only current `vault.v2` Slack credential envelopes and only
`slack-webhook.v2` inbox payloads. Legacy rows are handled by explicit startup
operations:

- credential rows are strictly decoded, resealed with the shared vault, and
  replaced by identity/version/original-value compare-and-swap;
- retryable Slack inbox rows are scanned in UUID order, authenticated against
  Slack team/event plus durable receipt identity, resealed with the dedicated
  payload key, and replaced by identity/original-ciphertext compare-and-swap.

The cutovers abort on malformed data and never grant compatibility to normal
runtime operations. See [provider credential vault](../security/provider-credential-vault.md)
and [Slack webhook security](../security/slack-webhooks.md).

## PostgreSQL verification

Integration tests run against migrated PostgreSQL 18 and must cover:

- inactive, removed, downgraded, wrong-tenant, and stale-generation actors;
- concurrent install/disconnect/reinstall and channel-sync races;
- transaction rollback after each meaningful statement boundary;
- uninstall lease exclusivity, exhaustion, and retained-secret clearing;
- credential/webhook cutover compare-and-swap races;
- stable pagination when rows share the same timestamp or display name; and
- `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` evidence for installation lookup,
  request-log ordering, recovery claims, and stable maintenance scans.

Tests use `internal/testkit.NewPostgres`; they must not silently skip because a
module-specific environment variable is absent.
