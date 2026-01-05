-- 000008_stripe_webhook_events.up.sql
CREATE TABLE public.stripe_webhook_events (
    event_id character varying(255) NOT NULL,
    event_type character varying(255) NOT NULL,
    workspace_id uuid,
    processed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    payload jsonb
);

ALTER TABLE ONLY public.stripe_webhook_events ADD CONSTRAINT stripe_webhook_events_pkey PRIMARY KEY (event_id);

ALTER TABLE ONLY public.stripe_webhook_events 
    ADD CONSTRAINT fk_webhook_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id);

CREATE INDEX idx_webhook_event_type ON public.stripe_webhook_events USING btree (event_type);
CREATE INDEX idx_webhook_workspace ON public.stripe_webhook_events USING btree (workspace_id);
