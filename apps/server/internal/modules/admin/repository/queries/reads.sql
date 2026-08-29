-- name: GetAdminUser :one
SELECT
    target.user_id,
    target.username,
    target.email,
    target.full_name,
    target.avatar_url,
    target.is_active,
    target.is_system,
    target.is_internal,
    target.login_reactivation_policy,
    target.last_login_at,
    target.last_used_workspace_id,
    last_workspace.name AS last_used_workspace,
    target.github_username,
    CAST((
        SELECT COUNT(*)
        FROM workspace_members AS membership
        WHERE membership.user_id = target.user_id
    ) AS bigint) AS workspace_count,
    target.created_at,
    target.updated_at
FROM users AS target
LEFT JOIN workspaces AS last_workspace
  ON last_workspace.workspace_id = target.last_used_workspace_id
WHERE target.user_id = CAST(sqlc.arg(user_id) AS uuid);

-- name: GetAdminDashboardSummary :one
SELECT
    CAST((SELECT COUNT(*) FROM workspaces AS active_workspace WHERE active_workspace.deleted_at IS NULL) AS bigint) AS total_workspaces,
    CAST((SELECT COUNT(*) FROM workspaces AS trial_workspace WHERE trial_workspace.deleted_at IS NULL AND trial_workspace.trial_ends_on > CAST(sqlc.arg(now_at) AS timestamptz)) AS bigint) AS active_trials,
    CAST((SELECT COUNT(*) FROM workspaces AS expired_workspace WHERE expired_workspace.deleted_at IS NULL AND expired_workspace.trial_ends_on IS NOT NULL AND expired_workspace.trial_ends_on <= CAST(sqlc.arg(now_at) AS timestamptz)) AS bigint) AS expired_trials,
    CAST((
        SELECT COUNT(DISTINCT paid_subscription.workspace_id)
        FROM workspace_subscriptions AS paid_subscription
        WHERE paid_subscription.subscription_status IN ('active', 'trialing', 'past_due')
          AND COALESCE(CAST(paid_subscription.subscription_tier AS text), 'free') <> 'free'
    ) AS bigint) AS paid_workspaces,
    CAST((SELECT COUNT(*) FROM workspaces AS deleted_workspace WHERE deleted_workspace.deleted_at IS NOT NULL) AS bigint) AS deleted_workspaces,
    CAST((SELECT COUNT(*) FROM users AS counted_user) AS bigint) AS total_users,
    CAST((SELECT COUNT(*) FROM users AS internal_user WHERE internal_user.is_internal = TRUE) AS bigint) AS internal_users,
    CAST((
        SELECT COUNT(*)
        FROM workspace_subscriptions AS active_subscription
        WHERE active_subscription.subscription_status IN ('active', 'trialing', 'past_due')
    ) AS bigint) AS active_subscriptions,
    CAST((SELECT COUNT(*) FROM slack_workspaces AS slack_installation WHERE slack_installation.is_active = TRUE) AS bigint) AS slack_installations,
    CAST((SELECT COUNT(*) FROM github_installations AS github_installation WHERE github_installation.is_active = TRUE) AS bigint) AS github_installations,
    CAST((
        SELECT COUNT(*)
        FROM admin_audit_logs AS recent_audit
        WHERE recent_audit.created_at >= CAST(sqlc.arg(now_at) AS timestamptz) - INTERVAL '30 days'
    ) AS bigint) AS recent_admin_audit_logs;

-- name: ListAdminWorkspaces :many
WITH latest_subscription AS (
    SELECT DISTINCT ON (candidate.workspace_id)
        candidate.workspace_id,
        candidate.subscription_tier,
        candidate.subscription_status,
        candidate.seat_count,
        candidate.stripe_customer_id,
        candidate.stripe_subscription_id
    FROM workspace_subscriptions AS candidate
    ORDER BY candidate.workspace_id, candidate.updated_at DESC NULLS LAST, candidate.stripe_subscription_id DESC
)
SELECT
    workspace.workspace_id,
    workspace.name,
    workspace.slug,
    workspace.avatar_url,
    workspace.color,
    workspace.team_size,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.last_accessed_at,
    workspace.created_by,
    creator.email AS created_by_email,
    creator.full_name AS created_by_name,
    CAST((SELECT COUNT(*) FROM workspace_members WHERE workspace_id = workspace.workspace_id) AS bigint) AS member_count,
    CAST((SELECT COUNT(*) FROM teams WHERE workspace_id = workspace.workspace_id) AS bigint) AS team_count,
    CAST((SELECT COUNT(*) FROM stories WHERE workspace_id = workspace.workspace_id AND deleted_at IS NULL) AS bigint) AS story_count,
    subscription.subscription_tier,
    subscription.subscription_status,
    subscription.seat_count AS subscription_seats,
    subscription.stripe_customer_id,
    subscription.stripe_subscription_id,
    EXISTS (
        SELECT 1 FROM slack_workspaces
        WHERE workspace_id = workspace.workspace_id AND is_active = TRUE
    ) AS slack_installed,
    EXISTS (
        SELECT 1 FROM github_installations
        WHERE workspace_id = workspace.workspace_id AND is_active = TRUE
    ) AS github_installed,
    workspace.created_at,
    workspace.updated_at,
    CAST(COUNT(*) OVER () AS bigint) AS total_count
FROM workspaces AS workspace
LEFT JOIN users AS creator ON creator.user_id = workspace.created_by
LEFT JOIN latest_subscription AS subscription
  ON subscription.workspace_id = workspace.workspace_id
