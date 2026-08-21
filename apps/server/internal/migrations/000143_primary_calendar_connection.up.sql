ALTER TABLE public.calendar_connections
    ADD COLUMN is_primary boolean NOT NULL DEFAULT FALSE;

-- Preserve the historical Google-first write behavior for existing accounts,
-- while only selecting connections that can actually receive scheduled work.
WITH ranked_connections AS (
    SELECT
        connection_id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY
                CASE provider WHEN 'google' THEN 0 ELSE 1 END,
                created_at,
                connection_id
        ) AS position
    FROM public.calendar_connections
    WHERE revoked_at IS NULL
      AND cleanup_pending_at IS NULL
      AND (
          (
              provider = 'google'
              AND 'https://www.googleapis.com/auth/calendar.events.readonly' = ANY(scopes)
              AND 'https://www.googleapis.com/auth/calendar.events.owned' = ANY(scopes)
          )
          OR (provider = 'microsoft' AND 'Calendars.ReadWrite' = ANY(scopes))
      )
)
UPDATE public.calendar_connections connection
SET is_primary = TRUE
FROM ranked_connections ranked
WHERE connection.connection_id = ranked.connection_id
  AND ranked.position = 1;

CREATE UNIQUE INDEX calendar_connections_one_primary_per_account
    ON public.calendar_connections (user_id)
    WHERE is_primary = TRUE;
