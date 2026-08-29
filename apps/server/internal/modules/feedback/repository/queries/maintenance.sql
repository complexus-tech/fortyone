-- name: PurgeFeedbackContributorVerifications :execrows
DELETE FROM feedback_contributor_verifications
WHERE expires_at < sqlc.arg(retained_before);

-- name: PurgeFeedbackContributorSessions :execrows
DELETE FROM feedback_contributor_sessions
WHERE expires_at < sqlc.arg(retained_before)
   OR (revoked_at IS NOT NULL AND revoked_at < sqlc.arg(retained_before));

-- name: PurgeFeedbackContributorUnsubscribeTokens :execrows
DELETE FROM feedback_contributor_unsubscribe_tokens
WHERE expires_at < sqlc.arg(retained_before)
   OR (consumed_at IS NOT NULL AND consumed_at < sqlc.arg(retained_before));

-- name: PurgeFeedbackWidgetAssertionNonces :execrows
DELETE FROM feedback_widget_assertion_nonces
WHERE expires_at < sqlc.arg(expired_before);

-- name: PurgeFeedbackWidgetSecretRotations :execrows
DELETE FROM feedback_widget_signing_secret_rotations
WHERE grace_expires_at < sqlc.arg(retained_before);

-- name: PurgeDeletedFeedback :one
WITH deleted_items AS (
    DELETE FROM feedback_items item
    WHERE item.deleted_at IS NOT NULL
      AND item.deleted_at < sqlc.arg(deleted_before)
    RETURNING item.contributor_id
), deleted_contributors AS (
    DELETE FROM feedback_contributors contributor
    WHERE contributor.id IN (SELECT DISTINCT deleted_items.contributor_id FROM deleted_items)
      AND contributor.kind = 'anonymous'
      AND NOT EXISTS (
          SELECT 1
          FROM feedback_items retained
          WHERE retained.contributor_id = contributor.id
      )
    RETURNING contributor.id
)
SELECT CAST((SELECT COUNT(*) FROM deleted_items) AS bigint) AS items_deleted,
       CAST((SELECT COUNT(*) FROM deleted_contributors) AS bigint) AS contributors_deleted;
