-- name: ListPresetsForActor :many
SELECT preset.id, preset.team_id, preset.owner_id, preset.kind, preset.visibility, preset.name, preset.configuration, preset.created_at, preset.updated_at, CAST(COALESCE(CAST(EXISTS (SELECT 1 FROM public.workspace_members AS editor WHERE editor.workspace_id = preset.workspace_id AND editor.user_id = sqlc.arg(actor_id) AND editor.role IN ('admin', 'member') AND (preset.owner_id = editor.user_id OR editor.role = 'admin')) AS boolean), FALSE) AS boolean) AS can_edit
FROM public.work_presets AS preset
WHERE preset.workspace_id = sqlc.arg(workspace_id)
  AND preset.team_id = sqlc.arg(team_id)
  AND preset.kind = sqlc.arg(kind)
  AND preset.archived_at IS NULL
  AND (preset.visibility = 'team' OR preset.owner_id = sqlc.arg(actor_id))
  AND EXISTS (
    SELECT 1 FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
    INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = membership.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.teams AS team ON team.team_id = preset.team_id AND team.workspace_id = membership.workspace_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND (membership.role = 'admin' OR EXISTS (
          SELECT 1 FROM public.team_members AS team_member
          WHERE team_member.team_id = team.team_id AND team_member.user_id = membership.user_id
      ))
)
  AND (CAST(sqlc.narg(before_id) AS uuid) IS NULL OR (preset.created_at, preset.id) < (CAST(sqlc.narg(before_created_at) AS timestamptz), CAST(sqlc.narg(before_id) AS uuid)))
ORDER BY preset.created_at DESC, preset.id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: CreatePresetForActor :one
INSERT INTO public.work_presets AS preset (workspace_id, team_id, owner_id, kind, visibility, name, configuration)
SELECT membership.workspace_id, team.team_id, membership.user_id, sqlc.arg(kind), sqlc.arg(visibility), sqlc.arg(name), sqlc.arg(configuration)
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = membership.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN public.teams AS team ON team.team_id = sqlc.arg(team_id) AND team.workspace_id = membership.workspace_id
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role IN ('admin', 'member')
  AND (membership.role = 'admin' OR EXISTS (
      SELECT 1 FROM public.team_members AS team_member
      WHERE team_member.team_id = team.team_id AND team_member.user_id = membership.user_id
  ))
RETURNING preset.id, preset.team_id, preset.owner_id, preset.kind, preset.visibility, preset.name, preset.configuration, preset.created_at, preset.updated_at, TRUE AS can_edit;

-- name: RenamePresetForActor :one
UPDATE public.work_presets AS preset
SET name = sqlc.arg(name), updated_at = now()
WHERE preset.id = sqlc.arg(id) AND preset.workspace_id = sqlc.arg(workspace_id)
  AND preset.archived_at IS NULL
  AND (preset.visibility = 'team' OR preset.owner_id = sqlc.arg(actor_id))
  AND (preset.owner_id = sqlc.arg(actor_id) OR EXISTS (SELECT 1 FROM public.workspace_members AS editor WHERE editor.workspace_id = preset.workspace_id AND editor.user_id = sqlc.arg(actor_id) AND editor.role = 'admin'))
  AND EXISTS (
    SELECT 1 FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
    INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = membership.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.teams AS team ON team.team_id = preset.team_id AND team.workspace_id = membership.workspace_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND membership.role IN ('admin', 'member')
      AND (membership.role = 'admin' OR EXISTS (
          SELECT 1 FROM public.team_members AS team_member
          WHERE team_member.team_id = team.team_id AND team_member.user_id = membership.user_id
      ))
)
RETURNING preset.id, preset.team_id, preset.owner_id, preset.kind, preset.visibility, preset.name, preset.configuration, preset.created_at, preset.updated_at, TRUE AS can_edit;

-- name: ArchivePresetForActor :execrows
UPDATE public.work_presets AS preset
SET archived_at = now(), updated_at = now()
WHERE preset.id = sqlc.arg(id) AND preset.workspace_id = sqlc.arg(workspace_id)
  AND preset.archived_at IS NULL
  AND (preset.visibility = 'team' OR preset.owner_id = sqlc.arg(actor_id))
  AND (preset.owner_id = sqlc.arg(actor_id) OR EXISTS (SELECT 1 FROM public.workspace_members AS editor WHERE editor.workspace_id = preset.workspace_id AND editor.user_id = sqlc.arg(actor_id) AND editor.role = 'admin'))
  AND EXISTS (
    SELECT 1 FROM public.workspace_members AS membership
    INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
    INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = membership.workspace_id AND workspace.deleted_at IS NULL
    INNER JOIN public.teams AS team ON team.team_id = preset.team_id AND team.workspace_id = membership.workspace_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(actor_id)
      AND membership.role IN ('admin', 'member')
      AND (membership.role = 'admin' OR EXISTS (
          SELECT 1 FROM public.team_members AS team_member
          WHERE team_member.team_id = team.team_id AND team_member.user_id = membership.user_id
      ))
);
