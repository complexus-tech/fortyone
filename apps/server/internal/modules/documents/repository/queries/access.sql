-- name: SetOwnedDocumentVisibility :one
UPDATE public.documents AS document
SET visibility = sqlc.arg(visibility),
    updated_at = CURRENT_TIMESTAMP,
    updated_by = sqlc.arg(actor_id)
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND document.archived_at IS NULL
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
RETURNING document.document_id;

-- name: DeleteOwnedDocumentMembers :exec
DELETE FROM public.document_members AS member
USING public.documents AS document,
      public.workspace_members AS actor_membership,
      public.users AS actor,
      public.workspaces AS workspace
WHERE member.document_id = document.document_id
  AND document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.created_by = sqlc.arg(actor_id)
  AND actor_membership.workspace_id = document.workspace_id
  AND actor_membership.user_id = sqlc.arg(actor_id)
  AND actor_membership.role <> CAST('guest' AS public.user_role)
  AND actor.user_id = actor_membership.user_id
  AND actor.is_active = TRUE
  AND workspace.workspace_id = document.workspace_id
  AND workspace.deleted_at IS NULL;

-- name: InsertActiveWorkspaceDocumentMembers :execrows
WITH requested_member AS (
    SELECT requested_id.user_id, requested_role.role
    FROM unnest(CAST(sqlc.arg(member_ids) AS uuid[]))
        WITH ORDINALITY AS requested_id(user_id, position)
    INNER JOIN unnest(CAST(sqlc.arg(member_roles) AS text[]))
        WITH ORDINALITY AS requested_role(role, position)
        USING (position)
)
INSERT INTO public.document_members (document_id, user_id, role)
SELECT
    document.document_id,
    requested.user_id,
    CASE
        WHEN membership.role = CAST('guest' AS public.user_role) THEN 'viewer'
        ELSE requested.role
    END
FROM requested_member AS requested
INNER JOIN public.documents AS document
    ON document.document_id = sqlc.arg(document_id)
   AND document.workspace_id = sqlc.arg(workspace_id)
   AND document.created_by = sqlc.arg(actor_id)
   AND document.visibility = 'restricted'
   AND document.archived_at IS NULL
INNER JOIN public.workspace_members AS actor_membership
    ON actor_membership.workspace_id = document.workspace_id
   AND actor_membership.user_id = sqlc.arg(actor_id)
   AND actor_membership.role <> CAST('guest' AS public.user_role)
INNER JOIN public.users AS actor ON actor.user_id = actor_membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = requested.user_id
INNER JOIN public.users AS target ON target.user_id = membership.user_id AND target.is_active = TRUE
WHERE requested.user_id <> sqlc.arg(actor_id)
ORDER BY requested.user_id
ON CONFLICT (document_id, user_id) DO UPDATE SET role = EXCLUDED.role;