WHERE (
    CAST(sqlc.arg(search_text) AS text) = ''
    OR workspace.name ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
    OR workspace.slug ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
    OR creator.email ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
)
AND (
    CAST(sqlc.arg(status_filter) AS text) = ''
    OR (CAST(sqlc.arg(status_filter) AS text) = 'active' AND workspace.deleted_at IS NULL)
    OR (CAST(sqlc.arg(status_filter) AS text) = 'trialing' AND workspace.deleted_at IS NULL AND workspace.trial_ends_on > CAST(sqlc.arg(now_at) AS timestamptz))
    OR (CAST(sqlc.arg(status_filter) AS text) = 'expired' AND workspace.deleted_at IS NULL AND workspace.trial_ends_on IS NOT NULL AND workspace.trial_ends_on <= CAST(sqlc.arg(now_at) AS timestamptz))
    OR (CAST(sqlc.arg(status_filter) AS text) = 'expiring' AND workspace.deleted_at IS NULL AND workspace.trial_ends_on > CAST(sqlc.arg(now_at) AS timestamptz) AND workspace.trial_ends_on <= CAST(sqlc.arg(now_at) AS timestamptz) + INTERVAL '7 days')
    OR (CAST(sqlc.arg(status_filter) AS text) = 'paid' AND workspace.deleted_at IS NULL AND subscription.subscription_status IN ('active', 'trialing', 'past_due') AND COALESCE(CAST(subscription.subscription_tier AS text), 'free') <> 'free')
    OR (CAST(sqlc.arg(status_filter) AS text) = 'past_due' AND workspace.deleted_at IS NULL AND subscription.subscription_status = 'past_due')
    OR (CAST(sqlc.arg(status_filter) AS text) = 'deleted' AND workspace.deleted_at IS NOT NULL)
)
ORDER BY workspace.created_at DESC, workspace.workspace_id DESC
LIMIT CAST(sqlc.arg(row_limit) AS integer)
OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: GetAdminWorkspace :one
WITH latest_subscription AS (
    SELECT DISTINCT ON (candidate.workspace_id)
        candidate.workspace_id,
        candidate.subscription_tier,
        candidate.subscription_status,
        candidate.seat_count,
        candidate.stripe_customer_id,
        candidate.stripe_subscription_id
    FROM workspace_subscriptions AS candidate
    ORDER BY candidate.workspace_id, candidate.updated_at DESC NULLS LAST, candidate.stripe_subscription_id DESC
)
SELECT
    workspace.workspace_id,
    workspace.name,
    workspace.slug,
    workspace.avatar_url,
    workspace.color AS workspace_color,
    workspace.team_size AS workspace_team_size,
    workspace.trial_ends_on,
    workspace.deleted_at,
    workspace.last_accessed_at,
    workspace.created_by,
    creator.email AS created_by_email,
    creator.full_name AS created_by_name,
    CAST((SELECT COUNT(*) FROM workspace_members WHERE workspace_id = workspace.workspace_id) AS bigint) AS member_count,
    CAST((SELECT COUNT(*) FROM teams WHERE workspace_id = workspace.workspace_id) AS bigint) AS team_count,
    CAST((SELECT COUNT(*) FROM stories WHERE workspace_id = workspace.workspace_id AND deleted_at IS NULL) AS bigint) AS story_count,
    subscription.subscription_tier,
    subscription.subscription_status,
    subscription.seat_count AS subscription_seats,
    subscription.stripe_customer_id,
    subscription.stripe_subscription_id,
    EXISTS (
        SELECT 1 FROM slack_workspaces
        WHERE workspace_id = workspace.workspace_id AND is_active = TRUE
    ) AS slack_installed,
    EXISTS (
        SELECT 1 FROM github_installations
        WHERE workspace_id = workspace.workspace_id AND is_active = TRUE
    ) AS github_installed,
    workspace.created_at,
    workspace.updated_at
FROM workspaces AS workspace
LEFT JOIN users AS creator ON creator.user_id = workspace.created_by
LEFT JOIN latest_subscription AS subscription
  ON subscription.workspace_id = workspace.workspace_id
WHERE workspace.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid);

-- name: ListAdminWorkspaceMembers :many
SELECT
    member.user_id,
    member.email,
    member.full_name,
    CAST(membership.role AS text) AS role,
    member.is_internal,
    membership.created_at AS joined_at
FROM workspace_members AS membership
JOIN users AS member ON member.user_id = membership.user_id
WHERE membership.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY
    CASE CAST(membership.role AS text) WHEN 'admin' THEN 0 ELSE 1 END,
    member.full_name NULLS LAST,
    member.email,
    member.user_id;

-- name: ListAdminUsers :many
SELECT
    target.user_id,
    target.username,
    target.email,
    target.full_name,
    target.avatar_url,
    target.is_active,
    target.is_system,
    target.is_internal,
    target.login_reactivation_policy,
    target.last_login_at,
    target.last_used_workspace_id,
    last_workspace.name AS last_used_workspace,
    target.github_username,
    CAST((SELECT COUNT(*) FROM workspace_members WHERE user_id = target.user_id) AS bigint) AS workspace_count,
    target.created_at,
    target.updated_at,
    CAST(COUNT(*) OVER () AS bigint) AS total_count
FROM users AS target
LEFT JOIN workspaces AS last_workspace ON last_workspace.workspace_id = target.last_used_workspace_id
WHERE CAST(sqlc.arg(search_text) AS text) = ''
   OR target.email ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
   OR target.username ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
   OR target.full_name ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
ORDER BY target.created_at DESC, target.user_id DESC
LIMIT CAST(sqlc.arg(row_limit) AS integer)
OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: ListAdminUserMemberships :many
SELECT
    workspace.workspace_id,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(membership.role AS text) AS role,
    membership.created_at AS joined_at
FROM workspace_members AS membership
JOIN workspaces AS workspace ON workspace.workspace_id = membership.workspace_id
WHERE membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
ORDER BY membership.created_at DESC, workspace.workspace_id DESC;
