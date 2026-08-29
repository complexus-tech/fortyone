-- name: GetWorkloadSummary :one
WITH workload_stories AS (
    SELECT
        story.id,
        story.team_id,
        team.name AS team_name,
        team.code AS team_code,
        story.assignee_id,
        story.priority,
        story.end_date,
        story.estimate_unit,
        status.category
    FROM stories AS story
    INNER JOIN teams AS team
        ON team.team_id = story.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
      AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR story.assignee_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
      AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
      AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR story.created_at >= sqlc.narg(start_date))
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR story.created_at <= sqlc.narg(end_date))
)
SELECT
    CAST(COUNT(*) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')) AS int) AS total_open_stories,
    CAST(COALESCE(SUM(estimate_unit) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')), 0) AS int) AS total_estimate,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND end_date < CURRENT_DATE) AS int) AS overdue_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND priority = 'Urgent') AS int) AS urgent_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND priority = 'High') AS int) AS high_priority_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND estimate_unit IS NULL) AS int) AS unestimated_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND assignee_id IS NULL) AS int) AS unassigned_stories
FROM workload_stories;

-- name: ListMemberWorkload :many
WITH workload_stories AS (
    SELECT
        story.id,
        story.team_id,
        story.assignee_id,
        story.priority,
        story.end_date,
        story.estimate_unit,
        story.updated_at,
        status.category
    FROM stories AS story
    INNER JOIN teams AS team
        ON team.team_id = story.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
      AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR story.assignee_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
      AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
      AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR story.created_at >= sqlc.narg(start_date))
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR story.created_at <= sqlc.narg(end_date))
)
SELECT
    account.user_id,
    COALESCE(account.full_name, '') AS full_name,
    account.username,
    COALESCE(account.avatar_url, '') AS avatar_url,
    CAST(COALESCE(MAX(membership.ai_role_title), '') AS text) AS team_ai_role_title,
    CAST(COALESCE(MAX(membership.ai_role_description), '') AS text) AS team_ai_role_description,
    CAST(COUNT(*) FILTER (WHERE story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')) AS int) AS open_stories,
    CAST(COUNT(*) FILTER (WHERE story.category = 'started') AS int) AS started_stories,
    CAST(COUNT(*) FILTER (WHERE story.category = 'paused') AS int) AS paused_stories,
    CAST(COUNT(*) FILTER (WHERE story.category = 'completed') AS int) AS completed_stories,
    CAST(COUNT(*) FILTER (WHERE (story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')) AND story.end_date < CURRENT_DATE) AS int) AS overdue_stories,
    CAST(COUNT(*) FILTER (WHERE (story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')) AND story.priority = 'Urgent') AS int) AS urgent_stories,
    CAST(COUNT(*) FILTER (WHERE (story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')) AND story.priority = 'High') AS int) AS high_priority_stories,
    CAST(COUNT(*) FILTER (WHERE (story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')) AND story.estimate_unit IS NULL) AS int) AS unestimated_stories,
    CAST(COALESCE(SUM(story.estimate_unit) FILTER (WHERE story.category IS NULL OR story.category NOT IN ('completed', 'cancelled')), 0) AS int) AS estimate_total,
    CAST(MAX(story.updated_at) AS timestamptz) AS last_story_activity_at
FROM workload_stories AS story
INNER JOIN users AS account
    ON account.user_id = story.assignee_id
   AND account.is_active = TRUE
INNER JOIN team_members AS membership
    ON membership.team_id = story.team_id
   AND membership.user_id = account.user_id
WHERE story.assignee_id IS NOT NULL
GROUP BY account.user_id, account.full_name, account.username, account.avatar_url
ORDER BY estimate_total DESC, open_stories DESC, account.username, account.user_id;

-- name: ListTeamWorkloadSummary :many
WITH workload_stories AS (
    SELECT
        story.id,
        story.team_id,
        team.name AS team_name,
        team.code AS team_code,
        story.assignee_id,
        story.end_date,
        story.estimate_unit,
        status.category
    FROM stories AS story
    INNER JOIN teams AS team
        ON team.team_id = story.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
      AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR story.assignee_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
      AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
      AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR story.created_at >= sqlc.narg(start_date))
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR story.created_at <= sqlc.narg(end_date))
)
SELECT
    team_id,
    team_name,
    team_code,
    CAST(COUNT(*) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')) AS int) AS open_stories,
    CAST(COALESCE(SUM(estimate_unit) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')), 0) AS int) AS estimate_total,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND end_date < CURRENT_DATE) AS int) AS overdue_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND assignee_id IS NULL) AS int) AS unassigned_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND estimate_unit IS NULL) AS int) AS unestimated_stories
FROM workload_stories
GROUP BY team_id, team_name, team_code
ORDER BY team_name, team_id;

-- name: GetUnassignedWorkload :one
WITH workload_stories AS (
    SELECT
        story.id,
        story.assignee_id,
        story.priority,
        story.end_date,
        story.estimate_unit,
        status.category
    FROM stories AS story
    INNER JOIN teams AS team
        ON team.team_id = story.team_id
       AND team.workspace_id = sqlc.arg(workspace_id)::uuid
    LEFT JOIN statuses AS status ON status.status_id = story.status_id
    WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
      AND story.deleted_at IS NULL
      AND story.archived_at IS NULL
      AND story.is_draft = FALSE
      AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR story.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
      AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR story.assignee_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
      AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR story.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
      AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR story.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR story.created_at >= sqlc.narg(start_date))
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR story.created_at <= sqlc.narg(end_date))
)
SELECT
    CAST(COUNT(*) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')) AS int) AS stories,
    CAST(COALESCE(SUM(estimate_unit) FILTER (WHERE category IS NULL OR category NOT IN ('completed', 'cancelled')), 0) AS int) AS estimate_total,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND end_date < CURRENT_DATE) AS int) AS overdue_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND priority = 'Urgent') AS int) AS urgent_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND priority = 'High') AS int) AS high_priority_stories,
    CAST(COUNT(*) FILTER (WHERE (category IS NULL OR category NOT IN ('completed', 'cancelled')) AND estimate_unit IS NULL) AS int) AS unestimated_stories
FROM workload_stories
WHERE assignee_id IS NULL;
