-- name: LockSlackInstallationLifecycle :exec
SELECT pg_advisory_xact_lock(CAST(sqlc.arg(lock_key) AS bigint));

-- name: ListActiveInstallationsForUpdate :many
SELECT workspace_id, slack_team_id, installation_generation
FROM public.slack_workspaces
WHERE is_active = TRUE
  AND (
      workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      OR slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  )
ORDER BY id
FOR UPDATE;

-- name: GetSlackUninstallState :one
SELECT
    EXISTS (
        SELECT 1
        FROM public.slack_uninstall_outbox
        WHERE slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
          AND status = 'processing'
    ) AS processing,
    EXISTS (
        SELECT 1
        FROM public.slack_uninstall_outbox
        WHERE slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
          AND status = 'revocation_required'
    ) AS resolution_required;

-- name: CompleteSupersededSlackUninstalls :execrows
UPDATE public.slack_uninstall_outbox
SET status = 'completed',
    credential_payload = NULL,
    last_error = 'superseded by Slack reinstall',
    next_attempt_at = NULL,
    processing_started_at = NULL,
    completed_at = NOW(),
    updated_at = NOW()
WHERE slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND status IN ('pending', 'failed');

-- name: DeleteInactiveSlackInstallations :execrows
DELETE FROM public.slack_workspaces
WHERE is_active = FALSE
  AND (
      workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      OR slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  );

-- name: CancelSlackInboundEvents :execrows
UPDATE public.messaging_inbound_events
SET status = 'cancelled',
    payload_encrypted = NULL,
    last_error = CAST(sqlc.arg(reason) AS text),
    recovery_enqueued_at = NULL,
    processed_at = NOW(),
    updated_at = NOW()
WHERE provider = 'slack'
  AND external_workspace_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND status IN ('pending', 'processing', 'failed');

-- name: CancelSlackOutboundDeliveries :execrows
UPDATE public.messaging_outbound_deliveries
SET status = 'cancelled',
    content = NULL,
    last_error = CAST(sqlc.arg(reason) AS text),
    updated_at = NOW()
WHERE provider = 'slack'
  AND external_workspace_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND status IN ('pending', 'delivering', 'failed');

-- name: UpsertSlackInstallation :one
INSERT INTO public.slack_workspaces (
    workspace_id,
    slack_team_id,
    slack_team_name,
    slack_team_domain,
    bot_user_id,
    bot_access_token,
    credential_payload,
    credential_key_version,
    installation_generation,
    slack_app_id,
    enterprise_id,
    authed_user_id,
    scope,
    is_active,
    installed_by_user_id,
    revoked_at
) VALUES (
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(slack_team_id) AS text),
    CAST(sqlc.arg(slack_team_name) AS text),
    CAST(sqlc.arg(slack_team_domain) AS text),
    sqlc.narg(bot_user_id),
    '',
    CAST(sqlc.arg(credential_payload) AS text),
    CAST(sqlc.arg(credential_key_version) AS smallint),
    CAST(sqlc.arg(installation_generation) AS uuid),
    sqlc.narg(slack_app_id),
    sqlc.narg(enterprise_id),
    sqlc.narg(authed_user_id),
    sqlc.narg(scope),
    TRUE,
    CAST(sqlc.arg(installed_by_user_id) AS uuid),
    NULL
)
ON CONFLICT (workspace_id) DO UPDATE SET
    slack_team_id = EXCLUDED.slack_team_id,
    slack_team_name = EXCLUDED.slack_team_name,
    slack_team_domain = EXCLUDED.slack_team_domain,
    bot_user_id = EXCLUDED.bot_user_id,
    bot_access_token = EXCLUDED.bot_access_token,
    credential_payload = EXCLUDED.credential_payload,
    credential_key_version = EXCLUDED.credential_key_version,
    installation_generation = EXCLUDED.installation_generation,
    installation_authorized_at = NOW(),
    slack_app_id = EXCLUDED.slack_app_id,
    enterprise_id = EXCLUDED.enterprise_id,
    authed_user_id = EXCLUDED.authed_user_id,
    scope = EXCLUDED.scope,
    is_active = TRUE,
    installed_by_user_id = EXCLUDED.installed_by_user_id,
    revoked_at = NULL,
    updated_at = NOW()
RETURNING
    id,
    workspace_id,
    slack_team_id,
    slack_team_name,
    slack_team_domain,
    bot_user_id,
    COALESCE(credential_payload, '') AS credential_payload,
    credential_key_version,
    installation_generation,
    installation_authorized_at,
    slack_app_id,
    enterprise_id,
    authed_user_id,
    scope,
    is_active,
    installed_by_user_id,
    revoked_at,
    created_at,
    updated_at;

