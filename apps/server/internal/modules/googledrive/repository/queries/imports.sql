-- name: CreateGoogleDriveImportOperation :one
INSERT INTO public.google_drive_document_import_operations (
    workspace_id,
    user_id,
    source_reference_id,
    document_id,
    idempotency_key,
    request_hash,
    visibility,
    attempt_generation
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(source_reference_id),
    sqlc.arg(document_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(request_hash),
    sqlc.arg(visibility),
    sqlc.arg(attempt_generation)
)
ON CONFLICT (workspace_id, user_id, idempotency_key) DO NOTHING
RETURNING operation_id, workspace_id, user_id, source_reference_id,
    document_id, idempotency_key, request_hash, visibility,
    attempt_generation, status, created_at, updated_at, completed_at;

-- name: GetGoogleDriveImportOperation :one
SELECT operation_id, workspace_id, user_id, source_reference_id,
    document_id, idempotency_key, request_hash, visibility,
    attempt_generation, status, created_at, updated_at, completed_at
FROM public.google_drive_document_import_operations
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND idempotency_key = sqlc.arg(idempotency_key);

-- name: ClaimGoogleDriveImportOperation :one
UPDATE public.google_drive_document_import_operations
SET status = 'pending',
    attempt_generation = sqlc.arg(attempt_generation),
    error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND updated_at = sqlc.arg(previous_updated_at)
  AND (
      status = 'failed'
      OR (status = 'pending' AND updated_at <= sqlc.arg(stale_before))
  )
RETURNING operation_id, workspace_id, user_id, source_reference_id,
    document_id, idempotency_key, request_hash, visibility,
    attempt_generation, status, created_at, updated_at, completed_at;

-- name: LockGoogleDriveImportOperation :one
SELECT operation_id, workspace_id, user_id, source_reference_id,
    document_id, idempotency_key, request_hash, visibility,
    attempt_generation, status, created_at, updated_at, completed_at
FROM public.google_drive_document_import_operations
WHERE operation_id = sqlc.arg(operation_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: GoogleDriveReferenceImportable :one
SELECT EXISTS (
    SELECT 1
    FROM public.google_drive_file_references AS reference
    INNER JOIN public.google_drive_files AS file
        ON file.file_id = reference.file_id
       AND file.workspace_id = reference.workspace_id
       AND file.unavailable_at IS NULL
    INNER JOIN public.google_drive_file_grants AS grant_record
        ON grant_record.file_id = file.file_id
       AND grant_record.user_id = sqlc.arg(user_id)
       AND grant_record.account_id = sqlc.arg(account_id)
       AND grant_record.verification_generation = sqlc.arg(grant_generation)
    INNER JOIN public.google_drive_workspace_connections AS connection
        ON connection.workspace_id = reference.workspace_id
       AND connection.user_id = sqlc.arg(user_id)
       AND connection.account_id = grant_record.account_id
    INNER JOIN public.google_drive_accounts AS account
        ON account.account_id = connection.account_id
       AND account.user_id = connection.user_id
       AND account.revoked_at IS NULL
    WHERE reference.reference_id = sqlc.arg(reference_id)
      AND reference.workspace_id = sqlc.arg(workspace_id)
      AND reference.target_type = sqlc.arg(target_type)
      AND COALESCE(reference.story_id, reference.objective_id, reference.document_id, reference.comment_id) = sqlc.arg(target_id)
      AND file.google_file_id = sqlc.arg(google_file_id)
) AS importable;

-- name: CreateGoogleDriveImportedDocument :one
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
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.workspace_id = sqlc.arg(workspace_id)
  AND membership.user_id = sqlc.arg(user_id)
  AND membership.role <> CAST('guest' AS public.user_role)
RETURNING document_id;

-- name: CompleteGoogleDriveImportOperation :execrows
UPDATE public.google_drive_document_import_operations
SET status = 'completed',
    error_code = NULL,
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND status = 'pending'
  AND attempt_generation = sqlc.arg(attempt_generation);

-- name: FailGoogleDriveImportOperation :execrows
UPDATE public.google_drive_document_import_operations
SET status = 'failed',
    error_code = sqlc.arg(error_code),
    completed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND status = 'pending'
  AND attempt_generation = sqlc.arg(attempt_generation);
