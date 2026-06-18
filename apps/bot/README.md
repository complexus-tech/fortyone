# FortyOne Slack Bot

Slack runtime for Maya, built with Chat SDK and AI SDK.

The bot has two runtime modes:

- **Production/API mode** when `FORTYONE_API_URL` and `FORTYONE_BOT_TOKEN` are
  set. Story forms, identity checks, installation tokens, and story creation go
  through the Go API.
- **Local fixture mode** when those API env vars are empty. This is only for
  ngrok/UI testing without running the Go API.

## Run locally

```bash
pnpm --filter bot dev
```

The bot runs on port `3002`.

## Required production env

- `OPENAI_API_KEY`
- `FORTYONE_API_URL`
- `FORTYONE_BOT_TOKEN`
- `REDIS_URL`
- `SLACK_SIGNING_SECRET`

Optional:

- `BOT_OPENAI_MODEL` defaults to `gpt-5.4-mini`
- `SLACK_BOT_TOKEN` enables single-workspace/local Slack testing. Production
  workspace installs should normally resolve bot tokens through the Go API
  installation provider.
- `SLACK_BOT_USERNAME` defaults to `maya`

## Slack request URL

Use the same URL for Slack events, interactivity, commands, and select menus:

```text
/api/webhooks/slack
```

For a local ngrok session, the full URL should look like:

```text
https://<ngrok-host>/api/webhooks/slack
```

Slack app configuration should point these surfaces to the same URL:

- Event Request URL
- Interactivity Request URL
- Select Menus Options Load URL
- Slash command Request URL

## Implemented workflows

- @mentions and DMs stream Maya replies through Chat SDK after Slack user
  identity is linked.
- `/fortyone` opens the create-story modal by default.
- `/fortyone create <title>` opens the same modal with a title prefilled.
- `/fortyone <prompt>` sends the prompt to Maya and replies through Slack.
- Slack message shortcuts open the create-story modal using the message text as
  the description.
- Modal option searches load teams, statuses, assignees, labels, and objectives
  from the Go API in production mode.
- Labels are submitted as a multi-select list.
- Modal submit calls the Go API create-story endpoint in production mode and
  posts the created story reference back to Slack.
