-- name: ListMayaSkills :many
SELECT id, name, description, instructions, created_at, updated_at
FROM public.maya_skills
WHERE workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id)
ORDER BY lower(name), id;

-- name: CreateMayaSkill :one
INSERT INTO public.maya_skills (workspace_id, user_id, name, description, instructions)
VALUES (sqlc.arg(workspace_id), sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(description), sqlc.arg(instructions))
RETURNING id, name, description, instructions, created_at, updated_at;

-- name: UpdateMayaSkill :one
UPDATE public.maya_skills
SET name = sqlc.arg(name), description = sqlc.arg(description), instructions = sqlc.arg(instructions),
    updated_at = clock_timestamp()
WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id)
    AND updated_at = sqlc.arg(expected_updated_at)
RETURNING id, name, description, instructions, created_at, updated_at;

-- name: MayaSkillExists :one
SELECT EXISTS (
    SELECT 1 FROM public.maya_skills
    WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id)
);

-- name: DeleteMayaSkill :execrows
DELETE FROM public.maya_skills
WHERE id = sqlc.arg(id) AND workspace_id = sqlc.arg(workspace_id) AND user_id = sqlc.arg(user_id);
