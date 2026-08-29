-- name: ListContributorFeedbackActivity :many
WITH activities AS (
    SELECT fi.id,
           'feedback' AS activity_type,
           fi.id AS feedback_id,
           fi.title AS feedback_title,
           fi.slug AS feedback_slug,
           LEFT(fi.description, 500) AS body,
           fb.name AS board_name,
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
           w.slug AS portal_slug,
           w.name AS workspace_name,
           w.slug AS workspace_slug,
           fi.created_at
    FROM feedback_items fi
    INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
    INNER JOIN feedback_boards fb ON fb.id = fi.board_id
    INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id AND w.deleted_at IS NULL
    LEFT JOIN LATERAL (
        SELECT fsl.story_id
        FROM feedback_story_links fsl
        WHERE fsl.item_id = fi.id AND fsl.is_primary = true
        ORDER BY fsl.created_at, fsl.id
        LIMIT 1
    ) primary_link ON true
    LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
    LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
    WHERE fi.author_id = sqlc.arg(user_id)
      AND fi.deleted_at IS NULL

    UNION ALL

    SELECT fc.id,
           'comment' AS activity_type,
           fi.id AS feedback_id,
           fi.title AS feedback_title,
           fi.slug AS feedback_slug,
           LEFT(fc.body, 500) AS body,
           '' AS board_name,
           '' AS status,
           CAST(0 AS integer) AS vote_count,
           CAST(0 AS integer) AS comment_count,
           w.slug AS portal_slug,
           w.name AS workspace_name,
           w.slug AS workspace_slug,
           fc.created_at
    FROM feedback_comments fc
    INNER JOIN feedback_items fi ON fi.id = fc.item_id AND fi.deleted_at IS NULL
    INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
    INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id AND w.deleted_at IS NULL
    WHERE fc.author_id = sqlc.arg(user_id)
)
SELECT activities.id,
       activities.activity_type,
       activities.feedback_id,
       activities.feedback_title,
       activities.feedback_slug,
       activities.body,
       activities.board_name,
       activities.status,
       activities.vote_count,
       activities.comment_count,
       activities.portal_slug,
       activities.workspace_name,
       activities.workspace_slug,
       activities.created_at
FROM activities
WHERE CAST(sqlc.arg(activity_type) AS text) = ''
   OR activities.activity_type = CAST(sqlc.arg(activity_type) AS text)
ORDER BY activities.created_at DESC, activities.id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: GetContributorFeedbackActivityStats :one
WITH contributions AS (
    SELECT 'feedback' AS activity_type,
           fi.portal_id,
           CAST(COALESCE((SELECT SUM(fv.direction) FROM feedback_votes fv WHERE fv.item_id = fi.id), 0) AS integer) AS vote_score
    FROM feedback_items fi
    INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
    INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id AND w.deleted_at IS NULL
    WHERE fi.author_id = sqlc.arg(user_id)
      AND fi.deleted_at IS NULL

    UNION ALL

    SELECT 'comment' AS activity_type,
           fi.portal_id,
           CAST(0 AS integer) AS vote_score
    FROM feedback_comments fc
    INNER JOIN feedback_items fi ON fi.id = fc.item_id AND fi.deleted_at IS NULL
    INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
    INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id AND w.deleted_at IS NULL
    WHERE fc.author_id = sqlc.arg(user_id)
)
SELECT CAST(COUNT(*) FILTER (WHERE activity_type = 'feedback') AS integer) AS feedback_count,
       CAST(COUNT(*) FILTER (WHERE activity_type = 'comment') AS integer) AS comment_count,
       CAST(COALESCE(SUM(vote_score), 0) AS integer) AS vote_score,
       CAST(COUNT(DISTINCT portal_id) AS integer) AS portal_count
FROM contributions;

