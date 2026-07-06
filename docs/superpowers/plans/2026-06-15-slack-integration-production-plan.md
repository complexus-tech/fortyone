# Slack Integration Production Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the Slack integration so installed Slack workspaces can talk to Maya, create FortyOne stories from Slack, resolve Slack users to FortyOne users, and leave enough logs to debug every failure.

**Architecture:** Keep `apps/bot` as the public chat-integration runtime powered by Chat SDK and AI SDK. Keep `apps/server` as the only database-facing product authority for workspace installation records, Slack/Teams user links, permissions, story creation, runtime options, and diagnostic logs. The bot may be deployed next to the API for low-latency internal calls, but it must not connect directly to the production database; it calls authenticated internal Go endpoints with `FORTYONE_BOT_TOKEN`, and the Go API resolves provider workspace/user identity from Slack or Teams payload IDs.

**Tech Stack:** Next.js 16 bot app, Chat SDK `chat@4.30.0`, `@chat-adapter/slack@4.30.0`, AI SDK `ai@6`, Go API, PostgreSQL Slack integration tables, Redis Chat SDK state in production.

---

## Current State

- `apps/bot/src/lib/bot.tsx` already handles `onNewMention`, `onDirectMessage`, `onSubscribedMessage`, `/fortyone`, message actions, option loads, and create-story modal submit.
- AI replies already use AI SDK in `apps/bot/src/lib/agent.ts`, but the request still includes `createTools(...)`. For this production slice, tools should be disabled or feature-gated because the user asked for AI text first, not tool execution.
- The Chat SDK create-story modal in `apps/bot/src/lib/create-story-modal.tsx` has team, status, assignee, objective, priority, title, and description, but no labels.
- The backend has label search and older raw Slack Block Kit support in `apps/server/internal/modules/slack/service/slack.go`, but the Chat SDK modal does not have a multi-select modal primitive. Labels therefore require either a small Chat SDK/Slack adapter extension or a single-label compromise.
- The backend already has `slack_workspaces`, `slack_user_links`, and `slack_request_logs` tables from migrations `000060`, `000061`, and `000062`.
- The backend already auto-links workspace members by email after OAuth install through `autoLinkWorkspaceMembers(...)` and can generate manual account-link URLs through `buildSlackUserLinkURL(...)`.
- There are two Slack webhook surfaces today: `apps/bot` Chat SDK `/api/webhooks/slack`, and older Go routes `/integrations/slack/events`, `/interactivity`, and `/commands`. Production should pick one public Slack surface to avoid misconfiguration. Use the Chat SDK route.

## Integration Ownership Model

Use a two-service split:

- `apps/bot`: Chat SDK adapter runtime. It owns provider webhooks, signature verification through adapters, Chat SDK routing, platform-specific message/card/modal rendering, AI SDK calls, streaming, and reply delivery.
- `apps/server`: FortyOne domain runtime. It owns Postgres access, install sessions, OAuth state, provider installation records, encrypted tokens, user linking, plan/permission checks, option data, story creation, labels, audit logs, and diagnostics.

The bot can be deployed on the same host/network as the API, but it should not receive `DATABASE_URL` or any direct Postgres credentials. Direct DB access would make the bot a second backend, duplicate business rules, increase secret blast radius, and make schema changes harder because both Go and TypeScript would depend on the same tables.

All product actions from the bot should cross an internal API boundary:

```text
Slack/Teams event
  -> apps/bot Chat SDK adapter
  -> apps/server internal integration API
  -> Postgres/domain services
  -> apps/bot reply through Chat SDK adapter
```

This same boundary should support future providers:

```text
provider = slack | teams | discord | ...
external_workspace_id = Slack team id or Teams tenant id
external_user_id = Slack user id or Teams user id
workspace_id = FortyOne workspace id
user_id = FortyOne user id
```

Account linking belongs in Go:

1. Go creates install/link sessions with signed state.
2. Go stores installation records and provider tokens.
3. Go auto-links provider users to FortyOne users by verified email when available.
4. If email matching fails, Go returns a signed connect URL to the bot.
5. The user authenticates in FortyOne, and Go writes the provider user link.

## Production Decisions

