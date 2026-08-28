ALTER TABLE public.messaging_inbound_events
    ADD COLUMN envelope_version smallint NOT NULL DEFAULT 1,
    ADD COLUMN installation_id uuid,
    ADD COLUMN trace_id text,
    ADD COLUMN payload_expires_at timestamptz,
    ADD CONSTRAINT messaging_inbound_events_envelope_version_check
        CHECK (envelope_version > 0),
    ADD CONSTRAINT messaging_inbound_events_trace_id_check
        CHECK (trace_id IS NULL OR char_length(trace_id) BETWEEN 1 AND 128);

UPDATE public.messaging_inbound_events AS delivery
SET installation_id = installation.id
FROM public.slack_workspaces AS installation
WHERE delivery.provider = 'slack'
  AND delivery.external_workspace_id = installation.slack_team_id
  AND delivery.workspace_id = installation.workspace_id
  AND delivery.installation_id IS NULL;

CREATE INDEX messaging_inbound_events_payload_expiry_idx
    ON public.messaging_inbound_events USING btree (payload_expires_at, id)
    WHERE payload_encrypted IS NOT NULL AND payload_expires_at IS NOT NULL;

COMMENT ON COLUMN public.messaging_inbound_events.envelope_version IS
    'Version of the provider-neutral webhook envelope; independent of provider payload versions.';
COMMENT ON COLUMN public.messaging_inbound_events.installation_id IS
    'Opaque first-party integration installation identifier. Provider-specific foreign keys remain in their adapter tables.';
COMMENT ON COLUMN public.messaging_inbound_events.trace_id IS
    'Safe correlation identifier only; never a credential or provider payload.';
COMMENT ON COLUMN public.messaging_inbound_events.payload_expires_at IS
    'Retention deadline for encrypted raw payload content. Safe delivery audit facts remain after expiry.';
