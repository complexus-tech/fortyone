-- name: GetPublicFeedbackPortalByWorkspaceSlug :one
SELECT fp.id,
       fp.workspace_id,
       w.name,
       w.slug,
       fp.is_public,
       CAST(fp.participation_mode AS text) AS participation_mode,
       CAST(fp.guest_identity_policy AS text) AS guest_identity_policy,
       EXISTS (
           SELECT 1
           FROM feedback_updates fu
           WHERE fu.portal_id = fp.id
             AND fu.status = 'published'
             AND fu.published_at IS NOT NULL
       ) AS has_published_updates,
       fp.created_at,
       fp.updated_at
FROM feedback_portals fp
INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
WHERE w.slug = sqlc.arg(workspace_slug)
  AND w.deleted_at IS NULL
  AND fp.is_public = true
LIMIT 1;

-- name: GetFeedbackPortal :one
SELECT fp.id,
       fp.workspace_id,
       w.name,
       w.slug,
       fp.is_public,
       CAST(fp.participation_mode AS text) AS participation_mode,
       CAST(fp.guest_identity_policy AS text) AS guest_identity_policy,
       EXISTS (
           SELECT 1
           FROM feedback_updates fu
           WHERE fu.portal_id = fp.id
             AND fu.status = 'published'
             AND fu.published_at IS NOT NULL
       ) AS has_published_updates,
       fp.created_at,
       fp.updated_at
FROM feedback_portals fp
INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
WHERE fp.workspace_id = sqlc.arg(workspace_id)
  AND fp.id = sqlc.arg(portal_id)
  AND w.deleted_at IS NULL;

-- name: ListFeedbackPortals :many
SELECT fp.id,
       fp.workspace_id,
       w.name,
       w.slug,
       fp.is_public,
       CAST(fp.participation_mode AS text) AS participation_mode,
       CAST(fp.guest_identity_policy AS text) AS guest_identity_policy,
       EXISTS (
           SELECT 1
           FROM feedback_updates fu
           WHERE fu.portal_id = fp.id
             AND fu.status = 'published'
             AND fu.published_at IS NOT NULL
       ) AS has_published_updates,
       fp.created_at,
       fp.updated_at
FROM feedback_portals fp
INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
WHERE fp.workspace_id = sqlc.arg(workspace_id)
  AND w.deleted_at IS NULL
ORDER BY fp.created_at ASC, fp.id ASC;

-- name: CreateFeedbackPortal :one
WITH inserted AS (
    INSERT INTO feedback_portals (
        workspace_id,
        is_public,
        participation_mode,
        guest_identity_policy
    )
    SELECT w.workspace_id,
           sqlc.arg(is_public),
           sqlc.arg(participation_mode),
           sqlc.arg(guest_identity_policy)
    FROM workspaces w
    WHERE w.workspace_id = sqlc.arg(workspace_id)
      AND w.deleted_at IS NULL
    RETURNING id,
              workspace_id,
              is_public,
              participation_mode,
              guest_identity_policy,
              created_at,
              updated_at
)
SELECT inserted.id,
       inserted.workspace_id,
       w.name,
       w.slug,
       inserted.is_public,
       CAST(inserted.participation_mode AS text) AS participation_mode,
       CAST(inserted.guest_identity_policy AS text) AS guest_identity_policy,
       false AS has_published_updates,
       inserted.created_at,
       inserted.updated_at
FROM inserted
INNER JOIN workspaces w ON w.workspace_id = inserted.workspace_id;

-- name: UpdateFeedbackPortal :one
WITH updated AS (
    UPDATE feedback_portals fp
    SET is_public = COALESCE(CAST(sqlc.narg(is_public) AS boolean), fp.is_public),
        participation_mode = COALESCE(CAST(sqlc.narg(participation_mode) AS text), fp.participation_mode),
        guest_identity_policy = COALESCE(CAST(sqlc.narg(guest_identity_policy) AS text), fp.guest_identity_policy),
        updated_at = NOW()
    WHERE fp.workspace_id = sqlc.arg(workspace_id)
      AND fp.id = sqlc.arg(portal_id)
    RETURNING fp.id,
              fp.workspace_id,
              fp.is_public,
              fp.participation_mode,
              fp.guest_identity_policy,
              fp.created_at,
              fp.updated_at
)
SELECT updated.id,
       updated.workspace_id,
       w.name,
       w.slug,
       updated.is_public,
       CAST(updated.participation_mode AS text) AS participation_mode,
       CAST(updated.guest_identity_policy AS text) AS guest_identity_policy,
       EXISTS (
           SELECT 1
           FROM feedback_updates fu
           WHERE fu.portal_id = updated.id
             AND fu.status = 'published'
             AND fu.published_at IS NOT NULL
       ) AS has_published_updates,
       updated.created_at,
       updated.updated_at
