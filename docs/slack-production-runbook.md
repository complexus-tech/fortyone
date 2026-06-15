# Slack Production Runbook

## Service Boundary

- Bot runtime: `apps/bot`
- Product/domain runtime: `apps/server`
- Database access: `apps/server` only
- Bot to API authentication: `Authorization: Bearer $FORTYONE_BOT_TOKEN`

`apps/bot` must not receive direct Postgres credentials. It should use
`FORTYONE_API_URL` and internal bot endpoints for every FortyOne read or write.

Slack installation is workspace-scoped. Slack's `team_id` identifies the Slack
workspace installation, not a FortyOne team. Create-story team options must be
loaded from the linked FortyOne user's team memberships inside that workspace.

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

## Local Test Flow

1. Start the Go API from `apps/server` with `make dev`.
2. Start the bot with `pnpm --filter bot dev`.
3. Start ngrok so Slack can reach the bot's Next.js server on port `3002`.
4. Use the same webhook URL for Slack Events, Interactivity, Select Menus, and the `/fortyone` slash command.

Current local ngrok URL:

- `https://1d58-216-234-215-72.ngrok-free.app/api/webhooks/slack`

Do not set `SLACK_BOT_TOKEN` in `apps/bot` for local multi-workspace testing.
The bot resolves the installed workspace bot token from
`GET /internal/bot/slack/installations/{teamId}` using the `team_id` in each
Slack payload.

## Production Slack App URLs

- Event Request URL: `https://<bot-domain>/api/webhooks/slack`
- Interactivity Request URL: `https://<bot-domain>/api/webhooks/slack`
- Select Menus Options Load URL: `https://<bot-domain>/api/webhooks/slack`
- Slash command `/fortyone` Request URL: `https://<bot-domain>/api/webhooks/slack`
- OAuth Redirect URL: `https://<api-domain>/integrations/slack/setup`

## Required Scopes

- `app_mentions:read`
- `channels:history`
- `channels:read`
- `chat:write`
- `commands`
- `groups:history`
- `groups:read`
- `im:history`
- `im:read`
- `mpim:history`
- `mpim:read`
- `users:read`
- `users:read.email`

## Manual Verification

- A linked user can DM Maya and receive an AI-generated response.
- An unlinked user receives a FortyOne account-link URL.
- `/fortyone create` opens the create-story modal.
- Team options only include FortyOne teams the linked user belongs to.
- Status, assignee, label, and objective options are scoped to the selected
  team and reject teams outside the linked user's membership.
- Submitting the modal creates a FortyOne story.
- Slack runtime failures appear in bot logs and `slack_request_logs`.
