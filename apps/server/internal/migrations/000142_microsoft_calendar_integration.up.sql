ALTER TABLE public.calendar_connections
    DROP CONSTRAINT calendar_connections_provider_check,
    ADD CONSTRAINT calendar_connections_provider_check
        CHECK (provider IN ('google', 'microsoft'));

ALTER TABLE public.calendar_busy_windows
    DROP CONSTRAINT calendar_busy_windows_provider_check,
    ADD CONSTRAINT calendar_busy_windows_provider_check
        CHECK (provider IN ('google', 'microsoft'));

ALTER TABLE public.calendar_events
    DROP CONSTRAINT calendar_events_provider_check,
    ADD CONSTRAINT calendar_events_provider_check
        CHECK (provider IN ('google', 'microsoft'));

ALTER TABLE public.calendar_schedule_event_outbox
    DROP CONSTRAINT calendar_schedule_event_outbox_provider_check,
    ADD CONSTRAINT calendar_schedule_event_outbox_provider_check
        CHECK (provider IN ('google', 'microsoft'));
