-- name: CreateContributorVerification :one
INSERT INTO feedback_contributor_verifications (
    portal_id,
    email,
    display_name,
    public_masked,
    token_hash,
    code_hash,
    source,
    expires_at
)
SELECT portal.id,
       sqlc.arg(email),
       NULLIF(CAST(sqlc.arg(display_name) AS text), ''),
       sqlc.arg(public_masked),
       sqlc.arg(token_hash),
       sqlc.arg(code_hash),
       sqlc.arg(source),
       sqlc.arg(expires_at)
FROM feedback_portals portal
WHERE portal.id = sqlc.arg(portal_id)
  AND portal.is_public = true
  AND (
      SELECT COUNT(*)
      FROM feedback_contributor_verifications recent
      WHERE recent.portal_id = portal.id
        AND recent.email = sqlc.arg(email)
        AND recent.created_at >= sqlc.arg(rate_limit_since)
  ) < CAST(sqlc.arg(rate_limit) AS integer)
RETURNING id, expires_at;

-- name: LockContributorVerificationByToken :one
SELECT id,
       portal_id,
       email,
       display_name,
       public_masked,
       token_hash,
       code_hash,
       source,
       expires_at,
       consumed_at,
       attempt_count,
       created_at
FROM feedback_contributor_verifications
WHERE portal_id = sqlc.arg(portal_id)
  AND token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- name: LockContributorVerificationByCode :one
SELECT id,
       portal_id,
       email,
       display_name,
       public_masked,
       token_hash,
       code_hash,
       source,
       expires_at,
       consumed_at,
       attempt_count,
       created_at
FROM feedback_contributor_verifications
WHERE portal_id = sqlc.arg(portal_id)
  AND email = sqlc.arg(email)
  AND source = sqlc.arg(source)
  AND consumed_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1
FOR UPDATE;

-- name: IncrementContributorVerificationAttempts :execrows
UPDATE feedback_contributor_verifications
SET attempt_count = sqlc.arg(attempt_count)
WHERE id = sqlc.arg(verification_id)
  AND consumed_at IS NULL;

-- name: ConsumeContributorVerification :execrows
UPDATE feedback_contributor_verifications
SET consumed_at = sqlc.arg(consumed_at)
WHERE id = sqlc.arg(verification_id)
  AND consumed_at IS NULL;

-- name: UpsertVerifiedFeedbackContributor :one
WITH existing AS (
    SELECT existing_contributor.id, existing_contributor.kind
    FROM feedback_contributors existing_contributor
    WHERE existing_contributor.portal_id = sqlc.arg(portal_id)
      AND LOWER(existing_contributor.email) = LOWER(sqlc.arg(email))
      AND existing_contributor.kind <> 'external'
    ORDER BY existing_contributor.created_at, existing_contributor.id
    LIMIT 1
    FOR UPDATE
), updated AS (
    UPDATE feedback_contributors contributor
    SET email = sqlc.arg(email),
        email_verified_at = COALESCE(contributor.email_verified_at, sqlc.arg(now)),
        display_name = COALESCE(NULLIF(CAST(sqlc.arg(display_name) AS text), ''), contributor.display_name),
        public_masked = CASE WHEN contributor.kind = 'verified_guest' THEN sqlc.arg(public_masked) ELSE false END,
        last_seen_at = sqlc.arg(now),
        updated_at = sqlc.arg(now)
    FROM existing
    WHERE contributor.id = existing.id
    RETURNING contributor.id,
              contributor.portal_id,
              contributor.user_id,
              contributor.kind,
              contributor.email,
              contributor.email_verified_at,
              contributor.display_name,
              contributor.avatar_url,
              contributor.external_id,
              contributor.public_masked,
              contributor.blocked_at,
              contributor.blocked_reason,
              contributor.last_seen_at,
              contributor.created_at,
              contributor.updated_at
), inserted AS (
    INSERT INTO feedback_contributors (
        portal_id,
        kind,
        email,
        email_verified_at,
        display_name,
        public_masked,
        last_seen_at
    )
    SELECT sqlc.arg(portal_id),
           'verified_guest',
           sqlc.arg(email),
           sqlc.arg(now),
           NULLIF(CAST(sqlc.arg(display_name) AS text), ''),
           sqlc.arg(public_masked),
           sqlc.arg(now)
    WHERE NOT EXISTS (SELECT 1 FROM existing)
    RETURNING id,
              portal_id,
              user_id,
              kind,
              email,
              email_verified_at,
              display_name,
              avatar_url,
              external_id,
              public_masked,
              blocked_at,
              blocked_reason,
              last_seen_at,
              created_at,
              updated_at
)
SELECT updated.id,
       updated.portal_id,
       updated.user_id,
       CAST(updated.kind AS text) AS kind,
       updated.email,
       updated.email_verified_at,
       CAST(updated.display_name AS text) AS display_name,
       updated.avatar_url,
       updated.external_id,
       updated.public_masked,
       updated.blocked_at,
       updated.blocked_reason,
       updated.last_seen_at,
       updated.created_at,
       updated.updated_at
