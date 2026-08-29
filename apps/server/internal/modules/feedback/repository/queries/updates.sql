-- name: ListWorkspaceFeedbackUpdates :many
SELECT update_record.id,
       update_record.workspace_id,
       update_record.portal_id,
       update_record.slug,
       update_record.title,
       update_record.summary,
       update_record.body,
       update_record.cover_image_url,
       CAST(update_record.status AS text) AS status,
       update_record.published_at,
       update_record.published_by_user_id,
       update_record.created_at,
       update_record.updated_at
FROM feedback_updates update_record
INNER JOIN feedback_portals portal
    ON portal.id = update_record.portal_id
   AND portal.workspace_id = update_record.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = update_record.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE update_record.workspace_id = sqlc.arg(workspace_id)
ORDER BY COALESCE(update_record.published_at, update_record.updated_at) DESC, update_record.id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: ListPublicFeedbackUpdates :many
SELECT update_record.id,
       update_record.workspace_id,
       update_record.portal_id,
       update_record.slug,
       update_record.title,
       update_record.summary,
       update_record.body,
       update_record.cover_image_url,
       CAST(update_record.status AS text) AS status,
       update_record.published_at,
       update_record.published_by_user_id,
       update_record.created_at,
       update_record.updated_at
FROM feedback_updates update_record
INNER JOIN feedback_portals portal
    ON portal.id = update_record.portal_id
   AND portal.is_public = true
WHERE update_record.portal_id = sqlc.arg(portal_id)
  AND update_record.status = 'published'
  AND update_record.published_at IS NOT NULL
ORDER BY update_record.published_at DESC, update_record.id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: GetWorkspaceFeedbackUpdate :one
SELECT update_record.id,
       update_record.workspace_id,
       update_record.portal_id,
       update_record.slug,
       update_record.title,
       update_record.summary,
       update_record.body,
       update_record.cover_image_url,
       CAST(update_record.status AS text) AS status,
       update_record.published_at,
       update_record.published_by_user_id,
       update_record.created_at,
       update_record.updated_at
FROM feedback_updates update_record
INNER JOIN feedback_portals portal
    ON portal.id = update_record.portal_id
   AND portal.workspace_id = update_record.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = update_record.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
LIMIT 1;

-- name: GetPublicFeedbackUpdate :one
SELECT update_record.id,
       update_record.workspace_id,
       update_record.portal_id,
       update_record.slug,
       update_record.title,
       update_record.summary,
       update_record.body,
       update_record.cover_image_url,
       CAST(update_record.status AS text) AS status,
       update_record.published_at,
       update_record.published_by_user_id,
       update_record.created_at,
       update_record.updated_at
FROM feedback_updates update_record
INNER JOIN feedback_portals portal
    ON portal.id = update_record.portal_id
   AND portal.is_public = true
WHERE update_record.portal_id = sqlc.arg(portal_id)
  AND update_record.slug = sqlc.arg(slug)
  AND update_record.status = 'published'
  AND update_record.published_at IS NOT NULL
LIMIT 1;

-- name: CreateFeedbackUpdate :one
INSERT INTO feedback_updates (
    workspace_id,
    portal_id,
    author_id,
    title,
    body,
    status,
    slug,
    summary,
    cover_image_url
)
SELECT portal.workspace_id,
       portal.id,
       wm.user_id,
       sqlc.arg(title),
       sqlc.arg(body),
       'draft',
       sqlc.arg(slug),
       CAST(sqlc.narg(summary) AS text),
       CAST(sqlc.narg(cover_image_url) AS text)
FROM feedback_portals portal
INNER JOIN workspace_members wm
    ON wm.workspace_id = portal.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE portal.workspace_id = sqlc.arg(workspace_id)
  AND portal.id = sqlc.arg(portal_id)
RETURNING id;

-- name: UpdateDraftFeedbackUpdate :execrows
UPDATE feedback_updates update_record
SET portal_id = target_portal.id,
    title = sqlc.arg(title),
    body = sqlc.arg(body),
    slug = sqlc.arg(slug),
    summary = CAST(sqlc.narg(summary) AS text),
    cover_image_url = CAST(sqlc.narg(cover_image_url) AS text),
    updated_at = NOW()
FROM feedback_portals target_portal,
     workspace_members wm,
     users current_actor
WHERE target_portal.workspace_id = update_record.workspace_id
  AND target_portal.id = sqlc.arg(portal_id)
  AND wm.workspace_id = update_record.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
  AND update_record.status = 'draft';

-- name: DeleteDraftFeedbackUpdate :execrows
DELETE FROM feedback_updates update_record
USING workspace_members wm, users current_actor
WHERE wm.workspace_id = update_record.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
  AND update_record.status = 'draft';

