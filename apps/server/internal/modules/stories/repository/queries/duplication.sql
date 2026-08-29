-- name: DuplicateAuthorizedStory :one
INSERT INTO public.stories (
    id,
    sequence_id,
    title,
    description,
    description_html,
    team_id,
    objective_id,
    status_id,
    assignee_id,
    priority,
    estimate_unit,
    estimated_duration_minutes,
    minimum_focus_block_minutes,
    sprint_id,
    workspace_id,
    reporter_id,
    created_at,
    updated_at
)
SELECT
    sqlc.arg(target_story_id),
    sqlc.arg(sequence_id),
    'Copy of ' || source.title,
    source.description,
    CASE
        WHEN source.description_html IS NULL THEN NULL
        ELSE REPLACE(source.description_html, sqlc.arg(source_media_path), sqlc.arg(target_media_path))
    END,
    source.team_id,
    source.objective_id,
    source.status_id,
    source.assignee_id,
    source.priority,
    source.estimate_unit,
    source.estimated_duration_minutes,
    source.minimum_focus_block_minutes,
    source.sprint_id,
    source.workspace_id,
    sqlc.arg(reporter_id),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
FROM public.stories AS source
WHERE source.id = sqlc.arg(source_story_id)
  AND source.workspace_id = sqlc.arg(workspace_id)
  AND source.deleted_at IS NULL
RETURNING id;

-- name: CopyDuplicatedStoryMediaLinks :exec
INSERT INTO public.story_inline_attachments (story_id, attachment_id, created_by)
SELECT
    sqlc.arg(target_story_id),
    media.attachment_id,
    sqlc.arg(created_by)
FROM public.story_inline_attachments AS media
INNER JOIN public.stories AS source_story
    ON source_story.id = media.story_id
INNER JOIN public.attachments AS attachment
    ON attachment.attachment_id = media.attachment_id
   AND attachment.workspace_id = source_story.workspace_id
WHERE media.story_id = sqlc.arg(source_story_id)
  AND source_story.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (story_id, attachment_id) DO NOTHING;
