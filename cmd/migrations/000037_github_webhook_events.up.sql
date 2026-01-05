-- 000037_github_webhook_events.up.sql
CREATE TABLE public.github_webhook_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    repository_id uuid,
    event_type text NOT NULL,
    github_delivery_id text NOT NULL,
    payload jsonb NOT NULL,
    processed boolean DEFAULT false,
    processed_at timestamp without time zone,
    error_message text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.github_webhook_events ADD CONSTRAINT github_webhook_events_github_delivery_id_key UNIQUE (github_delivery_id);
ALTER TABLE ONLY public.github_webhook_events ADD CONSTRAINT github_webhook_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.github_webhook_events 
    ADD CONSTRAINT github_webhook_events_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE;

CREATE INDEX idx_github_webhook_events_processed ON public.github_webhook_events USING btree (processed, created_at);
