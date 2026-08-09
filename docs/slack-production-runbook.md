# Slack Production Runbook

## Runtime Ownership

Slack terminates directly in the Go API at `apps/server`. There is no separate
Chat SDK or Node bot runtime. The Go integration owns OAuth, request
verification, interaction acknowledgement, Web API calls, installation and user
links, authorization, story creation, AI request dispatch, and diagnostics.

Slack's `team_id` identifies a Slack workspace installation. It is never a
FortyOne team ID. Every action resolves the installation first, then resolves
the Slack user to a FortyOne user in that workspace, and finally applies normal
FortyOne membership and permission checks.

## Source-Controlled Slack Configuration

The canonical Slack configuration is
[`apps/server/slack-app-manifest.yaml`](../apps/server/slack-app-manifest.yaml).
It is rendered for the production API at `https://api.fortyone.app`. Use a
separate development Slack app and replace only the host when applying the
manifest in another environment; do not change the paths.

| Slack setting             | Canonical Go endpoint                    |
| ------------------------- | ---------------------------------------- |
| OAuth redirect            | `GET /integrations/slack/setup`          |
| Events API                | `POST /integrations/slack/events`        |
| Interactivity and options | `POST /integrations/slack/interactivity` |
| `/fortyone` command       | `POST /integrations/slack/commands`      |

The message shortcut callback ID is `fortyone_create_task`. Modal submissions
use the same callback ID. External-select option requests are interaction
payloads and use the interactivity URL.

## Required Server Configuration

Set these in both the API and worker environments:

- `SLACK_CLIENT_ID`
- `SLACK_CLIENT_SECRET`
- `APP_AUTH_SECRET_KEY`

The worker needs the Slack client credentials to retry durable
`apps.uninstall` calls after an API instance has acknowledged a disconnect and
stopped serving that installation. Set these API-only values in the API secret
store:

- `SLACK_SIGNING_SECRET`

The API defaults `SLACK_REDIRECT_URL` to
`https://api.fortyone.app/integrations/slack/setup`. Set it only when running a
different environment, such as local development or staging.

The API and worker must share the same strong `APP_AUTH_SECRET_KEY`; changing it
without a credential rotation procedure makes encrypted installation,
recoverable inbox, and uninstall-outbox payloads unreadable.
The worker also requires `OPENAI_API_KEY` for Maya responses. `OPENAI_MODEL` is
optional and defaults to `gpt-5.4-mini`, including when the value passed to the
assistant is blank. Without an API key, task creation remains available and
Maya returns a clear configuration message.

The worker applies these launch safeguards by default; set the environment
variables only when an environment needs different limits:

- `OPENAI_ASSISTANT_USER_CALLS_PER_MINUTE=12`
- `OPENAI_ASSISTANT_WORKSPACE_CALLS_PER_MINUTE=120`
- `OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY=1000000`

The per-minute limits are atomic in Redis and idempotent for a replayed Slack
event. The token ceiling is shared by every messaging provider in a FortyOne
workspace and resets at midnight UTC. Responses API usage is recorded after
each model execution, including partial usage from a failed tool loop, using
the durable inbox receipt and attempt count so replaying one execution cannot
double-count it. Do not disable these checks when Redis or PostgreSQL is
unavailable; infrastructure failures must leave the event recoverable.

Never commit those values or a Slack access token. Bot tokens are obtained per
installation through OAuth, encrypted at rest, and selected by Slack
`team_id`/enterprise installation identity. Keep
`settings.token_rotation_enabled: false` until the persistence model stores and
refreshes Slack refresh tokens atomically; enabling it in Slack before that
cutover will create expiring access tokens the runtime cannot refresh.

## Permission Model

The core install requires only:

- `app_mentions:read` for `app_mention`
- `chat:write` for replies and task confirmations
- `commands` for `/fortyone` and the message shortcut
- `im:history` for `message.im`

