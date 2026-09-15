-- Team deletion owns the parent team lock before taking these locks. Both
-- archived and soft-deleted stories participate in permanent team deletion.
-- name: LockTeamDeletionStories :many
SELECT story.id FROM public.stories AS story
WHERE story.team_id = sqlc.arg(team_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
ORDER BY story.id
FOR UPDATE OF story;

-- The event payload matches the normal story.deleted integration contract.
-- Bulk insertion keeps deletion latency independent of per-story round trips.
-- name: DeleteTeamStoriesWithEvents :execrows
WITH deleted AS (
    DELETE FROM public.stories AS story
    WHERE story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))
      AND story.workspace_id = sqlc.arg(workspace_id)
    RETURNING story.id
)
INSERT INTO public.story_mutation_events (
    event_id, workspace_id, story_id, event_type, actor_kind, actor_id,
    actor_credential_id, payload, occurred_at, status, attempt_count,
    next_attempt_at, created_at, updated_at
)
SELECT gen_random_uuid(), sqlc.arg(workspace_id), deleted.id, 'story.deleted',
       sqlc.arg(actor_kind), sqlc.arg(actor_id), sqlc.narg(actor_credential_id),
       jsonb_build_object('storyId', deleted.id, 'workspaceId', CAST(sqlc.arg(workspace_id) AS uuid)),
       sqlc.arg(deleted_at), 'pending', 0,
       sqlc.arg(deleted_at), sqlc.arg(deleted_at), sqlc.arg(deleted_at)
FROM deleted;
