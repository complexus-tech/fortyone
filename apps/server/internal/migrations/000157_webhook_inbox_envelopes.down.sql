DROP INDEX IF EXISTS public.messaging_inbound_events_payload_expiry_idx;

ALTER TABLE public.messaging_inbound_events
    DROP CONSTRAINT IF EXISTS messaging_inbound_events_trace_id_check,
    DROP CONSTRAINT IF EXISTS messaging_inbound_events_envelope_version_check,
    DROP COLUMN IF EXISTS payload_expires_at,
    DROP COLUMN IF EXISTS trace_id,
    DROP COLUMN IF EXISTS installation_id,
    DROP COLUMN IF EXISTS envelope_version;
