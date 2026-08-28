-- name: CreateActivity :execrows
INSERT INTO public.story_activities (
    story_id,
    user_id,
    activity_type,
    field_changed,
    current_value,
    workspace_id
)
SELECT
    story.id,
    actor.user_id,
    sqlc.arg(activity_type),
    sqlc.arg(field_changed),
    sqlc.arg(current_value),
    story.workspace_id
FROM public.stories AS story
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(user_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS actor_membership
    ON actor_membership.workspace_id = story.workspace_id
   AND actor_membership.user_id = actor.user_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL;

-- name: ListActivitiesForMember :many
SELECT
    activity.activity_id,
    activity.story_id,
    activity.user_id,
    activity.activity_type,
    activity.field_changed,
    activity.current_value,
    activity.created_at,
    activity.workspace_id,
    account.username,
    account.full_name,
    account.avatar_url,
    account.is_active
FROM public.story_activities AS activity
INNER JOIN public.users AS account
    ON account.user_id = activity.user_id
   AND account.is_active = TRUE
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = activity.workspace_id
   AND membership.user_id = account.user_id
WHERE activity.user_id = sqlc.arg(user_id)
  AND activity.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND activity.created_at >= CAST(sqlc.arg(start_date) AS timestamp)
  AND activity.created_at <= CAST(sqlc.arg(end_date) AS timestamp)
ORDER BY activity.created_at DESC, activity.activity_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer);