Channel inventory and email-based account matching are intentionally excluded.
Account binding uses a short-lived, single-use link nonce instead of trusting
email identity. If channel routing is later enabled, add `channels:read` and
`groups:read` as optional scopes and handle `missing_scope` without breaking the
core command, shortcut, mention, and DM flows.

Do not add channel/message history scopes merely to create a task from a
message shortcut: Slack includes the selected message in the shortcut payload.
Untagged public/private-channel thread listening is a separate feature and
would require an explicit privacy review plus the appropriate history scopes
and message event subscriptions.

## Events and Interaction Timing

Subscribe to:

- `app_mention`
- `message.im`
- `app_uninstalled`
- `tokens_revoked`

For every Events API, command, shortcut, block action, option load, and modal
submission request:

1. Read the exact raw request body into bounded memory.
2. Reject timestamps more than five minutes old.
3. Verify `X-Slack-Signature` using HMAC-SHA256 over
   `v0:{timestamp}:{raw_body}` and compare in constant time.
4. Return a valid acknowledgement within three seconds.
5. Persist Events API payloads encrypted, queue only their tenant-scoped event
   identity, and move AI/provider delivery to durable asynchronous processing.

Do not reconstruct form or JSON payloads before signature verification. Raw
Slack bodies, signatures, response URLs, trigger IDs, and message text are not
written to request logs or Redis. The encrypted database inbox is the canonical
replay source and is deleted after 30 days. Treat Events API deliveries as
at-least-once, deduplicate by Slack workspace plus event identity, and return
`2xx` for a duplicate already accepted.

Commands, shortcuts, and team-change actions acknowledge before their bounded
modal/provider work. External option requests and modal submissions remain
synchronous because Slack requires their response bodies. Submission performs
authorization and one idempotent product write inside the deadline; source
links and confirmations have stable keys and recoverable delivery records.

When a user changes the team in the create-task modal, acknowledge the block
action first, recalculate status, assignee, label, and objective choices from
that linked user's permitted FortyOne data, then call `views.update` with both
the modal `view_id` and latest `hash`. Team search is an authorized external
select so users with more than 100 memberships can reach every team. Status
uses a static menu through Slack's limit and authorized external search beyond
it. The first permitted team and its first workflow status are selected by
default so the primary path creates a task; `Request` remains an explicit
status choice for triage. Never trust a rendered option; revalidate membership
and every dependent ID at submission.

## Disconnect and Remote Uninstall Recovery

Disconnect is locally authoritative and remotely recoverable. In one database
transaction, FortyOne snapshots the exact installation generation and its
versioned encrypted credential into `slack_uninstall_outbox`, disables local
authorization, removes identity and channel links, and cancels queued inbound
and outbound deliveries for that Slack workspace. New requests cannot use the
installation even when Slack's `apps.uninstall` endpoint is unavailable.

The API immediately attempts `apps.uninstall`. A worker recovery task runs
every minute and claims due `pending`, stale `processing`, or retryable `failed`
records. Transient provider errors and HTTP `429` responses use bounded
backoff. Terminal Slack responses such as an already revoked or invalid token
complete the record. Completion scrubs the encrypted credential snapshot. An
installation generation and active-install guard ensure an old disconnect can
never uninstall or deliver through a later reinstall, including a reinstall
for the same Slack `team_id`.

After the bounded retry budget is exhausted, the record moves to
`revocation_required`. This is the credential dead-letter state: an operator
must revoke the affected installation in Slack and resolve it through the
approved service procedure. Do not copy, decrypt, log, or manually edit the
credential payload. Use this metadata-only query for operational visibility:

```sql
SELECT id,
       workspace_id,
       slack_team_id,
       uninstall_kind,
       status,
       attempt_count,
       last_error,
       next_attempt_at,
       created_at,
       updated_at
FROM slack_uninstall_outbox
WHERE status IN ('failed', 'revocation_required')
ORDER BY updated_at ASC;
```

