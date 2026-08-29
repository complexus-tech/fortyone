-- name: ListObjectiveStatuses :many
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.objective_statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
ORDER BY status.order_index ASC, status.status_id ASC;

-- name: ListObjectiveStatusesForMember :many
SELECT
    status.status_id,
    status.name,
    status.category,
    status.order_index,
    status.workspace_id,
    status.is_default,
    status.color,
    status.created_at,
    status.updated_at
FROM public.objective_statuses AS status
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = status.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE status.workspace_id = sqlc.arg(workspace_id)
ORDER BY status.order_index ASC, status.status_id ASC;

-- name: ObjectiveStatusAdminAuthorized :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND membership.role = 'admin'
);

-- name: LockObjectiveStatusOrdering :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'objective-status-order:' || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text) || ':' || CAST(sqlc.arg(category) AS text),
        0
    )
);

-- name: LockObjectiveStatusDefaults :exec
SELECT pg_advisory_xact_lock(
    hashtextextended('objective-status-default:' || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text), 0)
);

-- name: ResetObjectiveStatusDefaults :exec
UPDATE public.objective_statuses
SET is_default = FALSE,
    updated_at = NOW()
WHERE workspace_id = sqlc.arg(workspace_id)
  AND is_default = TRUE;

-- name: NextObjectiveStatusOrderIndex :one
SELECT CAST(COALESCE(MAX(status.order_index), -1) + 10 AS integer)
FROM public.objective_statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.category = CAST(sqlc.arg(category) AS text);

-- name: InsertObjectiveStatus :one
INSERT INTO public.objective_statuses (
    name, category, order_index, workspace_id, is_default, color
) VALUES (
    sqlc.arg(name), CAST(sqlc.arg(category) AS text), CAST(sqlc.arg(order_index) AS integer),
    sqlc.arg(workspace_id), sqlc.arg(is_default), CAST(sqlc.arg(color) AS text)
)
RETURNING status_id, name, category, order_index, workspace_id,
          is_default, color, created_at, updated_at;

-- name: ObjectiveStatusExistsForAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM public.objective_statuses AS status
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = status.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
       AND membership.role = 'admin'
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE status.status_id = sqlc.arg(status_id)
      AND status.workspace_id = sqlc.arg(workspace_id)
);

-- name: UpdateObjectiveStatusForAdmin :one
UPDATE public.objective_statuses AS status
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
        AND membership.role = 'admin'
  )
RETURNING status_id, name, category, order_index, workspace_id,
          is_default, color, created_at, updated_at;

-- name: GetObjectiveStatusForDelete :one
SELECT status.category
FROM public.objective_statuses AS status
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = status.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
   AND membership.role = 'admin'
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE status.status_id = sqlc.arg(status_id)
  AND status.workspace_id = sqlc.arg(workspace_id)
FOR UPDATE OF status;

-- name: LockObjectiveStatusCategory :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'objective-status-delete:' || CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text) || ':' || CAST(sqlc.arg(category) AS text),
        0
    )
);

-- name: CountObjectivesWithStatus :one
SELECT CAST(COUNT(*) AS integer)
FROM public.objectives AS objective
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective.status_id = CAST(sqlc.arg(status_id) AS uuid);

-- name: CountObjectiveStatusesInCategory :one
SELECT CAST(COUNT(*) AS integer)
FROM public.objective_statuses AS status
WHERE status.workspace_id = sqlc.arg(workspace_id)
  AND status.category = CAST(sqlc.arg(category) AS text);

-- name: DeleteObjectiveStatus :execrows
DELETE FROM public.objective_statuses
WHERE status_id = sqlc.arg(status_id)
  AND workspace_id = sqlc.arg(workspace_id);
