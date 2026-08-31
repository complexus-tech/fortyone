-- name: ListAdminWorkspaceIntegrations :many
WITH latest_subscription AS (
    SELECT DISTINCT ON (candidate.workspace_id)
        candidate.workspace_id,
        candidate.subscription_tier
    FROM workspace_subscriptions AS candidate
    ORDER BY candidate.workspace_id, candidate.updated_at DESC NULLS LAST, candidate.stripe_subscription_id DESC
), integration_facts AS (
    SELECT
        workspace.workspace_id,
        workspace.name,
        workspace.slug,
        workspace.avatar_url,
        workspace.created_at,
        subscription.subscription_tier,
        CAST((SELECT COUNT(*) FROM workspace_members AS member WHERE member.workspace_id = workspace.workspace_id) AS bigint) AS member_count,
        slack.id AS slack_id,
        slack.slack_team_name,
        CAST((SELECT COUNT(DISTINCT link.user_id) FROM slack_user_links AS link WHERE link.workspace_id = workspace.workspace_id) AS bigint) AS slack_mapping_count,
        COALESCE((SELECT MAX(channel.last_synced_at) FROM slack_channels AS channel WHERE channel.workspace_id = workspace.workspace_id AND channel.is_active = TRUE), slack.created_at, workspace.created_at) AS slack_last_synced_at,
        CAST((SELECT COUNT(*) FROM github_installations AS installation WHERE installation.workspace_id = workspace.workspace_id AND installation.is_active = TRUE) AS bigint) AS github_connection_count,
        CAST(COALESCE((SELECT MIN(installation.account_login) FROM github_installations AS installation WHERE installation.workspace_id = workspace.workspace_id AND installation.is_active = TRUE), '') AS text) AS github_account_label,
        EXISTS (SELECT 1 FROM github_installations AS installation WHERE installation.workspace_id = workspace.workspace_id AND installation.is_active = TRUE AND installation.suspended_at IS NOT NULL) AS github_needs_attention,
        CAST((SELECT COUNT(*) FROM workspace_members AS member JOIN users AS account ON account.user_id = member.user_id WHERE member.workspace_id = workspace.workspace_id AND account.github_user_id IS NOT NULL) AS bigint) AS github_mapping_count,
        COALESCE(
            (SELECT MAX(repository.last_synced_at) FROM github_repositories AS repository WHERE repository.workspace_id = workspace.workspace_id AND repository.is_active = TRUE),
            (SELECT MAX(installation.created_at) FROM github_installations AS installation WHERE installation.workspace_id = workspace.workspace_id AND installation.is_active = TRUE),
            workspace.created_at
        ) AS github_last_synced_at,
        CAST((SELECT COUNT(*) FROM calendar_connections AS connection WHERE connection.workspace_id = workspace.workspace_id AND connection.revoked_at IS NULL) AS bigint) AS calendar_connection_count,
        EXISTS (SELECT 1 FROM calendar_connections AS connection WHERE connection.workspace_id = workspace.workspace_id AND connection.revoked_at IS NULL AND connection.sync_status = 'failed') AS calendar_needs_attention,
        COALESCE(
            (SELECT MAX(connection.last_synced_at) FROM calendar_connections AS connection WHERE connection.workspace_id = workspace.workspace_id AND connection.revoked_at IS NULL),
            (SELECT MAX(connection.created_at) FROM calendar_connections AS connection WHERE connection.workspace_id = workspace.workspace_id AND connection.revoked_at IS NULL),
            workspace.created_at
        ) AS calendar_last_synced_at,
        figma.id AS figma_id,
        COALESCE(figma.figma_handle, figma.figma_email) AS figma_account_label,
        figma.expires_at AS figma_expires_at,
        figma.created_at AS figma_created_at
    FROM workspaces AS workspace
    LEFT JOIN latest_subscription AS subscription ON subscription.workspace_id = workspace.workspace_id
    LEFT JOIN slack_workspaces AS slack ON slack.workspace_id = workspace.workspace_id AND slack.is_active = TRUE
    LEFT JOIN figma_connections AS figma ON figma.workspace_id = workspace.workspace_id AND figma.is_active = TRUE
    WHERE workspace.deleted_at IS NULL
)
SELECT
    facts.workspace_id,
    facts.name,
    facts.slug,
    facts.avatar_url,
    facts.created_at,
    facts.subscription_tier,
    facts.member_count,
    facts.slack_id,
    facts.slack_team_name,
    facts.slack_mapping_count,
    facts.slack_last_synced_at,
    facts.github_connection_count,
    facts.github_account_label,
    facts.github_needs_attention,
    facts.github_mapping_count,
    facts.github_last_synced_at,
    facts.calendar_connection_count,
    facts.calendar_needs_attention,
    facts.calendar_last_synced_at,
    facts.figma_id,
    facts.figma_account_label,
    facts.figma_expires_at,
    facts.figma_created_at,
    CAST(COUNT(*) OVER () AS bigint) AS total_count
