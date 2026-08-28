-- name: ListDeletedStoryRetentionCandidates :many
SELECT
    story.id,
    story.deleted_at
FROM public.stories AS story
WHERE story.deleted_at IS NOT NULL
  AND story.deleted_at < sqlc.arg(deleted_before)
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR story.deleted_at > sqlc.arg(after_deleted_at)
      OR (
          story.deleted_at = sqlc.arg(after_deleted_at)
          AND story.id > sqlc.arg(after_story_id)
      )
  )
ORDER BY story.deleted_at, story.id
LIMIT CAST(sqlc.arg(batch_size) AS integer)
FOR UPDATE OF story SKIP LOCKED;

-- name: ListStoryRetentionAttachmentCandidates :many
WITH relation AS MATERIALIZED (
    SELECT story_attachment.attachment_id
    FROM public.story_attachments AS story_attachment
    INNER JOIN public.stories AS story
        ON story.id = story_attachment.story_id
    INNER JOIN public.attachments AS attachment
        ON attachment.attachment_id = story_attachment.attachment_id
       AND attachment.workspace_id = story.workspace_id
    WHERE story_attachment.story_id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
    UNION
    SELECT inline_attachment.attachment_id
    FROM public.story_inline_attachments AS inline_attachment
    INNER JOIN public.stories AS story
        ON story.id = inline_attachment.story_id
    INNER JOIN public.attachments AS attachment
        ON attachment.attachment_id = inline_attachment.attachment_id
       AND attachment.workspace_id = story.workspace_id
    WHERE inline_attachment.story_id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
)
SELECT relation.attachment_id
FROM relation
ORDER BY relation.attachment_id
LIMIT CAST(sqlc.arg(maximum_attachment_count) AS integer);

-- name: DeleteStoryRetentionCandidates :many
DELETE FROM public.stories AS story
WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
  AND story.deleted_at IS NOT NULL
  AND story.deleted_at < sqlc.arg(deleted_before)
RETURNING story.id;

-- name: DeleteUnreferencedStoryRetentionAttachments :many
DELETE FROM public.attachments AS attachment
WHERE attachment.attachment_id = ANY(CAST(sqlc.arg(attachment_ids) AS uuid[]))
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_attachments AS story_relation
      WHERE story_relation.attachment_id = attachment.attachment_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.story_inline_attachments AS inline_relation
      WHERE inline_relation.attachment_id = attachment.attachment_id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM public.document_attachments AS document_relation
      WHERE document_relation.attachment_id = attachment.attachment_id
  )
RETURNING
    attachment.attachment_id,
    attachment.workspace_id,
    CAST(attachment.blob_name AS text) AS blob_name;

-- name: InsertAttachmentObjectDeletionOutbox :execrows
INSERT INTO public.attachment_object_deletion_outbox (
    attachment_id,
    workspace_id,
    storage_provider,
    container_name,
    blob_name,
    status,
    attempt_count,
    next_attempt_at,
    created_at,
    updated_at
)
VALUES (
    sqlc.arg(attachment_id),
    sqlc.arg(workspace_id),
    sqlc.arg(storage_provider),
    sqlc.arg(container_name),
    sqlc.arg(blob_name),
    'pending',
    0,
    sqlc.arg(enqueued_at),
    sqlc.arg(enqueued_at),
    sqlc.arg(enqueued_at)
);

-- name: ClaimAttachmentObjectDeletions :many
WITH candidates AS MATERIALIZED (
    SELECT deletion.outbox_id
    FROM public.attachment_object_deletion_outbox AS deletion
    WHERE (
        deletion.status IN ('pending', 'retrying')
        AND deletion.next_attempt_at <= sqlc.arg(as_of)
    ) OR (
        deletion.status = 'processing'
        AND deletion.processing_started_at <= sqlc.arg(lease_expired_before)
    )
    ORDER BY
        CASE
            WHEN deletion.status = 'processing' THEN deletion.processing_started_at
            ELSE deletion.next_attempt_at
        END,
        deletion.created_at,
        deletion.outbox_id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF deletion SKIP LOCKED
)
UPDATE public.attachment_object_deletion_outbox AS deletion
SET
    status = 'processing',
    attempt_count = deletion.attempt_count + 1,
    next_attempt_at = NULL,
    processing_started_at = sqlc.arg(as_of),
    claim_token = sqlc.arg(claim_token),
    completed_at = NULL,
    last_error = NULL,
    updated_at = sqlc.arg(as_of)
FROM candidates
WHERE deletion.outbox_id = candidates.outbox_id
RETURNING
    deletion.outbox_id,
    deletion.attachment_id,
    deletion.workspace_id,
    CAST(deletion.storage_provider AS text) AS storage_provider,
    CAST(deletion.container_name AS text) AS container_name,
    CAST(deletion.blob_name AS text) AS blob_name,
    deletion.claim_token,
    deletion.attempt_count;

-- name: CompleteAttachmentObjectDeletion :execrows
UPDATE public.attachment_object_deletion_outbox AS deletion
SET
    status = 'completed',
    next_attempt_at = NULL,
    processing_started_at = NULL,
    claim_token = NULL,
    completed_at = sqlc.arg(completed_at),
    last_error = NULL,
    updated_at = sqlc.arg(completed_at)
WHERE deletion.outbox_id = sqlc.arg(outbox_id)
  AND deletion.status = 'processing'
  AND deletion.claim_token = sqlc.arg(claim_token);

-- name: FailAttachmentObjectDeletion :execrows
UPDATE public.attachment_object_deletion_outbox AS deletion
SET
    status = 'retrying',
    next_attempt_at = sqlc.arg(next_attempt_at),
    processing_started_at = NULL,
    claim_token = NULL,
    completed_at = NULL,
    last_error = sqlc.arg(last_error),
    updated_at = sqlc.arg(failed_at)
WHERE deletion.outbox_id = sqlc.arg(outbox_id)
  AND deletion.status = 'processing'
  AND deletion.claim_token = sqlc.arg(claim_token);

-- name: PurgeCompletedAttachmentObjectDeletions :execrows
WITH candidates AS MATERIALIZED (
    SELECT deletion.outbox_id
    FROM public.attachment_object_deletion_outbox AS deletion
    WHERE deletion.status = 'completed'
      AND deletion.completed_at < sqlc.arg(completed_before)
    ORDER BY deletion.completed_at, deletion.outbox_id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF deletion SKIP LOCKED
)
DELETE FROM public.attachment_object_deletion_outbox AS deletion
USING candidates
WHERE deletion.outbox_id = candidates.outbox_id;