-- name: RebindSlackRequestThreads :execrows
UPDATE public.integration_request_threads
SET installation_generation = CAST(sqlc.arg(current_generation) AS uuid),
    updated_at = NOW()
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND provider = 'slack'
  AND external_workspace_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation_generation = CAST(sqlc.arg(previous_generation) AS uuid);

-- name: UpsertSlackOAuthInstallerLink :execrows
INSERT INTO public.slack_user_links (
    workspace_id,
    slack_workspace_id,
    slack_team_id,
    slack_user_id,
    user_id,
    linked_via,
    linked_at
)
SELECT
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(slack_workspace_id) AS uuid),
    CAST(sqlc.arg(slack_team_id) AS text),
    CAST(sqlc.arg(slack_user_id) AS text),
    CAST(sqlc.arg(actor_id) AS uuid),
    'oauth_installer',
    NOW()
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
ON CONFLICT (workspace_id, slack_team_id, slack_user_id) DO UPDATE SET
    slack_workspace_id = EXCLUDED.slack_workspace_id,
    user_id = EXCLUDED.user_id,
    linked_via = EXCLUDED.linked_via,
    linked_at = NOW(),
    updated_at = NOW();

-- name: GetSlackWorkspace :one
SELECT
    id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
    bot_user_id, COALESCE(credential_payload, '') AS credential_payload,
    credential_key_version, installation_generation, installation_authorized_at,
    slack_app_id, enterprise_id, authed_user_id, scope, is_active,
    installed_by_user_id, revoked_at, created_at, updated_at
FROM public.slack_workspaces
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND is_active = TRUE
LIMIT 1;

-- name: GetSlackWorkspaceByTeamID :one
SELECT
    id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
    bot_user_id, COALESCE(credential_payload, '') AS credential_payload,
    credential_key_version, installation_generation, installation_authorized_at,
    slack_app_id, enterprise_id, authed_user_id, scope, is_active,
    installed_by_user_id, revoked_at, created_at, updated_at
FROM public.slack_workspaces
WHERE slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND is_active = TRUE
LIMIT 1;

-- name: LockSlackInstallationForDisconnect :one
SELECT
    id,
    workspace_id,
    installation_generation,
    slack_team_id,
    COALESCE(credential_payload, '') AS credential_payload,
    credential_key_version
FROM public.slack_workspaces
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND is_active = TRUE
FOR UPDATE;

-- name: LockSlackInstallationByTeamID :one
SELECT
    id,
    workspace_id,
    installation_generation,
    slack_team_id,
    COALESCE(credential_payload, '') AS credential_payload,
    credential_key_version
FROM public.slack_workspaces
WHERE slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND is_active = TRUE
FOR UPDATE;

-- name: LockSlackInstallationForChannelSync :one
SELECT id
FROM public.slack_workspaces
WHERE id = CAST(sqlc.arg(slack_workspace_id) AS uuid)
  AND workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation_generation = CAST(sqlc.arg(installation_generation) AS uuid)
  AND is_active = TRUE
FOR UPDATE;

-- name: DeleteSlackUserLinksByWorkspace :execrows
DELETE FROM public.slack_user_links
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid);

-- name: DeleteSlackInstallationByID :execrows
DELETE FROM public.slack_workspaces
WHERE id = CAST(sqlc.arg(slack_workspace_id) AS uuid);

-- name: MarkSlackChannelsInactive :execrows
UPDATE public.slack_channels
SET is_active = FALSE,
    updated_at = NOW()
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid);

-- name: UpsertSlackChannel :exec
INSERT INTO public.slack_channels (
    workspace_id,
    slack_workspace_id,
    slack_channel_id,
    name,
    is_private,
    is_archived,
    is_member,
    is_active,
    last_synced_at
) VALUES (
    CAST(sqlc.arg(workspace_id) AS uuid),
    CAST(sqlc.arg(slack_workspace_id) AS uuid),
    CAST(sqlc.arg(slack_channel_id) AS text),
    CAST(sqlc.arg(name) AS text),
    CAST(sqlc.arg(is_private) AS boolean),
    CAST(sqlc.arg(is_archived) AS boolean),
    CAST(sqlc.arg(is_member) AS boolean),
    TRUE,
    NOW()
)
ON CONFLICT (workspace_id, slack_channel_id) DO UPDATE SET
    slack_workspace_id = EXCLUDED.slack_workspace_id,
    name = EXCLUDED.name,
    is_private = EXCLUDED.is_private,
    is_archived = EXCLUDED.is_archived,
    is_member = EXCLUDED.is_member,
    is_active = TRUE,
    last_synced_at = NOW(),
    updated_at = NOW();

-- name: ListChannels :many
SELECT
    id, workspace_id, slack_workspace_id, slack_channel_id, name,
    is_private, is_archived, is_member, is_active,
    is_assistant_configured, last_synced_at, created_at, updated_at
FROM public.slack_channels
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND is_active = TRUE
ORDER BY LOWER(name), name, id;