-- name: UnpublishFeedbackUpdate :execrows
UPDATE feedback_updates update_record
SET status = 'draft',
    published_at = NULL,
    published_by_user_id = NULL,
    updated_at = NOW()
FROM workspace_members wm, users current_actor
WHERE wm.workspace_id = update_record.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
  AND update_record.status = 'published';

-- name: DeleteFeedbackUpdateItems :exec
DELETE FROM feedback_update_items
WHERE update_id = sqlc.arg(update_id);

-- name: InsertFeedbackUpdateItems :execrows
INSERT INTO feedback_update_items (update_id, item_id)
SELECT update_record.id, item.id
FROM feedback_updates update_record
INNER JOIN feedback_items item
    ON item.workspace_id = update_record.workspace_id
   AND item.portal_id = update_record.portal_id
   AND item.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
   AND item.deleted_at IS NULL
   AND item.merged_into_item_id IS NULL
INNER JOIN feedback_boards fb ON fb.id = item.board_id AND fb.workspace_id = item.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
WHERE update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.portal_id = sqlc.arg(portal_id)
  AND update_record.id = sqlc.arg(update_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ON CONFLICT (update_id, item_id) DO NOTHING;

-- name: ListFeedbackUpdateItems :many
SELECT link.update_id,
       item.id,
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
       END AS text) AS status
FROM feedback_update_items link
INNER JOIN feedback_items item ON item.id = link.item_id
LEFT JOIN LATERAL (
    SELECT story_link.story_id
    FROM feedback_story_links story_link
    WHERE story_link.item_id = item.id AND story_link.is_primary = true
    ORDER BY story_link.created_at, story_link.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
WHERE link.update_id = ANY(CAST(sqlc.arg(update_ids) AS uuid[]))
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
ORDER BY link.update_id, item.title ASC, item.id ASC;

-- name: LockFeedbackUpdateForPublication :one
SELECT update_record.id,
       update_record.workspace_id,
       update_record.portal_id,
       workspace.slug AS portal_slug,
       update_record.slug,
       update_record.title,
       update_record.summary,
       update_record.body,
       update_record.cover_image_url,
       CAST(update_record.status AS text) AS status,
       update_record.published_at,
       update_record.published_by_user_id,
       update_record.publication_sequence,
       update_record.created_at,
       update_record.updated_at
FROM feedback_updates update_record
INNER JOIN feedback_portals portal
    ON portal.id = update_record.portal_id
   AND portal.workspace_id = update_record.workspace_id
INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id AND workspace.deleted_at IS NULL
INNER JOIN workspace_members wm
    ON wm.workspace_id = update_record.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
WHERE update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
FOR UPDATE OF update_record;

-- name: LockFeedbackPublicationItems :many
SELECT item.id
FROM feedback_update_items link
INNER JOIN feedback_items item ON item.id = link.item_id
WHERE link.update_id = sqlc.arg(update_id)
ORDER BY item.id
FOR SHARE OF item;

-- name: ListCanonicalFeedbackPublicationItems :many
WITH canonical_links AS (
    SELECT DISTINCT link.update_id, COALESCE(source.merged_into_item_id, source.id) AS item_id
    FROM feedback_update_items link
    INNER JOIN feedback_items source ON source.id = link.item_id
    WHERE link.update_id = sqlc.arg(update_id)
      AND source.deleted_at IS NULL
)
SELECT link.update_id,
       item.id,
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
       END AS text) AS status