- Public Slack Request URLs: use `https://<bot-domain>/api/webhooks/slack` for Event Request URL, Interactivity Request URL, Select Menus Options Load URL, and Slash command Request URL.
- Installation source of truth: use the existing Go OAuth/settings flow and database records. Do not move installation storage into Chat SDK Redis unless you intentionally replace the settings flow.
- Chat SDK token resolution: configure `@chat-adapter/slack` with an `installationProvider` that reads the installed Slack bot token from the Go API by Slack team ID. This lets Chat SDK handle multi-workspace webhooks while the backend remains the source of truth.
- Bot database access: do not give `apps/bot` direct access to Postgres in production. Add focused internal Go endpoints instead.
- Future providers: keep provider-specific webhook mechanics in `apps/bot`, but keep shared concepts such as installation, user link, permission check, option lookup, and story creation in `apps/server`.
- Story creation: deterministic form submit only. AI should not create stories in this slice.
- User authentication: every Slack interaction should resolve Slack team ID and Slack user ID against the backend before doing user-scoped work. If unlinked, respond with the existing account-link URL.
- Labels: implement proper multi-label support by extending the Chat SDK Slack modal conversion or, if time-constrained, ship without labels behind a visible plan item. Do not fake labels with a free-text field.

## Files

- Modify: `apps/bot/src/lib/bot.tsx` - handler orchestration, identity gate, option-load logging, create-story flow.
- Modify: `apps/bot/src/lib/agent.ts` - no-tool Maya text responses and optional thread-history prompt handling.
- Modify: `apps/bot/src/lib/config.ts` - production env validation for Redis, Slack signing, token provider mode, and OpenAI key.
- Modify: `apps/bot/src/lib/fortyone-client.ts` - internal API methods for installation lookup, identity resolution, label search, diagnostic logs.
- Modify: `apps/bot/src/lib/create-story-modal.tsx` - form fields and create-story payload shape.
- Create: `apps/bot/src/lib/logger.ts` - small structured logger helpers for bot-side events without leaking secrets.
- Modify: `apps/bot/src/app/api/webhooks/slack/route.ts` - wrap webhook call with request-level logging if Chat SDK does not surface enough context.
- Modify: `apps/server/internal/modules/slack/http/routes.go` - add internal bot endpoints.
- Modify: `apps/server/internal/modules/slack/http/slack.go` - handlers for installation lookup, identity resolution, labels, and runtime diagnostics.
- Modify: `apps/server/internal/modules/slack/http/models.go` - request/response DTOs.
- Modify: `apps/server/internal/modules/slack/service/models.go` - core models for installation, identity, labels, and create-story label IDs.
- Modify: `apps/server/internal/modules/slack/service/slack.go` - service methods and validation.
- Modify: `apps/server/internal/modules/slack/repository/repository.go` - only if existing repository methods do not expose needed data.
- Modify: `apps/server/internal/modules/slack/service/slack_test.go` - service tests for identity, labels, and story creation.
- Modify: `apps/bot/README.md` and `apps/bot/.env.example` - production setup and Slack URL checklist.
- Modify: `apps/projects/src/modules/settings/workspace/integrations/slack/slack-integration-settings.tsx` - surface request logs and account-link status if needed.
- Create: `docs/slack-production-runbook.md` - deployment, ownership, Slack app settings, and manual verification checklist.

---

### Task 0: Lock the Integration Boundary

**Files:**

- Modify: `apps/bot/README.md`
- Create: `docs/slack-production-runbook.md`
- Modify: `apps/bot/.env.example`

- [ ] **Step 1: Document the ownership split**

Add this to the bot README:

```md
## Runtime boundary

`apps/bot` owns provider webhooks, Chat SDK routing, AI SDK calls, and platform replies.
`apps/server` owns Postgres, installation records, provider user links, permissions,
story creation, option data, labels, audit logs, and diagnostics.

Do not give `apps/bot` direct database credentials in production. Add internal
Go API endpoints when the bot needs FortyOne data or writes.
```

- [ ] **Step 2: Remove DB env from the bot contract**

Ensure `apps/bot/.env.example` does not contain:

```bash
DATABASE_URL=
POSTGRES_URL=
```

If either appears later, remove it and add a comment pointing to `FORTYONE_API_URL`.

- [ ] **Step 3: Add the internal API contract to the runbook**

