-- name: ListTeamWorkload :many
SELECT
    team.team_id,
    team.name AS team_name,
    CAST(COUNT(DISTINCT story.id) AS int) AS assigned,
    CAST(COUNT(DISTINCT story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed,
    CAST(COUNT(DISTINCT member.user_id) * 40 AS int) AS capacity
FROM teams AS team
LEFT JOIN stories AS story
    ON story.team_id = team.team_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
   AND story.created_at >= sqlc.arg(start_date)
   AND story.created_at <= sqlc.arg(end_date)
LEFT JOIN statuses AS status ON status.status_id = story.status_id
LEFT JOIN team_members AS member ON member.team_id = team.team_id
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY team.team_id, team.name
ORDER BY team.name, team.team_id;

-- name: ListMemberContributions :many
SELECT
    account.user_id,
    account.username,
    COALESCE(account.avatar_url, '') AS avatar_url,
    membership.team_id,
    CAST(COUNT(story.id) AS int) AS assigned,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed
FROM users AS account
INNER JOIN team_members AS membership ON membership.user_id = account.user_id
INNER JOIN teams AS team
    ON team.team_id = membership.team_id
   AND team.workspace_id = sqlc.arg(workspace_id)::uuid
LEFT JOIN stories AS story
    ON story.assignee_id = account.user_id
   AND story.team_id = team.team_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
   AND story.created_at >= sqlc.arg(start_date)
   AND story.created_at <= sqlc.arg(end_date)
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE account.is_active = TRUE
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY account.user_id, account.username, account.avatar_url, membership.team_id
ORDER BY account.username, account.user_id, membership.team_id;

-- name: ListTeamVelocity :many
SELECT
    team.team_id,
    team.name AS team_name,
    CAST(0 AS int) AS week1,
    CAST(0 AS int) AS week2,
    CAST(0 AS int) AS week3,
    CAST(ROUND(COUNT(story.id) / 3.0, 2) AS double precision) AS average
FROM teams AS team
LEFT JOIN stories AS story
    ON story.team_id = team.team_id
   AND story.workspace_id = sqlc.arg(workspace_id)::uuid
   AND story.deleted_at IS NULL
   AND story.is_draft = FALSE
   AND CAST(story.updated_at AS date) >= CAST(sqlc.arg(start_date) AS date) - INTERVAL '3 weeks'
   AND story.updated_at <= sqlc.arg(end_date)
LEFT JOIN statuses AS status
    ON status.status_id = story.status_id
   AND status.category = 'completed'
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY team.team_id, team.name
ORDER BY team.name, team.team_id;

-- name: ListWorkloadTrend :many
SELECT
    CAST(story.created_at AS date) AS date,
    CAST(COUNT(story.id) AS int) AS assigned,
    CAST(COUNT(story.id) FILTER (WHERE status.category = 'completed') AS int) AS completed
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
  AND story.deleted_at IS NULL
  AND story.is_draft = FALSE
  AND story.created_at >= sqlc.arg(start_date)
  AND story.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY CAST(story.created_at AS date)
ORDER BY date;