Alert on any `revocation_required` row and on retryable rows older than the
normal backoff window. Logs and dashboards may include the uninstall ID,
workspace ID, Slack team ID, status, attempt count, and sanitized provider
error code. They must never include `credential_payload`, access tokens, client
secrets, raw Slack payloads, or decrypted message content.

## Deployment and Data Migration

1. Apply migrations `000107_messaging_integrations` and
   `000108_messaging_assistant_daily_usage` before deploying either the API or
   worker. Migration `000108` adds the UTC daily aggregate and the per-inbox-
   attempt usage ledger used for idempotent Responses API accounting.
2. Deploy the expand-compatible API with the shared `APP_AUTH_SECRET_KEY` and
   Slack secrets, and wait for the service to become stable so every legacy API
   task has drained. During this window, OAuth dual-writes the legacy plaintext
   column and the versioned encrypted payload. Lazy reads may add an encrypted
   payload but must not clear the legacy column while old API processes may
   still need it.
3. Deploy the worker with the same `APP_AUTH_SECRET_KEY`,
   `SLACK_CLIENT_ID`, `SLACK_CLIENT_SECRET`, and OpenAI configuration. After API
   stability, worker startup seals all version-0 credentials and scrubs the
   legacy plaintext column only for rows that already have a valid versioned
   encrypted payload. A daily job retries both bounded phases if startup is
   interrupted.
4. Verify no active plaintext credential remains:

   ```sql
   SELECT COUNT(*)
   FROM slack_workspaces
   WHERE is_active = true
     AND (
       credential_key_version = 0
       OR NULLIF(bot_access_token, '') IS NOT NULL
     );
   ```

   The expected count is `0`.

5. Keep the dual-read application path through an observation window. Drop the
   legacy plaintext column only in a later contract migration after the query
   remains at zero and rollback no longer depends on an old API binary.
6. Apply the manifest, reinstall the development workspace, and complete the
   live matrix before enabling customer installs.

Migration `000107` intentionally refuses to roll back while encrypted Slack
credentials exist. An application-assisted decrypt/disconnect procedure is
required before schema rollback; never force the migration version past that
guard.

## Applying the Manifest

1. Apply `apps/server/slack-app-manifest.yaml` directly to the production Slack
   app. For development or staging, copy it to a release-specific temporary file
   and replace `api.fortyone.app` with that environment's API host. Do not commit
   an environment-specific rendered copy.
2. In Slack app settings, open **App Manifest**, paste or upload the rendered
   YAML, and resolve every schema or Request URL validation error.
3. Confirm that Event Subscriptions can complete Slack's URL verification
   challenge against the deployed API.
4. Compare the live manifest with the committed manifest during release review.
5. If scopes or optional-scope classifications changed, reinstall/re-authorize
   every test installation. Existing tokens do not automatically gain new
   scopes. Removing a scope from the manifest also does not remove it from an
   existing token; revoke and reinstall when validating least privilege.

Apply the development Slack app first. Promote the identical path and callback
configuration to production only after the development app passes the live
checks below. Keep separate Slack app credentials per environment.

## Workspace Administration and Distribution

- A FortyOne workspace admin starts installation from the workspace integration
  settings page.
- OAuth state is an opaque, short-lived nonce bound to the initiating FortyOne
  user and workspace, stored only as a hash, consumed once, and verified before
  token exchange.
- If Slack requires admin approval, the UI must show the pending/denied state
  rather than treating it as a completed installation.
- Reinstall after any scope addition, optional-scope change, signing-secret
  rotation, or Slack app credential replacement.
- Before enabling installs outside development workspaces, enable Slack app
  distribution and complete the required redirect URL, privacy, support, and
  security configuration. Treat Slack Marketplace review as a production
  launch gate for broad commercial distribution and re-check Slack's current
  distribution and rate-limit policy before launch.

## Live Verification

Use a development Slack workspace, a FortyOne workspace with at least two teams
and different workflows, and two users: one linked and one unlinked.

- Install from FortyOne and confirm exactly one active installation maps to the
  expected Slack `team_id`.
