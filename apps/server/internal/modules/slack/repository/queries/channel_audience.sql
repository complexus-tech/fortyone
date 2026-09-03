-- name: ListAssistantChannelTeamAccess :many
SELECT access.slack_channel_id, access.team_id
FROM public.slack_channel_team_access AS access
JOIN public.slack_workspaces AS installation
  ON installation.id = access.slack_workspace_id
 AND installation.workspace_id = access.workspace_id
 AND installation.is_active = TRUE
JOIN public.teams AS team
  ON team.team_id = access.team_id
 AND team.workspace_id = access.workspace_id
 AND team.is_private = FALSE
WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY access.slack_channel_id, LOWER(team.name), team.name, team.team_id;

-- name: ListAssistantChannelTeamAccessForAdmin :many
SELECT access.slack_channel_id, access.team_id
FROM public.slack_channel_team_access AS access
JOIN public.slack_workspaces AS installation
  ON installation.id = access.slack_workspace_id
 AND installation.workspace_id = access.workspace_id
 AND installation.is_active = TRUE
JOIN public.teams AS team
  ON team.team_id = access.team_id
 AND team.workspace_id = access.workspace_id
 AND team.is_private = FALSE
JOIN public.workspace_members AS membership
  ON membership.workspace_id = access.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role = 'admin'
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
 AND workspace.deleted_at IS NULL
WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY access.slack_channel_id, LOWER(team.name), team.name, team.team_id;

-- name: UpdateSlackAssistantChannelConfiguration :execrows
UPDATE public.slack_channels AS channel
SET is_assistant_configured = CAST(sqlc.arg(is_configured) AS boolean),
    updated_at = NOW()
FROM public.slack_workspaces AS installation
WHERE channel.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND channel.slack_workspace_id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND channel.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
  AND channel.is_active = TRUE
  AND installation.id = channel.slack_workspace_id
  AND installation.workspace_id = channel.workspace_id
  AND installation.installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND installation.is_active = TRUE;

-- name: DeleteAssistantPublicChannelTeamAccess :execrows
DELETE FROM public.slack_channel_team_access AS access
USING public.teams AS team
WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND access.slack_workspace_id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND access.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
  AND team.team_id = access.team_id
  AND team.workspace_id = access.workspace_id
  AND team.is_private = FALSE;

-- name: InsertAssistantChannelTeamAccess :execrows
INSERT INTO public.slack_channel_team_access (
    workspace_id,
    slack_workspace_id,
    slack_channel_id,
    team_id,
    created_by_user_id
)
SELECT
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(slack_workspace_id) AS uuid),
    CAST(sqlc.arg(slack_channel_id) AS text),
    team.team_id,
    CAST(sqlc.arg(actor_id) AS uuid)
FROM public.teams AS team
WHERE team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team.team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND team.is_private = FALSE;

-- name: GetAuthorizedAssistantChannelTeamScope :many
WITH configured_public_teams AS (
    SELECT access.team_id
    FROM public.slack_channel_team_access AS access
    JOIN public.slack_channels AS configured_channel
      ON configured_channel.workspace_id = access.workspace_id
     AND configured_channel.slack_workspace_id = access.slack_workspace_id
     AND configured_channel.slack_channel_id = access.slack_channel_id
     AND configured_channel.is_active = TRUE
     AND configured_channel.is_assistant_configured = TRUE
    JOIN public.slack_workspaces AS installation
      ON installation.id = configured_channel.slack_workspace_id
     AND installation.workspace_id = configured_channel.workspace_id
     AND installation.is_active = TRUE
    JOIN public.teams AS mapped_team
      ON mapped_team.team_id = access.team_id
     AND mapped_team.workspace_id = access.workspace_id
     AND mapped_team.is_private = FALSE
    WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND access.slack_workspace_id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
      AND access.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
), configuration AS (
    SELECT EXISTS (SELECT 1 FROM configured_public_teams) AS is_configured
)
SELECT team.team_id, configuration.is_configured AS explicitly_mapped
FROM public.teams AS team
JOIN public.team_members AS team_membership
  ON team_membership.team_id = team.team_id
 AND team_membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
JOIN public.workspace_members AS workspace_membership
  ON workspace_membership.workspace_id = team.workspace_id
 AND workspace_membership.user_id = team_membership.user_id
JOIN public.users AS actor
  ON actor.user_id = team_membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = team.workspace_id
 AND workspace.deleted_at IS NULL
CROSS JOIN configuration
WHERE team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team.is_private = FALSE
  AND (
      (configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_public_teams))
      OR NOT configuration.is_configured
  )
ORDER BY LOWER(team.name), team.name, team.team_id;

-- name: ListAuthorizedChannelTeamIDs :many
WITH configured_teams AS (
    SELECT access.team_id
    FROM public.slack_channel_team_access AS access
    JOIN public.slack_workspaces AS installation
      ON installation.id = access.slack_workspace_id
     AND installation.workspace_id = access.workspace_id
     AND installation.is_active = TRUE
    WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND access.slack_workspace_id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
      AND access.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
), configuration AS (
    SELECT EXISTS (SELECT 1 FROM configured_teams) AS is_configured
)
SELECT team.team_id
FROM public.teams AS team
JOIN public.team_members AS team_membership
  ON team_membership.team_id = team.team_id
 AND team_membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
JOIN public.workspace_members AS workspace_membership
  ON workspace_membership.workspace_id = team.workspace_id
 AND workspace_membership.user_id = team_membership.user_id
JOIN public.users AS actor
  ON actor.user_id = team_membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = team.workspace_id
 AND workspace.deleted_at IS NULL
CROSS JOIN configuration
WHERE team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (
      (configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_teams))
      OR (NOT configuration.is_configured AND team.is_private = FALSE)
  )
ORDER BY LOWER(team.name), team.name, team.team_id;

-- name: ListInstallationAuthorizedChannelTeamIDs :many
WITH active_installation AS (
    SELECT installation.id
    FROM public.slack_workspaces AS installation
    WHERE installation.id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
      AND installation.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND installation.is_active = TRUE
), channel_configuration AS (
    SELECT channel.is_assistant_configured
    FROM public.slack_channels AS channel
    JOIN active_installation AS installation
      ON installation.id = channel.slack_workspace_id
    WHERE channel.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND channel.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
      AND channel.is_active = TRUE
), configured_teams AS (
    SELECT access.team_id
    FROM public.slack_channel_team_access AS access
    JOIN active_installation AS installation
      ON installation.id = access.slack_workspace_id
    JOIN public.teams AS mapped_team
      ON mapped_team.team_id = access.team_id
     AND mapped_team.workspace_id = access.workspace_id
     AND mapped_team.is_private = FALSE
    WHERE access.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND access.slack_channel_id = CAST(sqlc.arg(slack_channel_id) AS text)
), configuration AS (
    SELECT
        EXISTS (SELECT 1 FROM active_installation) AS installation_is_active,
        COALESCE(
            (SELECT is_assistant_configured FROM channel_configuration),
            FALSE
        ) AS is_configured
)
SELECT team.team_id
FROM public.teams AS team
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = team.workspace_id
 AND workspace.deleted_at IS NULL
CROSS JOIN configuration
WHERE team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team.is_private = FALSE
  AND configuration.installation_is_active
  AND (
      (configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_teams))
      OR NOT configuration.is_configured
  )
ORDER BY LOWER(team.name), team.name, team.team_id;
