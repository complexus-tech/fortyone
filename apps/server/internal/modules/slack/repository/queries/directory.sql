-- name: FindWorkspaceBySlug :one
SELECT workspace_id, slug, name
FROM public.workspaces
WHERE slug = CAST(sqlc.arg(slug) AS text)
  AND deleted_at IS NULL;

-- name: FindWorkspaceByID :one
SELECT workspace_id, slug, name
FROM public.workspaces
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND deleted_at IS NULL;

-- name: FindTeamByCode :one
SELECT team_id, code, name, color
FROM public.teams
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND LOWER(code) = LOWER(CAST(sqlc.arg(code) AS text))
ORDER BY team_id
LIMIT 1;

-- name: FindTeamByID :one
SELECT team_id, code, name, color
FROM public.teams
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team_id = CAST(sqlc.arg(team_id) AS uuid)
LIMIT 1;

-- name: GetWorkspaceBySlackTeamID :one
SELECT workspace.workspace_id, workspace.slug, workspace.name
FROM public.slack_workspaces AS installation
JOIN public.workspaces AS workspace
  ON workspace.workspace_id = installation.workspace_id
 AND workspace.deleted_at IS NULL
WHERE installation.slack_team_id = CAST(sqlc.arg(slack_team_id) AS text)
  AND installation.is_active = TRUE
LIMIT 1;

-- name: ListWorkspaceTeams :many
SELECT team_id, code, name, color
FROM public.teams
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
ORDER BY LOWER(name), name, team_id;

-- name: ListWorkspaceTeamsForUser :many
SELECT team.team_id, team.code, team.name, team.color
FROM public.teams AS team
JOIN public.team_members AS team_membership
  ON team_membership.team_id = team.team_id
JOIN public.workspace_members AS workspace_membership
  ON workspace_membership.workspace_id = team.workspace_id
 AND workspace_membership.user_id = team_membership.user_id
JOIN public.users AS actor
  ON actor.user_id = team_membership.user_id
 AND actor.is_active = TRUE
LEFT JOIN public.user_team_orders AS team_order
  ON team_order.team_id = team.team_id
 AND team_order.user_id = CAST(sqlc.arg(user_id) AS uuid)
 AND team_order.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
WHERE team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team_membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
ORDER BY
    CASE WHEN team_order.order_index IS NOT NULL THEN 0 ELSE 1 END,
    team_order.order_index ASC NULLS LAST,
    team.created_at DESC,
    team.team_id;

-- name: ListTeamStatuses :many
SELECT status_id, name, CAST(category AS text) AS category
FROM public.statuses
WHERE team_id = CAST(sqlc.arg(team_id) AS uuid)
ORDER BY order_index, name, status_id;

-- name: ListTeamMembers :many
SELECT user_record.user_id, user_record.username,
       COALESCE(user_record.full_name, '') AS full_name, user_record.email
FROM public.team_members AS team_membership
JOIN public.users AS user_record
  ON user_record.user_id = team_membership.user_id
 AND user_record.is_active = TRUE
WHERE team_membership.team_id = CAST(sqlc.arg(team_id) AS uuid)
ORDER BY COALESCE(NULLIF(TRIM(user_record.full_name), ''), NULLIF(TRIM(user_record.username), ''), user_record.email),
         user_record.user_id;

-- name: ListTeamLabels :many
SELECT label_id, name
FROM public.labels
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (team_id IS NULL OR team_id = CAST(sqlc.arg(team_id) AS uuid))
ORDER BY LOWER(name), name, label_id;

-- name: FindTeamMemberByID :one
SELECT user_record.user_id, user_record.username,
       COALESCE(user_record.full_name, '') AS full_name, user_record.email
FROM public.team_members AS team_membership
JOIN public.users AS user_record
  ON user_record.user_id = team_membership.user_id
 AND user_record.is_active = TRUE
WHERE team_membership.team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND team_membership.user_id = CAST(sqlc.arg(user_id) AS uuid)
LIMIT 1;

-- name: FindTeamLabelByID :one
SELECT label_id, name
FROM public.labels
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND label_id = CAST(sqlc.arg(label_id) AS uuid)
  AND (team_id IS NULL OR team_id = CAST(sqlc.arg(team_id) AS uuid))
LIMIT 1;

-- name: FindTeamObjectiveByID :one
SELECT objective_id, name
FROM public.objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND objective_id = CAST(sqlc.arg(objective_id) AS uuid)
LIMIT 1;

-- name: SearchTeamMembers :many
SELECT user_record.user_id, user_record.username,
       COALESCE(user_record.full_name, '') AS full_name, user_record.email
FROM public.team_members AS team_membership
JOIN public.users AS user_record
  ON user_record.user_id = team_membership.user_id
 AND user_record.is_active = TRUE
WHERE team_membership.team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND (
      COALESCE(user_record.full_name, '') ILIKE '%' || CAST(sqlc.arg(search_query) AS text) || '%'
      OR COALESCE(user_record.username, '') ILIKE '%' || CAST(sqlc.arg(search_query) AS text) || '%'
      OR COALESCE(user_record.email, '') ILIKE '%' || CAST(sqlc.arg(search_query) AS text) || '%'
  )
ORDER BY COALESCE(NULLIF(TRIM(user_record.full_name), ''), NULLIF(TRIM(user_record.username), ''), user_record.email),
         user_record.user_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: SearchTeamLabels :many
SELECT label_id, name
FROM public.labels
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (team_id IS NULL OR team_id = CAST(sqlc.arg(team_id) AS uuid))
  AND name ILIKE '%' || CAST(sqlc.arg(search_query) AS text) || '%'
ORDER BY LOWER(name), name, label_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: SearchTeamObjectives :many
SELECT objective_id, name
FROM public.objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND name ILIKE '%' || CAST(sqlc.arg(search_query) AS text) || '%'
ORDER BY LOWER(name), name, objective_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: FindFirstStatusByCategory :one
SELECT status_id
FROM public.statuses
WHERE team_id = CAST(sqlc.arg(team_id) AS uuid)
  AND category = CAST(sqlc.arg(category) AS text)
ORDER BY order_index, status_id
LIMIT 1;