FROM updated
INNER JOIN workspaces w ON w.workspace_id = updated.workspace_id
WHERE w.deleted_at IS NULL;

-- name: ListFeedbackBoards :many
SELECT fb.id,
       fb.workspace_id,
       fb.portal_id,
       fb.team_id,
       fb.name,
       fb.slug,
       fb.color,
       fb.order_index,
       fb.created_at,
       fb.updated_at
FROM feedback_boards fb
INNER JOIN feedback_portals fp
    ON fp.id = fb.portal_id
   AND fp.workspace_id = fb.workspace_id
WHERE fb.portal_id = sqlc.arg(portal_id)
ORDER BY fb.order_index ASC, fb.created_at ASC, fb.id ASC;

-- name: GetFeedbackBoard :one
SELECT fb.id,
       fb.workspace_id,
       fb.portal_id,
       fb.team_id,
       fb.name,
       fb.slug,
       fb.color,
       fb.order_index,
       fb.created_at,
       fb.updated_at
FROM feedback_boards fb
INNER JOIN feedback_portals fp
    ON fp.id = fb.portal_id
   AND fp.workspace_id = fb.workspace_id
WHERE fb.portal_id = sqlc.arg(portal_id)
  AND fb.id = sqlc.arg(board_id);

-- name: CreateFeedbackBoard :one
INSERT INTO feedback_boards (
    workspace_id,
    portal_id,
    team_id,
    name,
    slug,
    color,
    order_index
)
SELECT fp.workspace_id,
       fp.id,
       t.team_id,
       sqlc.arg(name),
       sqlc.arg(slug),
       sqlc.arg(color),
       sqlc.arg(order_index)
FROM feedback_portals fp
INNER JOIN teams t
    ON t.team_id = sqlc.arg(team_id)
   AND t.workspace_id = fp.workspace_id
WHERE fp.id = sqlc.arg(portal_id)
  AND fp.workspace_id = sqlc.arg(workspace_id)
RETURNING id,
          workspace_id,
          portal_id,
          team_id,
          name,
          slug,
          color,
          order_index,
          created_at,
          updated_at;

-- name: AddFeedbackBoardCreatorReviewer :one
INSERT INTO feedback_board_subscriptions (board_id, user_id, email_frequency)
SELECT fb.id,
       u.user_id,
       sqlc.arg(email_frequency)