FROM canonical_links link
INNER JOIN feedback_items item ON item.id = link.item_id
LEFT JOIN LATERAL (
    SELECT story_link.story_id
    FROM feedback_story_links story_link
    WHERE story_link.item_id = item.id AND story_link.is_primary = true
    ORDER BY story_link.created_at, story_link.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
WHERE item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
ORDER BY item.title, item.id;

-- name: PublishFeedbackUpdate :one
UPDATE feedback_updates update_record
SET status = 'published',
    published_at = NOW(),
    published_by_user_id = sqlc.arg(actor_id),
    publication_sequence = publication_sequence + 1,
    updated_at = NOW()
WHERE update_record.workspace_id = sqlc.arg(workspace_id)
  AND update_record.id = sqlc.arg(update_id)
  AND update_record.status = 'draft'
RETURNING published_at,
          published_by_user_id,
          publication_sequence,
          updated_at;

-- name: SnapshotFeedbackPublicationContributorAudience :many
WITH linked_items AS (
    SELECT item.id
    FROM feedback_items item
    WHERE item.portal_id = sqlc.arg(portal_id)
      AND item.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
      AND item.deleted_at IS NULL
      AND item.merged_into_item_id IS NULL
), recipient_ids AS (
    SELECT follower.contributor_id
    FROM linked_items linked
    INNER JOIN feedback_item_followers follower ON follower.item_id = linked.id AND follower.unsubscribed_at IS NULL

    UNION

    SELECT follower.contributor_id
    FROM feedback_portal_followers follower
    WHERE follower.portal_id = sqlc.arg(portal_id)
      AND follower.unsubscribed_at IS NULL
)
SELECT contributor.id
FROM (SELECT DISTINCT contributor_id FROM recipient_ids) recipient
INNER JOIN feedback_contributors contributor ON contributor.id = recipient.contributor_id
LEFT JOIN feedback_contributor_preferences preference ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.kind IN ('verified_guest', 'external')
  AND contributor.email IS NOT NULL
  AND contributor.blocked_at IS NULL
  AND preference.email_unsubscribed_at IS NULL
ORDER BY contributor.id;

-- name: SnapshotFeedbackPublicationAccountAudience :many
WITH linked_items AS (
    SELECT item.id AS item_id
    FROM feedback_items item
    WHERE item.portal_id = sqlc.arg(portal_id)
      AND item.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
      AND item.deleted_at IS NULL
      AND item.merged_into_item_id IS NULL
), candidates AS (
    SELECT contributor.user_id, follower.item_id
    FROM linked_items linked
    INNER JOIN feedback_item_followers follower ON follower.item_id = linked.item_id AND follower.unsubscribed_at IS NULL
    INNER JOIN feedback_contributors contributor ON contributor.id = follower.contributor_id AND contributor.portal_id = sqlc.arg(portal_id)
    WHERE contributor.kind = 'account' AND contributor.user_id IS NOT NULL AND contributor.blocked_at IS NULL

    UNION ALL

    SELECT contributor.user_id, linked.item_id
    FROM feedback_portal_followers follower
    INNER JOIN feedback_contributors contributor ON contributor.id = follower.contributor_id AND contributor.portal_id = follower.portal_id
    CROSS JOIN LATERAL (SELECT item_id FROM linked_items ORDER BY item_id LIMIT 1) linked
    WHERE follower.portal_id = sqlc.arg(portal_id)
      AND follower.unsubscribed_at IS NULL
      AND contributor.kind = 'account'
      AND contributor.user_id IS NOT NULL
      AND contributor.blocked_at IS NULL
)
SELECT DISTINCT candidate.user_id, candidate.item_id
FROM candidates candidate
INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
ORDER BY candidate.user_id, candidate.item_id;

-- name: InsertFeedbackPublicationOutbox :exec
INSERT INTO feedback_update_publication_outbox (
    publication_event_id,
    update_id,
    workspace_id,
    portal_id,
    published_by_user_id,
    publication_sequence,
    published_at,
    event_payload
)
VALUES (
    sqlc.arg(event_id),
    sqlc.arg(update_id),
    sqlc.arg(workspace_id),
    sqlc.arg(portal_id),
    sqlc.arg(actor_id),
    sqlc.arg(publication_sequence),
    sqlc.arg(published_at),
    CAST(sqlc.arg(event_payload) AS jsonb)
);

-- name: ClaimFeedbackPublicationOutboxEvents :many
WITH candidates AS (
    SELECT publication_event_id
    FROM feedback_update_publication_outbox candidate
    WHERE (candidate.status IN ('pending', 'retrying') AND candidate.next_attempt_at <= NOW())
       OR (candidate.status = 'processing' AND candidate.claimed_at <= sqlc.arg(stale_before))
    ORDER BY COALESCE(candidate.next_attempt_at, candidate.claimed_at), candidate.created_at, candidate.publication_event_id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(row_limit)
), claimed AS (
    UPDATE feedback_update_publication_outbox outbox
    SET status = 'processing',
        attempt_count = attempt_count + 1,
        next_attempt_at = NULL,
        claim_token = gen_random_uuid(),
        claimed_at = NOW(),
        completed_at = NULL,
        last_error = NULL,
        updated_at = NOW()
    FROM candidates
    WHERE outbox.publication_event_id = candidates.publication_event_id
    RETURNING outbox.publication_event_id,
              outbox.update_id,
              outbox.workspace_id,
              outbox.portal_id,
              outbox.published_by_user_id,
              outbox.publication_sequence,
              outbox.published_at,
              outbox.claim_token,
              outbox.attempt_count,
              outbox.event_payload
)
SELECT claimed.publication_event_id,
       claimed.update_id,
       claimed.workspace_id,
       claimed.portal_id,
       claimed.published_by_user_id,
       claimed.publication_sequence,
       claimed.published_at,
       claimed.claim_token,
       claimed.attempt_count,
       claimed.event_payload
FROM claimed
ORDER BY claimed.published_at, claimed.publication_event_id;

-- name: CompleteFeedbackPublicationOutboxEvent :execrows
UPDATE feedback_update_publication_outbox
SET status = 'completed',
    next_attempt_at = NULL,
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NOW(),
    last_error = NULL,
    updated_at = NOW()
WHERE publication_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: RetryFeedbackPublicationOutboxEvent :execrows
UPDATE feedback_update_publication_outbox
SET status = sqlc.arg(status),
    next_attempt_at = CAST(sqlc.narg(next_attempt_at) AS timestamptz),
    claim_token = NULL,
    claimed_at = NULL,
    completed_at = NULL,
    last_error = LEFT(sqlc.arg(failure), 4000),
    updated_at = NOW()
WHERE publication_event_id = sqlc.arg(event_id)
  AND claim_token = sqlc.arg(claim_token)
  AND status = 'processing';

-- name: ListFeedbackPublicationDeliveryRecipients :many
SELECT contributor.id AS contributor_id,
       contributor.email,
       CAST(COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'there') AS text) AS display_name,
       CAST(contributor.kind AS text) AS kind