FROM updated
UNION ALL
SELECT inserted.id,
       inserted.portal_id,
       inserted.user_id,
       CAST(inserted.kind AS text) AS kind,
       inserted.email,
       inserted.email_verified_at,
       CAST(inserted.display_name AS text) AS display_name,
       inserted.avatar_url,
       inserted.external_id,
       inserted.public_masked,
       inserted.blocked_at,
       inserted.blocked_reason,
       inserted.last_seen_at,
       inserted.created_at,
       inserted.updated_at
FROM inserted;

-- name: EnsureFeedbackContributorPreferences :exec
INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
SELECT contributor.portal_id, contributor.id
FROM feedback_contributors contributor
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = sqlc.arg(contributor_id)
ON CONFLICT (portal_id, contributor_id) DO NOTHING;

-- name: CreateFeedbackContributorSession :one
INSERT INTO feedback_contributor_sessions (
    portal_id,
    contributor_id,
    token_hash,
    source,
    expires_at,
    last_used_at
)
SELECT contributor.portal_id,
       contributor.id,
       sqlc.arg(token_hash),
       sqlc.arg(source),
       sqlc.arg(expires_at),
       NOW()
FROM feedback_contributors contributor
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = sqlc.arg(contributor_id)
  AND contributor.blocked_at IS NULL
RETURNING id,
          portal_id,
          contributor_id,
          source,
          expires_at,
          revoked_at,
          last_used_at,
          created_at;

