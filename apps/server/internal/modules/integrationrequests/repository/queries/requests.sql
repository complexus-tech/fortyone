-- name: UpsertPendingIntegrationRequest :one
INSERT INTO public.integration_requests (
    workspace_id,
    team_id,
    provider,
    source_type,
    source_external_id,
    source_number,
    source_url,
    title,
    description,
    status_id,
    priority,
    assignee_id,
    estimate_unit,
    estimated_duration_minutes,
    minimum_focus_block_minutes,
    objective_id,
    key_result_id,
    sprint_id,
    start_date,
    end_date,
    label_ids,
    metadata,
    created_by_user_id
) VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(team_id),
    sqlc.arg(provider),
    sqlc.arg(source_type),
    sqlc.arg(source_external_id),
    sqlc.narg(source_number),
    sqlc.narg(source_url),
    sqlc.arg(title),
    sqlc.narg(description),
    sqlc.narg(status_id),
    COALESCE(NULLIF(CAST(sqlc.arg(priority) AS text), ''), 'No Priority'),
    sqlc.narg(assignee_id),
    sqlc.narg(estimate_unit),
    sqlc.narg(estimated_duration_minutes),
    sqlc.narg(minimum_focus_block_minutes),
    sqlc.narg(objective_id),
    sqlc.narg(key_result_id),
    sqlc.narg(sprint_id),
    sqlc.narg(start_date),
    sqlc.narg(end_date),
    CAST(sqlc.arg(label_ids) AS uuid[]),
    CAST(sqlc.arg(metadata) AS jsonb),
    sqlc.narg(created_by_user_id)
)
ON CONFLICT (workspace_id, provider, source_type, source_external_id)
DO UPDATE SET
    team_id = EXCLUDED.team_id,
    source_number = EXCLUDED.source_number,
    source_url = EXCLUDED.source_url,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    status_id = COALESCE(integration_requests.status_id, EXCLUDED.status_id),
    priority = EXCLUDED.priority,
    assignee_id = COALESCE(integration_requests.assignee_id, EXCLUDED.assignee_id),
    estimate_unit = COALESCE(integration_requests.estimate_unit, EXCLUDED.estimate_unit),
    estimated_duration_minutes = COALESCE(integration_requests.estimated_duration_minutes, EXCLUDED.estimated_duration_minutes),
    minimum_focus_block_minutes = COALESCE(integration_requests.minimum_focus_block_minutes, EXCLUDED.minimum_focus_block_minutes),
    objective_id = COALESCE(integration_requests.objective_id, EXCLUDED.objective_id),
    key_result_id = COALESCE(integration_requests.key_result_id, EXCLUDED.key_result_id),
    sprint_id = COALESCE(integration_requests.sprint_id, EXCLUDED.sprint_id),
    start_date = COALESCE(integration_requests.start_date, EXCLUDED.start_date),
    end_date = COALESCE(integration_requests.end_date, EXCLUDED.end_date),
    label_ids = CASE
        WHEN cardinality(EXCLUDED.label_ids) > 0 THEN EXCLUDED.label_ids
        ELSE integration_requests.label_ids
    END,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
WHERE integration_requests.status = 'pending'
  AND integration_requests.acceptance_state = 'idle'
RETURNING *;

-- name: AuthorizeIntegrationRequestTeam :one
SELECT TRUE AS authorized
FROM public.teams AS team
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND team.team_id = sqlc.arg(team_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = team.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = team.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  );

-- name: ListIntegrationRequestsByTeam :many
SELECT *
FROM public.integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.team_id = sqlc.arg(team_id)
  AND request.status = sqlc.arg(request_status)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  )
  AND (
      NOT CAST(sqlc.arg(has_search) AS boolean)
      OR request.title ILIKE sqlc.arg(search_pattern)
      OR COALESCE(request.description, '') ILIKE sqlc.arg(search_pattern)
      OR request.source_external_id ILIKE sqlc.arg(search_pattern)
      OR COALESCE(CAST(request.source_number AS text), '') ILIKE sqlc.arg(search_pattern)
  )
  AND (NOT CAST(sqlc.arg(has_provider) AS boolean) OR request.provider = sqlc.arg(provider))
  AND (NOT CAST(sqlc.arg(has_priority) AS boolean) OR request.priority = sqlc.arg(priority))
  AND (NOT CAST(sqlc.arg(has_assignee) AS boolean) OR request.assignee_id = sqlc.narg(assignee_id))
  AND (NOT CAST(sqlc.arg(has_created_after) AS boolean) OR request.created_at >= sqlc.narg(created_after))
  AND (NOT CAST(sqlc.arg(has_created_before) AS boolean) OR request.created_at <= sqlc.narg(created_before))
ORDER BY request.created_at DESC
LIMIT CASE WHEN CAST(sqlc.arg(paginated) AS boolean) THEN CAST(sqlc.arg(row_limit) AS integer) ELSE NULL END
OFFSET CASE WHEN CAST(sqlc.arg(paginated) AS boolean) THEN CAST(sqlc.arg(row_offset) AS integer) ELSE 0 END;

-- name: CountIntegrationRequestsByTeam :one
SELECT COUNT(*)
FROM public.integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.team_id = sqlc.arg(team_id)
  AND request.status = sqlc.arg(request_status)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  )
  AND (
      NOT CAST(sqlc.arg(has_search) AS boolean)
      OR request.title ILIKE sqlc.arg(search_pattern)
      OR COALESCE(request.description, '') ILIKE sqlc.arg(search_pattern)
      OR request.source_external_id ILIKE sqlc.arg(search_pattern)
      OR COALESCE(CAST(request.source_number AS text), '') ILIKE sqlc.arg(search_pattern)
  )
  AND (NOT CAST(sqlc.arg(has_provider) AS boolean) OR request.provider = sqlc.arg(provider))
  AND (NOT CAST(sqlc.arg(has_priority) AS boolean) OR request.priority = sqlc.arg(priority))
  AND (NOT CAST(sqlc.arg(has_assignee) AS boolean) OR request.assignee_id = sqlc.narg(assignee_id))
  AND (NOT CAST(sqlc.arg(has_created_after) AS boolean) OR request.created_at >= sqlc.narg(created_after))
  AND (NOT CAST(sqlc.arg(has_created_before) AS boolean) OR request.created_at <= sqlc.narg(created_before));

-- name: GetIntegrationRequestForUser :one
SELECT *
FROM public.integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND (
      EXISTS (
          SELECT 1
          FROM public.team_members AS request_team_member
          WHERE request_team_member.team_id = request.team_id
            AND request_team_member.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM public.workspace_members AS request_workspace_member
          WHERE request_workspace_member.workspace_id = request.workspace_id
            AND request_workspace_member.user_id = sqlc.arg(actor_id)
            AND request_workspace_member.role = 'admin'
      )
  );

-- name: GetIntegrationRequest :one
SELECT *
FROM public.integration_requests
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(request_id);

-- name: GetIntegrationRequestByExternal :one
SELECT *
FROM public.integration_requests
WHERE workspace_id = sqlc.arg(workspace_id)
  AND provider = sqlc.arg(provider)
  AND source_type = sqlc.arg(source_type)
  AND source_external_id = sqlc.arg(source_external_id);

-- name: FindFirstIntegrationRequestStatusByCategory :one
SELECT status_id
FROM public.statuses
WHERE team_id = sqlc.arg(team_id)
  AND category = sqlc.arg(category)
ORDER BY order_index ASC
LIMIT 1;