FROM UNNEST(CAST(sqlc.arg(contributor_ids) AS uuid[])) audience(contributor_id)
INNER JOIN feedback_contributors contributor ON contributor.id = audience.contributor_id
LEFT JOIN feedback_contributor_preferences preference ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.kind IN ('verified_guest', 'external')
  AND contributor.email IS NOT NULL
  AND contributor.blocked_at IS NULL
  AND preference.email_unsubscribed_at IS NULL
  AND (
      EXISTS (
          SELECT 1
          FROM feedback_items published_item
          INNER JOIN feedback_items canonical_item ON canonical_item.id = COALESCE(published_item.merged_into_item_id, published_item.id)
          INNER JOIN feedback_item_followers follower
              ON follower.item_id = canonical_item.id
             AND follower.contributor_id = contributor.id
             AND follower.unsubscribed_at IS NULL
          WHERE published_item.portal_id = sqlc.arg(portal_id)
            AND published_item.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
            AND canonical_item.deleted_at IS NULL
            AND canonical_item.merged_into_item_id IS NULL
      )
      OR EXISTS (
          SELECT 1
          FROM feedback_portal_followers follower
          WHERE follower.portal_id = sqlc.arg(portal_id)
            AND follower.contributor_id = contributor.id
            AND follower.unsubscribed_at IS NULL
      )
  )
ORDER BY contributor.id;

-- name: ListAccountFeedbackPublicationRecipients :many
WITH audience AS (
    SELECT users.user_id, items.item_id
    FROM UNNEST(CAST(sqlc.arg(user_ids) AS uuid[])) WITH ORDINALITY users(user_id, ordinal)
    INNER JOIN UNNEST(CAST(sqlc.arg(item_ids) AS uuid[])) WITH ORDINALITY items(item_id, ordinal)
        ON items.ordinal = users.ordinal
), current_items AS (
    SELECT audience.user_id, canonical_item.id AS item_id
    FROM audience
    INNER JOIN feedback_items published_item ON published_item.id = audience.item_id AND published_item.portal_id = sqlc.arg(portal_id)
    INNER JOIN feedback_items canonical_item ON canonical_item.id = COALESCE(published_item.merged_into_item_id, published_item.id)
    WHERE canonical_item.portal_id = sqlc.arg(portal_id)
      AND canonical_item.deleted_at IS NULL
      AND canonical_item.merged_into_item_id IS NULL
)
SELECT DISTINCT CAST(current_item.user_id AS uuid) AS user_id, current_item.item_id
FROM current_items current_item
INNER JOIN users account ON account.user_id = current_item.user_id AND account.is_active = true
WHERE EXISTS (
    SELECT 1
    FROM feedback_contributors contributor
    WHERE contributor.portal_id = sqlc.arg(portal_id)
      AND contributor.kind = 'account'
      AND contributor.user_id = current_item.user_id
      AND contributor.blocked_at IS NULL
      AND (
          EXISTS (
              SELECT 1
              FROM feedback_item_followers follower
              WHERE follower.item_id = current_item.item_id
                AND follower.contributor_id = contributor.id
                AND follower.unsubscribed_at IS NULL
          )
          OR EXISTS (
              SELECT 1
              FROM feedback_portal_followers follower
              WHERE follower.portal_id = sqlc.arg(portal_id)
                AND follower.contributor_id = contributor.id
                AND follower.unsubscribed_at IS NULL
          )
      )
)
ORDER BY current_item.user_id, current_item.item_id;

-- name: ResolveFeedbackNotificationActor :one
SELECT CAST(candidate.user_id AS uuid) AS user_id
FROM UNNEST(CAST(sqlc.arg(candidate_ids) AS uuid[])) WITH ORDINALITY candidate(user_id, priority)
INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
WHERE candidate.user_id <> '00000000-0000-0000-0000-000000000000'
ORDER BY candidate.priority
LIMIT 1;
