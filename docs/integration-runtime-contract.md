# Integration Runtime Contract

## Provider Runtime

`apps/bot` handles provider adapters through Chat SDK. Providers can include
Slack, Microsoft Teams, Discord, and future chat surfaces.

For Slack, `team_id` is the installed Slack workspace identifier. It is not a
FortyOne team ID.

## Domain Runtime

`apps/server` owns installation records, tokens, workspace mapping, user
mapping, permissions, product data, and audit logs.

FortyOne team options must come from the linked user's memberships inside the
installed workspace. Team-scoped options and story creation must reject selected
team IDs that are outside that linked user's team memberships.

## Provider Identity

- `provider`: `slack`, `teams`, or another adapter key
- `externalWorkspaceId`: Slack team ID, Teams tenant ID, or equivalent
- `externalUserId`: Slack user ID, Teams user ID, or equivalent
- `workspaceId`: FortyOne workspace ID
- `userId`: FortyOne user ID

## Future Teams Flow

Teams should reuse the same internal API shape:

- resolve installation by provider and external workspace ID
- resolve user by provider, external workspace ID, and external user ID
- auto-link by verified email when the provider exposes email
- return a signed connect URL when no link exists
- create stories only through the Go API
