-- name: GetWorkspaceRole :one
SELECT CAST(membership.role AS text) AS role
FROM public.workspace_members AS membership
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
 AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid);

-- name: LockSlackWorkspaceAdmin :one
SELECT membership.user_id
FROM public.workspace_members AS membership
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = membership.workspace_id
 AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
  AND membership.role = 'admin'
FOR UPDATE OF membership, actor, workspace;

-- name: GetSlackWorkspaceForMember :one
SELECT
    installation.id,
    installation.workspace_id,
    installation.slack_team_id,
    installation.slack_team_name,
    installation.slack_team_domain,
    installation.bot_user_id,
    COALESCE(installation.credential_payload, '') AS credential_payload,
    installation.credential_key_version,
    installation.installation_generation,
    installation.installation_authorized_at,
    installation.slack_app_id,
    installation.enterprise_id,
    installation.authed_user_id,
    installation.scope,
    installation.is_active,
    installation.installed_by_user_id,
    installation.revoked_at,
    installation.created_at,
    installation.updated_at
FROM public.slack_workspaces AS installation
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = installation.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.workspace_members AS membership
  ON membership.workspace_id = installation.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role IN ('admin', 'member')
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
WHERE installation.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation.is_active = TRUE
LIMIT 1;

-- name: FindSlackUserLinkForMember :one
SELECT link.slack_user_id, link.user_id, link.linked_via, link.linked_at
FROM public.slack_user_links AS link
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = link.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.workspace_members AS membership
  ON membership.workspace_id = link.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role IN ('admin', 'member')
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
WHERE link.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND link.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND link.user_id = CAST(sqlc.arg(actor_id) AS uuid)
LIMIT 1;

-- name: ListChannelsForMember :many
SELECT
    channel.id,
    channel.workspace_id,
    channel.slack_workspace_id,
    channel.slack_channel_id,
    channel.name,
    channel.is_private,
    channel.is_archived,
    channel.is_member,
    channel.is_active,
    channel.is_assistant_configured,
    channel.last_synced_at,
    channel.created_at,
    channel.updated_at
FROM public.slack_channels AS channel
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = channel.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.workspace_members AS membership
  ON membership.workspace_id = channel.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role IN ('admin', 'member')
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
WHERE channel.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND channel.is_active = TRUE
ORDER BY LOWER(channel.name), channel.name, channel.id;

-- name: ListRequestLogsForAdmin :many
SELECT
    request.id,
    request.request_type,
    request.endpoint,
    request.workspace_id,
    request.slack_team_id,
    request.slack_user_id,
    request.slack_channel_id,
    request.command,
    request.trigger_id,
    request.request_body,
    request.headers,
    request.response_code,
    request.outcome,
    request.error_message,
    request.created_at
FROM public.slack_request_logs AS request
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = request.workspace_id
 AND workspace.deleted_at IS NULL
JOIN public.workspace_members AS membership
  ON membership.workspace_id = request.workspace_id
 AND membership.user_id = CAST(sqlc.arg(actor_id) AS uuid)
 AND membership.role = 'admin'
JOIN public.users AS actor
  ON actor.user_id = membership.user_id
 AND actor.is_active = TRUE
WHERE request.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY request.created_at DESC, request.id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer);