Create `docs/slack-production-runbook.md` with:

```md
# Slack Production Runbook

## Service Boundary

- Bot runtime: `apps/bot`
- Product/domain runtime: `apps/server`
- Database access: `apps/server` only
- Bot to API authentication: `Authorization: Bearer $FORTYONE_BOT_TOKEN`

## Required Internal Endpoints

- `GET /internal/bot/slack/installations/{teamId}`
- `POST /internal/bot/slack/identity/resolve`
- `POST /internal/bot/slack/options/teams`
- `POST /internal/bot/slack/options/statuses`
- `POST /internal/bot/slack/options/members`
- `POST /internal/bot/slack/options/objectives`
- `POST /internal/bot/slack/options/labels`
- `POST /internal/bot/slack/stories`
- `POST /internal/bot/slack/logs`
```

- [ ] **Step 4: Test**

Run:

```bash
rg -n "DATABASE_URL|POSTGRES_URL" apps/bot .env.example docs/slack-production-runbook.md
```

Expected: no bot production setup instructs direct database access.

### Task 1: Make Multi-Workspace Slack Token Resolution Production-Safe

**Files:**

- Modify: `apps/server/internal/modules/slack/http/routes.go`
- Modify: `apps/server/internal/modules/slack/http/slack.go`
- Modify: `apps/server/internal/modules/slack/http/models.go`
- Modify: `apps/server/internal/modules/slack/service/models.go`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Modify: `apps/bot/src/lib/fortyone-client.ts`
- Modify: `apps/bot/src/lib/bot.tsx`
- Test: `apps/server/internal/modules/slack/service/slack_test.go`

- [ ] **Step 1: Add a backend internal installation response**

Add DTO/core types that return exactly what `@chat-adapter/slack` needs:

```go
type AppRuntimeSlackInstallation struct {
	BotToken  string `json:"botToken"`
	BotUserID string `json:"botUserId,omitempty"`
	TeamName  string `json:"teamName,omitempty"`
}

type CoreRuntimeSlackInstallation struct {
	BotToken  string
	BotUserID string
	TeamName  string
}
```

- [ ] **Step 2: Add the internal endpoint**

Register:

```go
app.Get("/internal/bot/slack/installations/{teamId}", h.RuntimeGetInstallation, h.BotAuth)
```

Handler behavior:

```go
teamID := strings.TrimSpace(web.Params(r, "teamId"))
installation, err := h.service.RuntimeGetInstallation(ctx, teamID)
if err != nil {
	return web.RespondError(ctx, w, err, http.StatusNotFound)
}
return web.Respond(ctx, w, toAppRuntimeSlackInstallation(installation), http.StatusOK)
```

- [ ] **Step 3: Implement service lookup**

Use existing `repo.GetSlackWorkspaceByTeamID(ctx, teamID)`. Return `BotAccessToken`, `BotUserID`, and `SlackTeamName`. Never log or return the token anywhere except this BotAuth-only endpoint.

- [ ] **Step 4: Add bot client method**

In `FortyOneClient`:

```ts
export type SlackInstallation = {
  botToken: string;
  botUserId?: string;
  teamName?: string;
};

async getSlackInstallation(teamId: string): Promise<SlackInstallation | null> {
  try {
    return await this.request(`/internal/bot/slack/installations/${teamId}`);
  } catch (error) {
    return null;
  }
}
```

- [ ] **Step 5: Wire Chat SDK installation provider**

In `apps/bot/src/lib/bot.tsx`, configure Slack adapter with backend-backed lookup:

```ts
const slackAdapter = createSlackAdapter({
  botToken: config.slackBotToken || undefined,
  installationProvider: config.slackBotToken
    ? undefined
    : {
        getInstallation: async (installationId) =>
          client.getSlackInstallation(installationId),
      },
  signingSecret: config.slackSigningSecret || undefined,
  userName: config.botUserName,
});
```

Keep static `SLACK_BOT_TOKEN` for local single-workspace dev. Production should omit it and use the provider. Token lookup must cross the internal Go API boundary instead of reading the `slack_workspaces` table directly from the bot.

- [ ] **Step 6: Test**

Run:

```bash
pnpm --filter bot type-check
cd apps/server && go test ./internal/modules/slack/...
```

