-- name: ListPublicFeedbackStoryLinks :many
SELECT fsl.id,
       fsl.workspace_id,
       fsl.item_id,
       fsl.story_id,
       '' AS story_title,
       fsl.relationship,
       fsl.is_primary,
       fsl.created_by_user_id,
       fsl.created_at
FROM feedback_story_links fsl
INNER JOIN feedback_items fi ON fi.id = fsl.item_id
INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
WHERE fi.portal_id = sqlc.arg(portal_id)
  AND fi.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
  AND fi.deleted_at IS NULL
ORDER BY fsl.is_primary DESC, fsl.created_at ASC, fsl.id ASC;

-- name: ListInternalFeedbackStoryLinks :many
SELECT fsl.id,
       fsl.workspace_id,
       fsl.item_id,
       fsl.story_id,
       story.title AS story_title,
       fsl.relationship,
       fsl.is_primary,
       fsl.created_by_user_id,
       fsl.created_at
FROM feedback_story_links fsl
INNER JOIN feedback_items fi ON fi.id = fsl.item_id AND fi.workspace_id = fsl.workspace_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN stories story ON story.id = fsl.story_id AND story.workspace_id = fsl.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fi.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE fsl.workspace_id = sqlc.arg(workspace_id)
  AND fsl.item_id = sqlc.arg(item_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ORDER BY fsl.is_primary DESC, fsl.created_at ASC, fsl.id ASC;

-- name: ListStoryFeedbackLinks :many
SELECT fsl.id,
       fsl.workspace_id,
       fsl.item_id,
       fsl.story_id,
       fb.team_id,
       fi.title AS feedback_title,
       fsl.relationship,
       fsl.is_primary,
       fsl.created_at
FROM feedback_story_links fsl
INNER JOIN feedback_items fi ON fi.id = fsl.item_id AND fi.workspace_id = fsl.workspace_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fsl.workspace_id
INNER JOIN stories story ON story.id = fsl.story_id AND story.workspace_id = fsl.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fsl.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE fsl.workspace_id = sqlc.arg(workspace_id)
  AND fsl.story_id = sqlc.arg(story_id)
  AND fsl.is_primary = true
  AND fi.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ORDER BY fsl.created_at ASC, fsl.id ASC;

-- name: LinkFeedbackStory :one
INSERT INTO feedback_story_links (
    workspace_id,
    item_id,
    story_id,
    relationship,
    is_primary,
    created_by_user_id
)
SELECT item.workspace_id,
       item.id,
       story.id,
       sqlc.arg(relationship),
       sqlc.arg(is_primary),
       sqlc.arg(actor_id)
FROM feedback_items item
INNER JOIN feedback_boards fb ON fb.id = item.board_id AND fb.workspace_id = item.workspace_id
INNER JOIN stories story
    ON story.id = sqlc.arg(story_id)
   AND story.workspace_id = item.workspace_id
   AND story.deleted_at IS NULL
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members feedback_team ON feedback_team.team_id = fb.team_id AND feedback_team.user_id = wm.user_id
INNER JOIN team_members story_team ON story_team.team_id = story.team_id AND story_team.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND (
      CAST(sqlc.arg(all_teams) AS boolean)
      OR (
          fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[]))
          AND story.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[]))
      )
  )
ON CONFLICT (item_id, story_id) DO UPDATE
SET relationship = EXCLUDED.relationship,
    is_primary = EXCLUDED.is_primary
RETURNING id,
          workspace_id,
          item_id,
          story_id,
          relationship,
          is_primary,
          created_by_user_id,
          created_at;

-- name: FindFirstFeedbackStatusByCategory :one
SELECT status.status_id
FROM statuses status
INNER JOIN teams team ON team.team_id = status.team_id
WHERE status.team_id = sqlc.arg(team_id)
  AND status.category = sqlc.arg(category)
ORDER BY status.order_index ASC, status.status_id ASC
LIMIT 1;

-- name: GetFeedbackStatusCategory :one
SELECT CAST(status.category AS text)
FROM statuses status
WHERE status.team_id = sqlc.arg(team_id)
  AND status.status_id = sqlc.arg(status_id);

-- name: ListPrimaryStoryFeedbackItems :many
SELECT fi.id,
       fi.workspace_id,
       fi.portal_id,
       fi.contributor_id,
       fi.author_id,
       fi.title,
       fi.slug,
       CAST(CASE
           WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
           WHEN projected_state.category = 'backlog' THEN 'reviewing'
           WHEN projected_state.category = 'unstarted' THEN 'planned'
           WHEN projected_state.category = 'started' THEN 'in_progress'
           WHEN projected_state.category = 'paused' THEN 'planned'
           WHEN projected_state.category = 'completed' THEN 'completed'
           WHEN projected_state.category = 'cancelled' THEN 'closed'
           ELSE fi.status
       END AS text) AS status,
       fi.roadmap_summary,
       fi.updated_at
FROM feedback_story_links primary_link
INNER JOIN feedback_items fi ON fi.id = primary_link.item_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN stories projected_story ON projected_story.id = primary_link.story_id AND projected_story.workspace_id = primary_link.workspace_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
WHERE primary_link.workspace_id = sqlc.arg(workspace_id)
  AND primary_link.story_id = sqlc.arg(story_id)
  AND primary_link.is_primary = true
  AND fi.deleted_at IS NULL
  AND fi.merged_into_item_id IS NULL
ORDER BY fi.id;
