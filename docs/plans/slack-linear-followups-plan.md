# Slack Linear-Style Parity Status

## Document Status

- Status: Core implementation complete; live Slack verification required
- Canonical operations guide: `docs/slack-production-runbook.md`
- Canonical Slack configuration: `apps/server/slack-app-manifest.yaml`
- Primary surfaces:
  - `apps/server/internal/modules/slack`
  - `apps/server/internal/modules/messaging`
  - `apps/server/internal/modules/integrationrequests`
  - `apps/server/internal/modules/stories`
  - `apps/projects/src/modules/settings/workspace/integrations/slack`

## Implemented Product Contract

### Story creation

- **Create a story** is available from a Slack message and `/fortyone`.
- The modal starts with the linked user's authorized teams in their personal
  FortyOne order.
- Team changes refresh status, assignee, labels, and objective choices without
  retaining invalid state.
- Selecting a workflow status creates a story directly. Selecting `Request`
  creates a triage item that can be completed and converted later.
- Creation is idempotent and the source-conversation confirmation uses the
  compact copy `Joseph created WEB-123`, followed by an authorized rich story
  card. A shortcut on a channel root posts the receipt at the channel root;
  actions started inside an existing thread stay in that thread.

### Requests and thread synchronization

- Slack-backed Requests retain their canonical Slack workspace, channel, root
  thread, installation generation, and source link.
- Slack replies are stored once and appear in the Request activity.
- FortyOne comments are delivered through the durable Slack outbox.
- The Request detail uses the same editable intake pattern as Feedback,
  including title, description, workflow properties, labels, key result, and
  conversion actions.
- After conversion, the story keeps an expandable Slack thread history and
  composer backed by the same durable thread.

### Assistant conversations and actions

- Direct messages are private, actor-scoped conversations.
- An explicit channel mention starts one shared channel thread conversation;
  linked teammates may continue it without mentioning the bot again.
- Every turn rechecks the actor's active account link and team membership.
- Admin channel-audience settings restrict which FortyOne teams Maya can use.
  An unmapped channel is limited to public teams joined by the current actor.
- Slack Connect channel assistant conversations are rejected.
- Workspace settings control assistant enablement, confirmed workflow actions,
  and bounded Slack-specific guidance.
- Natural-language create/update requests produce Confirm and Cancel controls.
  Confirm rechecks settings, audience, membership, expiry, and story version;
  writes are idempotent and Cancel never mutates data.

### Links and rich Slack surfaces

- Authorized FortyOne story URLs support Slack `link_shared` unfurls.
- Story Work Object details expose a permission-gated compact view of the
  identifier, title, description, status, assignee, priority, and due date.
  Authorized users can edit those supported properties in Slack's detail pane;
  every submission rechecks installation, actor, team, channel audience, and
  story version before a compare-and-swap update.
- Rich Block Kit payloads are persisted with outbound deliveries so retry and
  recovery render the same result as the first attempt.
- Link previews fail closed for unlinked or unauthorized users and Slack
  Connect channels.

### Operations and safety

- Go is the only Slack runtime; the Chat SDK app and bridge are removed.
- Events are verified, encrypted, deduplicated, and queued by receipt identity.
- Inbox/outbox recovery is installation-generation fenced and rate-limit aware.
- Slack credentials are encrypted at rest and remote uninstall is recoverable.
- Maya has prompt/history bounds, per-user and per-workspace rate limits, and a
  daily workspace token budget with idempotent usage accounting.
- Admin-only settings manage channel audience, agent behavior, installation,
  and lifecycle actions.

## Deliberately Deferred Domains

These are not represented as partially working Slack features:

- **Projects in the creation modal:** FortyOne does not currently expose one
  first-class project entity in the story creation domain. Do not map this to
  objectives or epics merely to imitate Linear's label. Add it only after the
  product model and story persistence contract exist.
- **Project-created Slack channels:** depends on the same project domain and an
  explicit external-channel lifecycle policy.
- **Slack file ingestion:** requires a dedicated attachment ingestion pipeline,
  malware/content controls, provider download retries, retention policy, and
  the additional Slack file scope. The production manifest intentionally does
  not request file access until that path exists.

The deferred items do not block story creation, Request triage and conversion,
thread comments, Maya conversations and confirmed story mutations, channel
audience enforcement, or story link previews.

## Release Gate

Apply migrations through `000114`, deploy API before worker, apply the canonical
manifest, reinstall the development Slack workspace to grant the new scopes and
events, and complete every item in the runbook's live-verification matrix before
enabling customer installations.
