-- name: EnsureFeedbackWidgetSettings :exec
INSERT INTO feedback_widget_settings (portal_id, enabled, allowed_origins)
SELECT portal.id, false, CAST('{}' AS text[])
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
ON CONFLICT (portal_id) DO NOTHING;

-- name: GetFeedbackWidgetSettings :one
SELECT settings.portal_id,
       settings.enabled,
       settings.widget_key_id,
       settings.allowed_origins,
       settings.signing_secret_encrypted,
       settings.signing_secret_version,
       CAST(COALESCE((
           SELECT MAX(rotation.grace_expires_at)
           FROM feedback_widget_signing_secret_rotations rotation
           WHERE rotation.portal_id = settings.portal_id
             AND rotation.retired_at IS NULL
             AND rotation.grace_expires_at > NOW()
       ), CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS timestamptz) AS previous_version_expires_at,
       settings.created_at,
       settings.updated_at
FROM feedback_widget_settings settings
INNER JOIN feedback_portals portal ON portal.id = settings.portal_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = portal.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE portal.workspace_id = sqlc.arg(workspace_id)
  AND settings.portal_id = sqlc.arg(portal_id);

-- name: GetPublicFeedbackWidgetSettings :one
SELECT settings.portal_id,
       settings.enabled,
       settings.widget_key_id,
       settings.allowed_origins,
       settings.signing_secret_encrypted,
       settings.signing_secret_version,
       CAST(COALESCE((
           SELECT MAX(rotation.grace_expires_at)
           FROM feedback_widget_signing_secret_rotations rotation
           WHERE rotation.portal_id = settings.portal_id
             AND rotation.retired_at IS NULL
             AND rotation.grace_expires_at > NOW()
       ), CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS timestamptz) AS previous_version_expires_at,
       settings.created_at,
       settings.updated_at
FROM feedback_widget_settings settings
INNER JOIN feedback_portals portal
    ON portal.id = settings.portal_id
   AND portal.is_public = true
WHERE settings.portal_id = sqlc.arg(portal_id);

-- name: UpsertFeedbackWidgetSettings :one
WITH target AS (
    SELECT portal.id,
           settings.widget_key_id,
           settings.signing_secret_encrypted,
           settings.signing_secret_version
    FROM feedback_portals portal
    INNER JOIN workspace_members wm
        ON wm.workspace_id = portal.workspace_id
       AND wm.user_id = sqlc.arg(actor_id)
       AND wm.role IN ('admin', 'member')
    INNER JOIN users current_actor
        ON current_actor.user_id = wm.user_id
       AND current_actor.is_active = true
       AND current_actor.is_system = false
    LEFT JOIN feedback_widget_settings settings ON settings.portal_id = portal.id
    WHERE portal.workspace_id = sqlc.arg(workspace_id)
      AND portal.id = sqlc.arg(portal_id)
), updated AS (
    INSERT INTO feedback_widget_settings (
        portal_id,
        enabled,
        widget_key_id,
        allowed_origins,
        signing_secret_encrypted,
        signing_secret_version
    )
    SELECT target.id,
           sqlc.arg(enabled),
           COALESCE(target.widget_key_id, gen_random_uuid()),
           CAST(sqlc.arg(allowed_origins) AS text[]),
           target.signing_secret_encrypted,
           COALESCE(target.signing_secret_version, 0)
    FROM target
    ON CONFLICT (portal_id) DO UPDATE
    SET enabled = EXCLUDED.enabled,
        allowed_origins = EXCLUDED.allowed_origins,
        updated_at = NOW()
    RETURNING portal_id,
              enabled,
              widget_key_id,
              allowed_origins,
              signing_secret_encrypted,
              signing_secret_version,
              created_at,
              updated_at
)
SELECT updated.portal_id,
       updated.enabled,
       updated.widget_key_id,
       updated.allowed_origins,
       updated.signing_secret_encrypted,
       updated.signing_secret_version,
       CAST(COALESCE((
           SELECT MAX(rotation.grace_expires_at)
           FROM feedback_widget_signing_secret_rotations rotation
           WHERE rotation.portal_id = updated.portal_id
             AND rotation.retired_at IS NULL
             AND rotation.grace_expires_at > NOW()
       ), CAST('0001-01-01T00:00:00Z' AS timestamptz)) AS timestamptz) AS previous_version_expires_at,
       updated.created_at,
       updated.updated_at
FROM updated;

-- name: SetInitialFeedbackWidgetSecret :one
WITH target AS (
    SELECT portal.id
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
), updated AS (
    INSERT INTO feedback_widget_settings (
        portal_id,
        enabled,
        widget_key_id,
        allowed_origins,
        signing_secret_encrypted,
        signing_secret_version
    )
    SELECT target.id,
           false,
           sqlc.arg(widget_key_id),
           CAST('{}' AS text[]),
           sqlc.arg(signing_secret_encrypted),
           sqlc.arg(signing_secret_version)
    FROM target
    ON CONFLICT (portal_id) DO UPDATE
    SET widget_key_id = EXCLUDED.widget_key_id,
        signing_secret_encrypted = EXCLUDED.signing_secret_encrypted,
        signing_secret_version = EXCLUDED.signing_secret_version,
        updated_at = NOW()
    WHERE feedback_widget_settings.signing_secret_encrypted IS NULL
    RETURNING portal_id,
              enabled,
              widget_key_id,
              allowed_origins,
              signing_secret_encrypted,
              signing_secret_version,
              created_at,
              updated_at
)
SELECT updated.portal_id,
       updated.enabled,
       updated.widget_key_id,
       updated.allowed_origins,
       updated.signing_secret_encrypted,
       updated.signing_secret_version,
       CAST(NULL AS timestamptz) AS previous_version_expires_at,
       updated.created_at,
       updated.updated_at
FROM updated;

-- name: LockFeedbackWidgetSettings :one
SELECT settings.portal_id,
       settings.enabled,
       settings.widget_key_id,
       settings.allowed_origins,
       settings.signing_secret_encrypted,
       settings.signing_secret_version,
       CAST(NULL AS timestamptz) AS previous_version_expires_at,
       settings.created_at,
       settings.updated_at
FROM feedback_widget_settings settings
INNER JOIN feedback_portals portal ON portal.id = settings.portal_id
INNER JOIN workspace_members wm
    ON wm.workspace_id = portal.workspace_id
   AND wm.user_id = sqlc.arg(actor_id)
   AND wm.role IN ('admin', 'member')
INNER JOIN users current_actor
    ON current_actor.user_id = wm.user_id
   AND current_actor.is_active = true
   AND current_actor.is_system = false
WHERE portal.workspace_id = sqlc.arg(workspace_id)
  AND settings.portal_id = sqlc.arg(portal_id)
FOR UPDATE OF settings;

-- name: SavePreviousFeedbackWidgetSecret :exec
INSERT INTO feedback_widget_signing_secret_rotations (
    portal_id,
    signing_secret_version,
    signing_secret_encrypted,
    activated_at,
    grace_expires_at
)
VALUES (
    sqlc.arg(portal_id),
    sqlc.arg(signing_secret_version),
    sqlc.arg(signing_secret_encrypted),
    sqlc.arg(activated_at),
    sqlc.arg(grace_expires_at)
);

-- name: UpdateFeedbackWidgetSecret :execrows
UPDATE feedback_widget_settings
SET signing_secret_encrypted = sqlc.arg(signing_secret_encrypted),
    signing_secret_version = sqlc.arg(signing_secret_version),
    updated_at = NOW()
WHERE portal_id = sqlc.arg(portal_id)
  AND signing_secret_version = sqlc.arg(previous_version);

-- name: GetCurrentFeedbackWidgetSigningSecret :one
SELECT settings.signing_secret_encrypted
FROM feedback_widget_settings settings
INNER JOIN feedback_portals portal ON portal.id = settings.portal_id AND portal.is_public = true
WHERE settings.portal_id = sqlc.arg(portal_id)
  AND settings.widget_key_id = sqlc.arg(widget_key_id)
  AND settings.signing_secret_version = sqlc.arg(signing_secret_version)
  AND settings.enabled = true
  AND settings.signing_secret_encrypted IS NOT NULL;

-- name: GetPreviousFeedbackWidgetSigningSecret :one
SELECT rotation.signing_secret_encrypted
FROM feedback_widget_signing_secret_rotations rotation
INNER JOIN feedback_widget_settings settings ON settings.portal_id = rotation.portal_id
INNER JOIN feedback_portals portal ON portal.id = settings.portal_id AND portal.is_public = true
WHERE rotation.portal_id = sqlc.arg(portal_id)
  AND settings.widget_key_id = sqlc.arg(widget_key_id)
  AND settings.enabled = true
  AND rotation.signing_secret_version = sqlc.arg(signing_secret_version)
  AND rotation.retired_at IS NULL
  AND rotation.grace_expires_at > NOW();

-- name: ConsumeFeedbackWidgetAssertionNonce :exec
INSERT INTO feedback_widget_assertion_nonces (
    portal_id,
    widget_key_id,
    signing_secret_version,
    nonce_hash,
    parent_origin,
    expires_at
)
SELECT settings.portal_id,
       settings.widget_key_id,
       sqlc.arg(signing_secret_version),
       sqlc.arg(nonce_hash),
       sqlc.arg(parent_origin),
       sqlc.arg(expires_at)
FROM feedback_widget_settings settings
INNER JOIN feedback_portals portal ON portal.id = settings.portal_id AND portal.is_public = true
WHERE settings.portal_id = sqlc.arg(portal_id)
  AND settings.widget_key_id = sqlc.arg(widget_key_id)
  AND settings.enabled = true;