FROM feedback_boards fb
INNER JOIN workspace_members wm
    ON wm.workspace_id = fb.workspace_id
   AND wm.user_id = sqlc.arg(creator_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN team_members tm
    ON tm.team_id = fb.team_id
   AND tm.user_id = wm.user_id
INNER JOIN users u
    ON u.user_id = wm.user_id
   AND u.is_active = true
   AND u.is_system = false
WHERE fb.workspace_id = sqlc.arg(workspace_id)
  AND fb.id = sqlc.arg(board_id)
RETURNING user_id;

-- name: LockAnonymousFeedbackBoardContributors :many
SELECT fc.id
FROM feedback_contributors fc
WHERE fc.kind = 'anonymous'
  AND EXISTS (
      SELECT 1
      FROM feedback_items fi
      WHERE fi.portal_id = fc.portal_id
        AND fi.contributor_id = fc.id
        AND fi.workspace_id = sqlc.arg(workspace_id)
        AND fi.board_id = sqlc.arg(board_id)
  )
ORDER BY fc.id
FOR UPDATE OF fc;

-- name: DeleteFeedbackBoard :execrows
DELETE FROM feedback_boards fb
WHERE fb.workspace_id = sqlc.arg(workspace_id)
  AND fb.id = sqlc.arg(board_id);

-- name: DeleteOrphanAnonymousFeedbackContributors :execrows
DELETE FROM feedback_contributors fc
WHERE fc.id = ANY(CAST(sqlc.arg(contributor_ids) AS uuid[]))
  AND fc.kind = 'anonymous'
  AND NOT EXISTS (
      SELECT 1
      FROM feedback_items retained
      WHERE retained.contributor_id = fc.id
  );

-- name: FeedbackBoardExists :one
SELECT EXISTS (
    SELECT 1
    FROM feedback_boards fb
    WHERE fb.workspace_id = sqlc.arg(workspace_id)
      AND fb.id = sqlc.arg(board_id)
);

-- name: ListFeedbackBoardReviewers :many
SELECT u.user_id,
       COALESCE(NULLIF(TRIM(u.full_name), ''), u.email) AS name,
       u.email,
       u.avatar_url,
       CAST(wm.role AS text) AS role,
       CAST(CASE
           WHEN fbs.email_frequency IS NULL THEN CAST(sqlc.arg(default_frequency) AS text)
           ELSE fbs.email_frequency
       END AS text) AS email_frequency
FROM feedback_boards fb
INNER JOIN team_members tm ON tm.team_id = fb.team_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = fb.workspace_id
   AND wm.user_id = tm.user_id
   AND wm.role IN ('admin', 'member')
INNER JOIN users u
    ON u.user_id = tm.user_id
   AND u.is_active = true
   AND u.is_system = false
LEFT JOIN feedback_board_subscriptions fbs
    ON fbs.board_id = fb.id
   AND fbs.user_id = u.user_id
WHERE fb.workspace_id = sqlc.arg(workspace_id)
  AND fb.id = sqlc.arg(board_id)
ORDER BY LOWER(COALESCE(NULLIF(TRIM(u.full_name), ''), u.email)), u.user_id;

-- name: SetFeedbackBoardReviewer :one
WITH eligible AS (
    SELECT fb.id AS board_id,
           u.user_id,
           COALESCE(NULLIF(TRIM(u.full_name), ''), u.email) AS name,
           u.email,
           u.avatar_url,
           CAST(wm.role AS text) AS role
    FROM feedback_boards fb
    INNER JOIN team_members tm
        ON tm.team_id = fb.team_id
       AND tm.user_id = sqlc.arg(user_id)
    INNER JOIN workspace_members wm
        ON wm.workspace_id = fb.workspace_id
       AND wm.user_id = tm.user_id
       AND wm.role IN ('admin', 'member')
    INNER JOIN users u
        ON u.user_id = tm.user_id
       AND u.is_active = true
       AND u.is_system = false
    WHERE fb.workspace_id = sqlc.arg(workspace_id)
      AND fb.id = sqlc.arg(board_id)
), saved AS (
    INSERT INTO feedback_board_subscriptions (board_id, user_id, email_frequency)
    SELECT eligible.board_id,
           eligible.user_id,
           sqlc.arg(email_frequency)
    FROM eligible
    ON CONFLICT (board_id, user_id) DO UPDATE
    SET email_frequency = EXCLUDED.email_frequency,
        updated_at = NOW()
    RETURNING user_id, email_frequency
)
SELECT eligible.user_id,
       eligible.name,
       eligible.email,
       eligible.avatar_url,
       eligible.role,
       CAST(saved.email_frequency AS text) AS email_frequency
FROM eligible
INNER JOIN saved ON saved.user_id = eligible.user_id;

-- name: RemoveFeedbackBoardReviewer :one
WITH eligible AS (
    SELECT fb.id AS board_id,
           u.user_id,
           COALESCE(NULLIF(TRIM(u.full_name), ''), u.email) AS name,
           u.email,
           u.avatar_url,
           CAST(wm.role AS text) AS role
    FROM feedback_boards fb
    INNER JOIN team_members tm
        ON tm.team_id = fb.team_id
       AND tm.user_id = sqlc.arg(user_id)
    INNER JOIN workspace_members wm
        ON wm.workspace_id = fb.workspace_id
       AND wm.user_id = tm.user_id
       AND wm.role IN ('admin', 'member')
    INNER JOIN users u
        ON u.user_id = tm.user_id
       AND u.is_active = true
       AND u.is_system = false
    WHERE fb.workspace_id = sqlc.arg(workspace_id)
      AND fb.id = sqlc.arg(board_id)
), removed AS (
    DELETE FROM feedback_board_subscriptions fbs
    USING eligible
    WHERE fbs.board_id = eligible.board_id
      AND fbs.user_id = eligible.user_id
    RETURNING fbs.user_id
)
SELECT eligible.user_id,
       eligible.name,
       eligible.email,
       eligible.avatar_url,
       eligible.role,
       CAST(sqlc.arg(email_frequency) AS text) AS email_frequency
FROM eligible;
