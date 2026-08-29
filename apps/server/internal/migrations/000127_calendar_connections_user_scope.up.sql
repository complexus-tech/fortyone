-- Google Calendar is a personal connection. Keep one active provider account
-- per user so switching workspaces never requires another OAuth connection.
WITH ranked_connections AS (
    SELECT
        connection_id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id, provider
            ORDER BY updated_at DESC, created_at DESC, connection_id DESC
        ) AS row_number
    FROM public.calendar_connections
    WHERE revoked_at IS NULL
)
DELETE FROM public.calendar_connections cc
USING ranked_connections duplicate
WHERE cc.connection_id = duplicate.connection_id
  AND duplicate.row_number > 1;

DROP INDEX IF EXISTS public.calendar_connections_one_active_provider_per_user;

CREATE UNIQUE INDEX calendar_connections_one_active_provider_per_account
    ON public.calendar_connections (user_id, provider)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_calendar_connections_user_provider
    ON public.calendar_connections (user_id, provider)
    WHERE revoked_at IS NULL;
