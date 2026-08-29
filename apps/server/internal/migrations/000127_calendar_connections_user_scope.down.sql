DROP INDEX IF EXISTS public.idx_calendar_connections_user_provider;
DROP INDEX IF EXISTS public.calendar_connections_one_active_provider_per_account;

CREATE UNIQUE INDEX calendar_connections_one_active_provider_per_user
    ON public.calendar_connections (workspace_id, user_id, provider)
    WHERE revoked_at IS NULL;
