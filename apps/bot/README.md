# FortyOne Slack Bot

Production Slack runtime using Chat SDK + AI SDK.

The bot owns Slack mechanics: mentions, DMs, slash commands, message actions,
modals, option loading, thread subscriptions, and Slack delivery. FortyOne still
owns product state, permissions, story creation, notifications, audit logs, and
identity mapping through internal API endpoints.

## Run

```bash
pnpm --filter bot dev
```

## Required env

Copy `.env.example` to `.env` and set:

- `SLACK_SIGNING_SECRET`
- `SLACK_BOT_TOKEN` for a single-workspace dev install, or `SLACK_CLIENT_ID` and
  `SLACK_CLIENT_SECRET` for Slack OAuth installs
- `FORTYONE_API_URL`
- `FORTYONE_BOT_TOKEN`
- `OPENAI_API_KEY` for Maya AI replies

Production also requires:

- `CHATSDK_STATE_DRIVER=redis`
- `REDIS_URL`
- `SLACK_INSTALLATION_ENCRYPTION_KEY` when Chat SDK manages Slack installations

## Slack request URL

Use the same URL for Slack events/interactivity/commands:

- `/api/webhooks/slack`

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
  assignees, and objectives.
- Modal submits call FortyOne internal APIs to create stories without AI.
- Linked Slack threads are subscribed and synced back to Story comments.
- Story links posted in subscribed threads ask FortyOne for unfurl details.
- Mention notifications are available to Maya through a scoped tool so users can
  ask about recent @mentions in Slack.