Expected: bot type-check passes; Slack service tests pass.

---

### Task 2: Make Maya AI Text Responses Reliable and Tool-Free for This Slice

**Files:**

- Modify: `apps/bot/src/lib/agent.ts`
- Modify: `apps/bot/src/lib/bot.tsx`
- Modify: `apps/bot/src/lib/tools.ts`
- Test: `apps/bot/src/lib/agent.test.ts` if the bot test setup is added; otherwise validate with type-check and Slack manual tests.

- [ ] **Step 1: Disable tools by default**

Change `createAgentRequest` so default Slack replies are pure text:

```ts
const createAgentRequest = (prompt: string) => ({
  system:
    "You are Maya, the FortyOne Slack assistant. Keep answers concise, friendly, and useful for Slack. Do not claim to create, update, search, or read FortyOne data unless a deterministic integration flow did it.",
  prompt: prompt.trim() || "Say hello and ask how you can help.",
  maxOutputTokens: 600,
});
```

Keep `createTools(...)` behind a later feature flag such as `BOT_ENABLE_TOOLS=true`.

- [ ] **Step 2: Keep the Slack response shape consistent**

Create one helper in `bot.tsx`:

```ts
const replyWithMaya = async (thread, message) => {
  const actor = slackActorFromEvent(message.author, message.raw);
  await thread.post(createAgentStream(message.text, config, client, actor));
};
```

Use it from `onNewMention`, `onDirectMessage`, and `onSubscribedMessage`.

- [ ] **Step 3: Preserve subscriptions**

Keep:

```ts
await thread.subscribe();
```

for new mentions and DMs so follow-up messages work after restart when Redis state is enabled.

- [ ] **Step 4: Test the AI contract manually**

Run bot and API, then test in Slack:

```text
DM Maya: hi
Expected: AI-generated greeting from Maya.

Mention Maya in a channel: @Maya hi
Expected: AI-generated greeting, thread subscribed.

Reply in the subscribed thread: what can you do?
Expected: Maya replies again.
```

---

### Task 3: Resolve Slack User Identity on Every Interaction

**Files:**

- Modify: `apps/server/internal/modules/slack/http/routes.go`
- Modify: `apps/server/internal/modules/slack/http/slack.go`
- Modify: `apps/server/internal/modules/slack/http/models.go`
- Modify: `apps/server/internal/modules/slack/service/models.go`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Modify: `apps/bot/src/lib/fortyone-client.ts`
- Modify: `apps/bot/src/lib/bot.tsx`
- Test: `apps/server/internal/modules/slack/service/slack_test.go`

- [ ] **Step 1: Add identity endpoint**

Register:

```go
app.Post("/internal/bot/slack/identity/resolve", h.RuntimeResolveIdentity, h.BotAuth)
```

Response:

```go
type AppRuntimeIdentityResponse struct {
	WorkspaceID   string  `json:"workspaceId"`
	WorkspaceSlug string `json:"workspaceSlug"`
	UserID        *string `json:"userId,omitempty"`
	ConnectURL    *string `json:"connectUrl,omitempty"`
}
```

- [ ] **Step 2: Implement service method**

Method behavior:

1. Resolve `slack_workspaces` by Slack `team_id`.
2. Use existing `resolveLinkedSlackUser(...)`.
3. If linked, return `userId`.
4. If not linked, return `connectUrl`.
5. Return a clear error if the Slack workspace is not connected.

This resolution must stay in Go because it depends on workspace membership, signed connect links, and persisted provider user links.

- [ ] **Step 3: Add bot client method**

```ts
export type SlackIdentity = {
  workspaceId: string;
  workspaceSlug: string;
  userId?: string;
  connectUrl?: string;
};

async resolveSlackIdentity(actor: SlackActor): Promise<SlackIdentity> {
  return this.request("/internal/bot/slack/identity/resolve", {
    method: "POST",
    body: { actor },
  });
}
```

- [ ] **Step 4: Gate interactions**

In the bot, before AI replies and form submits:

```ts
const identity = await client.resolveSlackIdentity(actor);
if (!identity.userId && identity.connectUrl) {
  await thread.post(
    `Connect your FortyOne account to continue: ${identity.connectUrl}`,
  );
  return;
}
```

