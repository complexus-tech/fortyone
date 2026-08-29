-- Small story support reads stay explicit and tenant-scoped. The service has
-- already resolved the caller's actor and team-access policy; these queries
-- repeat current membership checks as a persistence-level tenant fence.

-- name: GetVisibleTeamEstimateScheme :one
SELECT CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme
FROM public.teams AS team
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = team.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = team.team_id
   AND team_member.user_id = actor.user_id
LEFT JOIN public.team_estimation_settings AS estimation
    ON estimation.team_id = team.team_id
   AND estimation.workspace_id = team.workspace_id
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR team.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  );

-- name: FindVisibleFirstStatusByCategory :one
SELECT status.status_id
FROM public.statuses AS status
INNER JOIN public.teams AS team
    ON team.team_id = status.team_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = team.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = team.team_id
   AND team_member.user_id = actor.user_id
WHERE status.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND status.category = sqlc.arg(category)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR team.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY status.order_index, status.status_id
LIMIT 1;

-- name: ResolveVisibleStoryKeyResult :one
SELECT key_result.objective_id, key_result.name
FROM public.key_results AS key_result
INNER JOIN public.objectives AS objective
    ON objective.objective_id = key_result.objective_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = objective.workspace_id
   AND workspace_member.user_id = actor.user_id
WHERE key_result.id = sqlc.arg(key_result_id)
  AND objective.workspace_id = sqlc.arg(workspace_id);

-- name: ListVisibleStoryLinks :many
SELECT
    story_link.link_id,
    story_link.title,
    story_link.url,
    story_link.story_id,
    story_link.created_at,
    story_link.updated_at
FROM public.story_links AS story_link
INNER JOIN public.stories AS story
    ON story.id = story_link.story_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY story_link.created_at, story_link.link_id;

-- name: ListVisibleStoryActivities :many
SELECT
    activity.activity_id,
    activity.story_id,
    activity.user_id,
    activity.activity_type,
    activity.field_changed,
    activity.current_value,
    activity.old_value,
    activity.new_value,
    activity.reason,
    activity.created_at,
    activity.workspace_id,
    account.username,
    CAST(COALESCE(account.full_name, '') AS text) AS full_name,
    CAST(COALESCE(account.avatar_url, '') AS text) AS avatar_url,
    account.is_active,
    account.is_system
FROM public.story_activities AS activity
INNER JOIN public.stories AS story
    ON story.id = activity.story_id
   AND story.workspace_id = activity.workspace_id
INNER JOIN public.users AS account
    ON account.user_id = activity.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY activity.created_at DESC, activity.activity_id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: GetVisibleStoryStatusCategory :one
SELECT CAST(status.category AS text) AS category
FROM public.statuses AS status
INNER JOIN public.teams AS team
    ON team.team_id = status.team_id
INNER JOIN public.users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = team.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = team.team_id
   AND team_member.user_id = actor.user_id
WHERE status.status_id = sqlc.arg(status_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR team.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  );
