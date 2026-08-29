-- name: GetActiveUserByID :one
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at
FROM public.users AS account
WHERE account.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE;

-- name: GetActiveUserByEmail :one
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at
FROM public.users AS account
WHERE account.email = CAST(sqlc.arg(email) AS text)
  AND account.is_active = TRUE;

-- name: GetUserByEmailAnyStatus :one
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at
FROM public.users AS account
WHERE account.email = CAST(sqlc.arg(email) AS text);

-- name: ListUsersByIDs :many
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at
FROM public.users AS account
WHERE account.user_id = ANY(CAST(sqlc.arg(user_ids) AS uuid[]))
ORDER BY account.user_id;

-- name: ListWorkspaceUsers :many
SELECT
    account.user_id,
    account.username,
    account.email,
    account.full_name,
    account.avatar_url,
    account.is_active,
    account.is_system,
    account.is_internal,
    account.has_seen_walkthrough,
    account.timezone,
    account.working_days,
    account.working_start_minute,
    account.working_end_minute,
    account.last_login_at,
    account.last_used_workspace_id,
    account.github_username,
    account.created_at,
    account.updated_at,
    CAST(workspace_membership.role AS text) AS role,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.ai_role_title, '')
        ELSE ''
    END AS team_ai_role_title,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.ai_role_description, '')
        ELSE ''
    END AS team_ai_role_description,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.inferred_ai_role_title, '')
        ELSE ''
    END AS inferred_team_ai_role_title,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.inferred_ai_role_description, '')
        ELSE ''
    END AS inferred_team_ai_role_description,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.inferred_ai_role_story_count, 0)
        ELSE 0
    END::integer AS inferred_team_ai_role_story_count,
    CASE
        WHEN CAST(sqlc.arg(has_team) AS boolean)
            THEN COALESCE(team_membership.inferred_ai_role_confidence, 0)
        ELSE 0
    END::real AS inferred_team_ai_role_confidence,
    team_membership.inferred_ai_role_generated_at,
    (last_activity.last_story_activity_at IS NOT NULL)::boolean AS has_last_story_activity,
    COALESCE(
        last_activity.last_story_activity_at,
        TIMESTAMPTZ '1970-01-01 00:00:00+00'
    )::timestamptz AS last_story_activity_at
FROM public.users AS account
INNER JOIN public.workspace_members AS workspace_membership
    ON workspace_membership.user_id = account.user_id
LEFT JOIN public.teams AS selected_team
    ON selected_team.team_id = sqlc.arg(team_id)
   AND selected_team.workspace_id = workspace_membership.workspace_id
LEFT JOIN public.team_members AS team_membership
    ON team_membership.team_id = selected_team.team_id
   AND team_membership.user_id = account.user_id
LEFT JOIN LATERAL (
    SELECT MAX(activity.created_at) AS last_story_activity_at
    FROM public.story_activities AS activity
    INNER JOIN public.stories AS story ON story.id = activity.story_id
    WHERE activity.user_id = account.user_id
      AND activity.workspace_id = workspace_membership.workspace_id
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND (
          CAST(sqlc.arg(has_team) AS boolean) = FALSE
          OR story.team_id = selected_team.team_id
      )
) AS last_activity ON TRUE
WHERE workspace_membership.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
  AND account.is_system = FALSE
  AND (
      CAST(sqlc.arg(has_team) AS boolean) = FALSE
      OR team_membership.user_id IS NOT NULL
  )
  AND (
      CAST(sqlc.arg(search) AS text) = ''
      OR account.full_name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
      OR account.username ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
      OR account.email ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
  )
ORDER BY account.full_name ASC NULLS LAST, account.user_id ASC
LIMIT CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_limit) AS integer)
    ELSE NULL
END
OFFSET CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_offset) AS integer)
    ELSE 0
END;