For option loads, workspace resolution is enough; user linkage should not block team/status/member/label loading.

- [ ] **Step 5: Test**

Service tests:

```go
func TestRuntimeResolveIdentityReturnsLinkedUser(t *testing.T) {}
func TestRuntimeResolveIdentityReturnsConnectURLWhenUnlinked(t *testing.T) {}
func TestRuntimeResolveIdentityErrorsWhenWorkspaceNotConnected(t *testing.T) {}
```

Manual Slack tests:

```text
Linked Slack user sends DM.
Expected: Maya replies.

Unlinked Slack user sends DM.
Expected: bot replies with connect link.

User opens link while logged into FortyOne.
Expected: settings page consumes slack_link_token and links account.
```

---

### Task 4: Fix Team-Dependent Form Options

**Files:**

- Modify: `apps/bot/src/lib/bot.tsx`
- Modify: `apps/bot/src/lib/create-story-modal.tsx`
- Modify: `apps/bot/src/lib/fortyone-client.ts`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Test: `apps/server/internal/modules/slack/service/slack_test.go`

- [ ] **Step 1: Make selected team extraction robust**

Replace `selectedTeamFromRaw(...)` with a helper that checks both current view state and selected option payload:

```ts
const selectedValueFromViewState = (
  raw: unknown,
  fieldId: string,
): string | undefined => {
  const stateValues = (raw as { view?: { state?: { values?: unknown } } })?.view
    ?.state?.values;
  if (!stateValues || typeof stateValues !== "object") return undefined;

  const block = (stateValues as Record<string, Record<string, unknown>>)[
    fieldId
  ];
  for (const action of Object.values(block ?? {})) {
    const selected = (action as { selected_option?: { value?: string } })
      .selected_option?.value;
    if (selected) return selected;
  }
  return undefined;
};
```

- [ ] **Step 2: Ensure field IDs match Chat SDK state**

Keep modal field IDs equal to `CREATE_STORY_FIELDS.team`, `status`, `assignee`, `objective`, and `labels`. Do not introduce separate block/action IDs unless Chat SDK requires them.

- [ ] **Step 3: Make empty-query loads fast**

Backend option methods should handle `query == ""` and return the first 25 options for teams and statuses. For members/objectives/labels, use `minQueryLength={1}` or return a small first page if Slack calls with an empty query.

- [ ] **Step 4: Validate team scoping**

Keep these backend checks:

```go
workspace, err := s.repo.GetWorkspaceBySlackTeamID(ctx, actor.SlackTeamID)
team, err := s.repo.FindTeamByID(ctx, workspace.ID, teamID)
```

Every dependent option endpoint must verify the selected team belongs to the connected FortyOne workspace.

- [ ] **Step 5: Test**

Manual Slack modal test:

```text
Open /fortyone create.
Select Team A.
Open Status, Assignee, Objective.
Expected: options are scoped to Team A.

Change to Team B.
Open Status, Assignee, Objective.
Expected: options are scoped to Team B, not Team A.
```

---

### Task 5: Add Labels to the Create Story Flow

**Files:**

- Modify: `apps/bot/src/lib/create-story-modal.tsx`
- Modify: `apps/bot/src/lib/bot.tsx`
- Modify: `apps/bot/src/lib/fortyone-client.ts`
- Modify: `apps/server/internal/modules/slack/http/routes.go`
- Modify: `apps/server/internal/modules/slack/http/slack.go`
- Modify: `apps/server/internal/modules/slack/http/models.go`
- Modify: `apps/server/internal/modules/slack/service/models.go`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Test: `apps/server/internal/modules/slack/service/slack_test.go`

- [ ] **Step 1: Add backend runtime label search**

Register:

```go
app.Post("/internal/bot/slack/options/labels", h.RuntimeSearchLabels, h.BotAuth)
```

Service logic:

```go
labels, err := s.repo.SearchTeamLabels(ctx, workspace.ID, teamID, query, 25)
```

Return `[]CoreRuntimeOption`.

- [ ] **Step 2: Add label IDs to create-story DTO**

Add:

```go
LabelIDs []string `json:"labelIds"`
```

to the app/core runtime create-story request.

- [ ] **Step 3: Validate and apply labels**