FROM integration_facts AS facts
WHERE (
    CAST(sqlc.arg(search_text) AS text) = ''
    OR facts.name ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
    OR facts.slug ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
)
AND (
    (
        CAST(sqlc.arg(status_filter) AS text) = 'not_connected'
        AND (
            (CAST(sqlc.arg(provider_filter) AS text) = 'slack' AND facts.slack_id IS NULL)
            OR (CAST(sqlc.arg(provider_filter) AS text) = 'github' AND facts.github_connection_count = 0)
            OR (CAST(sqlc.arg(provider_filter) AS text) = 'calendar' AND facts.calendar_connection_count = 0)
            OR (CAST(sqlc.arg(provider_filter) AS text) = 'figma' AND facts.figma_id IS NULL)
            OR (CAST(sqlc.arg(provider_filter) AS text) = '' AND facts.slack_id IS NULL AND facts.github_connection_count = 0 AND facts.calendar_connection_count = 0 AND facts.figma_id IS NULL)
        )
    )
    OR (
        CAST(sqlc.arg(status_filter) AS text) = 'attention'
        AND (
            (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'github') AND facts.github_needs_attention)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'calendar') AND facts.calendar_needs_attention)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'figma') AND facts.figma_id IS NOT NULL AND facts.figma_expires_at <= CAST(sqlc.arg(now_at) AS timestamptz))
        )
    )
    OR (
        CAST(sqlc.arg(status_filter) AS text) = ''
        AND (
            (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'slack') AND facts.slack_id IS NOT NULL)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'github') AND facts.github_connection_count > 0)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'calendar') AND facts.calendar_connection_count > 0)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'figma') AND facts.figma_id IS NOT NULL)
        )
    )
    OR (
        CAST(sqlc.arg(status_filter) AS text) = 'connected'
        AND (
            (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'slack') AND facts.slack_id IS NOT NULL)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'github') AND facts.github_connection_count > 0 AND NOT facts.github_needs_attention)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'calendar') AND facts.calendar_connection_count > 0 AND NOT facts.calendar_needs_attention)
            OR (CAST(sqlc.arg(provider_filter) AS text) IN ('', 'figma') AND facts.figma_id IS NOT NULL AND facts.figma_expires_at > CAST(sqlc.arg(now_at) AS timestamptz))
        )
    )
)
ORDER BY facts.created_at DESC, facts.workspace_id DESC
LIMIT CAST(sqlc.arg(row_limit) AS integer)
OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: GetAdminSlackIntegration :one
SELECT
    slack.id,
    slack.slack_team_id,
    slack.slack_team_name,
    slack.slack_team_domain,
    installer.full_name AS installed_by_name,
    installer.email AS installed_by_email,
    CAST((SELECT COUNT(*) FROM slack_channels AS channel WHERE channel.workspace_id = slack.workspace_id AND channel.is_active = TRUE) AS bigint) AS channel_count,
    CAST((SELECT COUNT(*) FROM slack_channel_team_access AS access WHERE access.workspace_id = slack.workspace_id) AS bigint) AS channel_mapping_count,
    COALESCE((SELECT MAX(channel.last_synced_at) FROM slack_channels AS channel WHERE channel.workspace_id = slack.workspace_id AND channel.is_active = TRUE), slack.created_at) AS last_synced_at,
    slack.created_at
FROM slack_workspaces AS slack
LEFT JOIN users AS installer ON installer.user_id = slack.installed_by_user_id
WHERE slack.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND slack.is_active = TRUE;

