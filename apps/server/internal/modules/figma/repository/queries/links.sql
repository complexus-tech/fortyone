-- name: ListStoryLinks :many
SELECT link.*
FROM public.story_figma_links AS link
INNER JOIN public.stories AS story
    ON story.id = link.story_id
   AND story.workspace_id = link.workspace_id
   AND story.deleted_at IS NULL
WHERE link.workspace_id = sqlc.arg(workspace_id)
  AND link.story_id = sqlc.arg(story_id)
ORDER BY link.created_at, link.id;

-- name: ListStoryHandoffStatuses :many
SELECT
    link.story_id,
    CASE
        WHEN BOOL_AND(link.dev_status = 'COMPLETED') THEN 'COMPLETED'
        ELSE 'READY_FOR_DEV'
    END AS status
FROM public.story_figma_links AS link
INNER JOIN public.stories AS story
    ON story.id = link.story_id
   AND story.workspace_id = link.workspace_id
   AND story.deleted_at IS NULL
WHERE link.workspace_id = sqlc.arg(workspace_id)
GROUP BY link.story_id
HAVING BOOL_OR(link.dev_status IN ('READY_FOR_DEV', 'COMPLETED'))
ORDER BY link.story_id;

-- name: ListLinksByFile :many
SELECT link.*
FROM public.story_figma_links AS link
INNER JOIN public.stories AS story
    ON story.id = link.story_id
   AND story.workspace_id = link.workspace_id
   AND story.deleted_at IS NULL
WHERE link.workspace_id = sqlc.arg(workspace_id)
  AND link.file_key = sqlc.arg(file_key)
ORDER BY link.created_at, link.id;

-- name: UpsertGenericStoryLink :one
INSERT INTO public.story_links (title, url, story_id, external_source_key)
SELECT
    CAST(sqlc.narg(title) AS text),
    sqlc.arg(url),
    story.id,
    sqlc.arg(external_source_key)
FROM public.stories AS story
INNER JOIN public.workspace_members AS member
    ON member.workspace_id = story.workspace_id
   AND member.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = member.user_id
   AND actor.is_active = TRUE
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
ON CONFLICT (external_source_key) WHERE external_source_key IS NOT NULL
DO UPDATE SET
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    updated_at = sqlc.arg(updated_at)
RETURNING link_id;

-- name: UpsertFigmaStoryLink :one
INSERT INTO public.story_figma_links (
    workspace_id, story_id, created_by_user_id, story_link_id,
    file_key, node_id, original_url, canonical_url, file_name,
    node_name, node_type, thumbnail_url, version, last_modified, metadata
)
SELECT
    story.workspace_id,
    story.id,
    actor.user_id,
    sqlc.arg(story_link_id),
    sqlc.arg(file_key),
    CAST(sqlc.narg(node_id) AS text),
    sqlc.arg(original_url),
    sqlc.arg(canonical_url),
    sqlc.arg(file_name),
    CAST(sqlc.narg(node_name) AS text),
    CAST(sqlc.narg(node_type) AS text),
    CAST(sqlc.narg(thumbnail_url) AS text),
    CAST(sqlc.narg(version) AS text),
    CAST(sqlc.narg(last_modified) AS timestamptz),
    sqlc.arg(metadata)
FROM public.stories AS story
INNER JOIN public.workspace_members AS member
    ON member.workspace_id = story.workspace_id
   AND member.user_id = sqlc.arg(actor_id)
INNER JOIN public.users AS actor
    ON actor.user_id = member.user_id
   AND actor.is_active = TRUE
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
ON CONFLICT (story_id, file_key, COALESCE(node_id, '')) DO UPDATE SET
    story_link_id = EXCLUDED.story_link_id,
    original_url = EXCLUDED.original_url,
    canonical_url = EXCLUDED.canonical_url,
    file_name = EXCLUDED.file_name,
    node_name = EXCLUDED.node_name,
    node_type = EXCLUDED.node_type,
    thumbnail_url = EXCLUDED.thumbnail_url,
    version = EXCLUDED.version,
    last_modified = EXCLUDED.last_modified,
    metadata = EXCLUDED.metadata,
    last_synced_at = sqlc.arg(updated_at),
    unavailable_at = NULL,
    updated_at = sqlc.arg(updated_at)
RETURNING *;

-- name: UpdateFigmaStoryLink :execrows
UPDATE public.story_figma_links
SET file_name = sqlc.arg(file_name),
    node_name = CAST(sqlc.narg(node_name) AS text),
    node_type = CAST(sqlc.narg(node_type) AS text),
    thumbnail_url = CAST(sqlc.narg(thumbnail_url) AS text),
    version = CAST(sqlc.narg(version) AS text),
    last_modified = CAST(sqlc.narg(last_modified) AS timestamptz),
    dev_status = CAST(sqlc.narg(dev_status) AS text),
    dev_resource_id = CAST(sqlc.narg(dev_resource_id) AS text),
    metadata = sqlc.arg(metadata),
    last_synced_at = sqlc.arg(updated_at),
    unavailable_at = CAST(sqlc.narg(unavailable_at) AS timestamptz),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND workspace_id = sqlc.arg(workspace_id);

-- name: GetFigmaStoryLink :one
SELECT link.*
FROM public.story_figma_links AS link
INNER JOIN public.stories AS story
    ON story.id = link.story_id
   AND story.workspace_id = link.workspace_id
   AND story.deleted_at IS NULL
WHERE link.workspace_id = sqlc.arg(workspace_id)
  AND link.id = sqlc.arg(id);

-- name: DeleteFigmaStoryLink :one
DELETE FROM public.story_figma_links
WHERE id = sqlc.arg(id)
  AND workspace_id = sqlc.arg(workspace_id)
RETURNING *;

-- name: DeleteGenericStoryLink :execrows
DELETE FROM public.story_links
WHERE link_id = sqlc.arg(link_id);
