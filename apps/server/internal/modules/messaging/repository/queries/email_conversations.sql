-- name: HasEarlierInboundEvent :one
SELECT EXISTS (
    SELECT 1
    FROM messaging_inbound_events AS candidate
    INNER JOIN messaging_inbound_events AS current
        ON current.id = sqlc.arg(current_id)
    WHERE candidate.provider = sqlc.arg(provider)
      AND candidate.external_workspace_id = sqlc.arg(external_workspace_id)
      AND candidate.id <> current.id
      AND (candidate.received_at, candidate.id) < (current.received_at, current.id)
      AND candidate.status IN ('pending', 'processing', 'failed')
      AND candidate.attempt_count < 20
);

-- name: TryEmailThreadAdvisoryLock :one
SELECT pg_try_advisory_lock(hashtextextended(sqlc.arg(lock_key), 0));

-- name: ReleaseEmailThreadAdvisoryLock :one
SELECT pg_advisory_unlock(hashtextextended(sqlc.arg(lock_key), 0));

-- name: InsertEmailThread :one
INSERT INTO messaging_email_threads (
    provider,
    workspace_id,
    user_id,
    recipient_email,
    external_thread_id,
    root_internet_message_id,
    latest_internet_message_id,
    context
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(recipient_email),
    sqlc.arg(external_thread_id),
    sqlc.arg(root_internet_message_id),
    sqlc.arg(root_internet_message_id),
    sqlc.arg(context)
)
ON CONFLICT (provider, workspace_id, external_thread_id) DO NOTHING
RETURNING id,
          workspace_id,
          user_id,
          provider,
          recipient_email,
          external_thread_id,
          root_internet_message_id,
          latest_internet_message_id,
          context,
          summary,
          summary_through_sequence,
          next_message_sequence,
          created_at,
          updated_at;

-- name: GetEmailThreadByExternalID :one
SELECT id,
       workspace_id,
       user_id,
       provider,
       recipient_email,
       external_thread_id,
       root_internet_message_id,
       latest_internet_message_id,
       context,
       summary,
       summary_through_sequence,
       next_message_sequence,
       created_at,
       updated_at
FROM messaging_email_threads
WHERE provider = sqlc.arg(provider)
  AND workspace_id = sqlc.arg(workspace_id)
  AND external_thread_id = sqlc.arg(external_thread_id);

-- name: InsertEmailReplyToken :one
INSERT INTO messaging_email_reply_tokens (
    thread_id,
    workspace_id,
    user_id,
    token_hash,
    expires_at
) VALUES (
    sqlc.arg(thread_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at)
)
ON CONFLICT (token_hash) DO NOTHING
RETURNING id, thread_id, expires_at, revoked_at, created_at;

-- name: GetEmailReplyToken :one
SELECT id, thread_id, expires_at, revoked_at, created_at
FROM messaging_email_reply_tokens
WHERE token_hash = sqlc.arg(token_hash);

-- name: FindEmailThreadByReplyToken :one
SELECT thread.id,
       thread.workspace_id,
       thread.user_id,
       thread.provider,
       thread.recipient_email,
       thread.external_thread_id,
       thread.root_internet_message_id,
       thread.latest_internet_message_id,
       thread.context,
       thread.summary,
       thread.summary_through_sequence,
       thread.next_message_sequence,
       thread.created_at,
       thread.updated_at,
       token.id AS reply_token_id,
       token.expires_at AS reply_token_expires_at
FROM messaging_email_reply_tokens AS token
INNER JOIN messaging_email_threads AS thread ON thread.id = token.thread_id
INNER JOIN workspaces AS workspace ON workspace.workspace_id = thread.workspace_id
INNER JOIN users AS actor ON actor.user_id = thread.user_id
INNER JOIN workspace_members AS member
    ON member.workspace_id = thread.workspace_id
    AND member.user_id = thread.user_id
WHERE token.token_hash = sqlc.arg(token_hash)
  AND thread.provider = sqlc.arg(provider)
  AND token.revoked_at IS NULL
  AND token.expires_at > sqlc.arg(now)
  AND workspace.deleted_at IS NULL
  AND actor.is_active = true
  AND actor.is_system = false
  AND LOWER(actor.email) = LOWER(thread.recipient_email)
  AND member.role IN ('admin', 'member', 'guest');

-- name: GetAuthorizedEmailThread :one
SELECT thread.id,
       thread.workspace_id,
       thread.user_id,
       thread.provider,
       thread.recipient_email,
       thread.external_thread_id,
       thread.root_internet_message_id,
       thread.latest_internet_message_id,
       thread.context,
       thread.summary,
       thread.summary_through_sequence,
       thread.next_message_sequence,
       thread.created_at,
       thread.updated_at