In `RuntimeCreateStory`, parse label IDs, verify they are valid for the selected team/workspace using `ListTeamLabels` or `FindTeamLabelByID`, then pass `LabelIDs` into `stories.CoreNewStory` or call `stories.UpdateLabels(...)` after story creation.

- [ ] **Step 4: Choose the Chat SDK label UI implementation**

Preferred path: add or use a Chat SDK modal primitive for `multi_external_select`. Current installed Chat SDK types do not expose one. If extending locally, keep the component focused:

```ts
type ExternalMultiSelectElement = {
  id: string;
  initialOptions?: SelectOptionElement[];
  label: string;
  minQueryLength?: number;
  optional?: boolean;
  placeholder?: string;
  type: "external_multi_select";
};
```

Then map it to Slack Block Kit `multi_external_select`.

Fallback path: ship story creation without labels and keep labels out of the modal until the Chat SDK primitive exists. Do not use free-text labels because it will create invalid or ambiguous story metadata.

- [ ] **Step 5: Bot client/search**

Add:

```ts
async searchLabels(actor: SlackActor, teamId: string, query: string): Promise<RuntimeOption[]> {
  return this.request("/internal/bot/slack/options/labels", {
    method: "POST",
    body: { actor, query, teamId },
  });
}
```

- [ ] **Step 6: Test**

Backend tests:

```go
func TestRuntimeSearchLabelsScopesToSelectedTeam(t *testing.T) {}
func TestRuntimeCreateStoryAppliesValidLabels(t *testing.T) {}
func TestRuntimeCreateStoryIgnoresOrErrorsInvalidLabels(t *testing.T) {}
```

Manual Slack test:

```text
Open create story form.
Select a team.
Search labels.
Expected: global workspace labels and team labels appear; labels from other teams do not.
Submit story with labels.
Expected: created FortyOne story has those labels.
```

---

### Task 6: Add Extensive, Searchable Diagnostics

**Files:**

- Create: `apps/bot/src/lib/logger.ts`
- Modify: `apps/bot/src/lib/bot.tsx`
- Modify: `apps/bot/src/lib/fortyone-client.ts`
- Modify: `apps/server/internal/modules/slack/http/slack.go`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Modify: `apps/projects/src/modules/settings/workspace/integrations/slack/slack-integration-settings.tsx`

- [ ] **Step 1: Add bot-side structured logger**

Create a helper that redacts tokens:

```ts
export const logBotError = (
  message: string,
  fields: Record<string, unknown>,
) => {
  console.error(JSON.stringify({ level: "error", message, ...fields }));
};
```

Never log `SLACK_BOT_TOKEN`, `FORTYONE_BOT_TOKEN`, `OPENAI_API_KEY`, or Slack OAuth access tokens.

- [ ] **Step 2: Wrap every handler boundary**

Log failures for:

- `onNewMention`
- `onDirectMessage`
- `onSubscribedMessage`
- `/fortyone`
- `onAction(CREATE_STORY_ACTION_ID)`
- `onOptionsLoad`
- `onModalSubmit`

Each log should include `action`, `teamId`, `userId`, `channelId`, `threadTs`, `actionId`, and the error message.

- [ ] **Step 3: Persist runtime request logs from bot calls**

Add an internal endpoint:

```go
app.Post("/internal/bot/slack/logs", h.RuntimeRecordLog, h.BotAuth)
```

Use existing `slack_request_logs` table. Record option load failures, modal submit failures, identity resolution failures, and AI generation failures.

- [ ] **Step 4: Make settings logs useful**

The UI already has types for `SlackRequestLog`. Add a compact log table to `SlackIntegrationSettings` showing recent outcome, endpoint/action, Slack user, response code, and error message.

- [ ] **Step 5: Test**

Trigger a bad team ID in an option request and confirm:

```text
Bot logs contain structured context.
Server slack_request_logs has a matching row.
Settings page shows the failure without exposing tokens.
```

---

### Task 7: Harden Create Story Submit Behavior

**Files:**

- Modify: `apps/bot/src/lib/create-story-modal.tsx`
- Modify: `apps/bot/src/lib/bot.tsx`
- Modify: `apps/server/internal/modules/slack/service/slack.go`
- Test: `apps/server/internal/modules/slack/service/slack_test.go`

