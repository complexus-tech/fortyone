-- name: GoogleDriveTargetAccessible :one
SELECT CASE CAST(sqlc.arg(target_type) AS text)
    WHEN 'story' THEN EXISTS (
        SELECT 1
        FROM public.stories AS story
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = story.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = story.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin', 'guest')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = story.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE story.id = sqlc.arg(target_id)
          AND story.workspace_id = sqlc.arg(workspace_id)
          AND story.deleted_at IS NULL
    )
    WHEN 'objective' THEN EXISTS (
        SELECT 1
        FROM public.objectives AS objective
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = objective.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = objective.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = objective.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE objective.objective_id = sqlc.arg(target_id)
          AND objective.workspace_id = sqlc.arg(workspace_id)
    )
    WHEN 'document' THEN EXISTS (
        SELECT 1
        FROM public.documents AS document
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = document.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = document.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin', 'guest')
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE document.document_id = sqlc.arg(target_id)
          AND document.workspace_id = sqlc.arg(workspace_id)
          AND document.archived_at IS NULL
          AND (
              document.visibility = 'workspace'
              OR document.created_by = sqlc.arg(user_id)
              OR EXISTS (
                  SELECT 1
                  FROM public.document_members AS document_member
                  WHERE document.visibility = 'restricted'
                    AND document_member.document_id = document.document_id
                    AND document_member.user_id = sqlc.arg(user_id)
              )
          )
    )
    WHEN 'comment' THEN EXISTS (
        SELECT 1
        FROM public.story_comments AS comment
        INNER JOIN public.stories AS story
            ON story.id = comment.story_id
           AND story.deleted_at IS NULL
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = story.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = story.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin', 'guest')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = story.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE comment.comment_id = sqlc.arg(target_id)
          AND story.workspace_id = sqlc.arg(workspace_id)
    )
    ELSE FALSE
END AS accessible;

-- name: GoogleDriveTargetMutable :one
SELECT CASE CAST(sqlc.arg(target_type) AS text)
    WHEN 'story' THEN EXISTS (
        SELECT 1
        FROM public.stories AS story
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = story.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = story.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = story.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE story.id = sqlc.arg(target_id)
          AND story.workspace_id = sqlc.arg(workspace_id)
          AND story.deleted_at IS NULL
    )
    WHEN 'objective' THEN EXISTS (
        SELECT 1
        FROM public.objectives AS objective
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = objective.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = objective.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = objective.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE objective.objective_id = sqlc.arg(target_id)
          AND objective.workspace_id = sqlc.arg(workspace_id)
    )
    WHEN 'document' THEN EXISTS (
        SELECT 1
        FROM public.documents AS document
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = document.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = document.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin')
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE document.document_id = sqlc.arg(target_id)
          AND document.workspace_id = sqlc.arg(workspace_id)
          AND document.archived_at IS NULL
          AND (
              document.visibility = 'workspace'
              OR document.created_by = sqlc.arg(user_id)
              OR EXISTS (
                  SELECT 1
                  FROM public.document_members AS document_member
                  WHERE document.visibility = 'restricted'
                    AND document_member.document_id = document.document_id
                    AND document_member.user_id = sqlc.arg(user_id)
                    AND document_member.role = 'editor'
              )
          )
    )
    WHEN 'comment' THEN EXISTS (
        SELECT 1
        FROM public.story_comments AS comment
        INNER JOIN public.stories AS story
            ON story.id = comment.story_id
           AND story.deleted_at IS NULL
        INNER JOIN public.workspaces AS active_workspace
            ON active_workspace.workspace_id = story.workspace_id
           AND active_workspace.deleted_at IS NULL
        INNER JOIN public.workspace_members AS workspace_member
            ON workspace_member.workspace_id = story.workspace_id
           AND workspace_member.user_id = sqlc.arg(user_id)
           AND workspace_member.role IN ('member', 'admin')
        INNER JOIN public.team_members AS team_member
            ON team_member.team_id = story.team_id
           AND team_member.user_id = sqlc.arg(user_id)
        INNER JOIN public.users AS actor
            ON actor.user_id = workspace_member.user_id
           AND actor.is_active = TRUE
        WHERE comment.comment_id = sqlc.arg(target_id)
          AND comment.commenter_id = sqlc.arg(user_id)
          AND story.workspace_id = sqlc.arg(workspace_id)
    )
    ELSE FALSE
