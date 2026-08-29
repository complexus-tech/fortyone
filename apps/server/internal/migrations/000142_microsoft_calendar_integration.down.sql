DELETE FROM public.calendar_schedule_event_outbox WHERE provider = 'microsoft';
UPDATE public.calendar_schedule_blocks
SET external_provider = NULL,
    external_calendar_id = NULL,
    external_event_id = NULL,
    external_sync_hash = NULL,
    external_synced_at = NULL
WHERE external_provider = 'microsoft';
DELETE FROM public.calendar_connections WHERE provider = 'microsoft';

ALTER TABLE public.calendar_schedule_event_outbox
    DROP CONSTRAINT calendar_schedule_event_outbox_provider_check,
    ADD CONSTRAINT calendar_schedule_event_outbox_provider_check
        CHECK (provider = 'google');

ALTER TABLE public.calendar_events
    DROP CONSTRAINT calendar_events_provider_check,
    ADD CONSTRAINT calendar_events_provider_check
        CHECK (provider IN ('google'));

ALTER TABLE public.calendar_busy_windows
    DROP CONSTRAINT calendar_busy_windows_provider_check,
    ADD CONSTRAINT calendar_busy_windows_provider_check
        CHECK (provider IN ('google'));

ALTER TABLE public.calendar_connections
    DROP CONSTRAINT calendar_connections_provider_check,
    ADD CONSTRAINT calendar_connections_provider_check
        CHECK (provider IN ('google'));
