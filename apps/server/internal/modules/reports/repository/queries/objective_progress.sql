-- name: ListObjectiveHealthDistribution :many
SELECT
    CAST(COALESCE(CAST(objective.health AS text), 'Not Set') AS text) AS status,
    CAST(COUNT(objective.objective_id) AS int) AS count
FROM objectives AS objective
WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
  AND objective.created_at >= sqlc.arg(start_date)
  AND objective.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR objective.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR objective.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY CAST(COALESCE(CAST(objective.health AS text), 'Not Set') AS text)
ORDER BY CASE CAST(COALESCE(CAST(objective.health AS text), 'Not Set') AS text)
    WHEN 'On Track' THEN 1
    WHEN 'At Risk' THEN 2
    WHEN 'Off Track' THEN 3
    WHEN 'Not Set' THEN 4
    ELSE 5
END;

-- name: ListObjectiveStatusBreakdown :many
SELECT
    status.name AS status_name,
    CAST(COUNT(objective.objective_id) AS int) AS count
FROM objectives AS objective
INNER JOIN objective_statuses AS status ON status.status_id = objective.status_id
WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
  AND objective.created_at >= sqlc.arg(start_date)
  AND objective.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR objective.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR objective.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY status.status_id, status.name, status.order_index
ORDER BY status.order_index;

-- name: ListKeyResultProgress :many
SELECT
    objective.objective_id,
    objective.name AS objective_name,
    CAST(COUNT(key_result.id) AS int) AS total,
    CAST(COUNT(key_result.id) FILTER (
        WHERE (key_result.measurement_type = 'percentage' AND key_result.current_value >= key_result.target_value)
           OR (key_result.measurement_type = 'number' AND key_result.current_value >= key_result.target_value)
           OR (key_result.measurement_type = 'boolean' AND key_result.current_value = key_result.target_value)
    ) AS int) AS completed,
    CAST(COALESCE(AVG(
        CASE
            WHEN key_result.measurement_type = 'percentage'
                THEN LEAST(key_result.current_value, 100)
            WHEN key_result.measurement_type = 'number' AND key_result.target_value != key_result.start_value
                THEN LEAST(
                    ((key_result.current_value - key_result.start_value)
                        / NULLIF(key_result.target_value - key_result.start_value, 0)) * 100,
                    100
                )
            WHEN key_result.measurement_type = 'boolean'
                THEN CASE WHEN key_result.current_value = key_result.target_value THEN 100 ELSE 0 END
            ELSE 0
        END
    ), 0) AS double precision) AS avg_progress
FROM objectives AS objective
LEFT JOIN key_results AS key_result ON key_result.objective_id = objective.objective_id
WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
  AND objective.created_at >= sqlc.arg(start_date)
  AND objective.created_at <= sqlc.arg(end_date)
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR objective.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR objective.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
GROUP BY objective.objective_id, objective.name
ORDER BY objective.name, objective.objective_id;

-- name: ListObjectiveProgressByTeam :many
SELECT
    team.team_id,
    team.name AS team_name,
    CAST(COUNT(objective.objective_id) AS int) AS objectives,
    CAST(COUNT(objective.objective_id) FILTER (WHERE status.category = 'completed') AS int) AS completed
FROM teams AS team
LEFT JOIN objectives AS objective
    ON objective.team_id = team.team_id
   AND objective.workspace_id = sqlc.arg(workspace_id)::uuid
   AND objective.created_at >= sqlc.arg(start_date)
   AND objective.created_at <= sqlc.arg(end_date)
   AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR objective.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
LEFT JOIN objective_statuses AS status ON status.status_id = objective.status_id
WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR team.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
GROUP BY team.team_id, team.name
ORDER BY team.name, team.team_id;
