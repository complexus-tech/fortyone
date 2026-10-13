-- All writes run after AuthorizeSecondaryStoryTargets locks the owning story.
-- That single aggregate lock serializes adds, edits, ordering, clearing and conversion.
-- name: ListStoryChecklist :many
SELECT item.id, item.story_id, item.title, item.completed, item.position,
       item.converted_at, item.created_at, item.updated_at,
       converted.id AS converted_id, converted.sequence_id AS converted_sequence_id,
       converted.title AS converted_title, converted.team_id AS converted_team_id,
       converted.status_id AS converted_status_id
FROM public.story_checklist_items AS item
INNER JOIN public.stories AS owner ON owner.id = item.story_id
LEFT JOIN public.stories AS converted
  ON converted.id = item.converted_story_id
 AND converted.workspace_id = owner.workspace_id
 AND converted.team_id = owner.team_id
 AND converted.deleted_at IS NULL
WHERE item.story_id = sqlc.arg(story_id)
  AND owner.workspace_id = sqlc.arg(workspace_id)
ORDER BY item.position, item.id;

-- name: InsertStoryChecklistItem :exec
INSERT INTO public.story_checklist_items(id, story_id, title, position, created_at, updated_at)
VALUES (sqlc.arg(id), sqlc.arg(story_id), sqlc.arg(title), sqlc.arg(position), sqlc.arg(now), sqlc.arg(now));

-- name: UpdateStoryChecklistItem :execrows
UPDATE public.story_checklist_items
SET title = COALESCE(sqlc.narg(title), title),
    completed = COALESCE(sqlc.narg(completed), completed), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND story_id = sqlc.arg(story_id);

-- name: DeleteStoryChecklistItem :execrows
DELETE FROM public.story_checklist_items WHERE id = sqlc.arg(id) AND story_id = sqlc.arg(story_id);

-- name: ClearStoryChecklist :exec
DELETE FROM public.story_checklist_items WHERE story_id = sqlc.arg(story_id);

-- name: ReorderStoryChecklist :execrows
UPDATE public.story_checklist_items AS item
SET position = CAST(ordered.ordinality - 1 AS integer), updated_at = sqlc.arg(now)
FROM unnest(CAST(sqlc.arg(item_ids) AS uuid[])) WITH ORDINALITY AS ordered(id, ordinality)
WHERE item.id = ordered.id AND item.story_id = sqlc.arg(story_id);

-- name: CompleteStoryChecklistConversion :exec
UPDATE public.story_checklist_items
SET completed = true, converted_story_id = sqlc.arg(converted_story_id),
    converted_at = sqlc.arg(now), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND story_id = sqlc.arg(story_id);

-- name: FindChecklistConversionStatus :one
SELECT status_id FROM public.statuses
WHERE team_id = sqlc.arg(team_id) AND category IN ('unstarted', 'backlog')
ORDER BY CASE WHEN category = 'unstarted' THEN 0 ELSE 1 END, order_index, status_id
LIMIT 1;

-- name: TouchChecklistStory :exec
UPDATE public.stories SET updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(story_id) AND workspace_id = sqlc.arg(workspace_id);

-- name: CopyDuplicatedStoryChecklist :exec
INSERT INTO public.story_checklist_items (id, story_id, title, position, created_at, updated_at)
SELECT gen_random_uuid(), sqlc.arg(target_story_id), item.title, item.position, sqlc.arg(now), sqlc.arg(now)
FROM public.story_checklist_items AS item
INNER JOIN public.stories AS source ON source.id = item.story_id
INNER JOIN public.stories AS target ON target.id = sqlc.arg(target_story_id)
    AND target.workspace_id = source.workspace_id
WHERE item.story_id = sqlc.arg(source_story_id)
  AND source.workspace_id = sqlc.arg(workspace_id);