FROM messaging_email_threads AS thread
INNER JOIN workspaces AS workspace ON workspace.workspace_id = thread.workspace_id
INNER JOIN users AS actor ON actor.user_id = thread.user_id
INNER JOIN workspace_members AS member
    ON member.workspace_id = thread.workspace_id
    AND member.user_id = thread.user_id
WHERE thread.id = sqlc.arg(thread_id)
  AND thread.workspace_id = sqlc.arg(workspace_id)
  AND thread.user_id = sqlc.arg(user_id)
  AND workspace.deleted_at IS NULL
  AND actor.is_active = true
  AND actor.is_system = false
  AND LOWER(actor.email) = LOWER(thread.recipient_email)
  AND member.role IN ('admin', 'member', 'guest');

-- name: LockEmailThreadSequence :one
SELECT next_message_sequence
FROM messaging_email_threads
WHERE id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
FOR UPDATE;

-- name: GetEmailMessageByIdempotencyKey :one
SELECT id,
       thread_id,
       sequence,
       inbound_event_id,
       idempotency_key,
       direction,
       role,
       kind,
       provider_message_id,
       internet_message_id,
       in_reply_to_message_id,
       subject,
       content,
       context,
       provider_metadata,
       created_at
FROM messaging_email_messages
WHERE thread_id = sqlc.arg(thread_id)
  AND idempotency_key = sqlc.arg(idempotency_key);

-- name: InsertEmailMessage :one
INSERT INTO messaging_email_messages (
    thread_id,
    workspace_id,
    user_id,
    sequence,
    inbound_event_id,
    idempotency_key,
    direction,
    role,
    kind,
    provider_message_id,
    internet_message_id,
    in_reply_to_message_id,
    subject,
    content,
    context,
    provider_metadata
) VALUES (
    sqlc.arg(thread_id),
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    sqlc.arg(sequence),
    sqlc.narg(inbound_event_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(direction),
    sqlc.arg(role),
    sqlc.arg(kind),
    NULLIF(CAST(sqlc.arg(provider_message_id) AS text), ''),
    NULLIF(CAST(sqlc.arg(internet_message_id) AS text), ''),
    NULLIF(CAST(sqlc.arg(in_reply_to_message_id) AS text), ''),
    sqlc.arg(subject),
    sqlc.arg(content),
    sqlc.arg(context),
    sqlc.arg(provider_metadata)
)
RETURNING id,
          thread_id,
          sequence,
          inbound_event_id,
          idempotency_key,
          direction,
          role,
          kind,
          provider_message_id,
          internet_message_id,
          in_reply_to_message_id,
          subject,
          content,
          context,
          provider_metadata,
          created_at;

-- name: AdvanceEmailThreadCursor :execrows
UPDATE messaging_email_threads
SET next_message_sequence = sqlc.arg(next_message_sequence),
    root_internet_message_id = CASE
        WHEN root_internet_message_id = '' THEN sqlc.arg(internet_message_id)
        ELSE root_internet_message_id
    END,
    latest_internet_message_id = CASE
        WHEN sqlc.arg(internet_message_id) = '' THEN latest_internet_message_id
        ELSE sqlc.arg(internet_message_id)
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(thread_id);

-- name: ListEmailMessages :many
SELECT message.id,
       message.thread_id,
       message.sequence,
       message.inbound_event_id,
       message.idempotency_key,
       message.direction,
       message.role,
       message.kind,
       message.provider_message_id,
       message.internet_message_id,
       message.in_reply_to_message_id,
       message.subject,
       message.content,
       message.context,
       message.provider_metadata,
       message.created_at
FROM messaging_email_messages AS message
INNER JOIN messaging_email_threads AS thread ON thread.id = message.thread_id
WHERE message.thread_id = sqlc.arg(thread_id)
  AND thread.workspace_id = sqlc.arg(workspace_id)
  AND thread.user_id = sqlc.arg(user_id)
  AND message.sequence > sqlc.arg(after_sequence)
ORDER BY message.sequence
LIMIT sqlc.arg(row_limit);

-- name: UpdateEmailThreadSummary :one
UPDATE messaging_email_threads
SET summary = sqlc.arg(summary),
    summary_through_sequence = sqlc.arg(through_sequence),
    updated_at = NOW()
WHERE id = sqlc.arg(thread_id)
  AND workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
  AND summary_through_sequence = sqlc.arg(expected_summary_through_sequence)
  AND sqlc.arg(through_sequence) < next_message_sequence
RETURNING id,
          workspace_id,
          user_id,
          provider,
          recipient_email,
          external_thread_id,
          root_internet_message_id,
          latest_internet_message_id,
          context,
          summary,
          summary_through_sequence,
          next_message_sequence,
          created_at,
          updated_at;
