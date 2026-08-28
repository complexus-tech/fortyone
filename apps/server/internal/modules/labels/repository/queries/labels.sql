-- name: ListLabelsForMember :many
SELECT
    label.label_id,
    label.name,
    label.team_id,
    label.workspace_id,
    label.color,
    label.created_at,
    label.updated_at
FROM public.labels AS label
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = label.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE label.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (
      NOT CAST(sqlc.arg(filter_team) AS boolean)
      OR label.team_id = CAST(sqlc.narg(team_id) AS uuid)
      OR label.team_id IS NULL
  )
  AND (
      NOT CAST(sqlc.arg(filter_search) AS boolean)
      OR label.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
  )
ORDER BY label.created_at DESC, label.label_id DESC;

-- name: ListLabelsPageForMember :many
SELECT
    label.label_id,
    label.name,
    label.team_id,
    label.workspace_id,
    label.color,
    label.created_at,
    label.updated_at
FROM public.labels AS label
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = label.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE label.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND (
      NOT CAST(sqlc.arg(filter_team) AS boolean)
      OR label.team_id = CAST(sqlc.narg(team_id) AS uuid)
      OR label.team_id IS NULL
  )
  AND (
      NOT CAST(sqlc.arg(filter_search) AS boolean)
      OR label.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
  )
ORDER BY label.created_at DESC, label.label_id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer)
OFFSET CAST(sqlc.arg(result_offset) AS integer);

-- name: GetLabelForMember :one
SELECT
    label.label_id,
    label.name,
    label.team_id,
    label.workspace_id,
    label.color,
    label.created_at,
    label.updated_at
FROM public.labels AS label
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = label.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE label.label_id = sqlc.arg(label_id)
  AND label.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid);

-- name: CreateLabelForMember :one
INSERT INTO public.labels (name, team_id, workspace_id, color)
SELECT
    sqlc.arg(name),
    CAST(sqlc.narg(team_id) AS uuid),
    membership.workspace_id,
    CAST(sqlc.arg(color) AS text)
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_id)
  AND (
      CAST(sqlc.narg(team_id) AS uuid) IS NULL
      OR EXISTS (
          SELECT 1
          FROM public.teams AS team
          WHERE team.team_id = CAST(sqlc.narg(team_id) AS uuid)
            AND team.workspace_id = membership.workspace_id
      )
  )
RETURNING label_id, name, team_id, workspace_id, color, created_at, updated_at;

-- name: UpdateLabelForMember :one
UPDATE public.labels AS label
SET name = sqlc.arg(name),
    color = CAST(sqlc.arg(color) AS text),
    updated_at = NOW()
WHERE label.label_id = sqlc.arg(label_id)
  AND label.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS actor
          ON actor.user_id = membership.user_id
         AND actor.is_active = TRUE
      WHERE membership.workspace_id = label.workspace_id
        AND membership.user_id = sqlc.arg(actor_id)
  )
RETURNING label_id, name, team_id, workspace_id, color, created_at, updated_at;

-- name: DeleteLabelForMember :execrows
DELETE FROM public.labels AS label
WHERE label.label_id = sqlc.arg(label_id)
  AND label.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS actor
          ON actor.user_id = membership.user_id
         AND actor.is_active = TRUE
      WHERE membership.workspace_id = label.workspace_id
        AND membership.user_id = sqlc.arg(actor_id)
  );
