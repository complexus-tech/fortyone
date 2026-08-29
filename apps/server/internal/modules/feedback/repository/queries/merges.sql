-- name: ListFeedbackMergeCandidates :many
SELECT item.id,
       item.slug,
       item.title,
       CAST(CASE
           WHEN projected_story.id IS NULL THEN item.status
           WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
           WHEN projected_state.category = 'backlog' THEN 'reviewing'
           WHEN projected_state.category = 'unstarted' THEN 'planned'
           WHEN projected_state.category = 'started' THEN 'in_progress'
           WHEN projected_state.category = 'paused' THEN 'planned'
           WHEN projected_state.category = 'completed' THEN 'completed'
           WHEN projected_state.category = 'cancelled' THEN 'closed'
           ELSE item.status
       END AS text) AS status,
       CAST(COALESCE(votes.vote_count, 0) AS integer) AS vote_count,
       CAST(COALESCE(comments.comment_count, 0) AS integer) AS comment_count
FROM feedback_items item
INNER JOIN feedback_boards board ON board.id = item.board_id AND board.workspace_id = item.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = board.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
LEFT JOIN LATERAL (
    SELECT story_link.story_id
    FROM feedback_story_links story_link
    WHERE story_link.item_id = item.id AND story_link.is_primary = true
    ORDER BY story_link.created_at, story_link.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(vote.direction), 0) AS vote_count
    FROM feedback_votes vote
    WHERE vote.item_id = item.id
) votes ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS comment_count
    FROM feedback_comments comment_record
    WHERE comment_record.item_id = item.id
) comments ON true
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.portal_id = sqlc.arg(portal_id)
  AND (CAST(sqlc.arg(excluded_item_id) AS uuid) = '00000000-0000-0000-0000-000000000000' OR item.id <> CAST(sqlc.arg(excluded_item_id) AS uuid))
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND (
      CAST(sqlc.arg(search) AS text) = ''
      OR to_tsvector('english', item.title || ' ' || item.description || ' ' || item.slug)
         @@ websearch_to_tsquery('english', sqlc.arg(search))
  )
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR board.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ORDER BY vote_count DESC, item.created_at DESC, item.id DESC
LIMIT sqlc.arg(row_limit);

-- name: LockFeedbackMergeItems :many
SELECT item.id,
       item.workspace_id,
       item.portal_id,
       item.title,
       item.slug,
       item.merged_into_item_id,
       item.deleted_at
FROM feedback_items item
INNER JOIN feedback_boards board ON board.id = item.board_id AND board.workspace_id = item.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = board.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR board.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ORDER BY item.id
FOR UPDATE OF item;

-- name: GetCompletedFeedbackMerge :one
SELECT source_item_id,
       target_item_id,
       portal_id,
       merged_at,
       merged_by_user_id,
       CAST(COALESCE(CAST(event_payload->>'movedFollowerCount' AS integer), 0) AS integer) AS moved_follower_count,
       CAST(COALESCE(CAST(event_payload->>'movedUpdateLinkCount' AS integer), 0) AS integer) AS moved_update_link_count,
       CAST(COALESCE(CAST(event_payload->>'movedStoryLinkCount' AS integer), 0) AS integer) AS moved_story_link_count
FROM feedback_item_merge_outbox
WHERE source_item_id = sqlc.arg(source_item_id)
  AND target_item_id = sqlc.arg(target_item_id);

-- name: FeedbackItemHasInboundMerges :one
SELECT EXISTS (
    SELECT 1
    FROM feedback_items item
    WHERE item.workspace_id = sqlc.arg(workspace_id)
      AND item.portal_id = sqlc.arg(portal_id)
      AND item.merged_into_item_id = sqlc.arg(item_id)
);

-- name: ListActiveFeedbackFollowerIDs :many
SELECT follower.contributor_id
FROM feedback_item_followers follower
WHERE follower.item_id = sqlc.arg(item_id)
  AND follower.unsubscribed_at IS NULL
