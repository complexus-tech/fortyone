-- name: CreateOKRActivity :execrows
INSERT INTO public.okr_activities (
    objective_id,
    key_result_id,
    user_id,
    activity_type,
    update_type,
    field_changed,
    current_value,
    comment,
    workspace_id
)
SELECT
    objective.objective_id,
    CAST(sqlc.narg(key_result_id) AS uuid),
    actor.user_id,
    CAST(sqlc.arg(activity_type) AS okr_activity_type),
    CAST(sqlc.arg(update_type) AS okr_update_type),
    CAST(sqlc.narg(field_changed) AS text),
    CAST(sqlc.narg(current_value) AS text),
    CAST(sqlc.narg(comment) AS text),
    objective.workspace_id
FROM public.objectives AS objective
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = actor.user_id
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = actor.user_id
WHERE objective.objective_id = sqlc.arg(objective_id)
  AND objective.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
  AND (
      CAST(sqlc.narg(key_result_id) AS uuid) IS NULL
      OR EXISTS (
          SELECT 1
          FROM public.key_results AS key_result
          WHERE key_result.id = CAST(sqlc.narg(key_result_id) AS uuid)
            AND key_result.objective_id = objective.objective_id
            AND key_result.team_id = objective.team_id
      )
  );

-- name: ListObjectiveActivities :many
SELECT
    activity.activity_id,
    activity.objective_id,
    activity.key_result_id,
    activity.user_id,
    CAST(activity.activity_type AS text) AS activity_type,
    CAST(activity.update_type AS text) AS update_type,
    COALESCE(activity.field_changed, '') AS field_changed,
    COALESCE(activity.current_value, '') AS current_value,
    COALESCE(activity.comment, '') AS comment,
    COALESCE(activity.created_at, timestamp 'epoch') AS created_at,
    activity.workspace_id,
    account.username,
    COALESCE(account.full_name, '') AS full_name,
    COALESCE(account.avatar_url, '') AS avatar_url,
    account.is_active
FROM public.okr_activities AS activity
INNER JOIN public.objectives AS objective
    ON objective.objective_id = activity.objective_id
   AND objective.workspace_id = activity.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.users AS account
    ON account.user_id = activity.user_id
   AND account.is_active = TRUE
WHERE activity.objective_id = sqlc.arg(objective_id)
  AND activity.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
ORDER BY activity.created_at DESC, activity.activity_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: ListKeyResultActivities :many
SELECT
    activity.activity_id,
    activity.objective_id,
    activity.key_result_id,
    activity.user_id,
    CAST(activity.activity_type AS text) AS activity_type,
    CAST(activity.update_type AS text) AS update_type,
    COALESCE(activity.field_changed, '') AS field_changed,
    COALESCE(activity.current_value, '') AS current_value,
    COALESCE(activity.comment, '') AS comment,
    COALESCE(activity.created_at, timestamp 'epoch') AS created_at,
    activity.workspace_id,
    account.username,
    COALESCE(account.full_name, '') AS full_name,
    COALESCE(account.avatar_url, '') AS avatar_url,
    account.is_active
FROM public.okr_activities AS activity
INNER JOIN public.key_results AS key_result
    ON key_result.id = activity.key_result_id
   AND key_result.objective_id = activity.objective_id
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
   AND objective.team_id = key_result.team_id
   AND objective.workspace_id = activity.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.team_members AS team_membership
    ON team_membership.team_id = objective.team_id
   AND team_membership.user_id = membership.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.users AS account
    ON account.user_id = activity.user_id
   AND account.is_active = TRUE
WHERE activity.key_result_id = sqlc.arg(key_result_id)
  AND activity.workspace_id = sqlc.arg(workspace_id)
  AND membership.role IN ('member', 'admin')
ORDER BY activity.created_at DESC, activity.activity_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);
