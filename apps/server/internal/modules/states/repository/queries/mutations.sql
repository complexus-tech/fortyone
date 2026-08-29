-- name: StateTeamExistsForMember :one
SELECT EXISTS (
    SELECT 1
    FROM public.teams AS team
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = team.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE team.team_id = sqlc.arg(team_id)
      AND team.workspace_id = sqlc.arg(workspace_id)
);

-- name: LockStateOrdering :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'states-order:' || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text) || ':' || CAST(sqlc.arg(category) AS text),
        0
    )
);

-- name: LockStateDefaults :exec
SELECT pg_advisory_xact_lock(
    hashtextextended('states-default:' || CAST(CAST(sqlc.arg(team_id) AS uuid) AS text), 0)
);

-- name: ResetStateDefaults :exec
UPDATE public.statuses
SET is_default = FALSE,
    updated_at = NOW()
WHERE workspace_id = sqlc.arg(workspace_id)
  AND team_id = sqlc.arg(team_id)
  AND is_default = TRUE;

-- name: NextStateOrderIndex :one
SELECT CAST(COALESCE(MAX(status.order_index), -1) + 10 AS integer)
FROM public.statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.category = CAST(sqlc.arg(category) AS text);

-- name: InsertState :one
INSERT INTO public.statuses (
    name, category, order_index, color, team_id, workspace_id, is_default
) VALUES (
    sqlc.arg(name), CAST(sqlc.arg(category) AS text), CAST(sqlc.arg(order_index) AS integer),
    CAST(sqlc.arg(color) AS text),
    sqlc.arg(team_id), sqlc.arg(workspace_id), sqlc.arg(is_default)
)
RETURNING status_id, name, category, order_index, team_id, workspace_id,
          is_default, color, created_at, updated_at;

-- name: GetStateTeamForMember :one
SELECT status.team_id
FROM public.statuses AS status
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = status.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE status.status_id = sqlc.arg(status_id)
  AND status.workspace_id = sqlc.arg(workspace_id);

-- name: UpdateStateForMember :one
UPDATE public.statuses AS status
SET name = CASE
        WHEN CAST(sqlc.arg(set_name) AS boolean) THEN CAST(sqlc.arg(name) AS text)
        ELSE status.name
    END,
    order_index = CASE
        WHEN CAST(sqlc.arg(set_order_index) AS boolean) THEN CAST(sqlc.arg(order_index) AS integer)
        ELSE status.order_index
    END,
    is_default = CASE
        WHEN CAST(sqlc.arg(set_is_default) AS boolean) THEN CAST(sqlc.arg(is_default) AS boolean)
        ELSE status.is_default
    END,
    color = CASE
        WHEN CAST(sqlc.arg(set_color) AS boolean) THEN CAST(sqlc.arg(color) AS text)
        ELSE status.color
    END,
    updated_at = NOW()
WHERE status.status_id = sqlc.arg(status_id)
  AND status.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS actor
          ON actor.user_id = membership.user_id
         AND actor.is_active = TRUE
      WHERE membership.workspace_id = status.workspace_id
        AND membership.user_id = sqlc.arg(actor_id)
  )
RETURNING status_id, name, category, order_index, team_id, workspace_id,
          is_default, color, created_at, updated_at;

-- name: GetStateForDelete :one
SELECT status.team_id, status.category
FROM public.statuses AS status
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = status.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE status.status_id = sqlc.arg(status_id)
  AND status.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF status;

-- name: LockStateCategory :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'states-delete:' || CAST(CAST(sqlc.arg(team_id) AS uuid) AS text) || ':' || CAST(sqlc.arg(category) AS text),
        0
    )
);

-- name: CountWorkspaceStoriesWithState :one
SELECT CAST(COUNT(*) AS integer)
FROM public.stories AS story
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.status_id = CAST(sqlc.arg(status_id) AS uuid)
  AND story.deleted_at IS NULL;

-- name: CountTeamStatesInCategory :one
SELECT CAST(COUNT(*) AS integer)
FROM public.statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.team_id = sqlc.arg(team_id)
  AND status.category = CAST(sqlc.arg(category) AS text);

-- name: DeleteState :execrows
DELETE FROM public.statuses
WHERE status_id = sqlc.arg(status_id)
  AND workspace_id = sqlc.arg(workspace_id);