- [ ] **Step 1: Keep validation user-facing**

Return modal field errors for missing title and team:

```ts
if (!input.title.trim())
  return modalError(CREATE_STORY_FIELDS.title, "Add a title.");
if (!input.teamId.trim())
  return modalError(CREATE_STORY_FIELDS.team, "Choose a team.");
```

- [ ] **Step 2: Normalize backend errors**

Map backend validation failures to Slack modal fields where possible:

```text
invalid team -> team
invalid status -> status
invalid assignee -> assignee
invalid objective -> objective
invalid labels -> labels
unlinked user -> title with account-link message
```

- [ ] **Step 3: Confirm story defaults**

If no status is selected, backend should keep current behavior: pick first `unstarted` status for the selected team. If no such status exists, return a clear validation error.

- [ ] **Step 4: Subscribe source threads only after successful creation**

Keep:

```ts
await event.relatedThread.subscribe();
await event.relatedThread.setState({ fortyOneStoryId: story.id });
```

only after the story has been created.

- [ ] **Step 5: Test**

Manual tests:

```text
Submit without title -> title error.
Submit without team -> team error.
Submit with stale status -> status error or default status.
Submit as unlinked user -> link prompt.
Successful submit from message action -> story created and Slack thread subscribed.
```

---

### Task 8: Production Configuration and Secret Hygiene

**Files:**

- Modify: `apps/bot/.env.example`
- Modify: `apps/server/.env.example`
- Modify: `apps/bot/README.md`
- Modify: deployment environment docs/scripts if present

- [ ] **Step 1: Update required production env**

Bot production:

```bash
NODE_ENV=production
CHATSDK_STATE_DRIVER=redis
REDIS_URL=<redis-url>
SLACK_SIGNING_SECRET=<slack-signing-secret>
FORTYONE_API_URL=<go-api-url>
FORTYONE_BOT_TOKEN=<shared-internal-token>
OPENAI_API_KEY=<openai-key>
BOT_OPENAI_MODEL=gpt-5.4-mini
```

Backend production:

```bash
FORTYONE_BOT_TOKEN=<same-shared-internal-token>
SLACK_CLIENT_ID=<slack-client-id>
SLACK_CLIENT_SECRET=<slack-client-secret>
SLACK_REDIRECT_URL=https://<api-domain>/integrations/slack/setup
APP_SECRET_KEY=<existing-secret-used-for-state-signing>
```

Do not add `DATABASE_URL`, `POSTGRES_URL`, or direct database credentials to the bot production environment. The bot should use `FORTYONE_API_URL` plus `FORTYONE_BOT_TOKEN` for all FortyOne data access.

- [ ] **Step 2: Rotate exposed local secrets before production**

Rotate Slack bot tokens, Slack signing secrets, and OpenAI keys that have appeared in local files or terminal output. Keep real values out of committed files and logs.

- [ ] **Step 3: Configure Slack app**

Slack URLs:

```text
Event Request URL: https://<bot-domain>/api/webhooks/slack
Interactivity Request URL: https://<bot-domain>/api/webhooks/slack
Select Menus Options Load URL: https://<bot-domain>/api/webhooks/slack
Slash command /fortyone Request URL: https://<bot-domain>/api/webhooks/slack
OAuth Redirect URL: https://<api-domain>/integrations/slack/setup
```

Scopes:

```text
app_mentions:read
channels:history
channels:read
chat:write
commands
groups:history
groups:read
im:history
im:read
mpim:history
mpim:read
users:read
users:read.email
```

Add `channels:join` only if the product requires posting into channels where the bot is not a member.

- [ ] **Step 4: Validate startup**

Run:

```bash
rg -n "DATABASE_URL|POSTGRES_URL" apps/bot apps/bot/.env.example
pnpm --filter bot build
pnpm --filter bot type-check
cd apps/server && go test ./internal/modules/slack/...
```

Expected: no bot production database dependency, no missing production config, no TypeScript errors, Go tests pass.

---

### Task 9: End-to-End Acceptance Test Matrix

**Files:**

- Modify: `apps/bot/README.md`
- Modify: `docs/slack-production-runbook.md`

- [ ] **Step 1: Workspace install**

