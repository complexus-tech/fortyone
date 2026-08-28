-- name: CreateUserMemoryForMember :one
INSERT INTO public.user_memories (
    workspace_id,
    user_id,
    content,
    created_at,
    updated_at
)
SELECT
    membership.workspace_id,
    membership.user_id,
    CAST(sqlc.arg(content) AS text),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
RETURNING
    id,
    workspace_id,
    user_id,
    content,
    created_at,
    updated_at;

-- name: UpdateUserMemoryForOwner :execrows
UPDATE public.user_memories AS memory
SET
    content = CAST(sqlc.arg(content) AS text),
    updated_at = CURRENT_TIMESTAMP
WHERE memory.id = sqlc.arg(memory_id)
  AND memory.user_id = sqlc.arg(user_id)
  AND memory.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account ON account.user_id = membership.user_id
      WHERE membership.workspace_id = memory.workspace_id
        AND membership.user_id = memory.user_id
        AND account.is_active = TRUE
  );

-- name: DeleteUserMemoryForOwner :execrows
DELETE FROM public.user_memories AS memory
WHERE memory.id = sqlc.arg(memory_id)
  AND memory.user_id = sqlc.arg(user_id)
  AND memory.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS membership
      INNER JOIN public.users AS account ON account.user_id = membership.user_id
      WHERE membership.workspace_id = memory.workspace_id
        AND membership.user_id = memory.user_id
        AND account.is_active = TRUE
  );

-- name: ListUserMemoriesForOwner :many
SELECT
    memory.id,
    memory.workspace_id,
    memory.user_id,
    memory.content,
    memory.created_at,
    memory.updated_at
FROM public.user_memories AS memory
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = memory.workspace_id
   AND membership.user_id = memory.user_id
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE memory.user_id = sqlc.arg(user_id)
  AND memory.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
ORDER BY memory.created_at DESC, memory.id DESC;
