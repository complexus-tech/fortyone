-- name: ListFeedbackItems :many
SELECT fi.id,
       fi.workspace_id,
       fi.portal_id,
       fi.board_id,
       fi.contributor_id,
       fi.author_id,
       CAST(CASE
           WHEN contributor.kind = 'anonymous'
             OR (
                 contributor.kind IN ('verified_guest', 'external')
                 AND (
                     contributor.public_masked
                     OR portal.guest_identity_policy = 'always_mask_guests'
                 )
             ) THEN 'Anonymous'
           WHEN contributor.kind IN ('verified_guest', 'external')
               THEN COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'Guest')
           ELSE COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous')
       END AS text) AS author_name,
       CAST(CASE
           WHEN contributor.kind = 'account' THEN COALESCE(account.email, '')
           ELSE COALESCE(contributor.email, '')
       END AS text) AS author_email,
       CAST(COALESCE(CASE
           WHEN contributor.kind = 'anonymous'
             OR (
                 contributor.kind IN ('verified_guest', 'external')
                 AND (
                     contributor.public_masked
                     OR portal.guest_identity_policy = 'always_mask_guests'
                 )
             ) THEN ''
           WHEN contributor.kind = 'account' THEN account.avatar_url
           ELSE contributor.avatar_url
       END, '') AS text) AS author_avatar,
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((
           contributor.kind = 'anonymous'
           OR (
               contributor.kind IN ('verified_guest', 'external')
               AND (
                   contributor.public_masked
                   OR portal.guest_identity_policy = 'always_mask_guests'
               )
           )
       ) AS boolean) AS author_masked,
       fi.merged_into_item_id,
       fi.merged_at,
       fi.merged_by_user_id,
       false AS following,
       fi.title,
       fi.description,
       fi.description_html,
       fi.slug,
       CAST(CASE
           WHEN projected_story.id IS NULL THEN fi.status
           WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
           WHEN projected_state.category = 'backlog' THEN 'reviewing'
           WHEN projected_state.category = 'unstarted' THEN 'planned'
           WHEN projected_state.category = 'started' THEN 'in_progress'
           WHEN projected_state.category = 'paused' THEN 'planned'
           WHEN projected_state.category = 'completed' THEN 'completed'
           WHEN projected_state.category = 'cancelled' THEN 'closed'
           ELSE fi.status
       END AS text) AS status,
       CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = 1) AS integer) AS upvote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = -1) AS integer) AS downvote_count,
       CAST((SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = fi.id) AS integer) AS comment_count,
       fi.roadmap_summary,
       fb.team_id AS board_team_id,
       fb.name AS board_name,
       fb.slug AS board_slug,
       fb.color AS board_color,
       fb.order_index AS board_order_index,
       fb.created_at AS board_created_at,
       fb.updated_at AS board_updated_at,
       COALESCE(primary_link.id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_link_id,
       COALESCE(primary_link.story_id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_story_id,
       projected_story.title AS primary_story_title,
       COALESCE(primary_link.relationship, '') AS primary_relationship,
       primary_link.created_by_user_id AS primary_created_by_user_id,
       COALESCE(primary_link.created_at, CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS primary_created_at,
       feedback_read.read_at,
       fi.deleted_at,
       fi.created_at,
       fi.updated_at
FROM feedback_items fi
INNER JOIN feedback_contributors contributor
    ON contributor.id = fi.contributor_id
   AND contributor.portal_id = fi.portal_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id
LEFT JOIN users account ON account.user_id = fi.author_id
INNER JOIN feedback_boards fb
    ON fb.id = fi.board_id
   AND fb.workspace_id = fi.workspace_id
   AND fb.portal_id = fi.portal_id
LEFT JOIN LATERAL (
    SELECT fsl.id,
           fsl.story_id,
           fsl.relationship,
           fsl.created_by_user_id,
           fsl.created_at
    FROM feedback_story_links fsl
    WHERE fsl.item_id = fi.id
      AND fsl.is_primary = true
    ORDER BY fsl.created_at, fsl.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
LEFT JOIN feedback_item_reads feedback_read
    ON feedback_read.item_id = fi.id
   AND feedback_read.user_id = CAST(sqlc.narg(viewer_id) AS uuid)
WHERE (NOT CAST(sqlc.arg(filter_portal) AS boolean) OR fi.portal_id = sqlc.arg(portal_id))
  AND (NOT CAST(sqlc.arg(filter_item) AS boolean) OR fi.id = sqlc.arg(item_id))
  AND (NOT CAST(sqlc.arg(filter_team) AS boolean) OR (
      fi.workspace_id = sqlc.arg(workspace_id)
      AND fb.team_id = sqlc.arg(team_id)
  ))
  AND (
      (
          CAST(sqlc.arg(deleted_only) AS boolean)
          AND fi.deleted_at IS NOT NULL
          AND fi.deleted_at >= sqlc.arg(recovery_cutoff)
          AND fi.merged_into_item_id IS NULL
      )
      OR (
          NOT CAST(sqlc.arg(deleted_only) AS boolean)
          AND fi.deleted_at IS NULL
          AND fi.merged_into_item_id IS NULL
      )
  )
  AND (
      CAST(sqlc.arg(status_mode) AS smallint) = 0
      OR (
          CAST(sqlc.arg(status_mode) AS smallint) = 1
          AND CAST(CASE
              WHEN projected_story.id IS NULL THEN fi.status
              WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
              WHEN projected_state.category = 'backlog' THEN 'reviewing'
              WHEN projected_state.category = 'unstarted' THEN 'planned'
              WHEN projected_state.category = 'started' THEN 'in_progress'
              WHEN projected_state.category = 'paused' THEN 'planned'
              WHEN projected_state.category = 'completed' THEN 'completed'
              WHEN projected_state.category = 'cancelled' THEN 'closed'
              ELSE fi.status
          END AS text) IN ('pending', 'reviewing', 'planned', 'in_progress')
      )
      OR (
          CAST(sqlc.arg(status_mode) AS smallint) = 2
          AND CAST(CASE
              WHEN projected_story.id IS NULL THEN fi.status
              WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
              WHEN projected_state.category = 'backlog' THEN 'reviewing'
              WHEN projected_state.category = 'unstarted' THEN 'planned'
              WHEN projected_state.category = 'started' THEN 'in_progress'
              WHEN projected_state.category = 'paused' THEN 'planned'
              WHEN projected_state.category = 'completed' THEN 'completed'
              WHEN projected_state.category = 'cancelled' THEN 'closed'
              ELSE fi.status
          END AS text) = CAST(sqlc.arg(status) AS text)
      )
  )
  AND (NOT CAST(sqlc.arg(filter_board) AS boolean) OR fi.board_id = sqlc.arg(board_id))
  AND (NOT CAST(sqlc.arg(filter_author) AS boolean) OR fi.author_id = sqlc.arg(author_id))
  AND (
      NOT CAST(sqlc.arg(filter_search) AS boolean)
      OR to_tsvector('english', fi.title || ' ' || fi.description || ' ' || fi.slug)
         @@ websearch_to_tsquery('english', sqlc.arg(search))
  )
  AND (
      NOT CAST(sqlc.arg(require_member) AS boolean)
      OR EXISTS (
          SELECT 1
          FROM workspace_members wm
          INNER JOIN team_members tm
              ON tm.team_id = fb.team_id
             AND tm.user_id = wm.user_id
          INNER JOIN users current_actor
              ON current_actor.user_id = wm.user_id
             AND current_actor.is_active = true
             AND current_actor.is_system = false
          WHERE wm.workspace_id = fi.workspace_id
            AND wm.user_id = sqlc.arg(actor_id)
            AND wm.role IN ('admin', 'member')
            AND (
                CAST(sqlc.arg(all_teams) AS boolean)
                OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[]))
            )
      )
  )
ORDER BY
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'newest' THEN fi.created_at END DESC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'oldest' THEN fi.created_at END ASC,
    CASE WHEN CAST(sqlc.arg(sort_key) AS text) = 'top' THEN CAST(COALESCE((SELECT SUM(v.direction) FROM feedback_votes v WHERE v.item_id = fi.id), 0) AS integer) END DESC,
    fi.created_at DESC,
    fi.id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: GetWorkspaceFeedbackItem :one
SELECT fi.id,
       fi.workspace_id,
       fi.portal_id,
       fi.board_id,
       fi.contributor_id,
       fi.author_id,
       CAST(CASE
           WHEN contributor.kind = 'anonymous'
             OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN 'Anonymous'
           WHEN contributor.kind IN ('verified_guest', 'external') THEN COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'Guest')
           ELSE COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous')
       END AS text) AS author_name,
       CAST(CASE WHEN contributor.kind = 'account' THEN COALESCE(account.email, '') ELSE COALESCE(contributor.email, '') END AS text) AS author_email,
       CAST(COALESCE(CASE
           WHEN contributor.kind = 'anonymous'
             OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN ''
           WHEN contributor.kind = 'account' THEN account.avatar_url
           ELSE contributor.avatar_url
       END, '') AS text) AS author_avatar,
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       fi.merged_into_item_id,
       fi.merged_at,
       fi.merged_by_user_id,
       false AS following,
       fi.title,
       fi.description,
       fi.description_html,
       fi.slug,
       CAST(CASE
           WHEN projected_story.id IS NULL THEN fi.status
           WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
           WHEN projected_state.category = 'backlog' THEN 'reviewing'
           WHEN projected_state.category = 'unstarted' THEN 'planned'
           WHEN projected_state.category = 'started' THEN 'in_progress'
           WHEN projected_state.category = 'paused' THEN 'planned'
           WHEN projected_state.category = 'completed' THEN 'completed'
           WHEN projected_state.category = 'cancelled' THEN 'closed'
           ELSE fi.status
       END AS text) AS status,
       CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = 1) AS integer) AS upvote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = -1) AS integer) AS downvote_count,
       CAST((SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = fi.id) AS integer) AS comment_count,
       fi.roadmap_summary,
       fb.team_id AS board_team_id,
       fb.name AS board_name,
       fb.slug AS board_slug,
       fb.color AS board_color,
       fb.order_index AS board_order_index,
       fb.created_at AS board_created_at,
       fb.updated_at AS board_updated_at,
       COALESCE(primary_link.id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_link_id,
       COALESCE(primary_link.story_id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_story_id,
       projected_story.title AS primary_story_title,
       COALESCE(primary_link.relationship, '') AS primary_relationship,
       primary_link.created_by_user_id AS primary_created_by_user_id,
       COALESCE(primary_link.created_at, CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS primary_created_at,
       CAST(NULL AS timestamptz) AS read_at,
       fi.deleted_at,
       fi.created_at,
       fi.updated_at
FROM feedback_items fi
INNER JOIN feedback_contributors contributor ON contributor.id = fi.contributor_id AND contributor.portal_id = fi.portal_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id
LEFT JOIN users account ON account.user_id = fi.author_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id AND fb.portal_id = fi.portal_id
LEFT JOIN LATERAL (
    SELECT fsl.id, fsl.story_id, fsl.relationship, fsl.created_by_user_id, fsl.created_at
    FROM feedback_story_links fsl
    WHERE fsl.item_id = fi.id AND fsl.is_primary = true
    ORDER BY fsl.created_at, fsl.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
WHERE fi.workspace_id = sqlc.arg(workspace_id)
  AND fi.id = sqlc.arg(item_id)
  AND fi.deleted_at IS NULL;

-- name: GetPublicFeedbackItem :one
SELECT fi.id,
       fi.workspace_id,
       fi.portal_id,
       fi.board_id,
       fi.contributor_id,
       fi.author_id,
       CAST(CASE
           WHEN contributor.kind = 'anonymous'
             OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN 'Anonymous'
           WHEN contributor.kind IN ('verified_guest', 'external') THEN COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'Guest')
           ELSE COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous')
       END AS text) AS author_name,
       CAST(CASE WHEN contributor.kind = 'account' THEN COALESCE(account.email, '') ELSE COALESCE(contributor.email, '') END AS text) AS author_email,
       CAST(COALESCE(CASE
           WHEN contributor.kind = 'anonymous'
             OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN ''
           WHEN contributor.kind = 'account' THEN account.avatar_url
           ELSE contributor.avatar_url
       END, '') AS text) AS author_avatar,
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       fi.merged_into_item_id,
       fi.merged_at,
       fi.merged_by_user_id,
       false AS following,
       fi.title,
       fi.description,
       fi.description_html,
       fi.slug,
       CAST(CASE
           WHEN projected_story.id IS NULL THEN fi.status
           WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
           WHEN projected_state.category = 'backlog' THEN 'reviewing'
           WHEN projected_state.category = 'unstarted' THEN 'planned'
           WHEN projected_state.category = 'started' THEN 'in_progress'
           WHEN projected_state.category = 'paused' THEN 'planned'
           WHEN projected_state.category = 'completed' THEN 'completed'
           WHEN projected_state.category = 'cancelled' THEN 'closed'
           ELSE fi.status
       END AS text) AS status,
       CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = 1) AS integer) AS upvote_count,
       CAST((SELECT COUNT(*) FROM feedback_votes fv WHERE fv.item_id = fi.id AND fv.direction = -1) AS integer) AS downvote_count,
       CAST((SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = fi.id) AS integer) AS comment_count,
       fi.roadmap_summary,
       fb.team_id AS board_team_id,
       fb.name AS board_name,
       fb.slug AS board_slug,
       fb.color AS board_color,
       fb.order_index AS board_order_index,
       fb.created_at AS board_created_at,
       fb.updated_at AS board_updated_at,
       COALESCE(primary_link.id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_link_id,
       COALESCE(primary_link.story_id, CAST('00000000-0000-0000-0000-000000000000' AS uuid)) AS primary_story_id,
       projected_story.title AS primary_story_title,
       COALESCE(primary_link.relationship, '') AS primary_relationship,
       primary_link.created_by_user_id AS primary_created_by_user_id,
       COALESCE(primary_link.created_at, CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS primary_created_at,
       CAST(NULL AS timestamptz) AS read_at,
       fi.deleted_at,
       fi.created_at,
       fi.updated_at
FROM feedback_items fi
INNER JOIN feedback_contributors contributor ON contributor.id = fi.contributor_id AND contributor.portal_id = fi.portal_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id AND portal.is_public = true
LEFT JOIN users account ON account.user_id = fi.author_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id AND fb.portal_id = fi.portal_id
LEFT JOIN LATERAL (
    SELECT fsl.id, fsl.story_id, fsl.relationship, fsl.created_by_user_id, fsl.created_at
    FROM feedback_story_links fsl
    WHERE fsl.item_id = fi.id AND fsl.is_primary = true
    ORDER BY fsl.created_at, fsl.id
    LIMIT 1
) primary_link ON true
LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
WHERE fi.portal_id = sqlc.arg(portal_id)
  AND fi.id = sqlc.arg(item_id)
  AND fi.deleted_at IS NULL
  AND fi.merged_into_item_id IS NULL;

-- name: CreateFeedbackItem :one
INSERT INTO feedback_items (
    workspace_id,
    portal_id,
    board_id,
    contributor_id,
    author_id,
    title,
    description,
    description_html,
    slug,
    submission_source
)
SELECT fp.workspace_id,
       fp.id,
       fb.id,
       contributor.id,
       contributor.user_id,
       sqlc.arg(title),
       sqlc.arg(description),
       sqlc.arg(description_html),
       sqlc.arg(slug),
       sqlc.arg(submission_source)
FROM feedback_portals fp
INNER JOIN feedback_boards fb
    ON fb.portal_id = fp.id
   AND fb.workspace_id = fp.workspace_id
   AND fb.id = sqlc.arg(board_id)
INNER JOIN feedback_contributors contributor
    ON contributor.portal_id = fp.id
   AND contributor.id = sqlc.arg(contributor_id)
   AND contributor.blocked_at IS NULL
WHERE fp.workspace_id = sqlc.arg(workspace_id)
  AND fp.id = sqlc.arg(portal_id)
  AND (
      NOT CAST(sqlc.arg(require_actor) AS boolean)
      OR (
          contributor.kind = 'account'
          AND contributor.user_id = sqlc.arg(actor_id)
          AND EXISTS (
              SELECT 1
              FROM workspace_members wm
              INNER JOIN team_members tm
                  ON tm.team_id = fb.team_id
                 AND tm.user_id = wm.user_id
              INNER JOIN users current_actor
                  ON current_actor.user_id = wm.user_id
                 AND current_actor.is_active = true
                 AND current_actor.is_system = false
              WHERE wm.workspace_id = fp.workspace_id
                AND wm.user_id = sqlc.arg(actor_id)
                AND wm.role IN ('admin', 'member')
          )
      )
  )
RETURNING id;

-- name: LinkFeedbackItemAttachment :one
INSERT INTO feedback_item_attachments (item_id, attachment_id)
SELECT item.id, attachment.attachment_id
FROM feedback_items AS item
INNER JOIN attachments AS attachment
    ON attachment.attachment_id = sqlc.arg(attachment_id)
   AND attachment.workspace_id = item.workspace_id
WHERE item.id = sqlc.arg(item_id)
  AND item.portal_id = sqlc.arg(portal_id)
  AND item.deleted_at IS NULL
ON CONFLICT (item_id, attachment_id) DO NOTHING
RETURNING attachment_id;

-- name: GetFeedbackItemAttachment :one
SELECT attachment.attachment_id,
       relation.item_id,
       attachment.workspace_id,
       attachment.filename,
       attachment.size,
       attachment.mime_type,
       relation.created_at
FROM feedback_item_attachments AS relation
INNER JOIN feedback_items AS item ON item.id = relation.item_id
INNER JOIN attachments AS attachment
    ON attachment.attachment_id = relation.attachment_id
   AND attachment.workspace_id = item.workspace_id
WHERE item.portal_id = sqlc.arg(portal_id)
  AND item.id = sqlc.arg(item_id)
  AND attachment.attachment_id = sqlc.arg(attachment_id)
  AND item.deleted_at IS NULL;

-- name: ListFeedbackItemAttachments :many
SELECT attachment.attachment_id,
       relation.item_id,
       attachment.workspace_id,
       attachment.filename,
       attachment.size,
       attachment.mime_type,
       relation.created_at
FROM feedback_item_attachments AS relation
INNER JOIN feedback_items AS item ON item.id = relation.item_id
INNER JOIN attachments AS attachment
    ON attachment.attachment_id = relation.attachment_id
   AND attachment.workspace_id = item.workspace_id
WHERE item.portal_id = sqlc.arg(portal_id)
  AND relation.item_id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
  AND item.deleted_at IS NULL
ORDER BY relation.created_at, attachment.attachment_id;

-- name: CreateAnonymousFeedbackContributor :one
INSERT INTO feedback_contributors (portal_id, kind)
SELECT fp.id, 'anonymous'
FROM feedback_portals fp
WHERE fp.id = sqlc.arg(portal_id)
  AND fp.is_public = true
RETURNING id;

-- name: GetOrCreateAccountFeedbackContributor :one
INSERT INTO feedback_contributors (portal_id, user_id, kind)
SELECT fp.id, account.user_id, 'account'
FROM feedback_portals fp
INNER JOIN users account
    ON account.user_id = sqlc.arg(user_id)
   AND account.is_active = true
   AND account.is_system = false
WHERE fp.id = sqlc.arg(portal_id)
ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL
DO UPDATE SET updated_at = NOW()
RETURNING id, portal_id, user_id, CAST(kind AS text) AS kind, created_at;

-- name: UpdateFeedbackItemStatus :one
WITH previous AS (
    SELECT fi.id, fi.status
    FROM feedback_items fi
    INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
    INNER JOIN workspace_members wm
        ON wm.workspace_id = fi.workspace_id
       AND wm.user_id = sqlc.arg(actor_id)
       AND wm.role IN ('admin', 'member')
    INNER JOIN team_members tm
        ON tm.team_id = fb.team_id
       AND tm.user_id = wm.user_id
    INNER JOIN users current_actor
        ON current_actor.user_id = wm.user_id
       AND current_actor.is_active = true
       AND current_actor.is_system = false
    WHERE fi.workspace_id = sqlc.arg(workspace_id)
      AND fi.id = sqlc.arg(item_id)
      AND fi.deleted_at IS NULL
      AND fi.merged_into_item_id IS NULL
      AND (
          CAST(sqlc.arg(allow_linked) AS boolean)
          OR NOT EXISTS (
              SELECT 1
              FROM feedback_story_links fsl
              WHERE fsl.item_id = fi.id
                AND fsl.is_primary = true
          )
      )
      AND (
          NOT CAST(sqlc.arg(require_unchanged) AS boolean)
          OR fi.updated_at = sqlc.arg(expected_updated_at)
      )
    FOR UPDATE OF fi
), updated AS (
    UPDATE feedback_items fi
    SET status = sqlc.arg(status),
        roadmap_summary = COALESCE(CAST(sqlc.narg(roadmap_summary) AS text), fi.roadmap_summary),
        updated_at = NOW()
    FROM previous
    WHERE fi.id = previous.id
    RETURNING fi.id, previous.status AS previous_status, fi.status
)
SELECT updated.id,
       updated.previous_status IS DISTINCT FROM updated.status AS status_changed
FROM updated;

-- name: FeedbackItemStoryManaged :one
SELECT EXISTS (
    SELECT 1
    FROM feedback_story_links fsl
    WHERE fsl.workspace_id = sqlc.arg(workspace_id)
      AND fsl.item_id = sqlc.arg(item_id)
      AND fsl.is_primary = true
);

-- name: TrashFeedbackItem :execrows
UPDATE feedback_items fi
SET deleted_at = NOW(),
    updated_at = NOW()
FROM feedback_boards fb,
     workspace_members wm,
     team_members tm,
     users current_actor
WHERE fb.id = fi.board_id
  AND fb.workspace_id = fi.workspace_id
  AND wm.workspace_id = fi.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND tm.team_id = fb.team_id
  AND tm.user_id = wm.user_id
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND fi.workspace_id = sqlc.arg(workspace_id)
  AND fi.id = sqlc.arg(item_id)
  AND fi.deleted_at IS NULL
  AND fi.merged_into_item_id IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM feedback_items merged_source
      WHERE merged_source.workspace_id = fi.workspace_id
        AND merged_source.portal_id = fi.portal_id
        AND merged_source.merged_into_item_id = fi.id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM feedback_story_links fsl
      WHERE fsl.item_id = fi.id
        AND fsl.is_primary = true
  );

-- name: GetFeedbackItemProtection :one
SELECT EXISTS (
           SELECT 1
           FROM feedback_items item
           WHERE item.workspace_id = sqlc.arg(workspace_id)
             AND item.id = sqlc.arg(item_id)
             AND (
                 item.merged_into_item_id IS NOT NULL
                 OR EXISTS (
                     SELECT 1
                     FROM feedback_items merged_source
                     WHERE merged_source.workspace_id = item.workspace_id
                       AND merged_source.portal_id = item.portal_id
                       AND merged_source.merged_into_item_id = item.id
                 )
             )
       ) AS merge_protected,
       EXISTS (
           SELECT 1
           FROM feedback_story_links fsl
           WHERE fsl.workspace_id = sqlc.arg(workspace_id)
             AND fsl.item_id = sqlc.arg(item_id)
             AND fsl.is_primary = true
       ) AS story_managed;

-- name: RestoreFeedbackItem :execrows
UPDATE feedback_items fi
SET deleted_at = NULL,
    updated_at = NOW()
FROM feedback_boards fb,
     workspace_members wm,
     team_members tm,
     users current_actor
WHERE fb.id = fi.board_id
  AND fb.workspace_id = fi.workspace_id
  AND wm.workspace_id = fi.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND tm.team_id = fb.team_id
  AND tm.user_id = wm.user_id
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND fi.workspace_id = sqlc.arg(workspace_id)
  AND fi.id = sqlc.arg(item_id)
  AND fi.deleted_at IS NOT NULL
  AND fi.merged_into_item_id IS NULL
  AND fi.deleted_at >= sqlc.arg(recovery_cutoff);

-- name: ResolveCanonicalFeedbackItem :one
SELECT canonical.id AS item_id,
       canonical.slug AS item_slug,
       CAST(source.merged_into_item_id IS NOT NULL AS boolean) AS merged
FROM feedback_items source
INNER JOIN feedback_items canonical
    ON canonical.id = COALESCE(source.merged_into_item_id, source.id)
   AND canonical.portal_id = source.portal_id
INNER JOIN feedback_portals portal
    ON portal.id = source.portal_id
   AND portal.is_public = true
WHERE source.portal_id = sqlc.arg(portal_id)
  AND (
      source.id = CASE
          WHEN CAST(sqlc.arg(item_reference) AS text) ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
              THEN CAST(sqlc.arg(item_reference) AS uuid)
          ELSE CAST(NULL AS uuid)
      END
      OR source.slug = CAST(sqlc.arg(item_reference) AS text)
  )
  AND source.deleted_at IS NULL
  AND canonical.deleted_at IS NULL;

-- name: ListSimilarFeedbackItems :many
WITH input AS (
    SELECT COALESCE(STRING_AGG(token, ' ' ORDER BY token_order), '') AS normalized_title
    FROM UNNEST(regexp_split_to_array(LOWER(sqlc.arg(title)), '[^a-z0-9]+')) WITH ORDINALITY AS tokens(token, token_order)
    WHERE token <> ''
      AND token NOT IN ('a', 'add', 'an', 'build', 'create', 'fix', 'for', 'implement', 'make', 'new', 'please', 'story', 'support', 'task', 'the', 'to', 'update')
), ranked AS (
    SELECT fi.id,
           fi.slug,
           fi.title,
           CASE WHEN contributor.kind = 'account' THEN fi.author_id ELSE CAST(NULL AS uuid) END AS author_id,
           CAST(CASE
               WHEN contributor.kind = 'anonymous'
                 OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN 'Anonymous'
               WHEN contributor.kind IN ('verified_guest', 'external') THEN COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'Guest')
               ELSE COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous')
           END AS text) AS author_name,
           CAST(COALESCE(CASE
               WHEN contributor.kind = 'anonymous'
                 OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN ''
               WHEN contributor.kind = 'account' THEN account.avatar_url
               ELSE contributor.avatar_url
           END, '') AS text) AS author_avatar,
           CAST(CASE
               WHEN projected_story.id IS NULL THEN fi.status
               WHEN projected_story.deleted_at IS NOT NULL THEN 'closed'
               WHEN projected_state.category = 'backlog' THEN 'reviewing'
               WHEN projected_state.category = 'unstarted' THEN 'planned'
               WHEN projected_state.category = 'started' THEN 'in_progress'
               WHEN projected_state.category = 'paused' THEN 'planned'
               WHEN projected_state.category = 'completed' THEN 'completed'
               WHEN projected_state.category = 'cancelled' THEN 'closed'
               ELSE fi.status
           END AS text) AS status,
           CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_count,
           CAST((SELECT COUNT(*) FROM feedback_comments fc WHERE fc.item_id = fi.id) AS integer) AS comment_count,
           CAST(GREATEST(
               similarity(normalized.normalized_title, input.normalized_title),
               CASE WHEN normalized.normalized_title = input.normalized_title THEN 1.0 ELSE 0.0 END,
               CASE
                   WHEN CAST(sqlc.arg(description) AS text) <> ''
                       THEN similarity(LOWER(normalized.normalized_title || ' ' || fi.description), LOWER(input.normalized_title || ' ' || CAST(sqlc.arg(description) AS text)))
                   ELSE 0.0
               END
           ) AS double precision) AS confidence
    FROM feedback_items fi
    INNER JOIN feedback_contributors contributor ON contributor.id = fi.contributor_id
    INNER JOIN feedback_portals portal ON portal.id = fi.portal_id AND portal.is_public = true
    LEFT JOIN users account ON account.user_id = fi.author_id
    LEFT JOIN LATERAL (
        SELECT fsl.id, fsl.story_id
        FROM feedback_story_links fsl
        WHERE fsl.item_id = fi.id AND fsl.is_primary = true
        ORDER BY fsl.created_at, fsl.id
        LIMIT 1
    ) primary_link ON true
    LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
    LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
    CROSS JOIN input
    CROSS JOIN LATERAL (
        SELECT COALESCE(STRING_AGG(token, ' ' ORDER BY token_order), '') AS normalized_title
        FROM UNNEST(regexp_split_to_array(LOWER(fi.title), '[^a-z0-9]+')) WITH ORDINALITY AS tokens(token, token_order)
        WHERE token <> ''
          AND token NOT IN ('a', 'add', 'an', 'build', 'create', 'fix', 'for', 'implement', 'make', 'new', 'please', 'story', 'support', 'task', 'the', 'to', 'update')
    ) normalized
    WHERE fi.portal_id = sqlc.arg(portal_id)
      AND fi.deleted_at IS NULL
      AND fi.merged_into_item_id IS NULL
      AND input.normalized_title <> ''
      AND normalized.normalized_title <> ''
)
SELECT ranked.id,
       ranked.slug,
       ranked.title,
       ranked.author_id,
       ranked.author_name,
       ranked.author_avatar,
       ranked.status,
       ranked.vote_count,
       ranked.comment_count,
       ranked.confidence
FROM ranked
WHERE ranked.confidence >= 0.45
ORDER BY ranked.confidence DESC, ranked.vote_count DESC, ranked.title ASC, ranked.id ASC
LIMIT sqlc.arg(row_limit);