-- name: CreateUser :one
INSERT INTO public.users (
    username,
    email,
    full_name,
    avatar_url,
    timezone,
    last_login_at
)
VALUES (
    CAST(sqlc.arg(username) AS text),
    CAST(sqlc.arg(email) AS text),
    CAST(sqlc.arg(full_name) AS text),
    CAST(sqlc.arg(avatar_url) AS text),
    CAST(sqlc.arg(timezone) AS text),
    sqlc.arg(last_login_at)
)
RETURNING
    user_id,
    username,
    email,
    full_name,
    avatar_url,
    is_active,
    is_system,
    is_internal,
    has_seen_walkthrough,
    timezone,
    working_days,
    working_start_minute,
    working_end_minute,
    last_login_at,
    last_used_workspace_id,
    github_username,
    created_at,
    updated_at;

-- name: UpdateActiveUser :one
UPDATE public.users
SET
    username = CASE
        WHEN CAST(sqlc.arg(set_username) AS boolean)
            THEN CAST(sqlc.arg(username) AS text)
        ELSE username
    END,
    full_name = CASE
        WHEN CAST(sqlc.arg(set_full_name) AS boolean)
            THEN CAST(sqlc.arg(full_name) AS text)
        ELSE full_name
    END,
    avatar_url = CASE
        WHEN CAST(sqlc.arg(set_avatar_url) AS boolean)
            THEN CAST(sqlc.arg(avatar_url) AS text)
        ELSE avatar_url
    END,
    has_seen_walkthrough = CASE
        WHEN CAST(sqlc.arg(set_has_seen_walkthrough) AS boolean)
            THEN CAST(sqlc.arg(has_seen_walkthrough) AS boolean)
        ELSE has_seen_walkthrough
    END,
    timezone = CASE
        WHEN CAST(sqlc.arg(set_timezone) AS boolean)
            THEN CAST(sqlc.arg(timezone) AS text)
        ELSE timezone
    END,
    working_days = CASE
        WHEN CAST(sqlc.arg(set_work_schedule) AS boolean)
            THEN CAST(sqlc.arg(working_days) AS smallint[])
        ELSE working_days
    END,
    working_start_minute = CASE
        WHEN CAST(sqlc.arg(set_work_schedule) AS boolean)
            THEN CAST(sqlc.narg(working_start_minute) AS smallint)
        ELSE working_start_minute
    END,
    working_end_minute = CASE
        WHEN CAST(sqlc.arg(set_work_schedule) AS boolean)
            THEN CAST(sqlc.narg(working_end_minute) AS smallint)
        ELSE working_end_minute
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
  AND is_active = TRUE
RETURNING
    user_id,
    username,
    email,
    full_name,
    avatar_url,
    is_active,
    is_system,
    is_internal,
    has_seen_walkthrough,
    timezone,
    working_days,
    working_start_minute,
    working_end_minute,
    last_login_at,
    last_used_workspace_id,
    github_username,
    created_at,
    updated_at;

-- name: ReactivateUserForVerifiedSignIn :one
UPDATE public.users
SET
    is_active = TRUE,
    login_reactivation_policy = 'verified_sign_in',
    last_login_at = CAST(sqlc.arg(signed_in_at) AS timestamptz),
    inactivity_warning_sent_at = NULL,
    updated_at = CAST(sqlc.arg(signed_in_at) AS timestamptz)
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND is_active = FALSE
  AND login_reactivation_policy = 'verified_sign_in'
RETURNING
    user_id,
    username,
    email,
    full_name,
    avatar_url,
    is_active,
    is_system,
    is_internal,
    has_seen_walkthrough,
    timezone,
    working_days,
    working_start_minute,
    working_end_minute,
    last_login_at,
    last_used_workspace_id,
    github_username,
    created_at,
    updated_at;

-- name: DeactivateUser :execrows
UPDATE public.users
SET
    is_active = FALSE,
    login_reactivation_policy = 'verified_sign_in',
    auth_session_version = auth_session_version + 1,
    updated_at = CAST(sqlc.arg(deactivated_at) AS timestamptz)
WHERE user_id = CAST(sqlc.arg(user_id) AS uuid)
  AND is_active = TRUE;

-- name: UpdateLastUsedWorkspaceForMember :execrows
UPDATE public.users AS account
SET
    last_used_workspace_id = sqlc.arg(workspace_id),
    updated_at = CURRENT_TIMESTAMP
WHERE account.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      WHERE membership.workspace_id = sqlc.arg(workspace_id)
        AND membership.user_id = account.user_id
  );
