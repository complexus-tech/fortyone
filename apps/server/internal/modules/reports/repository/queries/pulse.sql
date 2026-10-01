-- name: GetPulseStoryHealth :one
SELECT
    CAST(COUNT(*) FILTER (WHERE status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AS int) AS open_stories,
    CAST(COUNT(*) FILTER (WHERE status.category = 'started') AS int) AS started_stories,
    CAST(COUNT(*) FILTER (WHERE status.category = 'paused') AS int) AS paused_stories,
    CAST(COUNT(*) FILTER (WHERE status.category = 'completed') AS int) AS completed_stories,
    CAST(COUNT(*) FILTER (WHERE status.category = 'cancelled') AS int) AS cancelled_stories,
    CAST(COUNT(*) FILTER (
    WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled'))
          AND EXISTS (
              SELECT 1
              FROM stories AS blocker
              LEFT JOIN statuses AS blocker_status ON blocker_status.status_id = blocker.status_id
          WHERE blocker.workspace_id = story.workspace_id
                AND blocker.deleted_at IS NULL
                AND blocker.archived_at IS NULL
                AND blocker.is_draft = FALSE
                AND (blocker_status.category IS NULL OR blocker_status.category NOT IN ('completed', 'cancelled'))
                AND (
                    blocker.id = story.blocked_by_id
                    OR EXISTS (
                        SELECT 1
                        FROM story_associations AS association
                    WHERE association.workspace_id = story.workspace_id
                          AND association.from_story_id = blocker.id
                          AND association.to_story_id = story.id
                          AND association.association_type = 'blocking'
                    )
                )
          )
    ) AS int) AS blocked_stories,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND story.end_date < CURRENT_DATE) AS int) AS overdue_stories,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND story.priority = 'Urgent') AS int) AS urgent_stories,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND story.priority = 'High') AS int) AS high_priority_stories,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND story.assignee_id IS NULL) AS int) AS unassigned_stories,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND story.estimate_unit IS NULL) AS int) AS unestimated_stories
FROM stories AS story
LEFT JOIN statuses AS status ON status.status_id = story.status_id
WHERE story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND story.is_draft = FALSE
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR story.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR story.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR story.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR story.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR story.created_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR story.created_at <= sqlc.narg(end_date));

-- name: GetPulseSprintHealth :one
WITH sprint_scope AS (
    SELECT
        sprint.sprint_id,
        sprint.start_date,
        sprint.end_date,
        CAST(COUNT(story.id) FILTER (
        WHERE story_status.category IS NULL OR story_status.category NOT IN ('completed', 'cancelled')
        ) AS int) AS open_stories,
        CAST(COUNT(story.id) FILTER (
        WHERE (story_status.category IS NULL OR story_status.category NOT IN ('completed', 'cancelled'))
              AND story.end_date < CURRENT_DATE
        ) AS int) AS overdue_stories,
        CAST(COUNT(story.id) FILTER (
        WHERE (story_status.category IS NULL OR story_status.category NOT IN ('completed', 'cancelled'))
              AND story.estimate_unit IS NULL
        ) AS int) AS unestimated_stories
    FROM sprints AS sprint
    LEFT JOIN stories AS story
        ON story.sprint_id = sprint.sprint_id
       AND story.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
       AND story.deleted_at IS NULL
       AND story.archived_at IS NULL
       AND story.is_draft = FALSE
    LEFT JOIN statuses AS story_status ON story_status.status_id = story.status_id
    WHERE sprint.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
      AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR sprint.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR sprint.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
      AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR sprint.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR sprint.created_at >= sqlc.narg(start_date))
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR sprint.created_at <= sqlc.narg(end_date))
    GROUP BY sprint.sprint_id, sprint.start_date, sprint.end_date
)
SELECT
    CAST(COUNT(*) FILTER (WHERE start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE) AS int) AS active_sprints,
    CAST(COUNT(*) FILTER (WHERE start_date > CURRENT_DATE) AS int) AS upcoming_sprints,
    CAST(COUNT(*) FILTER (WHERE end_date < CURRENT_DATE AND open_stories = 0) AS int) AS completed_sprints,
    CAST(COUNT(*) FILTER (
    WHERE start_date <= CURRENT_DATE
          AND end_date >= CURRENT_DATE
          AND open_stories > 0
          AND (overdue_stories > 0 OR end_date <= CURRENT_DATE + INTERVAL '3 days')
    ) AS int) AS at_risk_sprints,
    CAST(COUNT(*) FILTER (WHERE end_date < CURRENT_DATE AND open_stories > 0) AS int) AS overdue_sprints,
    CAST(COALESCE(SUM(unestimated_stories), 0) AS int) AS unestimated_stories
FROM sprint_scope;

-- name: GetPulseObjectiveHealth :one
SELECT
    CAST(COUNT(*) FILTER (WHERE status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AS int) AS active_objectives,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND objective.health = 'At Risk') AS int) AS at_risk_objectives,
    CAST(COUNT(*) FILTER (WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled')) AND objective.health = 'Off Track') AS int) AS off_track_objectives,
    CAST(COUNT(*) FILTER (
    WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled'))
          AND objective.end_date < CURRENT_DATE
    ) AS int) AS overdue_objectives,
    CAST(COUNT(*) FILTER (
    WHERE (status.category IS NULL OR status.category NOT IN ('completed', 'cancelled'))
          AND objective.end_date >= CURRENT_DATE
          AND objective.end_date <= CURRENT_DATE + INTERVAL '7 days'
    ) AS int) AS objectives_due_soon
FROM objectives AS objective
LEFT JOIN objective_statuses AS status ON status.status_id = objective.status_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR objective.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR objective.lead_user_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR objective.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR objective.created_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR objective.created_at <= sqlc.narg(end_date));

-- name: GetPulseRequestHealth :one
SELECT
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending') AS int) AS pending_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.priority = 'Urgent') AS int) AS urgent_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.priority = 'High') AS int) AS high_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.provider = 'github') AS int) AS github_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.provider = 'slack') AS int) AS slack_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.provider = 'intercom') AS int) AS intercom_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending' AND request.created_at < NOW() - INTERVAL '7 days') AS int) AS stale_requests
FROM integration_requests AS request
WHERE request.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = 0 OR request.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(assignee_ids) AS uuid[])) = 0 OR request.assignee_id = ANY(CAST(sqlc.arg(assignee_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(sprint_ids) AS uuid[])) = 0 OR request.sprint_id = ANY(CAST(sqlc.arg(sprint_ids) AS uuid[])))
  AND (cardinality(CAST(sqlc.arg(objective_ids) AS uuid[])) = 0 OR request.objective_id = ANY(CAST(sqlc.arg(objective_ids) AS uuid[])))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR request.created_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR request.created_at <= sqlc.narg(end_date));
