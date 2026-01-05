-- 000027_notifications.up.sql
CREATE TABLE public.notifications (
    notification_id uuid DEFAULT gen_random_uuid() NOT NULL,
    recipient_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    type public.notification_type NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    title text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    read_at timestamp with time zone,
    message jsonb NOT NULL
);

ALTER TABLE ONLY public.notifications ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);

ALTER TABLE ONLY public.notifications 
    ADD CONSTRAINT fk_notifications_actor FOREIGN KEY (actor_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.notifications 
    ADD CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.notifications 
    ADD CONSTRAINT fk_notifications_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.notifications ADD CONSTRAINT notifications_recipient_workspace_entity_unique UNIQUE (recipient_id, workspace_id, entity_type, entity_id);

CREATE INDEX idx_notifications_created_at ON public.notifications USING btree (created_at);
CREATE INDEX idx_notifications_entity ON public.notifications USING btree (entity_type, entity_id);
CREATE INDEX idx_notifications_id_recipient ON public.notifications USING btree (notification_id, recipient_id);
CREATE INDEX idx_notifications_recipient ON public.notifications USING btree (recipient_id);
CREATE INDEX idx_notifications_recipient_workspace_created_at ON public.notifications USING btree (recipient_id, workspace_id, created_at DESC);
CREATE INDEX idx_notifications_type ON public.notifications USING btree (type);
CREATE INDEX idx_notifications_unread ON public.notifications USING btree (recipient_id, workspace_id, read_at) WHERE (read_at IS NULL);
CREATE INDEX idx_notifications_workspace ON public.notifications USING btree (workspace_id);
