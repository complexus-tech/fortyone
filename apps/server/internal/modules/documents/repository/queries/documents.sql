-- name: ListAccessibleDocuments :many
SELECT
    document.document_id,
    document.workspace_id,
    document.title,
    document.visibility,
    document.created_by,
    document.updated_by,
    document.created_at,
    document.updated_at,
    document.archived_at,
    CAST(document.public_token IS NOT NULL AND document.archived_at IS NULL AS boolean) AS public_enabled,
    (
        document.archived_at IS NULL
        AND actor_membership.role <> CAST('guest' AS public.user_role)
        AND (
            document.visibility = 'workspace'
            OR document.created_by = sqlc.arg(actor_id)
            OR EXISTS (
                SELECT 1
                FROM public.document_members AS editor
                WHERE document.visibility = 'restricted'
                  AND editor.document_id = document.document_id
                  AND editor.user_id = sqlc.arg(actor_id)
                  AND editor.role = 'editor'
            )
        )
    ) AS can_edit
FROM public.documents AS document
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS actor_membership
    ON actor_membership.workspace_id = document.workspace_id
   AND actor_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = actor_membership.user_id
   AND actor.is_active = TRUE
WHERE document.workspace_id = sqlc.arg(workspace_id)
  AND (
      (CAST(sqlc.arg(list_scope) AS text) <> 'archived' AND document.archived_at IS NULL)
      OR (
          CAST(sqlc.arg(list_scope) AS text) = 'archived'
          AND document.archived_at IS NOT NULL
          AND document.created_by = sqlc.arg(actor_id)
          AND actor_membership.role <> CAST('guest' AS public.user_role)
      )
  )
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE document.visibility = 'restricted'
            AND reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
  AND (
      CAST(sqlc.arg(search_text) AS text) = ''
      OR document.title ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
      OR document.content_text ILIKE '%' || CAST(sqlc.arg(search_text) AS text) || '%'
  )
  AND (
      CAST(sqlc.arg(list_scope) AS text) IN ('', 'all', 'archived')
      OR (CAST(sqlc.arg(list_scope) AS text) = 'mine' AND document.created_by = sqlc.arg(actor_id))
      OR (
          CAST(sqlc.arg(list_scope) AS text) = 'shared'
          AND document.created_by <> sqlc.arg(actor_id)
          AND EXISTS (
              SELECT 1
              FROM public.document_members AS shared_member
              WHERE document.visibility = 'restricted'
                AND shared_member.document_id = document.document_id
                AND shared_member.user_id = sqlc.arg(actor_id)
          )
      )
  )
ORDER BY document.updated_at DESC, document.document_id
LIMIT CAST(sqlc.narg(row_limit) AS integer) OFFSET CAST(sqlc.arg(row_offset) AS integer);

-- name: GetAccessibleDocument :one
SELECT
    document.document_id,
    document.workspace_id,
    document.title,
    document.content_html,
    document.content_text,
    document.visibility,
    document.created_by,
    document.updated_by,
    document.created_at,
    document.updated_at,
    document.archived_at,
    (
        actor_membership.role <> CAST('guest' AS public.user_role)
        AND (
            document.visibility = 'workspace'
            OR document.created_by = sqlc.arg(actor_id)
            OR EXISTS (
                SELECT 1
                FROM public.document_members AS editor
                WHERE document.visibility = 'restricted'
                  AND editor.document_id = document.document_id
                  AND editor.user_id = sqlc.arg(actor_id)
                  AND editor.role = 'editor'
            )
        )
    ) AS can_edit
FROM public.documents AS document
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS actor_membership
    ON actor_membership.workspace_id = document.workspace_id
   AND actor_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = actor_membership.user_id
   AND actor.is_active = TRUE
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE document.visibility = 'restricted'
            AND reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  );

-- name: ListAccessibleDocumentMembers :many
SELECT member.user_id, member.role
FROM public.document_members AS member
INNER JOIN public.documents AS document ON document.document_id = member.document_id
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = document.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS actor_membership
    ON actor_membership.workspace_id = document.workspace_id
   AND actor_membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = actor_membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.workspace_members AS target_membership
    ON target_membership.workspace_id = document.workspace_id
   AND target_membership.user_id = member.user_id
INNER JOIN public.users AS target
    ON target.user_id = target_membership.user_id
   AND target.is_active = TRUE
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.visibility = 'restricted'
  AND document.archived_at IS NULL
  AND (
      document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
ORDER BY member.created_at, member.user_id;

-- name: CreateDocument :one
INSERT INTO public.documents (
    workspace_id,
    title,
    content_html,
    content_text,
    visibility,
    created_by,
    updated_by
)
SELECT
    membership.workspace_id,
    sqlc.arg(title),
    sqlc.arg(content_html),
    sqlc.arg(content_text),
    sqlc.arg(visibility),
    membership.user_id,
    membership.user_id
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
RETURNING
    document_id,
    workspace_id,
    title,
    content_html,
    content_text,
    visibility,
    created_by,
    updated_by,
    created_at,
    updated_at,
    archived_at,
    TRUE AS can_edit;

-- name: UpdateEditableDocument :one
UPDATE public.documents AS document
SET title = COALESCE(sqlc.narg(title), document.title),
    content_html = COALESCE(sqlc.narg(content_html), document.content_html),
    content_text = COALESCE(sqlc.narg(content_text), document.content_text),
    updated_by = sqlc.arg(actor_id),
    updated_at = CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND document.collaboration_state IS NULL
  AND document.revision = sqlc.arg(expected_revision)
  AND membership.workspace_id = document.workspace_id
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS editor
          WHERE document.visibility = 'restricted'
            AND editor.document_id = document.document_id
            AND editor.user_id = sqlc.arg(actor_id)
            AND editor.role = 'editor'
      )
  )
RETURNING
    document.document_id,
    document.workspace_id,
    document.title,
    document.content_html,
    document.content_text,
    document.visibility,
    document.created_by,
    document.updated_by,
    document.created_at,
    document.updated_at,
    document.archived_at,
    TRUE AS can_edit;

-- name: CreateDocumentWithID :one
INSERT INTO public.documents (
    document_id,
    workspace_id,
    title,
    content_html,
    content_text,
    visibility,
    created_by,
    updated_by
)
SELECT
    sqlc.arg(document_id),
    membership.workspace_id,
    sqlc.arg(title),
    sqlc.arg(content_html),
    sqlc.arg(content_text),
    sqlc.arg(visibility),
    membership.user_id,
    membership.user_id
FROM public.workspace_members AS membership
INNER JOIN public.users AS actor ON actor.user_id = membership.user_id AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(actor_id)
  AND membership.role <> CAST('guest' AS public.user_role)
RETURNING
    document_id,
    workspace_id,
    title,
    content_html,
    content_text,
    visibility,
    created_by,
    updated_by,
    created_at,
    updated_at,
    archived_at,
    TRUE AS can_edit;
