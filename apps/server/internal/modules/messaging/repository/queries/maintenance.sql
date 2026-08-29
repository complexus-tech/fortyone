-- name: PurgeExpiredMessagingNonces :execrows
WITH candidates AS MATERIALIZED (
    SELECT nonce.id
    FROM public.messaging_nonces AS nonce
    WHERE nonce.expires_at < sqlc.arg(expired_before)
    ORDER BY nonce.expires_at, nonce.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF nonce SKIP LOCKED
)
DELETE FROM public.messaging_nonces AS nonce
USING candidates
WHERE nonce.id = candidates.id;

-- name: ExpireMessagingStoryMutationConfirmations :execrows
WITH candidates AS MATERIALIZED (
    SELECT confirmation.confirmation_id
    FROM public.messaging_story_mutation_confirmations AS confirmation
    WHERE confirmation.expires_at <= sqlc.arg(expired_at)
      AND (
          confirmation.status = 'pending'
          OR (
              confirmation.status = 'applied'
              AND confirmation.operation = 'create_stories'
              AND confirmation.proposal IS NOT NULL
          )
      )
    ORDER BY confirmation.expires_at, confirmation.confirmation_id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF confirmation SKIP LOCKED
)
UPDATE public.messaging_story_mutation_confirmations AS confirmation
SET status = 'expired',
    proposal = NULL,
    applied_at = NULL,
    expired_at = sqlc.arg(expired_at),
    updated_at = sqlc.arg(expired_at)
FROM candidates
WHERE confirmation.confirmation_id = candidates.confirmation_id;

-- name: PurgeOldMessagingOutboundDeliveries :execrows
WITH candidates AS MATERIALIZED (
    SELECT delivery.id
    FROM public.messaging_outbound_deliveries AS delivery
    WHERE delivery.created_at < sqlc.arg(created_before)
    ORDER BY delivery.created_at, delivery.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF delivery SKIP LOCKED
)
DELETE FROM public.messaging_outbound_deliveries AS delivery
USING candidates
WHERE delivery.id = candidates.id;

-- name: PurgeOldMessagingInboundEvents :execrows
WITH candidates AS MATERIALIZED (
    SELECT event.id
    FROM public.messaging_inbound_events AS event
    WHERE event.received_at < sqlc.arg(received_before)
    ORDER BY event.received_at, event.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF event SKIP LOCKED
)
DELETE FROM public.messaging_inbound_events AS event
USING candidates
WHERE event.id = candidates.id;

-- name: PurgeCompletedSlackUninstalls :execrows
WITH candidates AS MATERIALIZED (
    SELECT uninstall.id
    FROM public.slack_uninstall_outbox AS uninstall
    WHERE uninstall.status = 'completed'
      AND uninstall.completed_at < sqlc.arg(completed_before)
    ORDER BY uninstall.completed_at, uninstall.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF uninstall SKIP LOCKED
)
DELETE FROM public.slack_uninstall_outbox AS uninstall
USING candidates
WHERE uninstall.id = candidates.id;

-- name: PurgeOldMessagingMessages :execrows
WITH candidates AS MATERIALIZED (
    SELECT message.id
    FROM public.messaging_messages AS message
    WHERE message.created_at < sqlc.arg(created_before)
    ORDER BY message.created_at, message.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF message SKIP LOCKED
)
DELETE FROM public.messaging_messages AS message
USING candidates
WHERE message.id = candidates.id;

-- name: PurgeExpiredMessagingEmailReplyTokens :execrows
WITH candidates AS MATERIALIZED (
    SELECT token.id
    FROM public.messaging_email_reply_tokens AS token
    WHERE token.expires_at < sqlc.arg(retained_before)
       OR (token.revoked_at IS NOT NULL AND token.revoked_at < sqlc.arg(retained_before))
    ORDER BY LEAST(token.expires_at, token.revoked_at), token.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF token SKIP LOCKED
)
DELETE FROM public.messaging_email_reply_tokens AS token
USING candidates
WHERE token.id = candidates.id;

-- name: PurgeEmptyMessagingConversations :execrows
WITH candidates AS MATERIALIZED (
    SELECT conversation.id
    FROM public.messaging_conversations AS conversation
    WHERE conversation.updated_at < sqlc.arg(updated_before)
      AND NOT EXISTS (
          SELECT 1
          FROM public.messaging_messages AS message
          WHERE message.conversation_id = conversation.id
      )
    ORDER BY conversation.updated_at, conversation.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF conversation SKIP LOCKED
)
DELETE FROM public.messaging_conversations AS conversation
USING candidates
WHERE conversation.id = candidates.id;
