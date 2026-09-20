-- name: ListDocumentComments :many
SELECT
    thread.thread_id,
    thread.document_id,
    thread.quote,
    thread.anchor_start,
    thread.anchor_end,
    thread.created_by AS thread_created_by,
    thread.resolved_at,
    thread.resolved_by,
    thread.created_at AS thread_created_at,
    comment.comment_id,
    comment.body,
    comment.created_by,
    comment.created_at,
    comment.updated_at,
    COALESCE(author.full_name, author.username, 'Former user') AS author_name,
    author.avatar_url AS author_avatar
FROM public.document_comment_threads AS thread
INNER JOIN public.documents AS document
    ON document.document_id = thread.document_id
   AND document.workspace_id = thread.workspace_id
   AND document.archived_at IS NULL
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.document_comments AS comment ON comment.thread_id = thread.thread_id
INNER JOIN public.users AS author ON author.user_id = comment.created_by
WHERE document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
ORDER BY thread.created_at DESC, comment.created_at, comment.comment_id;

-- name: CreateDocumentCommentThread :one
WITH accessible_document AS (
    SELECT document.document_id, document.workspace_id
    FROM public.documents AS document
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = document.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
       AND membership.role <> CAST('guest' AS public.user_role)
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
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
              WHERE reader.document_id = document.document_id
                AND reader.user_id = sqlc.arg(actor_id)
          )
      )
), inserted_thread AS (
    INSERT INTO public.document_comment_threads (
        document_id, workspace_id, quote, anchor_start, anchor_end, created_by
    )
    SELECT
        document_id,
        workspace_id,
        sqlc.arg(quote),
        sqlc.arg(anchor_start),
        sqlc.arg(anchor_end),
        sqlc.arg(actor_id)
    FROM accessible_document
    RETURNING *
), inserted_comment AS (
    INSERT INTO public.document_comments (thread_id, body, created_by)
    SELECT thread_id, sqlc.arg(body), sqlc.arg(actor_id)
    FROM inserted_thread
    RETURNING *
)
SELECT
    thread.thread_id,
    thread.document_id,
    thread.quote,
    thread.anchor_start,
    thread.anchor_end,
    thread.created_by AS thread_created_by,
    thread.resolved_at,
    thread.resolved_by,
    thread.created_at AS thread_created_at,
    comment.comment_id,
    comment.body,
    comment.created_by,
    comment.created_at,
    comment.updated_at,
    COALESCE(author.full_name, author.username, 'Former user') AS author_name,
    author.avatar_url AS author_avatar
FROM inserted_thread AS thread
INNER JOIN inserted_comment AS comment ON comment.thread_id = thread.thread_id
INNER JOIN public.users AS author ON author.user_id = comment.created_by;

-- name: AddDocumentCommentReply :one
WITH accessible_thread AS (
    SELECT thread.thread_id
    FROM public.document_comment_threads AS thread
    INNER JOIN public.documents AS document
        ON document.document_id = thread.document_id
       AND document.workspace_id = thread.workspace_id
       AND document.archived_at IS NULL
    INNER JOIN public.workspace_members AS membership
        ON membership.workspace_id = document.workspace_id
       AND membership.user_id = sqlc.arg(actor_id)
       AND membership.role <> CAST('guest' AS public.user_role)
    INNER JOIN public.users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
    WHERE thread.thread_id = sqlc.arg(thread_id)
      AND document.document_id = sqlc.arg(document_id)
      AND document.workspace_id = sqlc.arg(workspace_id)
      AND (
          document.visibility = 'workspace'
          OR document.created_by = sqlc.arg(actor_id)
          OR EXISTS (
              SELECT 1
              FROM public.document_members AS reader
              WHERE reader.document_id = document.document_id
                AND reader.user_id = sqlc.arg(actor_id)
          )
      )
), inserted_comment AS (
    INSERT INTO public.document_comments (thread_id, body, created_by)
    SELECT thread_id, sqlc.arg(body), sqlc.arg(actor_id)
    FROM accessible_thread
    RETURNING *
)
SELECT
    comment.comment_id,
    comment.thread_id,
    comment.body,
    comment.created_by,
    comment.created_at,
    comment.updated_at,
    COALESCE(author.full_name, author.username, 'Former user') AS author_name,
    author.avatar_url AS author_avatar
FROM inserted_comment AS comment
INNER JOIN public.users AS author ON author.user_id = comment.created_by;

-- name: SetDocumentCommentResolved :one
UPDATE public.document_comment_threads AS thread
SET resolved_at = CASE WHEN sqlc.arg(resolved)::boolean THEN CURRENT_TIMESTAMP ELSE NULL END,
    resolved_by = CASE WHEN sqlc.arg(resolved)::boolean THEN sqlc.arg(actor_id) ELSE NULL END,
    updated_at = CURRENT_TIMESTAMP
FROM public.documents AS document
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = document.workspace_id
   AND membership.user_id = sqlc.arg(actor_id)
   AND membership.role <> CAST('guest' AS public.user_role)
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE thread.thread_id = sqlc.arg(thread_id)
  AND thread.document_id = document.document_id
  AND thread.workspace_id = document.workspace_id
  AND document.document_id = sqlc.arg(document_id)
  AND document.workspace_id = sqlc.arg(workspace_id)
  AND document.archived_at IS NULL
  AND (
      document.visibility = 'workspace'
      OR document.created_by = sqlc.arg(actor_id)
      OR EXISTS (
          SELECT 1
          FROM public.document_members AS reader
          WHERE reader.document_id = document.document_id
            AND reader.user_id = sqlc.arg(actor_id)
      )
  )
RETURNING thread.thread_id;