END AS mutable;

-- name: UpsertGoogleDriveFile :one
INSERT INTO public.google_drive_files (
    workspace_id,
    google_file_id,
    resource_key,
    name,
    mime_type,
    web_view_link,
    drive_id,
    version,
    size_bytes,
    modified_at,
    metadata
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(google_file_id),
    CAST(sqlc.narg(resource_key) AS text),
    sqlc.arg(name),
    sqlc.arg(mime_type),
    sqlc.arg(web_view_link),
    CAST(sqlc.narg(drive_id) AS text),
    CAST(sqlc.narg(version) AS text),
    CAST(sqlc.narg(size_bytes) AS bigint),
    CAST(sqlc.narg(modified_at) AS timestamptz),
    sqlc.arg(metadata)
)
ON CONFLICT (workspace_id, google_file_id)
DO UPDATE SET
    resource_key = EXCLUDED.resource_key,
    name = EXCLUDED.name,
    mime_type = EXCLUDED.mime_type,
    web_view_link = EXCLUDED.web_view_link,
    drive_id = EXCLUDED.drive_id,
    version = EXCLUDED.version,
    size_bytes = EXCLUDED.size_bytes,
    modified_at = EXCLUDED.modified_at,
    metadata = EXCLUDED.metadata,
    last_synced_at = CURRENT_TIMESTAMP,
    unavailable_at = NULL,
    updated_at = CURRENT_TIMESTAMP
RETURNING file_id, workspace_id, google_file_id, resource_key, name, mime_type,
    web_view_link, drive_id, version, size_bytes, modified_at, metadata,
    last_synced_at, unavailable_at, created_at, updated_at;

-- name: UpsertGoogleDriveFileGrant :execrows
INSERT INTO public.google_drive_file_grants (
    file_id,
    user_id,
    account_id,
    verification_generation
)
SELECT
    sqlc.arg(file_id),
    connection.user_id,
    connection.account_id,
    sqlc.arg(grant_generation)
FROM public.google_drive_workspace_connections AS connection
INNER JOIN public.google_drive_accounts AS account
    ON account.account_id = connection.account_id
   AND account.user_id = connection.user_id
   AND account.revoked_at IS NULL
WHERE connection.workspace_id = sqlc.arg(workspace_id)
  AND connection.user_id = sqlc.arg(user_id)
  AND connection.account_id = sqlc.arg(account_id)
ON CONFLICT (file_id, user_id)
DO UPDATE SET
    account_id = EXCLUDED.account_id,
    verification_generation = EXCLUDED.verification_generation,
    last_verified_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP;

-- name: RevalidateGoogleDriveFileReference :one
UPDATE public.google_drive_files AS file
SET resource_key = CAST(sqlc.narg(resource_key) AS text),
    name = sqlc.arg(name),
    mime_type = sqlc.arg(mime_type),
    web_view_link = sqlc.arg(web_view_link),
    drive_id = CAST(sqlc.narg(drive_id) AS text),
    version = CAST(sqlc.narg(version) AS text),
    size_bytes = CAST(sqlc.narg(size_bytes) AS bigint),
    modified_at = CAST(sqlc.narg(modified_at) AS timestamptz),
    metadata = sqlc.arg(metadata),
    last_synced_at = CURRENT_TIMESTAMP,
    unavailable_at = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM public.google_drive_file_references AS reference
WHERE reference.reference_id = sqlc.arg(reference_id)
  AND reference.workspace_id = sqlc.arg(workspace_id)
  AND reference.file_id = file.file_id
  AND file.workspace_id = reference.workspace_id
  AND file.google_file_id = sqlc.arg(google_file_id)
RETURNING file.file_id;

-- name: DeleteGoogleDriveFileGrantForActor :execrows
DELETE FROM public.google_drive_file_grants AS grant_record
USING public.google_drive_files AS file,
      public.google_drive_file_references AS reference
WHERE reference.reference_id = sqlc.arg(reference_id)
  AND reference.workspace_id = sqlc.arg(workspace_id)
  AND reference.file_id = file.file_id
  AND file.workspace_id = reference.workspace_id
  AND grant_record.file_id = file.file_id
  AND grant_record.user_id = sqlc.arg(user_id)
  AND grant_record.account_id = sqlc.arg(account_id)
  AND grant_record.verification_generation = sqlc.arg(grant_generation);