- Confirm an invalid/expired OAuth state and an invalid Slack signature are
  rejected without side effects.
- Run `/fortyone create Example title`; confirm acknowledgement is under three
  seconds and the modal opens.
- Use **Create a story** on an existing message; confirm its text and
  permalink context are prefilled without message-history scopes.
- Confirm the team selector contains only teams the linked user belongs to.
- Switch teams and confirm statuses, assignees, labels, and objectives refresh
  immediately and stale selections are cleared.
- Submit once, then replay the same delivery; confirm one story/request, one
  source mapping, and one private creation confirmation.
- Mention the app in a channel; confirm Maya sends the linked user a normal,
  durable direct message and does not expose workspace data in the channel. DM
  it and confirm continuity in the app conversation.
- In a test environment with reduced limits, exhaust the user, workspace, and
  daily token budgets independently. Confirm Maya makes no additional model
  call, persists a safe outbox reply, completes the inbound event, and keeps a
  channel mention's reply in the user's direct message. Replay the same event
  and confirm it does not consume another rate-limit slot or duplicate recorded
  usage for the same inbox attempt.
- Send a message larger than 16 KiB and confirm it is neither stored in the
  conversation nor sent to OpenAI. Confirm Maya returns the durable shortening
  prompt privately for a channel mention.
- As an unlinked channel user, confirm the one-time account URL is sent only to
  that user's app conversation and never appears in the public channel.
- As the unlinked user, invoke every entry point and confirm the single-use
  account link returns to the intended FortyOne workspace.
- Disconnect while Slack is reachable; confirm local authorization, links, and
  queued deliveries are disabled immediately and the uninstall record completes
  with its credential snapshot scrubbed.
- Repeat while Slack returns a transient error or `429`; confirm the every-minute
  worker recovery honors backoff and eventually completes without exposing a
  token in logs. Exhaust the bounded budget in a test environment and confirm a
  visible `revocation_required` dead-letter record.
- Reinstall the same Slack workspace before an old uninstall retry runs; confirm
  the old generation is completed as superseded and never targets the new token.
- Revoke tokens and uninstall the app from Slack; confirm the installation is
  disabled, secrets are no longer usable, and queued outbound work is cancelled
  rather than retried through a replacement installation.
- Check request logs, encrypted inbox/recoverable outbox state, worker logs,
  Slack `429` handling, and retry headers for correlation IDs and absence of
  tokens or plaintext sensitive message content. Confirm the daily cleanup
  removes inbox, outbox, and conversation content older than 30 days.

## Rollback and Incident Response

If the integration is sending incorrect or unauthorized messages:

1. Disable new installs in FortyOne.
2. Disable Slack Event Subscriptions and Interactivity, or block the Slack ingress
   routes at the edge if immediate containment is required.
3. Stop the Slack worker/outbox consumer while preserving queued records for
   investigation.
4. Revoke affected installation tokens and mark installations disconnected.
5. Rotate the signing secret/client secret when compromise is possible, update
   server secrets, redeploy, and reinstall affected workspaces.
6. Re-enable against a development workspace and repeat the live-verification
   matrix before restoring production traffic.

## Slack References

- [App manifest reference](https://docs.slack.dev/reference/app-manifest/)
- [Installing with OAuth](https://docs.slack.dev/authentication/installing-with-oauth/)
- [Binding accounts across services](https://docs.slack.dev/authentication/binding-accounts-across-services/)
- [Using token rotation](https://docs.slack.dev/authentication/using-token-rotation/)
- [Verifying requests from Slack](https://docs.slack.dev/authentication/verifying-requests-from-slack/)
- [Events API](https://docs.slack.dev/apis/events-api/)
- [Handling user interaction](https://docs.slack.dev/interactivity/handling-user-interaction/)
- [Modals and `views.update`](https://docs.slack.dev/surfaces/modals/)
- [Web API rate limits](https://docs.slack.dev/apis/web-api/rate-limits/)
- [App lifecycle and distribution](https://docs.slack.dev/app-management/distribution/)
