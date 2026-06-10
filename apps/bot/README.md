# FortyOne Slack Bot (Chat SDK)

Slack-only bot runtime using Chat SDK + AI SDK.

## Run

```bash
pnpm --filter bot dev
```

## Required env

Copy `.env.example` to `.env` and set:

- `OPENAI_API_KEY`
- `SLACK_SIGNING_SECRET`
- `SLACK_BOT_TOKEN`

Optional:

- `BOT_OPENAI_MODEL` (default: `gpt-5.4-mini`)
- `SLACK_BOT_USERNAME` (default: `maya`)

## Slack request URL

Use the same URL for Slack events/interactivity/commands:

- `/api/webhooks/slack`

This runtime uses `@chat-adapter/state-memory`.