-- name: MarkGoogleDriveFileUnavailable :execrows
UPDATE public.google_drive_files AS file
SET unavailable_at = COALESCE(unavailable_at, CURRENT_TIMESTAMP),
    last_synced_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
FROM public.google_drive_file_references AS reference
WHERE reference.reference_id = sqlc.arg(reference_id)
  AND reference.workspace_id = sqlc.arg(workspace_id)
  AND reference.file_id = file.file_id
  AND file.workspace_id = reference.workspace_id;

-- name: UpsertGoogleDriveFileReference :one
INSERT INTO public.google_drive_file_references (
    workspace_id,
    file_id,
    target_type,
    story_id,
    objective_id,
    document_id,
    comment_id,
    created_by_user_id
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(file_id),
    sqlc.arg(target_type),
    CAST(sqlc.narg(story_id) AS uuid),
    CAST(sqlc.narg(objective_id) AS uuid),
    CAST(sqlc.narg(document_id) AS uuid),
    CAST(sqlc.narg(comment_id) AS uuid),
    sqlc.arg(created_by_user_id)
)
ON CONFLICT DO NOTHING
RETURNING reference_id;

-- name: FindGoogleDriveFileReference :one
SELECT reference_id
FROM public.google_drive_file_references
WHERE file_id = sqlc.arg(file_id)
  AND target_type = sqlc.arg(target_type)
  AND (
      (target_type = 'story' AND story_id = sqlc.arg(target_id))
      OR (target_type = 'objective' AND objective_id = sqlc.arg(target_id))
      OR (target_type = 'document' AND document_id = sqlc.arg(target_id))
      OR (target_type = 'comment' AND comment_id = sqlc.arg(target_id))
  );

-- name: ListGoogleDriveFileReferences :many
SELECT
    reference.reference_id,
    file.file_id AS internal_file_id,
    file.google_file_id,
    file.resource_key,
    file.name,
    file.mime_type,
    file.web_view_link,
    file.version,
    file.modified_at,
    file.unavailable_at,
    reference.target_type,
    COALESCE(reference.story_id, reference.objective_id, reference.document_id, reference.comment_id) AS target_id,
    actor_account.email AS connection_email,
    actor_account.requires_reauthorization,
    reference.created_at,
    reference.updated_at
FROM public.google_drive_file_references AS reference
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = reference.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.google_drive_files AS file
    ON file.file_id = reference.file_id
   AND file.workspace_id = reference.workspace_id
LEFT JOIN public.google_drive_file_grants AS actor_grant
    ON actor_grant.file_id = file.file_id
   AND actor_grant.user_id = sqlc.arg(user_id)
LEFT JOIN public.google_drive_workspace_connections AS actor_connection
    ON actor_connection.workspace_id = reference.workspace_id
   AND actor_connection.user_id = sqlc.arg(user_id)
   AND actor_connection.account_id = actor_grant.account_id
LEFT JOIN public.google_drive_accounts AS actor_account
    ON actor_account.account_id = actor_connection.account_id
   AND actor_account.user_id = actor_connection.user_id
   AND actor_account.revoked_at IS NULL
WHERE reference.workspace_id = sqlc.arg(workspace_id)
  AND reference.target_type = sqlc.arg(target_type)
  AND (
      (reference.target_type = 'story' AND reference.story_id = sqlc.arg(target_id))
      OR (reference.target_type = 'objective' AND reference.objective_id = sqlc.arg(target_id))
      OR (reference.target_type = 'document' AND reference.document_id = sqlc.arg(target_id))
      OR (reference.target_type = 'comment' AND reference.comment_id = sqlc.arg(target_id))
  )
ORDER BY reference.created_at, reference.reference_id;

-- name: GetGoogleDriveFileReference :one
SELECT
    reference.reference_id,
    file.file_id AS internal_file_id,
    file.google_file_id,
    file.resource_key,
    file.name,
    file.mime_type,
    file.web_view_link,
    file.version,
    file.modified_at,
    file.unavailable_at,
    reference.target_type,
    COALESCE(reference.story_id, reference.objective_id, reference.document_id, reference.comment_id) AS target_id,
    account.account_id,
    account.user_id AS account_user_id,
    account.google_subject,
    account.email AS connection_email,
    account.display_name,
    account.credential_payload,
    account.credential_key_version,
    account.installation_generation,
    account.scopes,
    account.expires_at,
    account.requires_reauthorization,
    account.created_at AS account_created_at,
    account.updated_at AS account_updated_at,
    grant_record.verification_generation AS grant_generation,
    reference.created_at,
    reference.updated_at