ORDER BY follower.contributor_id;

-- name: CountFeedbackFollowersMovedByMerge :one
SELECT CAST(COUNT(*) AS integer)
FROM feedback_item_followers source
LEFT JOIN feedback_item_followers target
    ON target.item_id = sqlc.arg(target_item_id)
   AND target.contributor_id = source.contributor_id
WHERE source.item_id = sqlc.arg(source_item_id)
  AND source.unsubscribed_at IS NULL
  AND (target.contributor_id IS NULL OR target.unsubscribed_at IS NOT NULL);

-- name: CopyFeedbackFollowersForMerge :exec
INSERT INTO feedback_item_followers (item_id, contributor_id, created_at, unsubscribed_at)
SELECT sqlc.arg(target_item_id), follower.contributor_id, follower.created_at, NULL
FROM feedback_item_followers follower
WHERE follower.item_id = sqlc.arg(source_item_id)
  AND follower.unsubscribed_at IS NULL
ON CONFLICT (item_id, contributor_id) DO UPDATE
SET unsubscribed_at = NULL,
    created_at = LEAST(feedback_item_followers.created_at, EXCLUDED.created_at);

-- name: CopyFeedbackUpdateLinksForMerge :one
WITH moved AS (
    INSERT INTO feedback_update_items (update_id, item_id)
    SELECT link.update_id, sqlc.arg(target_item_id)
    FROM feedback_update_items link
    WHERE link.item_id = sqlc.arg(source_item_id)
    ON CONFLICT (update_id, item_id) DO NOTHING
    RETURNING update_id
)
SELECT CAST(COUNT(*) AS integer) FROM moved;

-- name: GetPrimaryFeedbackStoryID :one
SELECT story_id
FROM feedback_story_links
WHERE item_id = sqlc.arg(item_id)
  AND is_primary = true;

-- name: FeedbackTargetAlreadyLinksStory :one
SELECT EXISTS (
    SELECT 1
    FROM feedback_story_links
    WHERE item_id = sqlc.arg(target_item_id)
      AND story_id = sqlc.arg(story_id)
);

-- name: MovePrimaryFeedbackStoryLink :execrows
UPDATE feedback_story_links
SET item_id = sqlc.arg(target_item_id)
WHERE item_id = sqlc.arg(source_item_id)
  AND is_primary = true;

-- name: CopyNonPrimaryFeedbackStoryLinksForMerge :one
WITH moved AS (
    INSERT INTO feedback_story_links (
        workspace_id,
        item_id,
        story_id,
        relationship,
        is_primary,
        created_by_user_id,
        created_at
    )
    SELECT link.workspace_id,
           sqlc.arg(target_item_id),
           link.story_id,
           link.relationship,
           false,
           link.created_by_user_id,
           link.created_at
    FROM feedback_story_links link
    WHERE link.item_id = sqlc.arg(source_item_id)
      AND link.is_primary = false
    ON CONFLICT (item_id, story_id) DO NOTHING
    RETURNING id
)
SELECT CAST(COUNT(*) AS integer) FROM moved;

-- name: MarkFeedbackItemMerged :one
UPDATE feedback_items source
SET merged_into_item_id = target.id,
    merged_at = NOW(),
    merged_by_user_id = sqlc.arg(actor_id),
    updated_at = NOW()
FROM feedback_items target
WHERE source.id = sqlc.arg(source_item_id)
  AND target.id = sqlc.arg(target_item_id)
  AND source.workspace_id = target.workspace_id
  AND source.portal_id = target.portal_id
  AND source.merged_into_item_id IS NULL
  AND target.merged_into_item_id IS NULL
RETURNING source.merged_at;

-- name: UnsubscribeSourceFeedbackFollowersAfterMerge :exec
UPDATE feedback_item_followers
SET unsubscribed_at = sqlc.arg(merged_at)
WHERE item_id = sqlc.arg(source_item_id)
  AND unsubscribed_at IS NULL;

