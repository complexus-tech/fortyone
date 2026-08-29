-- name: ListRequestSourcePerformance :many
SELECT
    request.provider,
    CAST(COUNT(*) AS int) AS total_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'pending') AS int) AS pending_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'accepted') AS int) AS accepted_requests,
    CAST(COUNT(*) FILTER (WHERE request.status = 'declined') AS int) AS declined_requests,
    CAST(COUNT(*) FILTER (WHERE request.priority = 'Urgent') AS int) AS urgent_requests,
    CAST(COUNT(*) FILTER (WHERE request.priority = 'High') AS int) AS high_requests,
    CAST(COUNT(*) FILTER (
        WHERE request.status = 'pending'
          AND request.created_at < NOW() - INTERVAL '7 days'
    ) AS int) AS stale_requests,
    CAST(COALESCE(
        CAST(COUNT(*) FILTER (WHERE request.status = 'accepted') AS double precision)
            / NULLIF(COUNT(*), 0),
        0
    ) AS double precision) AS acceptance_rate
FROM integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)::uuid
  AND (cardinality(sqlc.arg(team_ids)::uuid[]) = 0 OR request.team_id = ANY(sqlc.arg(team_ids)::uuid[]))
  AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR request.assignee_id = ANY(sqlc.arg(assignee_ids)::uuid[]))
  AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR request.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
  AND (cardinality(sqlc.arg(objective_ids)::uuid[]) = 0 OR request.objective_id = ANY(sqlc.arg(objective_ids)::uuid[]))
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR request.created_at >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR request.created_at <= sqlc.narg(end_date))
GROUP BY request.provider
ORDER BY total_requests DESC, request.provider ASC;

