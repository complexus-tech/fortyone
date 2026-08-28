-- name: PurgeDeletedChatSessions :execrows
WITH candidates AS MATERIALIZED (
    SELECT session.id
    FROM public.chat_sessions AS session
    WHERE session.deleted_at IS NOT NULL
      AND session.deleted_at < sqlc.arg(deleted_before)
      AND NOT EXISTS (
          SELECT 1
          FROM public.chat_mutation_approval_executions AS execution
          WHERE execution.session_id = session.id
            AND execution.status IN ('ready', 'retry_ready', 'executing', 'failed_uncertain')
      )
    ORDER BY session.deleted_at, session.id
    LIMIT CAST(sqlc.arg(batch_size) AS integer)
    FOR UPDATE OF session SKIP LOCKED
)
DELETE FROM public.chat_sessions AS session
USING candidates
WHERE session.id = candidates.id;