-- name: InsertFeedbackMergeOutbox :exec
INSERT INTO feedback_item_merge_outbox (
    merge_event_id,
    source_item_id,
    target_item_id,
    workspace_id,
    portal_id,
    merged_by_user_id,
    merged_at,
    event_payload
)
VALUES (
    sqlc.arg(event_id),
    sqlc.arg(source_item_id),
    sqlc.arg(target_item_id),
    sqlc.arg(workspace_id),
    sqlc.arg(portal_id),
    sqlc.arg(actor_id),
    sqlc.arg(merged_at),
    CAST(sqlc.arg(event_payload) AS jsonb)
);

-- name: ClaimFeedbackMergeOutboxEvents :many
WITH candidates AS (
    SELECT candidate.merge_event_id
    FROM feedback_item_merge_outbox candidate
    WHERE (candidate.status IN ('pending', 'retrying') AND candidate.next_attempt_at <= NOW())
       OR (candidate.status = 'processing' AND candidate.claimed_at <= sqlc.arg(stale_before))
    ORDER BY COALESCE(candidate.next_attempt_at, candidate.claimed_at), candidate.created_at, candidate.merge_event_id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(row_limit)
), claimed AS (
    UPDATE feedback_item_merge_outbox outbox
    SET status = 'processing',
        attempt_count = attempt_count + 1,
        next_attempt_at = NULL,
        claim_token = gen_random_uuid(),
        claimed_at = NOW(),
        completed_at = NULL,
        last_error = NULL,
        updated_at = NOW()
    FROM candidates
    WHERE outbox.merge_event_id = candidates.merge_event_id
    RETURNING outbox.merge_event_id,
              outbox.workspace_id,
              outbox.portal_id,
              outbox.source_item_id,
              outbox.target_item_id,
              outbox.merged_by_user_id,
              outbox.merged_at,
              outbox.claim_token,
              outbox.attempt_count,
              outbox.event_payload
)
SELECT claimed.merge_event_id AS event_id,
       claimed.workspace_id,
       claimed.portal_id,
       claimed.source_item_id,
       claimed.target_item_id,
       claimed.merged_by_user_id AS actor_id,
       claimed.merged_at,
       claimed.claim_token,
       claimed.attempt_count,
       claimed.event_payload
FROM claimed
ORDER BY claimed.merged_at, claimed.merge_event_id;

-- name: ListFeedbackMergeRecipients :many
SELECT contributor.id AS contributor_id,
       contributor.user_id,
       CAST(contributor.kind AS text) AS kind
FROM feedback_contributors contributor
INNER JOIN feedback_item_followers target_follower
    ON target_follower.item_id = sqlc.arg(target_item_id)
   AND target_follower.contributor_id = contributor.id
LEFT JOIN users account ON account.user_id = contributor.user_id
LEFT JOIN feedback_contributor_preferences preference
    ON preference.portal_id = contributor.portal_id
   AND preference.contributor_id = contributor.id
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = ANY(CAST(sqlc.arg(contributor_ids) AS uuid[]))
  AND target_follower.unsubscribed_at IS NULL
  AND contributor.blocked_at IS NULL
  AND (
      (contributor.kind = 'account' AND contributor.user_id IS NOT NULL AND account.is_active = true)
      OR (
          contributor.kind IN ('verified_guest', 'external')
          AND contributor.email IS NOT NULL
          AND preference.email_unsubscribed_at IS NULL
      )
  )
ORDER BY contributor.id;

-- name: CompleteFeedbackMergeOutboxEvent :execrows
UPDATE feedback_item_merge_outbox
SET status = 'completed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NOW(),
    last_error = NULL,
    updated_at = NOW()
WHERE merge_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: RetryFeedbackMergeOutboxEvent :execrows
UPDATE feedback_item_merge_outbox
SET status = sqlc.arg(status),
    next_attempt_at = CAST(sqlc.narg(next_attempt_at) AS timestamptz),
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(sqlc.arg(failure), 4000),
    updated_at = NOW()
WHERE merge_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';