-- name: ListAdminGitHubInstallations :many
SELECT
    installation.id,
    installation.account_login,
    installation.account_type,
    installation.repository_selection,
    CASE WHEN installation.suspended_at IS NOT NULL THEN 'suspended' ELSE 'connected' END AS state,
    installer.full_name AS installed_by_name,
    installer.email AS installed_by_email,
    CAST((SELECT COUNT(*) FROM github_repositories AS repository WHERE repository.installation_id = installation.id AND repository.is_active = TRUE) AS bigint) AS repository_count,
    CAST((SELECT COUNT(*) FROM github_issue_sync_links AS link JOIN github_repositories AS repository ON repository.id = link.repository_id WHERE repository.installation_id = installation.id AND link.is_active = TRUE) AS bigint) AS team_mapping_count,
    COALESCE((SELECT MAX(repository.last_synced_at) FROM github_repositories AS repository WHERE repository.installation_id = installation.id AND repository.is_active = TRUE), installation.created_at) AS last_synced_at,
    installation.created_at
FROM github_installations AS installation
LEFT JOIN users AS installer ON installer.user_id = installation.installed_by_user_id
WHERE installation.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND installation.is_active = TRUE
ORDER BY installation.created_at DESC, installation.id DESC;

-- name: ListAdminCalendarConnections :many
SELECT
    connection.connection_id,
    connection.user_id,
    account.full_name,
    account.email,
    connection.provider,
    connection.connected_email,
    connection.sync_status,
    connection.last_synced_at,
    connection.created_at
FROM calendar_connections AS connection
JOIN users AS account ON account.user_id = connection.user_id
WHERE connection.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND connection.revoked_at IS NULL
ORDER BY connection.created_at DESC, connection.connection_id DESC;

-- name: GetAdminFigmaIntegration :one
SELECT
    connection.id,
    COALESCE(connection.figma_handle, connection.figma_email, connection.figma_user_id) AS account_label,
    connector.full_name AS connected_by_name,
    connector.email AS connected_by_email,
    CASE WHEN connection.expires_at <= CAST(sqlc.arg(now_at) AS timestamptz) THEN 'reauthorization_required' ELSE 'connected' END AS state,
    CAST((SELECT COUNT(DISTINCT link.file_key) FROM story_figma_links AS link WHERE link.workspace_id = connection.workspace_id AND link.unavailable_at IS NULL) AS bigint) AS linked_file_count,
    CAST((SELECT COUNT(*) FROM figma_webhooks AS webhook WHERE webhook.connection_id = connection.id AND webhook.is_active = TRUE) AS bigint) AS webhook_count,
    connection.expires_at,
    connection.created_at
FROM figma_connections AS connection
JOIN users AS connector ON connector.user_id = connection.connected_by_user_id
WHERE connection.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND connection.is_active = TRUE;

-- name: ListAdminIntegrationMemberMappings :many
SELECT
    account.user_id,
    account.full_name,
    account.email,
    CAST(member.role AS text) AS role,
    COALESCE(slack_link.slack_user_id, '') AS slack_user_id,
    COALESCE(slack_link.linked_via, '') AS slack_linked_via,
    COALESCE(slack_link.linked_at, TIMESTAMPTZ 'epoch') AS slack_linked_at,
    account.github_username,
    calendar.provider AS calendar_provider,
    calendar.connected_email AS calendar_email,
    calendar.sync_status AS calendar_state
FROM workspace_members AS member
JOIN users AS account ON account.user_id = member.user_id
LEFT JOIN LATERAL (
    SELECT
        candidate.slack_user_id,
        candidate.linked_via,
        candidate.linked_at
    FROM slack_user_links AS candidate
    WHERE candidate.workspace_id = member.workspace_id
      AND candidate.user_id = member.user_id
    ORDER BY candidate.linked_at DESC, candidate.id DESC
    LIMIT 1
) AS slack_link ON TRUE
LEFT JOIN calendar_connections AS calendar
  ON calendar.workspace_id = member.workspace_id
 AND calendar.user_id = member.user_id
 AND calendar.revoked_at IS NULL
WHERE member.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY
    CASE CAST(member.role AS text) WHEN 'admin' THEN 0 ELSE 1 END,
    account.full_name NULLS LAST,
    account.email,
    account.user_id;