```text
From FortyOne settings, click Connect Slack.
Complete Slack OAuth.
Expected: settings page shows connected Slack team and channel sync works.
```

- [ ] **Step 2: AI DM**

```text
A linked Slack user DMs "hi".
Expected: Maya replies with AI-generated text.
```

- [ ] **Step 3: AI mention**

```text
A linked Slack user mentions Maya in a public channel.
Expected: Maya replies in thread and subscribes.
```

- [ ] **Step 4: Unlinked user**

```text
An unlinked Slack user DMs Maya or submits create story.
Expected: Maya returns a FortyOne account-link URL.
```

- [ ] **Step 5: Create story slash command**

```text
/fortyone create Fix onboarding crash
Expected: modal opens with title prefilled; team list loads; dependent fields scope by team; submit creates story.
```

- [ ] **Step 6: Create story from message action**

```text
Use Slack message action on a message.
Expected: modal opens with description prefilled from message; submit creates story and posts story URL.
```

- [ ] **Step 7: Failure observability**

```text
Break FORTYONE_BOT_TOKEN or use an unconnected Slack team.
Expected: Slack receives a graceful response; bot logs and slack_request_logs show actionable error details.
```

- [ ] **Step 8: Boundary check**

```text
Temporarily revoke any bot database env vars.
Expected: bot still boots because it only depends on Redis, Slack credentials, OpenAI, FORTYONE_API_URL, and FORTYONE_BOT_TOKEN.
```

---

### Task 10: Prepare the Same Contract for Microsoft Teams

**Files:**

- Create: `docs/integration-runtime-contract.md`
- Modify: `docs/slack-production-runbook.md`

- [ ] **Step 1: Define provider-neutral concepts**

Create `docs/integration-runtime-contract.md`:

```md
# Integration Runtime Contract

## Provider Runtime

`apps/bot` handles provider adapters through Chat SDK. Providers can include Slack,
Microsoft Teams, Discord, and future chat surfaces.

## Domain Runtime

`apps/server` owns installation records, tokens, workspace mapping, user mapping,
permissions, product data, and audit logs.

## Provider Identity

- `provider`: `slack`, `teams`, or another adapter key
- `externalWorkspaceId`: Slack team ID, Teams tenant ID, or equivalent
- `externalUserId`: Slack user ID, Teams user ID, or equivalent
- `workspaceId`: FortyOne workspace ID
- `userId`: FortyOne user ID
```

- [ ] **Step 2: Document future Teams flow**

Add:

```md
## Future Teams Flow

Teams should reuse the same internal API shape:

- resolve installation by provider and external workspace ID
- resolve user by provider, external workspace ID, and external user ID
- auto-link by verified email when the provider exposes email
- return a signed connect URL when no link exists
- create stories only through the Go API
```

- [ ] **Step 3: Test**

Run:

```bash
rg -n "DATABASE_URL|direct database|Postgres" docs/integration-runtime-contract.md docs/slack-production-runbook.md apps/bot/README.md
```

Expected: docs consistently say the bot does not access Postgres directly.

---

## Suggested Commit Sequence

1. `docs(integrations): lock bot api database boundary`
2. `feat(slack): resolve chat sdk installations from api`
3. `feat(slack): require linked identity for user scoped bot actions`
4. `feat(bot): make maya slack replies text-only`
5. `feat(slack): support team scoped labels in create story`
6. `chore(slack): add production diagnostics and runbook`
7. `docs(integrations): define provider runtime contract`

## Open Questions

- Should unlinked users be allowed to receive generic AI answers, or should every AI message require a linked FortyOne user? The plan assumes linked identity is required because the requested production behavior mentions authentication for every Slack message.
- Should labels block launch? Current Chat SDK does not expose a modal multi-select primitive, while labels naturally need multi-select. The staff-level option is to implement that primitive cleanly or ship labels in the next slice, not to fake it.
- Should the older Go public Slack command/interactivity endpoints remain documented? The plan assumes they become legacy/internal and the Slack app points only at the Chat SDK bot webhook.
- Should provider installation endpoints be Slack-specific first (`/internal/bot/slack/...`) or generalized immediately (`/internal/bot/integrations/{provider}/...`)? The plan keeps Slack-specific endpoints for the shipping slice and adds a provider-neutral contract before Teams work.
