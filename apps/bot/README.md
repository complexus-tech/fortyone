# FortyOne Slack Bot

Production Slack runtime using Chat SDK + AI SDK.

The bot owns Slack mechanics: mentions, DMs, slash commands, message actions,
modals, option loading, thread subscriptions, and Slack delivery. FortyOne still
owns product state, permissions, story creation, notifications, audit logs, and
identity mapping through internal API endpoints.

## Runtime boundary

`apps/bot` owns provider webhooks, Chat SDK routing, AI SDK calls, and platform
replies. `apps/server` owns Postgres, installation records, provider user links,
permissions, story creation, option data, labels, audit logs, and diagnostics.

Do not give `apps/bot` direct database credentials in production. Add internal
Go API endpoints when the bot needs FortyOne data or writes.

## Run

```bash
pnpm --filter bot dev
```

## Required env

Copy `.env.example` to `.env` and set:

- `SLACK_SIGNING_SECRET`
- `FORTYONE_API_URL`
- `FORTYONE_BOT_TOKEN`
- `OPENAI_API_KEY` for Maya AI replies

Production also requires:

- `CHATSDK_STATE_DRIVER=redis`
- `REDIS_URL`

Do not set `DATABASE_URL` or `POSTGRES_URL` for the bot. The bot reaches
FortyOne data through `FORTYONE_API_URL` and authenticates with
`FORTYONE_BOT_TOKEN`.

## Slack request URL

Use the same URL for Slack events/interactivity/commands:

- `/api/webhooks/slack`

For the current local ngrok session, use:

- `https://1d58-216-234-215-72.ngrok-free.app/api/webhooks/slack`

Slack app configuration should point these surfaces to the same URL:

- Event Request URL
- Interactivity Request URL
- Select Menus Options Load URL
- Slash command Request URL

## Implemented runtime workflows

- @mentions and DMs stream Maya replies through Chat SDK.
- `/fortyone` opens a deterministic create-story modal by default.
- Non-create slash command text is treated as an AI prompt.
- Slack message actions can open the same create-story modal using the message
  text as the description.
- Modal option searches call FortyOne internal APIs for teams, statuses,
  assignees, labels, and objectives.
- Modal submits call FortyOne internal APIs to create stories without AI.
- Linked Slack threads are subscribed and synced back to Story comments.
- Story links posted in subscribed threads ask FortyOne for unfurl details.