FROM public.google_drive_file_references AS reference
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = reference.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.google_drive_files AS file
    ON file.file_id = reference.file_id
   AND file.workspace_id = reference.workspace_id
LEFT JOIN public.google_drive_file_grants AS grant_record
    ON grant_record.file_id = file.file_id
   AND grant_record.user_id = sqlc.arg(user_id)
LEFT JOIN public.google_drive_workspace_connections AS connection
    ON connection.workspace_id = reference.workspace_id
   AND connection.user_id = sqlc.arg(user_id)
   AND connection.account_id = grant_record.account_id
LEFT JOIN public.google_drive_accounts AS account
    ON account.account_id = connection.account_id
   AND account.user_id = connection.user_id
   AND account.revoked_at IS NULL
WHERE reference.reference_id = sqlc.arg(reference_id)
  AND reference.workspace_id = sqlc.arg(workspace_id);

-- name: DeleteGoogleDriveFileReference :one
DELETE FROM public.google_drive_file_references
WHERE reference_id = sqlc.arg(reference_id)
  AND workspace_id = sqlc.arg(workspace_id)
RETURNING file_id;

-- name: DeleteOrphanGoogleDriveFile :execrows
DELETE FROM public.google_drive_files AS file
WHERE file.file_id = sqlc.arg(file_id)
  AND file.workspace_id = sqlc.arg(workspace_id)
  AND NOT EXISTS (
      SELECT 1
      FROM public.google_drive_file_references AS reference
      WHERE reference.file_id = file.file_id
  );

-- name: CreateGoogleDriveOperation :one
INSERT INTO public.google_drive_create_operations (
    workspace_id,
    user_id,
    idempotency_key,
    request_hash,
    target_type,
    target_id,
    file_type,
    title
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(request_hash),
    sqlc.arg(target_type),
    sqlc.arg(target_id),
    sqlc.arg(file_type),
    sqlc.arg(title)
)
ON CONFLICT (workspace_id, user_id, idempotency_key) DO NOTHING
RETURNING operation_id, workspace_id, user_id, idempotency_key, request_hash,
    target_type, target_id, file_type, title, status, google_file_id,
    reference_id, created_at, updated_at;

-- name: GetGoogleDriveOperation :one
SELECT operation_id, workspace_id, user_id, idempotency_key, request_hash,
    target_type, target_id, file_type, title, status, google_file_id,
    reference_id, created_at, updated_at
FROM public.google_drive_create_operations
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND idempotency_key = sqlc.arg(idempotency_key);

-- name: ClaimGoogleDriveOperation :one
UPDATE public.google_drive_create_operations
SET status = 'pending',
    error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND updated_at = sqlc.arg(previous_updated_at)
  AND (
      status = 'failed'
      OR (
          status = 'pending'
          AND updated_at <= sqlc.arg(stale_before)
      )
  )
RETURNING operation_id, workspace_id, user_id, idempotency_key, request_hash,
    target_type, target_id, file_type, title, status, google_file_id,
    reference_id, created_at, updated_at;

-- name: CompleteGoogleDriveOperation :execrows
UPDATE public.google_drive_create_operations
SET status = 'completed',
    google_file_id = sqlc.arg(google_file_id),
    reference_id = sqlc.arg(reference_id),
    error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND status = 'pending';

-- name: FailGoogleDriveOperation :execrows
UPDATE public.google_drive_create_operations
SET status = 'failed',
    error_code = sqlc.arg(error_code),
    updated_at = CURRENT_TIMESTAMP
WHERE operation_id = sqlc.arg(operation_id)
  AND status = 'pending';

-- name: SaveGoogleDriveDocumentImport :execrows
INSERT INTO public.google_drive_document_imports (
    workspace_id,
    document_id,
    reference_id,
    google_file_id,
    source_version,
    imported_by_user_id
)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(document_id),
    sqlc.arg(reference_id),
    sqlc.arg(google_file_id),
    CAST(sqlc.narg(source_version) AS text),
    sqlc.arg(imported_by_user_id)
);
