-- name: GetWorkspaceEngagementTotals :one
SELECT
    CAST(COUNT(*) AS int) AS total_events,
    CAST(COUNT(DISTINCT event.user_id) AS int) AS unique_users
FROM workspace_analytics_events AS event
WHERE event.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR event.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR event.user_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR event.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR event.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR event.occurred_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR event.occurred_at <= sqlc.narg(end_date));

-- name: ListWorkspaceEngagementByName :many
SELECT
    event.event_name AS name,
    CAST(COUNT(*) AS int) AS count
FROM workspace_analytics_events AS event
WHERE event.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR event.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR event.user_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR event.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR event.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR event.occurred_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR event.occurred_at <= sqlc.narg(end_date))
GROUP BY event.event_name
ORDER BY count DESC, name ASC
LIMIT 20;

-- name: ListWorkspaceEngagementBySurface :many
SELECT
    event.surface AS name,
    CAST(COUNT(*) AS int) AS count
FROM workspace_analytics_events AS event
WHERE event.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR event.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR event.user_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR event.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR event.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR event.occurred_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR event.occurred_at <= sqlc.narg(end_date))
GROUP BY event.surface
ORDER BY count DESC, name ASC
LIMIT 20;

-- name: ListWorkspaceEngagementTopUsers :many
SELECT
    account.user_id,
    COALESCE(account.full_name, '') AS full_name,
    account.username,
    COALESCE(account.avatar_url, '') AS avatar_url,
    CAST(COUNT(*) AS int) AS events
FROM workspace_analytics_events AS event
INNER JOIN users AS account
    ON account.user_id = event.user_id
   AND account.is_active = TRUE
INNER JOIN workspace_members AS membership
    ON membership.workspace_id = event.workspace_id
   AND membership.user_id = account.user_id
WHERE event.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR event.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR event.user_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR event.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR event.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR event.occurred_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR event.occurred_at <= sqlc.narg(end_date))
GROUP BY account.user_id, account.full_name, account.username, account.avatar_url
ORDER BY events DESC, account.username ASC, account.user_id
LIMIT 10;

