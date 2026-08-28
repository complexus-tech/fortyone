-- name: LockPendingIntegrationRequestForUpdate :one
SELECT *
FROM public.integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND request.status = 'pending'
  AND request.acceptance_state = 'idle'
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
FOR UPDATE;

-- name: UpdatePendingIntegrationRequest :one
UPDATE public.integration_requests AS request
SET
    title = sqlc.arg(title),
    description = sqlc.narg(description),
    status_id = sqlc.narg(status_id),
    priority = sqlc.arg(priority),
    assignee_id = sqlc.narg(assignee_id),
    estimate_unit = sqlc.narg(estimate_unit),
    estimated_duration_minutes = sqlc.narg(estimated_duration_minutes),
    minimum_focus_block_minutes = sqlc.narg(minimum_focus_block_minutes),
    objective_id = sqlc.narg(objective_id),
    key_result_id = sqlc.narg(key_result_id),
    sprint_id = sqlc.narg(sprint_id),
    start_date = sqlc.narg(start_date),
    end_date = sqlc.narg(end_date),
    label_ids = CAST(sqlc.arg(label_ids) AS uuid[]),
    updated_at = NOW()
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND request.status = 'pending'
  AND request.acceptance_state = 'idle'
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
RETURNING *;

-- name: LockPendingIntegrationRequestForAcceptance :one
SELECT *
FROM public.integration_requests AS request
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND request.status = 'pending'
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
FOR UPDATE;

-- name: ReserveIntegrationRequestAcceptance :one
UPDATE public.integration_requests AS request
SET
    acceptance_state = 'reserved',
    acceptance_started_by_user_id = sqlc.arg(actor_id),
    acceptance_started_at = NOW(),
    updated_at = NOW()
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND request.status = 'pending'
  AND request.acceptance_state = 'idle'
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
RETURNING *;

-- name: MarkIntegrationRequestAccepted :one
UPDATE public.integration_requests
SET
    status = 'accepted',
    accepted_story_id = sqlc.arg(story_id),
    accepted_by_user_id = sqlc.arg(actor_id),
    accepted_at = NOW(),
    acceptance_state = 'idle',
    acceptance_started_by_user_id = NULL,
    acceptance_started_at = NULL,
    updated_at = NOW()
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(request_id)
  AND status = 'pending'
  AND acceptance_state = 'reserved'
  AND acceptance_started_by_user_id = sqlc.arg(actor_id)
RETURNING *;

-- name: MarkIntegrationRequestDeclined :one
UPDATE public.integration_requests AS request
SET
    status = 'declined',
    declined_by_user_id = sqlc.arg(actor_id),
    declined_at = NOW(),
    updated_at = NOW()
WHERE request.workspace_id = sqlc.arg(workspace_id)
  AND request.id = sqlc.arg(request_id)
  AND request.status = 'pending'
  AND request.acceptance_state = 'idle'
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
RETURNING *;