-- name: GetFeedbackContributorSession :one
WITH active AS (
    UPDATE feedback_contributor_sessions session
    SET last_used_at = NOW()
    WHERE session.portal_id = sqlc.arg(portal_id)
      AND session.token_hash = sqlc.arg(token_hash)
      AND session.revoked_at IS NULL
      AND session.expires_at > NOW()
      AND (
          (CAST(sqlc.arg(source) AS text) = '' AND session.source IN ('portal', 'widget'))
          OR session.source = CAST(sqlc.arg(source) AS text)
      )
    RETURNING session.id,
              session.portal_id,
              session.contributor_id,
              session.source,
              session.expires_at,
              session.revoked_at,
              session.last_used_at,
              session.created_at
)
SELECT contributor.id,
       contributor.portal_id,
       contributor.user_id,
       CAST(contributor.kind AS text) AS kind,
       COALESCE(contributor.email, account.email) AS email,
       contributor.email_verified_at,
       CAST(COALESCE(contributor.display_name, NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS text) AS display_name,
       COALESCE(contributor.avatar_url, account.avatar_url) AS avatar_url,
       contributor.external_id,
       contributor.public_masked,
       contributor.blocked_at,
       contributor.blocked_reason,
       contributor.last_seen_at,
       contributor.created_at,
       contributor.updated_at,
       active.id AS session_id,
       CAST(active.source AS text) AS session_source,
       active.expires_at AS session_expires_at,
       active.revoked_at AS session_revoked_at,
       active.last_used_at AS session_last_used_at,
       active.created_at AS session_created_at
FROM active
INNER JOIN feedback_contributors contributor
    ON contributor.id = active.contributor_id
   AND contributor.portal_id = active.portal_id
LEFT JOIN users account ON account.user_id = contributor.user_id;

-- name: RevokeFeedbackContributorSession :execrows
UPDATE feedback_contributor_sessions
SET revoked_at = NOW()
WHERE portal_id = sqlc.arg(portal_id)
  AND token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL;

-- name: GetOrCreateAccountParticipant :one
WITH contributor AS (
    INSERT INTO feedback_contributors (portal_id, user_id, kind, last_seen_at)
    SELECT portal.id, account.user_id, 'account', NOW()
    FROM feedback_portals portal
    INNER JOIN users account
        ON account.user_id = sqlc.arg(user_id)
       AND account.is_active = true
       AND account.is_system = false
    WHERE portal.id = sqlc.arg(portal_id)
    ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL
    DO UPDATE SET last_seen_at = NOW(), updated_at = NOW()
    RETURNING id,
              portal_id,
              user_id,
              kind,
              external_id,
              public_masked,
              blocked_at,
              blocked_reason,
              last_seen_at,
              created_at,
              updated_at
)
SELECT contributor.id,
       contributor.portal_id,
       contributor.user_id,
       CAST(contributor.kind AS text) AS kind,
       account.email,
       CAST(NULL AS timestamptz) AS email_verified_at,
       CAST(COALESCE(NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS text) AS display_name,
       account.avatar_url,
       contributor.external_id,
       contributor.public_masked,
       contributor.blocked_at,
       contributor.blocked_reason,
       contributor.last_seen_at,
       contributor.created_at,
       contributor.updated_at
FROM contributor
INNER JOIN users account ON account.user_id = contributor.user_id;

-- name: GetFeedbackParticipant :one
SELECT contributor.id,
       contributor.portal_id,
       contributor.user_id,
       CAST(contributor.kind AS text) AS kind,
       COALESCE(contributor.email, account.email) AS email,
       contributor.email_verified_at,
       CAST(COALESCE(contributor.display_name, NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS text) AS display_name,
       COALESCE(contributor.avatar_url, account.avatar_url) AS avatar_url,
       contributor.external_id,
       contributor.public_masked,
       contributor.blocked_at,
       contributor.blocked_reason,
       contributor.last_seen_at,
       contributor.created_at,
       contributor.updated_at
FROM feedback_contributors contributor
LEFT JOIN users account ON account.user_id = contributor.user_id
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = sqlc.arg(contributor_id);

-- name: LockFeedbackUnsubscribeToken :one
SELECT id,
       contributor_id,
       item_id,
       purpose,
       expires_at,
       consumed_at
FROM feedback_contributor_unsubscribe_tokens
WHERE portal_id = sqlc.arg(portal_id)
  AND token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- name: ConsumeFeedbackUnsubscribeToken :execrows
UPDATE feedback_contributor_unsubscribe_tokens
SET consumed_at = NOW()
WHERE id = sqlc.arg(token_id)
  AND consumed_at IS NULL;

-- name: UnsubscribeFeedbackItem :exec
UPDATE feedback_item_followers follower
SET unsubscribed_at = COALESCE(follower.unsubscribed_at, NOW())
FROM feedback_items item
WHERE item.id = follower.item_id
  AND item.portal_id = sqlc.arg(portal_id)
  AND follower.item_id = sqlc.arg(item_id)
  AND follower.contributor_id = sqlc.arg(contributor_id);

-- name: UnsubscribeFeedbackPortal :exec
UPDATE feedback_portal_followers
SET unsubscribed_at = COALESCE(unsubscribed_at, NOW())
WHERE portal_id = sqlc.arg(portal_id)
  AND contributor_id = sqlc.arg(contributor_id);

-- name: UnsubscribeAllFeedbackEmail :exec
INSERT INTO feedback_contributor_preferences (portal_id, contributor_id, email_unsubscribed_at)
SELECT contributor.portal_id, contributor.id, NOW()
FROM feedback_contributors contributor
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = sqlc.arg(contributor_id)
ON CONFLICT (portal_id, contributor_id)
DO UPDATE SET email_unsubscribed_at = NOW(), updated_at = NOW();

-- name: UpsertExternalFeedbackContributor :one
WITH existing AS (
    SELECT existing_contributor.id
    FROM feedback_contributors existing_contributor
    WHERE existing_contributor.portal_id = sqlc.arg(portal_id)
      AND existing_contributor.external_id = sqlc.arg(external_id)
      AND existing_contributor.kind = 'external'
    ORDER BY existing_contributor.created_at, existing_contributor.id
    LIMIT 1
    FOR UPDATE
), updated AS (
    UPDATE feedback_contributors contributor
    SET kind = 'external',
        user_id = NULL,
        external_id = sqlc.arg(external_id),
        email = sqlc.arg(email),
        email_verified_at = sqlc.arg(now),
        display_name = CAST(sqlc.arg(display_name) AS text),
        avatar_url = CAST(sqlc.narg(avatar_url) AS text),
        public_masked = false,
        last_seen_at = sqlc.arg(now),
        updated_at = sqlc.arg(now)
    FROM existing
    WHERE contributor.id = existing.id
    RETURNING contributor.id,
              contributor.portal_id,
              contributor.user_id,
              contributor.kind,
              contributor.email,
              contributor.email_verified_at,
              contributor.display_name,
              contributor.avatar_url,
              contributor.external_id,
              contributor.public_masked,
              contributor.blocked_at,
              contributor.blocked_reason,
              contributor.last_seen_at,
              contributor.created_at,
              contributor.updated_at
), inserted AS (
    INSERT INTO feedback_contributors (
        portal_id,
        kind,
        external_id,
        email,
        email_verified_at,
        display_name,
        avatar_url,
        last_seen_at
    )
    SELECT portal.id,
           'external',
           sqlc.arg(external_id),
           sqlc.arg(email),
           sqlc.arg(now),
           CAST(sqlc.arg(display_name) AS text),
           CAST(sqlc.narg(avatar_url) AS text),
           sqlc.arg(now)
    FROM feedback_portals portal
    WHERE portal.id = sqlc.arg(portal_id)
      AND portal.is_public = true
      AND NOT EXISTS (SELECT 1 FROM existing)
    RETURNING id,
              portal_id,
              user_id,
              kind,
              email,
              email_verified_at,
              display_name,
              avatar_url,
              external_id,
              public_masked,
              blocked_at,
              blocked_reason,
              last_seen_at,
              created_at,
              updated_at
)
SELECT updated.id,
       updated.portal_id,
       updated.user_id,
       CAST(updated.kind AS text) AS kind,
       updated.email,
       updated.email_verified_at,
       CAST(updated.display_name AS text) AS display_name,
       updated.avatar_url,
       updated.external_id,
       updated.public_masked,
       updated.blocked_at,
       updated.blocked_reason,
       updated.last_seen_at,
       updated.created_at,
       updated.updated_at
FROM updated
UNION ALL
SELECT inserted.id,
       inserted.portal_id,
       inserted.user_id,
       CAST(inserted.kind AS text) AS kind,
       inserted.email,
       inserted.email_verified_at,
       CAST(inserted.display_name AS text) AS display_name,
       inserted.avatar_url,
       inserted.external_id,
       inserted.public_masked,
       inserted.blocked_at,
       inserted.blocked_reason,
       inserted.last_seen_at,
       inserted.created_at,
       inserted.updated_at
FROM inserted;

-- name: FollowFeedbackItem :execrows
INSERT INTO feedback_item_followers (item_id, contributor_id)
SELECT item.id, contributor.id
FROM feedback_items item
INNER JOIN feedback_portals portal ON portal.id = item.portal_id AND portal.is_public = true
INNER JOIN feedback_contributors contributor
    ON contributor.portal_id = item.portal_id
   AND contributor.id = sqlc.arg(contributor_id)
   AND contributor.blocked_at IS NULL
WHERE item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
ON CONFLICT (item_id, contributor_id)
DO UPDATE SET unsubscribed_at = NULL;

-- name: UnfollowFeedbackItem :execrows
UPDATE feedback_item_followers follower
SET unsubscribed_at = COALESCE(follower.unsubscribed_at, NOW())
FROM feedback_items item,
     feedback_portals portal
WHERE item.id = follower.item_id
  AND portal.id = item.portal_id
  AND portal.is_public = true
  AND follower.item_id = sqlc.arg(item_id)
  AND follower.contributor_id = sqlc.arg(contributor_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL;

-- name: GetFeedbackItemFollow :one
SELECT item.id AS item_id,
       item.slug AS item_slug,
       item.title,
       CAST(sqlc.arg(contributor_id) AS uuid) AS contributor_id,
       CAST(follower.item_id IS NOT NULL AND follower.unsubscribed_at IS NULL AS boolean) AS following,
       follower.created_at
FROM feedback_items item
INNER JOIN feedback_portals portal ON portal.id = item.portal_id AND portal.is_public = true
LEFT JOIN feedback_item_followers follower
    ON follower.item_id = item.id
   AND follower.contributor_id = sqlc.arg(contributor_id)
WHERE item.id = sqlc.arg(item_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL;

-- name: GetFeedbackContributorPreference :one
WITH preference AS (
    INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
    SELECT contributor.portal_id, contributor.id
    FROM feedback_contributors contributor
    WHERE contributor.portal_id = sqlc.arg(portal_id)
      AND contributor.id = sqlc.arg(contributor_id)
      AND contributor.blocked_at IS NULL
    ON CONFLICT (portal_id, contributor_id)
    DO UPDATE SET updated_at = feedback_contributor_preferences.updated_at
    RETURNING email_unsubscribed_at, updated_at
)
SELECT CAST(email_unsubscribed_at IS NULL AS boolean) AS portal_emails_enabled, updated_at
FROM preference;

-- name: ListFeedbackItemFollows :many
SELECT item.id AS item_id,
       item.slug AS item_slug,
       item.title,
       CAST(follower.unsubscribed_at IS NULL AS boolean) AS following,
       follower.created_at
FROM feedback_item_followers follower
INNER JOIN feedback_items item ON item.id = follower.item_id
INNER JOIN feedback_portals portal ON portal.id = item.portal_id AND portal.is_public = true
WHERE item.portal_id = sqlc.arg(portal_id)
  AND follower.contributor_id = sqlc.arg(contributor_id)
  AND item.deleted_at IS NULL
  AND item.merged_into_item_id IS NULL
ORDER BY item.created_at DESC, item.id DESC;

-- name: SetFeedbackPortalEmailPreference :one
WITH eligible AS (
    SELECT contributor.portal_id, contributor.id
    FROM feedback_contributors contributor
    WHERE contributor.portal_id = sqlc.arg(portal_id)
      AND contributor.id = sqlc.arg(contributor_id)
      AND contributor.blocked_at IS NULL
), saved AS (
    INSERT INTO feedback_contributor_preferences (portal_id, contributor_id, email_unsubscribed_at)
    SELECT eligible.portal_id,
           eligible.id,
           CAST(sqlc.narg(email_unsubscribed_at) AS timestamptz)
    FROM eligible
    ON CONFLICT (portal_id, contributor_id)
    DO UPDATE SET email_unsubscribed_at = EXCLUDED.email_unsubscribed_at,
                  updated_at = NOW()
    RETURNING contributor_id
)
SELECT CAST(COUNT(*) AS integer) FROM saved;

-- name: GetUnreadFeedbackUpdateCount :one
SELECT CAST(COUNT(*) AS integer)
FROM feedback_updates update_record
INNER JOIN feedback_portals portal ON portal.id = update_record.portal_id AND portal.is_public = true
INNER JOIN feedback_contributors contributor
    ON contributor.portal_id = portal.id
   AND contributor.id = sqlc.arg(contributor_id)
   AND contributor.blocked_at IS NULL
LEFT JOIN feedback_contributor_preferences preference
    ON preference.portal_id = portal.id
   AND preference.contributor_id = contributor.id
WHERE update_record.portal_id = sqlc.arg(portal_id)
  AND update_record.status = 'published'
  AND update_record.published_at IS NOT NULL
  AND update_record.published_at > COALESCE(preference.last_seen_update_published_at, CAST('-infinity' AS timestamptz));

-- name: MarkFeedbackUpdatesSeen :one
INSERT INTO feedback_contributor_preferences (
    portal_id,
    contributor_id,
    last_seen_update_published_at
)
SELECT contributor.portal_id, contributor.id, NOW()
FROM feedback_contributors contributor
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id = sqlc.arg(contributor_id)
  AND contributor.blocked_at IS NULL
ON CONFLICT (portal_id, contributor_id)
DO UPDATE SET last_seen_update_published_at = NOW(), updated_at = NOW()
RETURNING last_seen_update_published_at;

-- name: ListFeedbackDeliveryRecipients :many
WITH recipient_ids AS (
    SELECT follower.contributor_id
    FROM feedback_item_followers follower
    INNER JOIN feedback_items item ON item.id = follower.item_id
    WHERE item.portal_id = sqlc.arg(portal_id)
      AND follower.unsubscribed_at IS NULL
      AND CAST(sqlc.narg(item_id) AS uuid) IS NOT NULL
      AND follower.item_id = CAST(sqlc.narg(item_id) AS uuid)

    UNION

    SELECT follower.contributor_id
    FROM feedback_item_followers follower
    INNER JOIN feedback_items item ON item.id = follower.item_id
    INNER JOIN feedback_update_items linked ON linked.item_id = item.id
    WHERE item.portal_id = sqlc.arg(portal_id)
      AND follower.unsubscribed_at IS NULL
      AND CAST(sqlc.narg(update_id) AS uuid) IS NOT NULL
      AND linked.update_id = CAST(sqlc.narg(update_id) AS uuid)

    UNION

    SELECT follower.contributor_id
    FROM feedback_portal_followers follower
    WHERE follower.portal_id = sqlc.arg(portal_id)
      AND follower.unsubscribed_at IS NULL
)
SELECT contributor.id AS contributor_id,
       contributor.email,
       CAST(COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'there') AS text) AS display_name,
       CAST(contributor.kind AS text) AS kind
FROM (SELECT DISTINCT contributor_id FROM recipient_ids) recipient
INNER JOIN feedback_contributors contributor ON contributor.id = recipient.contributor_id
LEFT JOIN feedback_contributor_preferences preference
    ON preference.portal_id = contributor.portal_id
   AND preference.contributor_id = contributor.id
WHERE contributor.portal_id = sqlc.arg(portal_id)
  AND contributor.id <> sqlc.arg(actor_contributor_id)
  AND contributor.kind IN ('verified_guest', 'external')
  AND contributor.email IS NOT NULL
  AND contributor.blocked_at IS NULL
  AND preference.email_unsubscribed_at IS NULL
ORDER BY contributor.id;

-- name: ListAccountFeedbackUpdateRecipients :many
WITH linked_items AS (
    SELECT link.item_id
    FROM feedback_update_items link
    INNER JOIN feedback_items item ON item.id = link.item_id
    WHERE link.update_id = sqlc.arg(update_id)
      AND item.portal_id = sqlc.arg(portal_id)
      AND item.deleted_at IS NULL
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
SELECT DISTINCT ON (candidate.user_id) candidate.user_id, candidate.item_id
FROM candidates candidate
INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
ORDER BY candidate.user_id, candidate.item_id;

-- name: ListAccountFeedbackItemFollowers :many
SELECT DISTINCT contributor.user_id
FROM feedback_item_followers follower
INNER JOIN feedback_items item ON item.id = follower.item_id AND item.portal_id = sqlc.arg(portal_id)
INNER JOIN feedback_contributors contributor ON contributor.id = follower.contributor_id AND contributor.portal_id = item.portal_id
INNER JOIN users account ON account.user_id = contributor.user_id AND account.is_active = true
WHERE follower.item_id = sqlc.arg(item_id)
  AND follower.unsubscribed_at IS NULL
  AND contributor.kind = 'account'
  AND contributor.user_id IS NOT NULL
  AND contributor.blocked_at IS NULL
ORDER BY contributor.user_id;

-- name: CreateFeedbackContributorDelivery :one
WITH inserted AS (
    INSERT INTO feedback_contributor_deliveries (
        id,
        portal_id,
        contributor_id,
        item_id,
        update_id,
        event_type,
        dedupe_key,
        subject,
        message,
        destination_url,
        recipient_email,
        event_payload
    )
    SELECT sqlc.arg(delivery_id),
           contributor.portal_id,
           contributor.id,
           CASE WHEN EXISTS (
               SELECT 1 FROM feedback_items item
               WHERE item.id = CAST(sqlc.narg(item_id) AS uuid)
                 AND item.portal_id = contributor.portal_id
           ) THEN CAST(sqlc.narg(item_id) AS uuid) ELSE CAST(NULL AS uuid) END,
           CASE WHEN EXISTS (
               SELECT 1 FROM feedback_updates update_record
               WHERE update_record.id = CAST(sqlc.narg(update_id) AS uuid)
                 AND update_record.portal_id = contributor.portal_id
           ) THEN CAST(sqlc.narg(update_id) AS uuid) ELSE CAST(NULL AS uuid) END,
           sqlc.arg(event_type),
           sqlc.arg(dedupe_key),
           sqlc.arg(subject),
           sqlc.arg(message),
           sqlc.arg(destination_url),
           contributor.email,
           CAST(sqlc.arg(event_payload) AS jsonb)
    FROM feedback_contributors contributor
    LEFT JOIN feedback_contributor_preferences preference
        ON preference.portal_id = contributor.portal_id
       AND preference.contributor_id = contributor.id
    WHERE contributor.portal_id = sqlc.arg(portal_id)
      AND contributor.id = sqlc.arg(contributor_id)
      AND contributor.kind IN ('verified_guest', 'external')
      AND contributor.email IS NOT NULL
      AND contributor.blocked_at IS NULL
      AND preference.email_unsubscribed_at IS NULL
    ON CONFLICT (portal_id, contributor_id, channel, dedupe_key) DO NOTHING
    RETURNING id,
              portal_id,
              contributor_id,
              recipient_email,
              item_id,
              update_id,
              event_type,
              dedupe_key,
              subject,
              message,
              destination_url,
              status,
              attempt_count,
              final_failure_reason,
              created_at
), selected AS (
    SELECT inserted.id,
           inserted.portal_id,
           inserted.contributor_id,
           inserted.recipient_email,
           inserted.item_id,
           inserted.update_id,
           inserted.event_type,
           inserted.dedupe_key,
           inserted.subject,
           inserted.message,
           inserted.destination_url,
           inserted.status,
           inserted.attempt_count,
           inserted.final_failure_reason,
           inserted.created_at,
           true AS was_created
    FROM inserted

    UNION ALL

    SELECT existing.id,
           existing.portal_id,
           existing.contributor_id,
           existing.recipient_email,
           existing.item_id,
           existing.update_id,
           existing.event_type,
           existing.dedupe_key,
           existing.subject,
           existing.message,
           existing.destination_url,
           existing.status,
           existing.attempt_count,
           existing.final_failure_reason,
           existing.created_at,
           false AS was_created
    FROM feedback_contributor_deliveries existing
    WHERE existing.portal_id = sqlc.arg(portal_id)
      AND existing.contributor_id = sqlc.arg(contributor_id)
      AND existing.channel = 'email'
      AND existing.dedupe_key = sqlc.arg(dedupe_key)
      AND NOT EXISTS (SELECT 1 FROM inserted)
    LIMIT 1
)
SELECT selected.id,
       selected.portal_id,
       selected.contributor_id,
       selected.recipient_email,
       CAST(COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'there') AS text) AS display_name,
       workspace.name AS portal_name,
       workspace.slug AS portal_slug,
       selected.item_id,
       selected.update_id,
       selected.event_type,
       selected.dedupe_key,
       selected.subject,
       selected.message,
       selected.destination_url,
       CAST(selected.status AS text) AS status,
       selected.attempt_count,
       selected.final_failure_reason,
       selected.created_at,
       selected.was_created
FROM selected
INNER JOIN feedback_contributors contributor ON contributor.id = selected.contributor_id
LEFT JOIN feedback_contributor_preferences preference ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
INNER JOIN feedback_portals portal ON portal.id = selected.portal_id
INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id AND workspace.deleted_at IS NULL
WHERE contributor.kind IN ('verified_guest', 'external')
  AND contributor.email IS NOT NULL
  AND contributor.blocked_at IS NULL
  AND preference.email_unsubscribed_at IS NULL;

-- name: CreateFeedbackUnsubscribeToken :exec
INSERT INTO feedback_contributor_unsubscribe_tokens (
    portal_id,
    contributor_id,
    delivery_id,
    purpose,
    token_hash,
    expires_at
)
SELECT delivery.portal_id,
       delivery.contributor_id,
       delivery.id,
       'all_email',
       sqlc.arg(token_hash),
       sqlc.arg(expires_at)
FROM feedback_contributor_deliveries delivery
WHERE delivery.id = sqlc.arg(delivery_id)
  AND delivery.portal_id = sqlc.arg(portal_id)
  AND delivery.contributor_id = sqlc.arg(contributor_id);

-- name: ClaimFeedbackContributorDelivery :one
WITH candidate AS (
    SELECT delivery.id,
           token.token_hash,
           (
               contributor.kind IN ('verified_guest', 'external')
               AND contributor.email IS NOT NULL
               AND contributor.blocked_at IS NULL
               AND preference.email_unsubscribed_at IS NULL
               AND token.token_hash IS NOT NULL
               AND token.consumed_at IS NULL
               AND token.expires_at > NOW()
           ) AS eligible
    FROM feedback_contributor_deliveries delivery
    INNER JOIN feedback_contributors contributor ON contributor.id = delivery.contributor_id
    LEFT JOIN feedback_contributor_preferences preference ON preference.portal_id = delivery.portal_id AND preference.contributor_id = delivery.contributor_id
    LEFT JOIN feedback_contributor_unsubscribe_tokens token ON token.delivery_id = delivery.id
    WHERE delivery.id = sqlc.arg(delivery_id)
      AND (
          delivery.status IN ('queued', 'retrying')
          OR (delivery.status = 'processing' AND delivery.last_attempt_at <= sqlc.arg(stale_before))
      )
    FOR UPDATE OF delivery
), claimed AS (
    UPDATE feedback_contributor_deliveries delivery
    SET status = CASE WHEN candidate.eligible THEN 'processing' ELSE 'suppressed' END,
        next_attempt_at = NULL,
        last_attempt_at = CASE WHEN candidate.eligible THEN NOW() ELSE delivery.last_attempt_at END,
        final_failure_reason = CASE WHEN candidate.eligible THEN NULL ELSE 'recipient blocked or unsubscribed before delivery' END,
        updated_at = NOW()
    FROM candidate
    WHERE delivery.id = candidate.id
    RETURNING delivery.id,
              delivery.portal_id,
              delivery.contributor_id,
              delivery.recipient_email,
              delivery.subject,
              delivery.message,
              delivery.destination_url,
              candidate.eligible,
              candidate.token_hash
)
SELECT claimed.id,
       claimed.recipient_email,
       CAST(COALESCE(NULLIF(TRIM(contributor.display_name), ''), 'there') AS text) AS display_name,
       workspace.name AS portal_name,
       workspace.slug AS portal_slug,
       claimed.subject,
       claimed.message,
       claimed.destination_url,
       claimed.token_hash
FROM claimed
INNER JOIN feedback_contributors contributor ON contributor.id = claimed.contributor_id
INNER JOIN feedback_portals portal ON portal.id = claimed.portal_id
INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id AND workspace.deleted_at IS NULL
WHERE claimed.eligible;

-- name: MarkFeedbackContributorDeliverySent :execrows
UPDATE feedback_contributor_deliveries
SET status = 'sent',
    attempt_count = attempt_count + 1,
    last_attempt_at = NOW(),
    next_attempt_at = NULL,
    sent_at = NOW(),
    final_failure_reason = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(delivery_id)
  AND status = 'processing';

-- name: MarkFeedbackContributorDeliveryFailed :execrows
UPDATE feedback_contributor_deliveries
SET status = sqlc.arg(status),
    attempt_count = attempt_count + 1,
    last_attempt_at = NOW(),
    next_attempt_at = CASE WHEN sqlc.arg(status) = 'retrying' THEN NOW() ELSE NULL END,
    final_failure_reason = CASE WHEN sqlc.arg(status) = 'failed' THEN LEFT(sqlc.arg(reason), 1000) ELSE final_failure_reason END,
    updated_at = NOW()
WHERE id = sqlc.arg(delivery_id)
  AND (
      (sqlc.arg(status) = 'failed' AND status IN ('queued', 'processing', 'retrying'))
      OR (sqlc.arg(status) = 'retrying' AND status = 'processing')
  );

-- name: ListRecoverableFeedbackContributorDeliveries :many
SELECT delivery.id AS delivery_id, token.token_hash
FROM feedback_contributor_deliveries delivery
INNER JOIN feedback_contributor_unsubscribe_tokens token ON token.delivery_id = delivery.id
WHERE (
        (
            delivery.status IN ('queued', 'retrying')
            AND (delivery.next_attempt_at IS NULL OR delivery.next_attempt_at <= NOW())
        )
        OR (
            delivery.status = 'processing'
            AND delivery.last_attempt_at <= sqlc.arg(stale_before)
        )
      )
  AND token.consumed_at IS NULL
  AND token.expires_at > NOW()
ORDER BY delivery.created_at, delivery.id
LIMIT sqlc.arg(row_limit);