-- name: GetPublicFeedbackContributor :one
SELECT account.user_id AS id,
       CAST(COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous') AS text) AS name,
       account.avatar_url,
       account.created_at AS joined_at,
       CAST((
           SELECT COUNT(*)
           FROM feedback_items authored
           WHERE authored.portal_id = sqlc.arg(portal_id)
             AND authored.author_id = account.user_id
             AND authored.deleted_at IS NULL
       ) AS integer) AS feedback_count,
       CAST((
           SELECT COUNT(*)
           FROM feedback_comments authored_comment
           INNER JOIN feedback_items item ON item.id = authored_comment.item_id
           WHERE item.portal_id = sqlc.arg(portal_id)
             AND authored_comment.author_id = account.user_id
             AND item.deleted_at IS NULL
       ) AS integer) AS comment_count,
       CAST(COALESCE((
           SELECT SUM(received_vote.direction)
           FROM feedback_votes received_vote
           INNER JOIN feedback_items voted_item ON voted_item.id = received_vote.item_id
           WHERE voted_item.portal_id = sqlc.arg(portal_id)
             AND voted_item.author_id = account.user_id
             AND voted_item.deleted_at IS NULL
       ), 0) AS integer) AS vote_score
FROM users account
INNER JOIN feedback_portals portal
    ON portal.id = sqlc.arg(portal_id)
   AND portal.is_public = true
WHERE account.user_id = sqlc.arg(author_id)
  AND account.is_active = true
  AND (
      EXISTS (
          SELECT 1
          FROM feedback_items contributed
          WHERE contributed.portal_id = portal.id
            AND contributed.author_id = account.user_id
            AND contributed.deleted_at IS NULL
      )
      OR EXISTS (
          SELECT 1
          FROM feedback_comments contributed_comment
          INNER JOIN feedback_items item ON item.id = contributed_comment.item_id
          WHERE item.portal_id = portal.id
            AND contributed_comment.author_id = account.user_id
            AND item.deleted_at IS NULL
      )
  );

-- name: PublicFeedbackContributorExists :one
SELECT EXISTS (
    SELECT 1
    FROM feedback_portals portal
    WHERE portal.id = sqlc.arg(portal_id)
      AND portal.is_public = true
      AND (
          EXISTS (
              SELECT 1
              FROM feedback_items item
              WHERE item.portal_id = portal.id
                AND item.author_id = sqlc.arg(author_id)
                AND item.deleted_at IS NULL
          )
          OR EXISTS (
              SELECT 1
              FROM feedback_comments comment_record
              INNER JOIN feedback_items item ON item.id = comment_record.item_id
              WHERE item.portal_id = portal.id
                AND comment_record.author_id = sqlc.arg(author_id)
                AND item.deleted_at IS NULL
          )
      )
);

-- name: ListPublicContributorComments :many
SELECT fc.id,
       fc.item_id,
       fi.title AS feedback_title,
       fi.slug AS feedback_slug,
       fc.body,
       fc.created_at,
       fc.updated_at
FROM feedback_comments fc
INNER JOIN feedback_items fi ON fi.id = fc.item_id
INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.is_public = true
WHERE fi.portal_id = sqlc.arg(portal_id)
  AND fc.author_id = sqlc.arg(author_id)
  AND fi.deleted_at IS NULL
ORDER BY fc.created_at DESC, fc.id DESC
LIMIT sqlc.arg(row_limit)
OFFSET sqlc.arg(row_offset);

-- name: ListPublicFeedbackComments :many
SELECT fc.id,
       fc.workspace_id,
       fc.item_id,
       fc.author_id,
       fc.contributor_id,
       fc.parent_id,
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
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       fc.body,
       fc.created_at,
       fc.updated_at
FROM feedback_comments fc
INNER JOIN feedback_items fi ON fi.id = fc.item_id
INNER JOIN feedback_contributors contributor ON contributor.id = fc.contributor_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id AND portal.is_public = true
LEFT JOIN users account ON account.user_id = fc.author_id
WHERE fi.portal_id = sqlc.arg(portal_id)
  AND fi.id = ANY(CAST(sqlc.arg(item_ids) AS uuid[]))
  AND fi.deleted_at IS NULL
ORDER BY fc.created_at DESC, fc.id DESC;

-- name: ListInternalFeedbackComments :many
SELECT fc.id,
       fc.workspace_id,
       fc.item_id,
       fc.author_id,
       fc.contributor_id,
       fc.parent_id,
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
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       fc.body,
       fc.created_at,
       fc.updated_at
FROM feedback_comments fc
INNER JOIN feedback_items fi ON fi.id = fc.item_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN feedback_contributors contributor ON contributor.id = fc.contributor_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id
LEFT JOIN users account ON account.user_id = fc.author_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fi.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE fc.workspace_id = sqlc.arg(workspace_id)
  AND fc.item_id = sqlc.arg(item_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ORDER BY fc.created_at DESC, fc.id DESC;

-- name: GetInternalFeedbackComment :one
SELECT fc.id,
       fc.workspace_id,
       fc.item_id,
       fc.author_id,
       fc.contributor_id,
       fc.parent_id,
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
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       fc.body,
       fc.created_at,
       fc.updated_at
FROM feedback_comments fc
INNER JOIN feedback_items fi ON fi.id = fc.item_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN feedback_contributors contributor ON contributor.id = fc.contributor_id
INNER JOIN feedback_portals portal ON portal.id = fi.portal_id
LEFT JOIN users account ON account.user_id = fc.author_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fi.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE fc.workspace_id = sqlc.arg(workspace_id)
  AND fc.item_id = sqlc.arg(item_id)
  AND fc.id = sqlc.arg(comment_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: CreateAccountFeedbackComment :one
WITH eligible_item AS (
    SELECT fi.workspace_id, fi.portal_id, fi.id
    FROM feedback_items fi
    INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
    INNER JOIN workspace_members wm
        ON wm.workspace_id = fi.workspace_id
       AND wm.user_id = sqlc.arg(actor_id)
       AND wm.role IN ('admin', 'member')
    INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
    INNER JOIN users current_actor
        ON current_actor.user_id = wm.user_id
       AND current_actor.is_active = true
       AND current_actor.is_system = false
    WHERE fi.workspace_id = sqlc.arg(workspace_id)
      AND fi.id = sqlc.arg(item_id)
      AND fi.deleted_at IS NULL
      AND fi.merged_into_item_id IS NULL
      AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
      AND (
          CAST(sqlc.narg(parent_id) AS uuid) IS NULL
          OR EXISTS (
              SELECT 1
              FROM feedback_comments parent
              WHERE parent.id = CAST(sqlc.narg(parent_id) AS uuid)
                AND parent.item_id = fi.id
                AND parent.workspace_id = fi.workspace_id
          )
      )
), contributor AS (
    INSERT INTO feedback_contributors (portal_id, user_id, kind)
    SELECT eligible_item.portal_id, sqlc.arg(actor_id), 'account'
    FROM eligible_item
    ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL
    DO UPDATE SET updated_at = NOW()
    RETURNING id
), inserted AS (
    INSERT INTO feedback_comments (workspace_id, item_id, author_id, contributor_id, parent_id, body)
    SELECT eligible_item.workspace_id,
           eligible_item.id,
           sqlc.arg(actor_id),
           contributor.id,
           CAST(sqlc.narg(parent_id) AS uuid),
           sqlc.arg(body)
    FROM eligible_item
    CROSS JOIN contributor
    RETURNING id, workspace_id, item_id, author_id, contributor_id, parent_id, body, created_at, updated_at
)
SELECT inserted.id,
       inserted.workspace_id,
       inserted.item_id,
       inserted.author_id,
       inserted.contributor_id,
       inserted.parent_id,
       CAST(COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Anonymous') AS text) AS author_name,
       COALESCE(account.avatar_url, '') AS author_avatar,
       'account' AS participant_kind,
       false AS author_masked,
       inserted.body,
       inserted.created_at,
       inserted.updated_at
FROM inserted
LEFT JOIN users account ON account.user_id = inserted.author_id;

-- name: CreateContributorFeedbackComment :one
WITH inserted AS (
    INSERT INTO feedback_comments (workspace_id, item_id, author_id, contributor_id, parent_id, body)
    SELECT item.workspace_id,
           item.id,
           contributor.user_id,
           contributor.id,
           CAST(sqlc.narg(parent_id) AS uuid),
           sqlc.arg(body)
    FROM feedback_items item
    INNER JOIN feedback_portals portal ON portal.id = item.portal_id AND portal.is_public = true
    INNER JOIN feedback_contributors contributor
        ON contributor.portal_id = item.portal_id
       AND contributor.id = sqlc.arg(contributor_id)
       AND contributor.blocked_at IS NULL
    WHERE item.workspace_id = sqlc.arg(workspace_id)
      AND item.portal_id = sqlc.arg(portal_id)
      AND item.id = sqlc.arg(item_id)
      AND item.deleted_at IS NULL
      AND item.merged_into_item_id IS NULL
      AND (
          CAST(sqlc.narg(parent_id) AS uuid) IS NULL
          OR EXISTS (
              SELECT 1
              FROM feedback_comments parent
              WHERE parent.id = CAST(sqlc.narg(parent_id) AS uuid)
                AND parent.item_id = item.id
                AND parent.workspace_id = item.workspace_id
          )
      )
    RETURNING id, workspace_id, item_id, author_id, contributor_id, parent_id, body, created_at, updated_at
)
SELECT inserted.id,
       inserted.workspace_id,
       inserted.item_id,
       inserted.author_id,
       inserted.contributor_id,
       inserted.parent_id,
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
       CAST(contributor.kind AS text) AS participant_kind,
       CAST((contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS boolean) AS author_masked,
       inserted.body,
       inserted.created_at,
       inserted.updated_at
FROM inserted
INNER JOIN feedback_items item ON item.id = inserted.item_id
INNER JOIN feedback_portals portal ON portal.id = item.portal_id
INNER JOIN feedback_contributors contributor ON contributor.id = inserted.contributor_id
LEFT JOIN users account ON account.user_id = contributor.user_id;

-- name: GetFeedbackPrivateAuthor :one
SELECT contributor.id AS contributor_id,
       contributor.user_id,
       CAST(contributor.kind AS text) AS kind,
       CAST(CASE
           WHEN contributor.kind = 'anonymous' THEN 'Anonymous'
           WHEN contributor.kind = 'account' THEN COALESCE(NULLIF(TRIM(account.full_name), ''), NULLIF(TRIM(account.username), ''), 'Account user')
           ELSE COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'Guest')
       END AS text) AS display_name,
       CAST(COALESCE(CASE WHEN contributor.kind = 'account' THEN account.email ELSE contributor.email END, '') AS text) AS email,
       CAST(COALESCE(CASE WHEN contributor.kind = 'account' THEN account.avatar_url ELSE contributor.avatar_url END, '') AS text) AS avatar_url,
       (contributor.kind IN ('verified_guest', 'external') AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) AS public_masked
FROM feedback_items item
INNER JOIN feedback_boards fb ON fb.id = item.board_id AND fb.workspace_id = item.workspace_id
INNER JOIN feedback_contributors contributor ON contributor.id = item.contributor_id AND contributor.portal_id = item.portal_id
INNER JOIN feedback_portals portal ON portal.id = item.portal_id
LEFT JOIN users account ON account.user_id = contributor.user_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: GetFeedbackItemReadAt :one
SELECT fir.read_at
FROM feedback_item_reads fir
INNER JOIN feedback_items fi ON fi.id = fir.item_id
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fi.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
WHERE fi.workspace_id = sqlc.arg(workspace_id)
  AND fir.item_id = sqlc.arg(item_id)
  AND fir.user_id = sqlc.arg(actor_id)
  AND fi.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: ListFeedbackTeamSummaries :many
SELECT fb.team_id,
       true AS enabled,
       CAST(COUNT(fi.id) AS integer) AS total_count,
       CAST(COUNT(fi.id) FILTER (WHERE fi.id IS NOT NULL AND fir.item_id IS NULL) AS integer) AS unread_count
FROM feedback_boards fb
INNER JOIN workspace_members wm
    ON wm.workspace_id = fb.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
LEFT JOIN feedback_items fi
    ON fi.board_id = fb.id
   AND fi.workspace_id = fb.workspace_id
   AND fi.deleted_at IS NULL
   AND fi.merged_into_item_id IS NULL
LEFT JOIN feedback_item_reads fir ON fir.item_id = fi.id AND fir.user_id = sqlc.arg(actor_id)
WHERE fb.workspace_id = sqlc.arg(workspace_id)
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
GROUP BY fb.team_id
ORDER BY fb.team_id;

-- name: MarkFeedbackItemRead :one
INSERT INTO feedback_item_reads (item_id, user_id)
SELECT fi.id, wm.user_id
FROM feedback_items fi
INNER JOIN feedback_boards fb ON fb.id = fi.board_id AND fb.workspace_id = fi.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fi.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
WHERE fi.workspace_id = sqlc.arg(workspace_id)
  AND fi.id = sqlc.arg(item_id)
  AND fi.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ON CONFLICT (item_id, user_id)
DO UPDATE SET read_at = feedback_item_reads.read_at
RETURNING read_at;

-- name: MarkFeedbackItemUnread :execrows
DELETE FROM feedback_item_reads fir
USING feedback_items fi,
      feedback_boards fb,
      workspace_members wm,
      team_members tm,
      users current_actor
WHERE fi.id = fir.item_id
  AND fb.id = fi.board_id
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
  AND fir.item_id = sqlc.arg(item_id)
  AND fir.user_id = sqlc.arg(actor_id)
  AND fi.deleted_at IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: GetFeedbackVote :one
SELECT CAST(COALESCE((
    SELECT vote.direction
    FROM feedback_votes vote
    INNER JOIN feedback_items item ON item.id = vote.item_id
    WHERE vote.workspace_id = sqlc.arg(workspace_id)
      AND vote.item_id = sqlc.arg(item_id)
      AND vote.contributor_id = sqlc.arg(contributor_id)
      AND item.deleted_at IS NULL
      AND item.merged_into_item_id IS NULL
), 0) AS integer);

-- name: DeleteFeedbackVote :execrows
DELETE FROM feedback_votes vote
USING feedback_items item
WHERE item.id = vote.item_id
  AND vote.workspace_id = sqlc.arg(workspace_id)
  AND vote.item_id = sqlc.arg(item_id)
  AND vote.contributor_id = sqlc.arg(contributor_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL;

-- name: UpsertContributorFeedbackVote :execrows
INSERT INTO feedback_votes (workspace_id, item_id, user_id, contributor_id, direction)
SELECT item.workspace_id,
       item.id,
       contributor.user_id,
       contributor.id,
       sqlc.arg(direction)
FROM feedback_items item
INNER JOIN feedback_portals portal ON portal.id = item.portal_id AND portal.is_public = true
INNER JOIN feedback_contributors contributor
    ON contributor.portal_id = item.portal_id
   AND contributor.id = sqlc.arg(contributor_id)
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND contributor.blocked_at IS NULL
  AND contributor.kind <> 'anonymous'
ON CONFLICT (item_id, contributor_id)
DO UPDATE SET direction = EXCLUDED.direction;

-- name: UpsertAccountFeedbackVote :execrows
INSERT INTO feedback_votes (workspace_id, item_id, user_id, contributor_id, direction)
SELECT item.workspace_id,
       item.id,
       contributor.user_id,
       contributor.id,
       sqlc.arg(direction)
FROM feedback_items item
INNER JOIN feedback_boards fb ON fb.id = item.board_id AND fb.workspace_id = item.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor ON current_actor.user_id = wm.user_id AND current_actor.is_active = true AND current_actor.is_system = false
INNER JOIN feedback_contributors contributor
    ON contributor.portal_id = item.portal_id
   AND contributor.user_id = wm.user_id
   AND contributor.kind = 'account'
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])))
ON CONFLICT (item_id, contributor_id)
DO UPDATE SET direction = EXCLUDED.direction;

-- name: GetAccountFeedbackVote :one
SELECT CAST(COALESCE(vote.direction, 0) AS integer)
FROM feedback_items item
INNER JOIN feedback_boards fb ON fb.id = item.board_id AND fb.workspace_id = item.workspace_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = item.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm ON tm.team_id = fb.team_id AND tm.user_id = wm.user_id
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
LEFT JOIN feedback_contributors contributor
    ON contributor.portal_id = item.portal_id
   AND contributor.user_id = wm.user_id
   AND contributor.kind = 'account'
LEFT JOIN feedback_votes vote
    ON vote.item_id = item.id
   AND vote.contributor_id = contributor.id
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: DeleteAccountFeedbackVote :execrows
DELETE FROM feedback_votes vote
USING feedback_items item,
      feedback_boards fb,
      workspace_members wm,
      team_members tm,
      users current_actor,
      feedback_contributors contributor
WHERE item.id = vote.item_id
  AND fb.id = item.board_id
  AND fb.workspace_id = item.workspace_id
  AND wm.workspace_id = item.workspace_id
  AND wm.user_id = sqlc.arg(actor_id)
  AND wm.role IN ('admin', 'member')
  AND tm.team_id = fb.team_id
  AND tm.user_id = wm.user_id
  AND current_actor.user_id = wm.user_id
  AND current_actor.is_active = true
  AND current_actor.is_system = false
  AND contributor.portal_id = item.portal_id
  AND contributor.user_id = wm.user_id
  AND contributor.kind = 'account'
  AND vote.contributor_id = contributor.id
  AND item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
  AND (CAST(sqlc.arg(all_teams) AS boolean) OR fb.team_id = ANY(CAST(sqlc.arg(credential_team_ids) AS uuid[])));

-- name: GetFeedbackVoteCount :one
SELECT CAST(COALESCE(SUM(direction), 0) AS integer)
FROM feedback_votes
WHERE item_id = sqlc.arg(item_id);
